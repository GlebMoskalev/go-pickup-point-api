package httpresponse

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	body := map[string]string{"message": "hello"}

	JSON(rec, http.StatusOK, body)

	resp := rec.Result()
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["message"] != "hello" {
		t.Errorf("body[message] = %q, want %q", result["message"], "hello")
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()

	Error(rec, http.StatusBadRequest, "invalid input")

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	var result ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if result.Error != "invalid input" {
		t.Errorf("Error = %q, want %q", result.Error, "invalid input")
	}
}
