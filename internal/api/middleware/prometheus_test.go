package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GlebMoskalev/go-pickup-point-api/internal/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMiddleware(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	r := chi.NewRouter()
	r.Use(PrometheusMiddleware)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	r.Post("/another-test", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET request",
			method:         "GET",
			path:           "/test",
			expectedStatus: http.StatusOK,
			expectedBody:   "test response",
		},
		{
			name:           "POST request",
			method:         "POST",
			path:           "/another-test",
			expectedStatus: http.StatusCreated,
			expectedBody:   "created",
		},
		{
			name:           "Not found route",
			method:         "GET",
			path:           "/not-exists",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code, "HTTP status code should match")
			body, err := io.ReadAll(rec.Body)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedBody, string(body), "Response body should match")
		})
	}

	httpRequestCount := testutil.ToFloat64(metrics.HttpRequest.WithLabelValues("GET", "/test", "200"))
	assert.Equal(t, 1.0, httpRequestCount, "HTTP request counter should be incremented for GET /test")

	httpRequestCount = testutil.ToFloat64(metrics.HttpRequest.WithLabelValues("POST", "/another-test", "201"))
	assert.Equal(t, 1.0, httpRequestCount, "HTTP request counter should be incremented for POST /another-test")

	httpRequestCount = testutil.ToFloat64(metrics.HttpRequest.WithLabelValues("GET", "unknown", "404"))
	assert.Equal(t, 1.0, httpRequestCount, "HTTP request counter should be incremented for not found route")

	samples, err := testutil.GatherAndCount(
		prometheus.DefaultGatherer,
		"http_request_duration_seconds",
	)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, samples, 1, "HTTP duration histogram should have samples")
}

func TestPrometheusMiddlewareWithoutChiContext(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "HTTP status code should be OK")
	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)
	assert.Equal(t, "test response", string(body), "Response body should match")

	httpRequestCount := testutil.ToFloat64(metrics.HttpRequest.WithLabelValues("GET", "unknown", "200"))
	assert.Equal(t, 1.0, httpRequestCount, "HTTP request counter should be incremented with 'unknown' path")
}
