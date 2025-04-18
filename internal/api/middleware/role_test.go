package middleware

import (
	"context"
	"encoding/json"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoleMiddleware(t *testing.T) {
	testCases := []struct {
		name               string
		claims             *entity.UserClaims
		allowedRoles       []string
		expectedHTTPStatus int
		expectedBody       any
		shouldCallNext     bool
	}{
		{
			name: "success - user has allowed role (employee)",
			claims: &entity.UserClaims{
				UserID: uuid.New(),
				Role:   entity.RoleEmployee,
			},
			allowedRoles:       []string{entity.RoleEmployee},
			expectedHTTPStatus: http.StatusOK,
			shouldCallNext:     true,
		},
		{
			name: "success - user has one of multiple allowed roles (moderator)",
			claims: &entity.UserClaims{
				UserID: uuid.New(),
				Role:   entity.RoleModerator,
			},
			allowedRoles:       []string{entity.RoleEmployee, entity.RoleModerator},
			expectedHTTPStatus: http.StatusOK,
			shouldCallNext:     true,
		},
		{
			name:               "error - missing claims in context",
			claims:             nil,
			allowedRoles:       []string{entity.RoleEmployee},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedBody:       httpresponse.ErrorResponse{Error: "unauthorized"},
			shouldCallNext:     false,
		},
		{
			name: "error - user does not have allowed role",
			claims: &entity.UserClaims{
				UserID: uuid.New(),
				Role:   entity.RoleEmployee,
			},
			allowedRoles:       []string{entity.RoleModerator},
			expectedHTTPStatus: http.StatusForbidden,
			expectedBody:       httpresponse.ErrorResponse{Error: "access denied"},
			shouldCallNext:     false,
		},
		{
			name: "error - no allowed roles provided",
			claims: &entity.UserClaims{
				UserID: uuid.New(),
				Role:   entity.RoleEmployee,
			},
			allowedRoles:       []string{},
			expectedHTTPStatus: http.StatusForbidden,
			expectedBody:       httpresponse.ErrorResponse{Error: "access denied"},
			shouldCallNext:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextHandlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextHandlerCalled = true
				if tc.shouldCallNext {
					claims, ok := r.Context().Value(ClaimsContext).(*entity.UserClaims)
					assert.True(t, ok, "claims should be in context")
					assert.Equal(t, tc.claims, claims, "claims should match")
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			})

			middleware := RoleMiddleware(tc.allowedRoles...)
			handler := middleware(nextHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tc.claims != nil {
				ctx := context.WithValue(req.Context(), ClaimsContext, tc.claims)
				req = req.WithContext(ctx)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code, "unexpected HTTP status code")
			if tc.expectedBody != nil {
				var actualResponse httpresponse.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				assert.Equal(t, tc.expectedBody, actualResponse, "unexpected response body")
			}
			assert.Equal(t, tc.shouldCallNext, nextHandlerCalled, "next handler call status mismatch")
		})
	}
}
