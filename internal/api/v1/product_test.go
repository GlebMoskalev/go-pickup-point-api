package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service/mocks"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/httpresponse"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateProduct(t *testing.T) {
	testCases := []struct {
		name                  string
		request               any
		prepareProductService func(mockService *mocks.Product)
		expectedHTTPStatus    int
		expectedResponse      any
	}{
		{
			name:    "successful creation",
			request: createProductRequest{PVZID: uuid.New().String(), Type: "электроника"},
			prepareProductService: func(mockService *mocks.Product) {
				productID := uuid.New()
				receptionID := uuid.New()
				mockService.On("Create", mock.Anything, mock.AnythingOfType("string"), "электроника").
					Return(&entity.Product{
						ID:          productID,
						DateTime:    time.Now(),
						Type:        "электроника",
						ReceptionID: receptionID,
					}, nil)
			},
			expectedHTTPStatus: http.StatusCreated,
			expectedResponse: createProductResponse{
				ID:          "id",
				DateTime:    "date",
				Type:        "электроника",
				ReceptionID: "reception_id",
			},
		},
		{
			name:                  "invalid pvz id",
			request:               createProductRequest{PVZID: "not-a-uuid", Type: "электроника"},
			prepareProductService: func(mockService *mocks.Product) {},
			expectedHTTPStatus:    http.StatusBadRequest,
			expectedResponse:      httpresponse.ErrorResponse{Error: "invalid pvz id"},
		},
		{
			name:    "invalid product type",
			request: createProductRequest{PVZID: uuid.New().String(), Type: "неизвестный"},
			prepareProductService: func(mockService *mocks.Product) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("string"), "неизвестный").
					Return(nil, service.ErrInvalidProductType)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid product type"},
		},
		{
			name:    "no open reception",
			request: createProductRequest{PVZID: uuid.New().String(), Type: "электроника"},
			prepareProductService: func(mockService *mocks.Product) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("string"), "электроника").
					Return(nil, service.ErrNoOpenReception)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "no open reception exists"},
		},
		{
			name:    "internal server error",
			request: createProductRequest{PVZID: uuid.New().String(), Type: "электроника"},
			prepareProductService: func(mockService *mocks.Product) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("string"), "электроника").
					Return(nil, errors.New("database error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server error"},
		},
		{
			name:                  "invalid request body",
			request:               "invalid json",
			prepareProductService: func(mockService *mocks.Product) {},
			expectedHTTPStatus:    http.StatusBadRequest,
			expectedResponse:      httpresponse.ErrorResponse{Error: "invalid request body"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			productService := mocks.NewProduct(t)
			tc.prepareProductService(productService)

			handler := newProductHandler(productService)

			reqBody, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}
			req := httptest.NewRequest("POST", "/products", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.createProduct(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusCreated {
				var actualResponse createProductResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				_, err = uuid.Parse(actualResponse.ID)
				assert.NoError(t, err, "ID should be a valid UUID")
				_, err = uuid.Parse(actualResponse.ReceptionID)
				assert.NoError(t, err, "ReceptionID should be a valid UUID")
				_, err = time.Parse(time.RFC3339, actualResponse.DateTime)
				assert.NoError(t, err, "DateTime should be in correct format")
				assert.Equal(t, tc.expectedResponse.(createProductResponse).Type, actualResponse.Type)
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

func TestDeleteProduct(t *testing.T) {
	testCases := []struct {
		name                  string
		pvzID                 string
		prepareProductService func(mockService *mocks.Product)
		expectedHTTPStatus    int
		expectedResponse      any
	}{
		{
			name:  "successful deletion",
			pvzID: uuid.New().String(),
			prepareProductService: func(mockService *mocks.Product) {
				mockService.On("DeleteLastProduct", mock.Anything, mock.AnythingOfType("string")).
					Return(nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   deleteProductResponse{Message: "successfully delete"},
		},
		{
			name:                  "invalid pvz id",
			pvzID:                 "not-a-uuid",
			prepareProductService: func(mockService *mocks.Product) {},
			expectedHTTPStatus:    http.StatusBadRequest,
			expectedResponse:      httpresponse.ErrorResponse{Error: "invalid pvz id"},
		},
		{
			name:  "no open reception",
			pvzID: uuid.New().String(),
			prepareProductService: func(mockService *mocks.Product) {
				mockService.On("DeleteLastProduct", mock.Anything, mock.AnythingOfType("string")).
					Return(service.ErrNoOpenReception)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "no open reception exists"},
		},
		{
			name:  "no products in reception",
			pvzID: uuid.New().String(),
			prepareProductService: func(mockService *mocks.Product) {
				mockService.On("DeleteLastProduct", mock.Anything, mock.AnythingOfType("string")).
					Return(service.ErrNoProducts)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "no products in open reception"},
		},
		{
			name:  "internal server error",
			pvzID: uuid.New().String(),
			prepareProductService: func(mockService *mocks.Product) {
				mockService.On("DeleteLastProduct", mock.Anything, mock.AnythingOfType("string")).
					Return(errors.New("database error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			productService := mocks.NewProduct(t)
			tc.prepareProductService(productService)

			handler := newProductHandler(productService)

			r := chi.NewRouter()
			r.Post("/pvz/{pvzId}/delete_last_product", handler.deleteProduct)
			req := httptest.NewRequest("POST", "/pvz/"+tc.pvzID+"/delete_last_product", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusOK {
				var actualResponse deleteProductResponse
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
