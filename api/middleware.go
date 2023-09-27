package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/common"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/database"
	"go.uber.org/zap"
)

var defaultCredsInsecure = "default_credentails_insecure"

func (s *Server) AuthMiddleware(next http.Handler, permission common.Permission) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")

		if token == "" {
			token = r.Header.Get("Authorization")
		}

		if token == "" {
			WriteError(w, "Unauthorized", nil, http.StatusUnauthorized)
			return
		}

		data, err := s.db.CheckAuthValid(token)

		if err != nil && errors.Is(err, database.ErrDefaultToken) {
			WriteError(w, "Security: Change the default admin token (you must do this through the database) - new token must be 32 characters", &defaultCredsInsecure, http.StatusForbidden)
			return
		}

		if err != nil {
			zap.L().Error("error checking auth", zap.Error(err))
			WriteError(w, "Internal Server Error", nil, http.StatusInternalServerError)
			return
		}

		if !data.Valid {
			WriteError(w, "Unauthorized", nil, http.StatusUnauthorized)
			return
		}

		if permission != 0 && !data.Permissions.Has(permission) {
			WriteError(w, "Forbidden", nil, http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "user", data)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
