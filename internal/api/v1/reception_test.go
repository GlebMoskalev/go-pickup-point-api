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

func TestCreateReception(t *testing.T) {
	testCases := []struct {
		name                    string
		request                 any
		prepareReceptionService func(mockService *mocks.Reception)
		expectedHTTPStatus      int
		expectedResponse        any
	}{
		{
			name:    "successful creation",
			request: createReceptionRequest{PVZID: uuid.New().String()},
			prepareReceptionService: func(mockService *mocks.Reception) {
				receptionID := uuid.New()
				pvzID := uuid.New()
				mockService.On("Create", mock.Anything, mock.AnythingOfType("string")).
					Return(&entity.Reception{
						ID:       receptionID,
						DateTime: time.Now(),
						PVZID:    pvzID,
						Status:   "in_progress",
					}, nil)
			},
			expectedHTTPStatus: http.StatusCreated,
			expectedResponse: createReceptionResponse{
				ID:       "id",
				DateTime: "date",
				PVZID:    "pvz_id",
				Status:   "in_progress",
			},
		},
		{
			name:                    "invalid pvz id",
			request:                 createReceptionRequest{PVZID: "not-a-uuid"},
			prepareReceptionService: func(mockService *mocks.Reception) {},
			expectedHTTPStatus:      http.StatusBadRequest,
			expectedResponse:        httpresponse.ErrorResponse{Error: "invalid pvz id"},
		},
		{
			name:    "open reception exists",
			request: createReceptionRequest{PVZID: uuid.New().String()},
			prepareReceptionService: func(mockService *mocks.Reception) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, service.ErrOpenReceptionExists)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "open reception already exists"},
		},
		{
			name:    "invalid pvz id from service",
			request: createReceptionRequest{PVZID: uuid.New().String()},
			prepareReceptionService: func(mockService *mocks.Reception) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, service.ErrInvalidPVZID)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid pvz id"},
		},
		{
			name:    "internal server error",
			request: createReceptionRequest{PVZID: uuid.New().String()},
			prepareReceptionService: func(mockService *mocks.Reception) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, errors.New("database error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server error"},
		},
		{
			name:                    "invalid request body",
			request:                 "invalid json",
			prepareReceptionService: func(mockService *mocks.Reception) {},
			expectedHTTPStatus:      http.StatusBadRequest,
			expectedResponse:        httpresponse.ErrorResponse{Error: "invalid request body"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			receptionService := mocks.NewReception(t)
			tc.prepareReceptionService(receptionService)

			handler := newReceptionHandler(receptionService)

			reqBody, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}
			req := httptest.NewRequest("POST", "/receptions", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.createReception(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusCreated {
				var actualResponse createReceptionResponse
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				_, err = uuid.Parse(actualResponse.ID)
				assert.NoError(t, err, "ID should be a valid UUID")
				_, err = uuid.Parse(actualResponse.PVZID)
				assert.NoError(t, err, "PVZID should be a valid UUID")
				_, err = time.Parse(time.RFC3339, actualResponse.DateTime)
				assert.NoError(t, err, "DateTime should be in correct format")
				assert.Equal(t, tc.expectedResponse.(createReceptionResponse).Status, actualResponse.Status)
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

func TestCloseLastReception(t *testing.T) {
	testCases := []struct {
		name                    string
		pvzID                   string
		prepareReceptionService func(mockService *mocks.Reception)
		expectedHTTPStatus      int
		expectedResponse        any
	}{
		{
			name:  "successful closure",
			pvzID: uuid.New().String(),
			prepareReceptionService: func(mockService *mocks.Reception) {
				mockService.On("CloseLastReception", mock.Anything, mock.AnythingOfType("string")).
					Return(nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   closeReceptionResponse{Message: "close reception"},
		},
		{
			name:                    "invalid pvz id",
			pvzID:                   "not-a-uuid",
			prepareReceptionService: func(mockService *mocks.Reception) {},
			expectedHTTPStatus:      http.StatusBadRequest,
			expectedResponse:        httpresponse.ErrorResponse{Error: "invalid pvz id"},
		},
		{
			name:  "no open reception",
			pvzID: uuid.New().String(),
			prepareReceptionService: func(mockService *mocks.Reception) {
				mockService.On("CloseLastReception", mock.Anything, mock.AnythingOfType("string")).
					Return(service.ErrNoOpenReception)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "no open reception exists"},
		},
		{
			name:  "invalid pvz id from service",
			pvzID: uuid.New().String(),
			prepareReceptionService: func(mockService *mocks.Reception) {
				mockService.On("CloseLastReception", mock.Anything, mock.AnythingOfType("string")).
					Return(service.ErrInvalidPVZID)
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   httpresponse.ErrorResponse{Error: "invalid pvz id"},
		},
		{
			name:  "internal server error",
			pvzID: uuid.New().String(),
			prepareReceptionService: func(mockService *mocks.Reception) {
				mockService.On("CloseLastReception", mock.Anything, mock.AnythingOfType("string")).
					Return(errors.New("database error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   httpresponse.ErrorResponse{Error: "internal server error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			receptionService := mocks.NewReception(t)
			tc.prepareReceptionService(receptionService)

			handler := newReceptionHandler(receptionService)

			r := chi.NewRouter()
			r.Post("/pvz/{pvzId}/close_last_reception", handler.closeLastReception)
			req := httptest.NewRequest("POST", "/pvz/"+tc.pvzID+"/close_last_reception", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedHTTPStatus, rec.Code)

			if tc.expectedHTTPStatus == http.StatusOK {
				var actualResponse closeReceptionResponse
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
