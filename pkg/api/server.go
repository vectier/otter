package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/vectier/otter/pkg/seaweedfs"
)

type server struct {
	s  *http.Server
	sc seaweedfs.Client
}

func NewServer(addr string, sc seaweedfs.Client) *server {
	return &server{
		s:  &http.Server{Addr: addr},
		sc: sc,
	}
}

func (s *server) setupRoutes() {
	r := chi.NewRouter()
	s.s.Handler = r

	r.Get("/*", s.GetFile)
}

func response(w http.ResponseWriter, statusCode int, body string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(body))
}

func (s *server) Serve(stopCh <-chan struct{}) <-chan struct{} {
	s.setupRoutes()

	go func() {
		log.Info().Str("addr", s.s.Addr).Msg("server is listening")
		if err := s.s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Msg("cannot listen and serve the server")
		}
	}()

	stoppedCh := make(chan struct{})
	go func() {
		<-stopCh
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer close(stoppedCh)

		shutdownErr := make(chan error)
		go func() {
			shutdownErr <- s.s.Shutdown(ctx)
		}()

		select {
		case err := <-shutdownErr:
			if err != nil {
				log.Error().Err(err).Msg("cannot shutdown server gracefully")
			} else {
				log.Info().Msg("shutdown server gracefully")
			}
		case <-ctx.Done():
			log.Warn().Msg("shutdown timeout exceed, force shutdown server")
			_ = s.s.Close()
		}
	}()

	return stoppedCh
}
