package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

func (s *server) Serve() <-chan struct{} {
	s.setupRoutes()

	shutdownCh := make(chan struct{})
	closeCh := make(chan os.Signal, 1)
	signal.Notify(closeCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info().Msg("server is listening")
		_ = s.s.ListenAndServe()
	}()

	go func() {
		<-closeCh
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_ = s.s.Shutdown(ctx)
		log.Info().Msg("server is gracefully shutdown")
		shutdownCh <- struct{}{}
	}()

	return shutdownCh
}
