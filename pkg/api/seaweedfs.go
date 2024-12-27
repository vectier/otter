package api

import (
	"net/http"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/vectier/otter/pkg/auth"
)

func (s *server) GetFile(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.GetClaims(r)
	if err != nil {
		log.Err(err).Msg("cannot get claims")
		response(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	path := chi.URLParam(r, "*")

	for _, glob := range claims.GrantedDirectories {
		match, err := doublestar.Match(glob, path)
		if err != nil {
			log.Err(err).Msg("invalid match pattern from granted directories")
			response(w, http.StatusInternalServerError, "invalid granted directories")
			return
		}
		if match {
			if err := s.sc.PipeFile(r.Context(), path, w); err != nil {
				log.Err(err).Msg("cannot pipe file")
			}
			return
		}
	}

	response(w, http.StatusForbidden, "forbidden")
}
