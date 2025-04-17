package middleware

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"log/slog"
	"net/http"
	"strings"
)

const ClaimsContext = "claims"

func AuthMiddleware(authService service.Auth) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := slog.With("layer", "middleware", "middleware", "AuthMiddleware")

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("missing authorization header")
				httpresponse.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Warn("invalid authorization header format", "header", authHeader)
				httpresponse.Error(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			token := parts[1]
			claims, err := authService.ValidateToken(token)
			if err != nil {
				switch {
				case errors.Is(err, service.ErrInvalidToken):
					log.Warn("invalid token", "error", err)
					httpresponse.Error(w, http.StatusUnauthorized, "invalid token")
				case errors.Is(err, service.ErrTokenExpired):
					log.Warn("token expired", "error", err)
					httpresponse.Error(w, http.StatusUnauthorized, "token expired")
				default:
					log.Error("unexpected token validation error", "error", err)
					httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsContext, claims)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
