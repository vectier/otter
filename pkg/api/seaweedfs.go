package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (s *server) GetFile(w http.ResponseWriter, r *http.Request) error {
	path := chi.URLParam(r, "*")
	defer measureGetFile(path)()

	if err := s.sc.PipeFile(r.Context(), path, w); err != nil {
		log.Err(err).Msg("cannot pipe file")
		return err
	}
	return nil
}

func measureGetFile(path string) func() {
	start := time.Now()
	return func() {
		log.Debug().
			Str("executionTime", time.Since(start).String()).
			Str("path", path).
			Msg("pipe file")
	}
}
