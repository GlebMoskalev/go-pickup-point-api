package middleware

import (
	"encoding/json"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service/mocks"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name               string
		setupRequest       func(req *http.Request)
		prepareAuthService func(mockService *mocks.Auth)
		expectedHTTPStatus int
		expectedBody       any
		shouldCallNext     bool
	}{
		{
			name: "success - valid token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer valid-token")
			},
			prepareAuthService: func(mockService *mocks.Auth) {
				claims := &entity.UserClaims{
					UserID: uuid.New(),
					Role:   "employee",
				}
				mockService.On("ValidateToken", "valid-token").
					Return(claims, nil)
			},
			expectedHTTPStatus: http.StatusOK,
			shouldCallNext:     true,
		},
		{
			name:               "error - missing authorization header",
			setupRequest:       func(req *http.Request) {},
			prepareAuthService: func(mock *mocks.Auth) {},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedBody:       httpresponse.ErrorResponse{Error: "missing authorization header"},
			shouldCallNext:     false,
		},
		{
			name: "error - invalid authorization header format (no bearer)",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "invalid-token")
			},
			prepareAuthService: func(mock *mocks.Auth) {},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedBody:       httpresponse.ErrorResponse{Error: "invalid authorization header format"},
			shouldCallNext:     false,
		},
		{
			name: "error - invalid token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid-token")
			},
			prepareAuthService: func(mock *mocks.Auth) {
				mock.On("ValidateToken", "invalid-token").Return(nil, service.ErrInvalidToken)
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedBody:       httpresponse.ErrorResponse{Error: "invalid token"},
			shouldCallNext:     false,
		},
		{
			name: "error - token expired",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer expired-token")
			},
			prepareAuthService: func(mock *mocks.Auth) {
				mock.On("ValidateToken", "expired-token").Return(nil, service.ErrTokenExpired)
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedBody:       httpresponse.ErrorResponse{Error: "token expired"},
			shouldCallNext:     false,
		},
		{
			name: "error - internal server error",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer some-token")
			},
			prepareAuthService: func(mock *mocks.Auth) {
				mock.On("ValidateToken", "some-token").Return(nil, errors.New("unexpected error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedBody:       httpresponse.ErrorResponse{Error: "internal server error"},
			shouldCallNext:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService := mocks.NewAuth(t)
			tc.prepareAuthService(authService)

			nextHandlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextHandlerCalled = true
				if tc.shouldCallNext {
					claims, ok := r.Context().Value(ClaimsContext).(*entity.UserClaims)
					assert.True(t, ok, "claims should be in context")
					assert.NotNil(t, claims, "claims should not be nil")
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			})

			middleware := AuthMiddleware(authService)
			handler := middleware(nextHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			tc.setupRequest(req)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedBody != nil {
				var actualResponse httpresponse.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				assert.Equal(t, tc.expectedBody, actualResponse)
			}
			assert.Equal(t, tc.shouldCallNext, nextHandlerCalled)
		})
	}
}
