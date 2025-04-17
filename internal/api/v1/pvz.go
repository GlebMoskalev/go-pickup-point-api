package v1

import (
	"encoding/json"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/api/middleware"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

// @Description Запрос для создания ПВЗ
type createPVZRequest struct {
	// Город
	// enum: Москва, Санкт-Петербург, Казань
	City string `json:"city"`
}

// @Description Ответ с данными о созданном ПВЗ
type createPVZResponse struct {
	// Уникальный идентификатор ПВЗ
	// format: uuid
	ID uuid.UUID `json:"id"`
	// Дата регистрации ПВЗ
	// format: date-time
	RegistrationDate string `json:"registration_date"`
	// Город
	// enum: Москва, Санкт-Петербург, Казань
	City string `json:"city"`
}

func SetupPVZRoutes(r chi.Router, pvzService service.PVZ, productService service.Product, receptionService service.Reception) {
	pvzHandler := newPVZHandler(pvzService)
	productHandler := newProductHandler(productService)
	receptionHandler := newReceptionHandler(receptionService)
	r.Post("/pvz", pvzHandler.createPVZ)
	r.Post("/pvz/{pvzId}/delete_last_product", productHandler.deleteProduct)
	r.Post("/pvz/{pvzId}/close_last_reception", receptionHandler.closeLastReception)
}

type pvzHandler struct {
	pvzService service.PVZ
}

func newPVZHandler(pvzService service.PVZ) *pvzHandler {
	return &pvzHandler{pvzService: pvzService}
}

// @Summary Создание ПВЗ
// @Description Только для модераторов. Создаёт пункт выдачи заказов (ПВЗ) в одном из поддерживаемых городов: Москва, Санкт-Петербург, Казань.
// @Tags pvz
// @Accept json
// @Produce json
// @Param input body createPVZRequest true "Данные для создания ПВЗ"
// @Success 201 {object} createPVZResponse "ПВЗ успешно создан"
// @Failure 400 {object} httpresponse.ErrorResponse "Неверный город или некорректное тело запроса"
// @Failure 401 {object} httpresponse.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} httpresponse.ErrorResponse "Доступ запрещён: требуется роль модератора"
// @Failure 500 {object} httpresponse.ErrorResponse "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/v1/pvz [post]
func (h *pvzHandler) createPVZ(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsContext).(*entity.UserClaims)
	if !ok {
		httpresponse.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if claims.Role != entity.RoleModerator {
		httpresponse.Error(w, http.StatusForbidden, "access denied")
		return
	}

	var req createPVZRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pvz, err := h.pvzService.Create(r.Context(), req.City)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCity):
			httpresponse.Error(w, http.StatusBadRequest, "invalid city")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	resp := createPVZResponse{
		ID:               pvz.ID,
		RegistrationDate: pvz.RegistrationDate.Format("2006-01-02 15:04"),
		City:             pvz.City,
	}
	httpresponse.JSON(w, http.StatusCreated, resp)
}
