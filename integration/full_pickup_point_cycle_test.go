package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/GlebMoskalev/go-pickup-point-api/integration/helperstest"
	v1 "github.com/GlebMoskalev/go-pickup-point-api/internal/api/v1"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFullPickupPointCycle(t *testing.T) {
	ctx := context.Background()
	postgresContainer, dbConfig := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbConfig)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbConfig)

	repositories := repo.NewRepositories(dbPool)
	services := service.NewServices(repositories, helperstest.CreateTestConfig(dbConfig))

	router := v1.NewRouter(services)

	employeeToken := getEmployeeToken(t, router, "employee")
	moderatorToken := getEmployeeToken(t, router, "moderator")

	pvzID := createPickupPoint(t, router, moderatorToken)
	_ = createReception(t, router, employeeToken, pvzID)

	addProducts(t, router, employeeToken, pvzID, 50)

	closeReception(t, router, employeeToken, pvzID)

	verifyCannotAddProductWhenReceptionClosed(t, router, employeeToken, pvzID)
}

func getEmployeeToken(t *testing.T, router http.Handler, role string) string {
	loginReq := map[string]string{"role": role}
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err, "failed to marshal login request")

	req := httptest.NewRequest("POST", "/api/v1/dummyLogin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, "failed to get employee token")

	var resp struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err, "failed to unmarshal login response")
	require.NotEmpty(t, resp.Token, "token should not be empty")

	return resp.Token
}

func createPickupPoint(t *testing.T, router http.Handler, token string) string {
	createPVZReq := map[string]string{"city": "Москва"}
	reqBody, err := json.Marshal(createPVZReq)
	require.NoError(t, err, "failed to marshal create PVZ request")

	req := httptest.NewRequest("POST", "/api/v1/pvz", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code, "failed to create pickup point")

	var resp struct {
		ID               string `json:"id"`
		RegistrationDate string `json:"registration_date"`
		City             string `json:"city"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err, "failed to unmarshal create PVZ response")

	_, err = uuid.Parse(resp.ID)
	require.NoError(t, err, "PVZ ID should be a valid UUID")

	return resp.ID
}

func addProducts(t *testing.T, router http.Handler, token string, pvzID string, count int) {
	productTypes := []string{"электроника", "одежда", "обувь"}

	for i := 0; i < count; i++ {
		createProductReq := map[string]string{"pvzId": pvzID, "type": productTypes[i%len(productTypes)]}

		reqBody, err := json.Marshal(createProductReq)
		require.NoError(t, err, "failed to marshal create product request")

		req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusCreated, recorder.Code,
			fmt.Sprintf("Failed to create product %d: %s", i+1, recorder.Body.String()))

		var resp struct {
			ID          string `json:"id"`
			DateTime    string `json:"dateTime"`
			Type        string `json:"type"`
			ReceptionID string `json:"receptionId"`
		}
		err = json.Unmarshal(recorder.Body.Bytes(), &resp)
		require.NoError(t, err, "failed to unmarshal create product response")

		_, err = uuid.Parse(resp.ID)
		require.NoError(t, err, "Product ID should be a valid UUID")
	}
}

func createReception(t *testing.T, router http.Handler, token string, pvzID string) string {
	createReceptionReq := map[string]string{
		"pvz_id": pvzID,
	}
	reqBody, err := json.Marshal(createReceptionReq)
	require.NoError(t, err, "failed to marshal create reception request")

	req := httptest.NewRequest("POST", "/api/v1/receptions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code, "failed to create reception")

	var resp struct {
		ID       string `json:"id"`
		DateTime string `json:"dateTime"`
		PVZID    string `json:"pvzId"`
		Status   string `json:"status"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err, "failed to unmarshal create reception response")

	_, err = uuid.Parse(resp.ID)
	require.NoError(t, err, "reception ID should be a valid UUID")

	require.Equal(t, "in_progress", resp.Status, "reception status should be 'open'")

	return resp.ID
}

func closeReception(t *testing.T, router http.Handler, token string, pvzID string) {
	req := httptest.NewRequest("POST", "/api/v1/pvz/"+pvzID+"/close_last_reception", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, "failed to close reception")

	var resp struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err, "failed to unmarshal close reception response")

	require.Equal(t, "close reception", resp.Message, "Message should be 'close reception'")
}

func verifyCannotAddProductWhenReceptionClosed(t *testing.T, router http.Handler, token string, pvzID string) {
	createProductReq := map[string]string{
		"pvzId": pvzID,
		"type":  "электроника",
	}
	reqBody, err := json.Marshal(createProductReq)
	require.NoError(t, err, "Failed to marshal create product request")

	req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code, "should not be able to add product when reception is closed")

	var resp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err, "Failed to unmarshal error response")

	require.Equal(t, "no open reception exists", resp.Error, "error message should be 'no open reception exists'")
}
