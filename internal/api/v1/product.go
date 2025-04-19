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

// @Description Запрос для добавления товара
type createProductRequest struct {
	// Идентификатор ПВЗ
	// format: uuid
	PVZID string `json:"pvzId"`
	// Тип товара
	// enum: электроника, одежда, обувь
	Type string `json:"type"`
}

// @Description Ответ с данными о добавленном товаре
type createProductResponse struct {
	// Уникальный идентификатор товара
	// format: uuid
	ID string `json:"id"`
	// Дата и время добавления товара
	// format: date-time
	DateTime string `json:"dateTime"`
	// Тип товара
	Type string `json:"type"`
	// Идентификатор приёмки
	// format: uuid
	ReceptionID string `json:"receptionId"`
}

// @Description Ответ с сообщением об удалении товара
type deleteProductResponse struct {
	// Сообщение об успешном удалении товара
	Message string `json:"message"`
}

func SetupProductRoutes(r chi.Router, productService service.Product) {
	handler := newProductHandler(productService)

	r.With(middleware.RoleMiddleware(entity.RoleEmployee)).
		Post("/", handler.createProduct)
}

type productHandler struct {
	productService service.Product
}

func newProductHandler(productService service.Product) *productHandler {
	return &productHandler{productService: productService}
}

// @Summary Добавление товара в приёмку
// @Description Добавляет товар в последнюю незакрытую приёмку в указанном ПВЗ. Доступно только для сотрудников ПВЗ. Требуется незакрытая приёмка.
// @Tags products
// @Accept json
// @Produce json
// @Param input body createProductRequest true "Данные для добавления товара"
// @Success 201 {object} createProductResponse
// @Failure 400 {object} httpresponse.ErrorResponse "Неверный идентификатор ПВЗ, тип товара или отсутствие открытой приёмки"
// @Failure 401 {object} httpresponse.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} httpresponse.ErrorResponse "Доступ запрещён"
// @Failure 500 {object} httpresponse.ErrorResponse "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/v1/products [post]
func (h *productHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if _, err := uuid.Parse(req.PVZID); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid pvz id")
		return
	}

	product, err := h.productService.Create(r.Context(), req.PVZID, req.Type)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPVZID):
			httpresponse.Error(w, http.StatusBadRequest, "invalid pvz id")
		case errors.Is(err, service.ErrNoOpenReception):
			httpresponse.Error(w, http.StatusBadRequest, "no open reception exists")
		case errors.Is(err, service.ErrInvalidProductType):
			httpresponse.Error(w, http.StatusBadRequest, "invalid product type")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	resp := createProductResponse{
		ID:          product.ID.String(),
		DateTime:    product.DateTime.Format(time.RFC3339),
		Type:        product.Type,
		ReceptionID: product.ReceptionID.String(),
	}
	httpresponse.JSON(w, http.StatusCreated, resp)
}

// @Summary Удаление последнего добавленного товара
// @Description Удаляет последний добавленный товар в последней незакрытой приёмке указанного ПВЗ. Доступно только для сотрудников ПВЗ. Требуется наличие незакрытой приёмки и хотя бы одного товара в ней.
// @Tags pvz
// @Produce json
// @Param pvzId path string true "Идентификатор ПВЗ (uuid)"
// @Success 200 {object} deleteProductResponse "Сообщение об успешном удалении"
// @Failure 400 {object} httpresponse.ErrorResponse "Неверный идентификатор ПВЗ, отсутствие открытой приёмки или отсутствие товаров в приёмке"
// @Failure 401 {object} httpresponse.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} httpresponse.ErrorResponse "Доступ запрещён"
// @Failure 500 {object} httpresponse.ErrorResponse "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/v1/pvz/{pvzId}/delete_last_product [post]
func (h *productHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	pvzID := chi.URLParam(r, "pvzId")
	if _, err := uuid.Parse(pvzID); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid pvz id")
		return
	}

	err := h.productService.DeleteLastProduct(r.Context(), pvzID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPVZID):
			httpresponse.Error(w, http.StatusBadRequest, "invalid pvz id")
		case errors.Is(err, service.ErrNoOpenReception):
			httpresponse.Error(w, http.StatusBadRequest, "no open reception exists")
		case errors.Is(err, service.ErrNoProducts):
			httpresponse.Error(w, http.StatusBadRequest, "no products in open reception")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httpresponse.JSON(w, http.StatusOK, deleteProductResponse{Message: "successfully delete"})
}
