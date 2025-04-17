package middleware

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"net/http"
	"strings"
)

func AuthMiddleware(authService service.Auth) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpresponse.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				httpresponse.Error(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			token := parts[1]
			claims, err := authService.ValidateToken(token)
			if err != nil {
				switch {
				case errors.Is(err, service.ErrInvalidToken):
					httpresponse.Error(w, http.StatusUnauthorized, "invalid token")
				case errors.Is(err, service.ErrTokenExpired):
					httpresponse.Error(w, http.StatusUnauthorized, "token expired")
				default:
					httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			ctx := context.WithValue(r.Context(), "claims", claims)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
