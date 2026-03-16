package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adafia/solid-fortnight/internal/protocol"
)

type mockProcessor struct {
	events []protocol.EvaluationEvent
	err    error
}

func (m *mockProcessor) Process(events []protocol.EvaluationEvent) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, events...)
	return nil
}

func TestAnalyticsHandler_HandleBulkEvents(t *testing.T) {
	processor := &mockProcessor{}
	handler := NewAnalyticsHandler(processor)

	events := []protocol.EvaluationEvent{
		{
			ProjectID:     "proj-1",
			EnvironmentID: "env-1",
			FlagKey:       "flag-1",
			UserID:        "user-1",
			VariationKey:  "on",
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/bulk", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	if len(processor.events) != 1 {
		t.Errorf("Expected 1 event processed, got %d", len(processor.events))
	}
}

func TestAnalyticsHandler_InvalidMethod(t *testing.T) {
	processor := &mockProcessor{}
	handler := NewAnalyticsHandler(processor)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/bulk", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
