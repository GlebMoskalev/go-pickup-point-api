package v1

import (
	"encoding/json"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// @Description Запрос для получения тестового токена авторизации
type dummyLoginRequest struct {
	// Роль пользователя (employee или moderator)
	// enum: employee,moderator
	Role string `json:"role"`
}

// @Description Ответ с тестовым токеном авторизации
type dummyLoginResponse struct {
	// JWT-токен для аутентификации
	Token string `json:"token"`
}

// @Description Запрос для аутентификации пользователя
type loginRequest struct {
	// Электронная почта пользователя
	// format: email
	Email string `json:"email"`
	// Пароль пользователя
	Password string `json:"password"`
}

// @Description Ответ с токеном авторизации после успешной аутентификации
type loginResponse struct {
	// JWT-токен для аутентификации
	Token string `json:"token"`
}

// @Description Запрос для регистрации нового пользователя
type registerRequest struct {
	// Электронная почта пользователя
	// format: email
	Email string `json:"email"`
	// Пароль пользователя
	Password string `json:"password"`
	// Роль пользователя (employee или moderator)
	// enum: employee,moderator
	Role string `json:"role"`
}

// @Description Ответ с данными зарегистрированного пользователя
type registerResponse struct {
	// Уникальный идентификатор пользователя
	// format: uuid
	ID string `json:"id"`
	// Электронная почта пользователя
	// format: email
	Email string `json:"email"`
	// Роль пользователя (employee или moderator)
	// enum: employee,moderator
	Role string `json:"role"`
}

func SetupAuthRoutes(r chi.Router, authService service.Auth) {
	handler := newAuthHandler(authService)
	r.Post("/dummyLogin", handler.dummyLogin)
	r.Post("/login", handler.login)
	r.Post("/register", handler.register)
}

type authHandler struct {
	authService service.Auth
}

func newAuthHandler(authService service.Auth) *authHandler {
	return &authHandler{authService: authService}
}

// @Summary Dummy login
// @Description Получение тестового токена авторизации по роли
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dummyLoginRequest true "Данные для входа"
// @Success 200 {object} dummyLoginResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /api/v1/dummyLogin [post]
func (h *authHandler) dummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := h.authService.DummyLogin(r.Context(), req.Role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRole):
			httpresponse.Error(w, http.StatusBadRequest, "invalid role")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpresponse.JSON(w, http.StatusOK, dummyLoginResponse{Token: token})
}

// @Summary Login
// @Description Аутентификация пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param input body loginRequest true "Учетные данные"
// @Success 200 {object} loginResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 401 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /api/v1/login [post]
func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			httpresponse.Error(w, http.StatusUnauthorized, "invalid credentials")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpresponse.JSON(w, http.StatusOK, loginResponse{Token: token})
}

// @Summary Register
// @Description Регистрация нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param input body registerRequest true "Данные для регистрации"
// @Success 201 {object} registerResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 409 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /api/v1/register [post]
func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRole):
			httpresponse.Error(w, http.StatusBadRequest, "invalid role")
		case errors.Is(err, service.ErrUserExists):
			httpresponse.Error(w, http.StatusConflict, "user already exists")
		case errors.Is(err, service.ErrInvalidEmail):
			httpresponse.Error(w, http.StatusBadRequest, "invalid email")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	resp := registerResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Role:  user.Role,
	}
	httpresponse.JSON(w, http.StatusCreated, resp)
}
