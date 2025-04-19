package middleware

import (
	"github.com/GlebMoskalev/go-pickup-point-api/internal/metrics"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/statuswriter"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		statusWriter := statuswriter.NewResponseWriter(w)
		next.ServeHTTP(statusWriter, r)

		duration := time.Since(start).Seconds()

		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = "unknown"
		}

		metrics.HttpRequest.WithLabelValues(
			r.Method,
			routePattern,
			strconv.Itoa(statusWriter.Status()),
		).Inc()

		metrics.HttpDuration.WithLabelValues(
			r.Method,
			routePattern,
		).Observe(duration)
	})
}
