package statuswriter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	w := NewResponseWriter(rr)

	if w.statusCode != http.StatusOK {
		t.Errorf("default status should be %d, got %d", http.StatusOK, w.Status())
	}

	w.WriteHeader(http.StatusBadRequest)

	if w.Status() != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusTeapot, w.Status())
	}
}
