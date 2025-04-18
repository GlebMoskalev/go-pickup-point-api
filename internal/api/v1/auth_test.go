package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service/mocks"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDummyLogin(t *testing.T) {
	testCases := []struct {
		name               string
		request            any
		prepareAuthService func(mockService *mocks.Auth)
		expectedHTTPStatus int
		expectedResponse   any
	}{
		{
			name:    "invalid role",
			request: dummyLoginRequest{Role: "admin"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("DummyLogin", mock.Anything, "admin").
					Return("", service.ErrInvalidRole)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid role"},
		},
		{
			name:    "success with moderator",
			request: dummyLoginRequest{Role: "moderator"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("DummyLogin", mock.Anything, "moderator").
					Return("valid token", nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   dummyLoginResponse{Token: "valid token"},
		},
		{
			name:    "success with employee",
			request: dummyLoginRequest{Role: "employee"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("DummyLogin", mock.Anything, "employee").
					Return("valid token", nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   dummyLoginResponse{Token: "valid token"},
		},
		{
			name:    "internal server",
			request: dummyLoginRequest{Role: "employee"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("DummyLogin", mock.Anything, "employee").
					Return("", service.ErrInternal)
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server error"},
		},
		{
			name:               "invalid json in request body",
			request:            "invalid json",
			prepareAuthService: func(mockService *mocks.Auth) {},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid request body"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService := mocks.NewAuth(t)
			tc.prepareAuthService(authService)

			handler := newAuthHandler(authService)

			reqBody, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}
			req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.dummyLogin(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusOK {
				var actualResponse dummyLoginResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				assert.Equal(t, tc.expectedResponse, actualResponse)
			} else {
				var actualResponse httpresponse.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				assert.Equal(t, tc.expectedResponse, actualResponse)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	testCases := []struct {
		name               string
		request            any
		prepareAuthService func(mockService *mocks.Auth)
		expectedHTTPStatus int
		expectedResponse   any
	}{
		{
			name:    "success login",
			request: loginRequest{Email: "user@example.com", Password: "password123"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Login", mock.Anything, "user@example.com", "password123").
					Return("valid token", nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   loginResponse{Token: "valid token"},
		},
		{
			name:    "invalid credentials",
			request: loginRequest{Email: "user@example.com", Password: "wrongpassword"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Login", mock.Anything, "user@example.com", "wrongpassword").
					Return("", service.ErrInvalidCredentials)
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid credentials"},
		},
		{
			name:    "internal server error",
			request: loginRequest{Email: "user@example.com", Password: "password123"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Login", mock.Anything, "user@example.com", "password123").
					Return("", errors.New("database connection error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server error"},
		},
		{
			name:               "invalid request body",
			request:            "not a valid json",
			prepareAuthService: func(mockService *mocks.Auth) {},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid request body"},
		},
		{
			name:    "empty email",
			request: loginRequest{Email: "", Password: "password123"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Login", mock.Anything, "", "password123").
					Return("", service.ErrInvalidCredentials)
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid credentials"},
		},
		{
			name:    "empty password",
			request: loginRequest{Email: "user@example.com", Password: ""},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Login", mock.Anything, "user@example.com", "").
					Return("", service.ErrInvalidCredentials)
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid credentials"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService := mocks.NewAuth(t)
			tc.prepareAuthService(authService)

			handler := newAuthHandler(authService)

			reqBody, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}
			req := httptest.NewRequest("POST", "/login", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.login(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusOK {
				var actualResponse loginResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				assert.Equal(t, tc.expectedResponse, actualResponse)
			} else {
				var actualResponse httpresponse.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				assert.Equal(t, tc.expectedResponse, actualResponse)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	id := uuid.New()
	testCases := []struct {
		name               string
		request            any
		prepareAuthService func(mockService *mocks.Auth)
		expectedHTTPStatus int
		expectedResponse   any
	}{
		{
			name:    "successful registration",
			request: registerRequest{Email: "new@example.com", Password: "secure123", Role: "employee"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Register", mock.Anything, "new@example.com", "secure123", "employee").
					Return(&entity.User{
						ID:           id,
						Email:        "new@example.com",
						PasswordHash: "hash",
						Role:         "employee",
						CreatedAt:    time.Now(),
					}, nil)
			},
			expectedHTTPStatus: http.StatusCreated,
			expectedResponse:   registerResponse{ID: id.String(), Email: "new@example.com", Role: "employee"},
		},
		{
			name:    "user already exists",
			request: registerRequest{Email: "exists@example.com", Password: "password123", Role: "employee"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Register", mock.Anything, "exists@example.com", "password123", "employee").
					Return(nil, service.ErrUserExists)
			},
			expectedHTTPStatus: http.StatusConflict,
			expectedResponse:   httpresponse.ErrorResponse{Error: "user already exists"},
		},
		{
			name:    "invalid role",
			request: registerRequest{Email: "new@example.com", Password: "password123", Role: "admin"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Register", mock.Anything, "new@example.com", "password123", "admin").
					Return(nil, service.ErrInvalidRole)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid role"},
		},
		{
			name:    "invalid email",
			request: registerRequest{Email: "notanemail", Password: "password123", Role: "employee"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Register", mock.Anything,
					"notanemail", "password123", "employee").
					Return(nil, service.ErrInvalidEmail)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid email"},
		},
		{
			name:    "internal server error",
			request: registerRequest{Email: "new@example.com", Password: "password123", Role: "employee"},
			prepareAuthService: func(mockService *mocks.Auth) {
				mockService.On("Register", mock.Anything,
					"new@example.com", "password123", "employee").
					Return(nil, errors.New("database connection error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server error"},
		},
		{
			name:               "invalid request body",
			request:            "not a valid json",
			prepareAuthService: func(mock *mocks.Auth) {},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid request body"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService := mocks.NewAuth(t)
			tc.prepareAuthService(authService)

			handler := newAuthHandler(authService)

			reqBody, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}
			req := httptest.NewRequest("POST", "/register", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.register(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusCreated {
				var actualResponse registerResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				assert.Equal(t, tc.expectedResponse, actualResponse)
			} else {
				var actualResponse httpresponse.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				assert.Equal(t, tc.expectedResponse, actualResponse)
			}
		})
	}
}
