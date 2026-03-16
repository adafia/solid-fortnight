package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adafia/solid-fortnight/internal/engine"
)

// Mock stores for testing would be better, but we can do a quick check
// of the structure and basic logic if we had mocks.
// For now, let's just ensure it compiles and has basic structure.

func TestNewEvaluatorHandler(t *testing.T) {
	h := NewEvaluatorHandler(nil, nil, nil, nil)
	if h == nil {
		t.Fatal("Expected NewEvaluatorHandler to return a handler")
	}
}

func TestEvaluatorHandler_ServeHTTP_MethodNotAllowed(t *testing.T) {
	h := NewEvaluatorHandler(nil, nil, nil, nil)
	req, _ := http.NewRequest(http.MethodGet, "/evaluate", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
