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
	"strconv"
	"time"
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

// @Description Ответ с данными о ПВЗ, включая приёмки и товары
type listPVZWithDetailsResponse struct {
	PVZs []pvzWithDetails `json:"pvzs"`
}

// @Description Детали ПВЗ
type pvzWithDetails struct {
	// Уникальный идентификатор ПВЗ
	// format: uuid
	ID uuid.UUID `json:"id"`
	// Дата регистрации ПВЗ
	// format: date-time
	RegistrationDate string `json:"registration_date"`
	// Город
	// enum: Москва, Санкт-Петербург, Казань
	City string `json:"city"`
	// Список приёмок
	Receptions []receptionDetails `json:"receptions"`
}

// @Description Детали приёмки
type receptionDetails struct {
	// Уникальный идентификатор приёмки
	// format: uuid
	ID uuid.UUID `json:"id"`
	// Дата и время приёмки
	// format: date-time
	DateTime string `json:"date_time"`
	// Идентификатор ПВЗ, к которому относится приёмка
	// format: uuid
	PVZID uuid.UUID `json:"pvz_id"`
	// Статус приёмки
	// enum: in_progress, closed
	Status string `json:"status"`
	// Список товаров в приёмке
	Products []productDetails `json:"products"`
}

// @Description Детали товара
type productDetails struct {
	// Уникальный идентификатор товара
	// format: uuid
	ID uuid.UUID `json:"id"`
	// Дата и время добавления товара
	// format: date-time
	DateTime string `json:"date_time"`
	// Тип товара
	// enum: электроника, одежда, продукты
	Type string `json:"type"`
	// Идентификатор приёмки, к которой относится товар
	// format: uuid
	ReceptionID uuid.UUID `json:"reception_id"`
}

func SetupPVZRoutes(r chi.Router, pvzService service.PVZ, productService service.Product, receptionService service.Reception) {
	pvzHandler := newPVZHandler(pvzService)
	productHandler := newProductHandler(productService)
	receptionHandler := newReceptionHandler(receptionService)
	r.Post("/pvz", pvzHandler.createPVZ)
	r.Post("/pvz/{pvzId}/delete_last_product", productHandler.deleteProduct)
	r.Post("/pvz/{pvzId}/close_last_reception", receptionHandler.closeLastReception)
	r.Get("/pvz", pvzHandler.listPVZWithDetails)
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

// @Summary Получение списка ПВЗ с приёмками и товарами
// @Description Доступно для сотрудников и модераторов. Возвращает список ПВЗ с информацией о приёмках и товарах, с поддержкой пагинации и фильтрации по датам приёмок.
// @Tags pvz
// @Accept json
// @Produce json
// @Param startDate query string false "Начальная дата приёмок (формат: RFC3339)" example "2025-04-01T00:00:00Z"
// @Param endDate query string false "Конечная дата приёмок (формат: RFC3339)" example "2025-04-18T23:59:59Z"
// @Param page query int false "Номер страницы (начинается с 1)" example 1
// @Param limit query int false "Количество записей на страницу (1-30)" example 10
// @Success 200 {object} listPVZWithDetailsResponse "Список ПВЗ с приёмками и товарами"
// @Failure 400 {object} httpresponse.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} httpresponse.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} httpresponse.ErrorResponse "Доступ запрещён: требуется роль сотрудника или модератора"
// @Failure 500 {object} httpresponse.ErrorResponse "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/v1/pvz [get]
func (h *pvzHandler) listPVZWithDetails(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsContext).(*entity.UserClaims)
	if !ok {
		httpresponse.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if claims.Role != entity.RoleModerator && claims.Role != entity.RoleEmployee {
		httpresponse.Error(w, http.StatusForbidden, "access denied")
		return
	}

	var (
		startDate *time.Time
		endDate   *time.Time
		page      int
		limit     int

		err error
	)

	startDateQuery := r.URL.Query().Get("startDate")
	if startDateQuery != "" {
		date, err := time.Parse(time.RFC3339, startDateQuery)
		if err != nil {
			httpresponse.Error(w, http.StatusBadRequest, "invalid start date")
			return
		}
		startDate = &date
	}

	endDateQuery := r.URL.Query().Get("endDate")
	if endDateQuery != "" {
		date, err := time.Parse(time.RFC3339, endDateQuery)
		if err != nil {
			httpresponse.Error(w, http.StatusBadRequest, "invalid end date")
			return
		}
		endDate = &date
	}

	pageQuery := r.URL.Query().Get("page")
	page, err = strconv.Atoi(pageQuery)
	if pageQuery != "" {
		if err != nil {
			httpresponse.Error(w, http.StatusBadRequest, "invalid page")
			return
		}
	}

	limitQuery := r.URL.Query().Get("limit")
	limit, err = strconv.Atoi(limitQuery)
	if limitQuery != "" {
		if err != nil {
			httpresponse.Error(w, http.StatusBadRequest, "invalid limit")
			return
		}
	}

	pvzs, err := h.pvzService.ListWithDetails(r.Context(), startDate, endDate, page, limit)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "internal server")
		return
	}
	resp := listPVZWithDetailsResponse{
		PVZs: make([]pvzWithDetails, len(pvzs)),
	}

	for i, pvz := range pvzs {
		receptions := make([]receptionDetails, len(pvz.Receptions))
		for j, r := range pvz.Receptions {
			products := make([]productDetails, len(r.Products))
			for k, p := range r.Products {
				products[k] = productDetails{
					ID:          p.ID,
					DateTime:    p.DateTime.Format(time.RFC3339),
					Type:        p.Type,
					ReceptionID: p.ReceptionID,
				}
			}
			receptions[j] = receptionDetails{
				ID:       r.Reception.ID,
				DateTime: r.Reception.DateTime.Format(time.RFC3339),
				PVZID:    r.Reception.PVZID,
				Status:   r.Reception.Status,
				Products: products,
			}
		}
		resp.PVZs[i] = pvzWithDetails{
			ID:               pvz.PVZ.ID,
			RegistrationDate: pvz.PVZ.RegistrationDate.Format(time.RFC3339),
			City:             pvz.PVZ.City,
			Receptions:       receptions,
		}
	}

	httpresponse.JSON(w, http.StatusOK, resp)
}
