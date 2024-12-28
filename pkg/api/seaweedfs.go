package api

import (
	"net/http"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/vectier/otter/pkg/auth"
)

func (s *server) GetFile(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	claims, err := auth.GetClaims(r)
	if err != nil {
		log.Error().Err(err).Msg("cannot get claims")
		response(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	path := chi.URLParam(r, "*")
	logger := log.With().Str("path", path).Str("client", claims.Client).Logger()

	for _, glob := range claims.GrantedDirectories {
		match, err := doublestar.Match(glob, path)
		if err != nil {
			logger.Error().Err(err).Str("glob", glob).Msg("failed to match pattern of granted directories")
			response(w, http.StatusInternalServerError, "invalid granted directories")
			return
		}
		if match {
			if err := s.sc.PipeFile(r.Context(), path, w); err != nil {
				logger.Error().Err(err).Msg("cannot pipe file")
			} else {
				logger.Debug().Str("path", path).Dur("execTime", time.Since(start)).Msg("pipe file")
			}
			return
		}
	}

	response(w, http.StatusForbidden, "forbidden")
}
