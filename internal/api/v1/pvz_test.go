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

func TestCreatePVZ(t *testing.T) {
	testCases := []struct {
		name               string
		request            any
		preparePVZService  func(mockService *mocks.PVZ)
		expectedHTTPStatus int
		expectedResponse   any
	}{
		{
			name:    "successful creation",
			request: createPVZRequest{City: "Москва"},
			preparePVZService: func(mockService *mocks.PVZ) {
				pvzID := uuid.New()
				mockService.On("Create", mock.Anything, "Москва").
					Return(&entity.PVZ{
						ID:               pvzID,
						RegistrationDate: time.Now(),
						City:             "Москва",
					}, nil)
			},
			expectedHTTPStatus: http.StatusCreated,
			expectedResponse: createPVZResponse{
				ID:               "id",
				RegistrationDate: "date",
				City:             "Москва",
			},
		},
		{
			name:    "invalid city",
			request: createPVZRequest{City: "НекорректныйГород"},
			preparePVZService: func(mockService *mocks.PVZ) {
				mockService.On("Create", mock.Anything, "НекорректныйГород").
					Return(nil, service.ErrInvalidCity)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid city"},
		},
		{
			name:    "internal server error",
			request: createPVZRequest{City: "Москва"},
			preparePVZService: func(mockService *mocks.PVZ) {
				mockService.On("Create", mock.Anything, "Москва").
					Return(nil, errors.New("database error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server error"},
		},
		{
			name:               "invalid request body",
			request:            "invalid json",
			preparePVZService:  func(mockService *mocks.PVZ) {},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid request body"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pvzService := mocks.NewPVZ(t)
			tc.preparePVZService(pvzService)

			handler := newPVZHandler(pvzService)

			reqBody, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}
			req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.createPVZ(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusCreated {
				var actualResponse createPVZResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				_, err = uuid.Parse(actualResponse.ID)
				assert.NoError(t, err, "ID should be a valid UUID")
				_, err = time.Parse(time.RFC3339, actualResponse.RegistrationDate)
				assert.NoError(t, err, "RegistrationDate should be in correct format")
				assert.Equal(t, tc.expectedResponse.(createPVZResponse).City, actualResponse.City)
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

func TestListPVZWithDetails(t *testing.T) {
	testCases := []struct {
		name               string
		queryParams        map[string]string
		preparePVZService  func(mockService *mocks.PVZ)
		expectedHTTPStatus int
		expectedResponse   any
	}{
		{
			name: "successful list with details",
			queryParams: map[string]string{
				"startDate": "2025-04-01T00:00:00Z",
				"endDate":   "2025-04-18T23:59:59Z",
				"page":      "1",
				"limit":     "10",
			},
			preparePVZService: func(mockService *mocks.PVZ) {
				pvzID := uuid.New()
				receptionID := uuid.New()
				productID := uuid.New()
				startDate, _ := time.Parse(time.RFC3339, "2025-04-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-04-18T23:59:59Z")
				mockService.On("ListWithDetails", mock.Anything, &startDate, &endDate, 1, 10).
					Return([]entity.PVZWithDetails{
						{
							PVZ: entity.PVZ{
								ID:               pvzID,
								RegistrationDate: time.Now(),
								City:             "Москва",
							},
							Receptions: []entity.ReceptionDetails{
								{
									Reception: entity.Reception{
										ID:       receptionID,
										DateTime: time.Now(),
										PVZID:    pvzID,
										Status:   "open",
									},
									Products: []entity.Product{
										{
											ID:          productID,
											DateTime:    time.Now(),
											Type:        "электроника",
											ReceptionID: receptionID,
										},
									},
								},
							},
						},
					}, nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse: listPVZWithDetailsResponse{
				PVZs: []pvzWithDetails{
					{
						ID:               uuid.UUID{},
						RegistrationDate: "date",
						City:             "Москва",
						Receptions: []receptionDetails{
							{
								ID:       uuid.UUID{},
								DateTime: "date",
								PVZID:    uuid.UUID{},
								Status:   "in_progress",
								Products: []productDetails{
									{
										ID:          uuid.UUID{},
										DateTime:    "date",
										Type:        "электроника",
										ReceptionID: uuid.UUID{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "invalid start date",
			queryParams: map[string]string{
				"startDate": "invalid-date",
			},
			preparePVZService:  func(mockService *mocks.PVZ) {},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid start date"},
		},
		{
			name: "invalid end date",
			queryParams: map[string]string{
				"endDate": "invalid-date",
			},
			preparePVZService:  func(mockService *mocks.PVZ) {},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid end date"},
		},
		{
			name: "invalid page",
			queryParams: map[string]string{
				"page": "invalid",
			},
			preparePVZService:  func(mockService *mocks.PVZ) {},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid page"},
		},
		{
			name: "invalid limit",
			queryParams: map[string]string{
				"limit": "invalid",
			},
			preparePVZService:  func(mockService *mocks.PVZ) {},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid limit"},
		},
		{
			name: "internal server error",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "10",
			},
			preparePVZService: func(mockService *mocks.PVZ) {
				mockService.On("ListWithDetails", mock.Anything, (*time.Time)(nil), (*time.Time)(nil), 1, 10).
					Return(nil, errors.New("database error"))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server"},
		},
		{
			name:        "no query params",
			queryParams: map[string]string{},
			preparePVZService: func(mockService *mocks.PVZ) {
				mockService.On("ListWithDetails", mock.Anything, (*time.Time)(nil), (*time.Time)(nil), 0, 0).
					Return([]entity.PVZWithDetails{}, nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   listPVZWithDetailsResponse{PVZs: []pvzWithDetails{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pvzService := mocks.NewPVZ(t)
			tc.preparePVZService(pvzService)

			handler := newPVZHandler(pvzService)

			req := httptest.NewRequest("GET", "/pvz", nil)
			q := req.URL.Query()
			for k, v := range tc.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			rec := httptest.NewRecorder()

			handler.listPVZWithDetails(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusOK {
				var actualResponse listPVZWithDetailsResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				for _, pvz := range actualResponse.PVZs {
					_, err := uuid.Parse(pvz.ID.String())
					assert.NoError(t, err, "PVZ ID should be a valid UUID")
					_, err = time.Parse(time.RFC3339, pvz.RegistrationDate)
					assert.NoError(t, err, "RegistrationDate should be in RFC3339 format")
					for _, reception := range pvz.Receptions {
						_, err := uuid.Parse(reception.ID.String())
						assert.NoError(t, err, "Reception ID should be a valid UUID")
						_, err = time.Parse(time.RFC3339, reception.DateTime)
						assert.NoError(t, err, "Reception DateTime should be in RFC3339 format")
						_, err = uuid.Parse(reception.PVZID.String())
						assert.NoError(t, err, "Reception PVZID should be a valid UUID")
						for _, product := range reception.Products {
							_, err := uuid.Parse(product.ID.String())
							assert.NoError(t, err, "Product ID should be a valid UUID")
							_, err = time.Parse(time.RFC3339, product.DateTime)
							assert.NoError(t, err, "Product DateTime should be in RFC3339 format")
							_, err = uuid.Parse(product.ReceptionID.String())
							assert.NoError(t, err, "Product ReceptionID should be a valid UUID")
						}
					}
				}
				assert.Equal(t, len(tc.expectedResponse.(listPVZWithDetailsResponse).PVZs), len(actualResponse.PVZs))
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
