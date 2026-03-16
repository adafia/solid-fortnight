package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/adafia/solid-fortnight/internal/protocol"
)

type AnalyticsHandler struct {
	// For now, we'll just log or batch them. In a real system,
	// this would push to a message queue (like Redis Streams or Kafka)
	// which is then consumed by a worker that writes to TimescaleDB.
	// For simplicity in this iteration, we'll provide an interface.
	eventProcessor EventProcessor
}

type EventProcessor interface {
	Process(events []protocol.EvaluationEvent) error
}

func NewAnalyticsHandler(processor EventProcessor) *AnalyticsHandler {
	return &AnalyticsHandler{eventProcessor: processor}
}

func (h *AnalyticsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simple path routing
	if r.URL.Path == "/api/v1/events/bulk" {
		h.HandleBulkEvents(w, r)
		return
	}

	http.NotFound(w, r)
}

func (h *AnalyticsHandler) HandleBulkEvents(w http.ResponseWriter, r *http.Request) {
	var events []protocol.EvaluationEvent
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.eventProcessor.Process(events); err != nil {
		// Log the error, but return 202 Accepted to the client so they don't retry unnecessarily
		// if the problem is on our end (unless we want strict at-least-once delivery).
		// For a robust system, we might return 500.
		http.Error(w, "Failed to process events", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
