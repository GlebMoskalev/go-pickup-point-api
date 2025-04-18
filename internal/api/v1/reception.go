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
	"time"
)

// @Description Запрос для создания приёмки
type createReceptionRequest struct {
	// Идентификатор ПВЗ
	// format: uuid
	PVZID string `json:"pvz_id"`
}

// @Description Ответ с данными о созданной приёмке
type createReceptionResponse struct {
	// Уникальный идентификатор приёмки
	// format: uuid
	ID string `json:"id"`
	// Дата и время создания приёмки
	// format: date-time
	DateTime string `json:"dateTime"`
	// Идентификатор ПВЗ
	// format: uuid
	PVZID string `json:"pvzId"`
	// Статус приёмки
	// enum: open, close
	Status string `json:"status"`
}

// @Description Ответ с сообщением о закрытие приемки
type closeReceptionResponse struct {
	// Сообщение о статусе закрытия приёмки
	Message string `json:"message"`
}

func SetupReceptionRoutes(r chi.Router, receptionService service.Reception) {
	handler := newReceptionHandler(receptionService)

	r.With(middleware.RoleMiddleware(entity.RoleEmployee)).
		Post("/receptions", handler.createReception)
}

type receptionHandler struct {
	receptionService service.Reception
}

func newReceptionHandler(receptionService service.Reception) *receptionHandler {
	return &receptionHandler{receptionService: receptionService}
}

// @Summary Создание приёмки товаров
// @Description Создаёт новую приёмку товаров в указанном ПВЗ. Доступно только для сотрудников ПВЗ. Нельзя создать, если есть открытая приёмка.
// @Tags receptions
// @Accept json
// @Produce json
// @Param input body createReceptionRequest true "Данные для создания приёмки"
// @Success 201 {object} createReceptionResponse
// @Failure 400 {object} httpresponse.ErrorResponse "Неверный идентификатор ПВЗ или открытая приёмка существует"
// @Failure 401 {object} httpresponse.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} httpresponse.ErrorResponse "Доступ запрещён"
// @Failure 500 {object} httpresponse.ErrorResponse "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/v1/receptions [post]
func (h *receptionHandler) createReception(w http.ResponseWriter, r *http.Request) {
	var req createReceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if _, err := uuid.Parse(req.PVZID); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid pvz id")
		return
	}
	reception, err := h.receptionService.Create(r.Context(), req.PVZID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPVZID):
			httpresponse.Error(w, http.StatusBadRequest, "invalid pvz id")
		case errors.Is(err, service.ErrOpenReceptionExists):
			httpresponse.Error(w, http.StatusBadRequest, "open reception already exists")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	resp := createReceptionResponse{
		ID:       reception.ID.String(),
		DateTime: reception.DateTime.Format(time.RFC3339),
		PVZID:    reception.PVZID.String(),
		Status:   reception.Status,
	}
	httpresponse.JSON(w, http.StatusCreated, resp)
}

// @Summary Закрытие последней приёмки
// @Description Закрывает последнюю открытое приёмку в ПВЗ. Доступно только для сотрудников ПВЗ. Приёмка должна быть открытой.
// @Tags pvz
// @Accept json
// @Produce json
// @Param pvzId path string true "Идентификатор ПВЗ"
// @Success 200 {object} closeReceptionResponse
// @Failure 400 {object} httpresponse.ErrorResponse "Неверный идентификатор ПВЗ или приёмка не найдена"
// @Failure 401 {object} httpresponse.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} httpresponse.ErrorResponse "Доступ запрещён"
// @Failure 500 {object} httpresponse.ErrorResponse "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/v1/pvz/{pvzId}/close_last_reception [post]
func (h *receptionHandler) closeLastReception(w http.ResponseWriter, r *http.Request) {
	pvzID := chi.URLParam(r, "pvzId")
	if _, err := uuid.Parse(pvzID); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid pvz id")
		return
	}

	err := h.receptionService.CloseLastReception(r.Context(), pvzID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPVZID):
			httpresponse.Error(w, http.StatusBadRequest, "invalid pvz id")
		case errors.Is(err, service.ErrNoOpenReception):
			httpresponse.Error(w, http.StatusBadRequest, "no open reception exists")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httpresponse.JSON(w, http.StatusOK, closeReceptionResponse{Message: "close reception"})
}
