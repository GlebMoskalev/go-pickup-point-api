package middleware

import (
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"net/http"
)

func RoleMiddleware(allowedRoles ...string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContext).(*entity.UserClaims)
			if !ok {
				httpresponse.Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			hasRole := false
			for _, role := range allowedRoles {
				if role == claims.Role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				httpresponse.Error(w, http.StatusForbidden, "access denied")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
