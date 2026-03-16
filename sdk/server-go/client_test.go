package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/adafia/solid-fortnight/internal/engine"
)

func TestClient_LocalEvaluation(t *testing.T) {
	// 1. Setup Mock Evaluator
	flagKey := "test-flag"
	mockFlag := engine.FlagConfig{
		ID:      "flag-1",
		Key:     flagKey,
		Enabled: true,
		Variations: []engine.Variation{
			{ID: "v1", Key: "on", Value: json.RawMessage(`true`)},
			{ID: "v2", Key: "off", Value: json.RawMessage(`false`)},
		},
		DefaultVariationID: stringPtr("v1"),
	}

	evalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/evaluate" {
			json.NewEncoder(w).Encode([]engine.FlagConfig{mockFlag})
			return
		}
		http.NotFound(w, r)
	}))
	defer evalServer.Close()

	// 2. Setup Mock Streamer
	streamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, ": keep-alive\n\n")
		// Keep connection open
		select {
		case <-r.Context().Done():
			return
		case <-time.After(1 * time.Second):
			return
		}
	}))
	defer streamServer.Close()

	// 3. Create SDK Client
	client, err := NewClient(Config{
		EvaluatorURL:  evalServer.URL,
		StreamerURL:   streamServer.URL,
		EnvironmentID: "env-1",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 4. Test Evaluation
	ctx := engine.UserContext{ID: "user-1"}
	val := client.BoolVariation(flagKey, ctx, false)
	if val != true {
		t.Errorf("Expected true, got %v", val)
	}

	// 5. Test missing flag
	val = client.BoolVariation("missing", ctx, false)
	if val != false {
		t.Errorf("Expected false (default), got %v", val)
	}
}

func TestClient_RealtimeUpdates(t *testing.T) {
	var mu sync.Mutex
	flagKey := "test-flag"
	mockFlag := engine.FlagConfig{
		ID:      "flag-1",
		Key:     flagKey,
		Enabled: true,
		Variations: []engine.Variation{
			{ID: "v1", Key: "on", Value: json.RawMessage(`true`)},
			{ID: "v2", Key: "off", Value: json.RawMessage(`false`)},
		},
		DefaultVariationID: stringPtr("v1"),
	}

	// 1. Setup Mock Evaluator with state
	evalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		json.NewEncoder(w).Encode([]engine.FlagConfig{mockFlag})
	}))
	defer evalServer.Close()

	// 2. Setup Mock Streamer with broadcast capability
	updateChan := make(chan struct{})
	streamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		f, _ := w.(http.Flusher)
		f.Flush()

		for {
			select {
			case <-updateChan:
				fmt.Fprintf(w, "data: update\n\n")
				f.Flush()
			case <-r.Context().Done():
				return
			case <-time.After(5 * time.Second):
				return
			}
		}
	}))
	defer streamServer.Close()

	// 3. Create SDK Client
	client, err := NewClient(Config{
		EvaluatorURL:  evalServer.URL,
		StreamerURL:   streamServer.URL,
		EnvironmentID: "env-1",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 4. Initial check
	ctx := engine.UserContext{ID: "user-1"}
	if client.BoolVariation(flagKey, ctx, false) != true {
		t.Errorf("Expected true")
	}

	// 5. Update mock flag and trigger update
	mu.Lock()
	mockFlag.DefaultVariationID = stringPtr("v2") // Change default to 'off'
	mu.Unlock()
	
	updateChan <- struct{}{}

	// 6. Wait for SDK to sync
	time.Sleep(200 * time.Millisecond)

	// 7. Check updated evaluation
	if client.BoolVariation(flagKey, ctx, true) != false {
		t.Errorf("Expected false after update")
	}
}

func TestClient_DeltaUpdates(t *testing.T) {
	flagKey := "test-flag"
	mockFlag := engine.FlagConfig{
		ID:      "flag-1",
		Key:     flagKey,
		Enabled: true,
		Variations: []engine.Variation{
			{ID: "v1", Key: "on", Value: json.RawMessage(`true`)},
			{ID: "v2", Key: "off", Value: json.RawMessage(`false`)},
		},
		DefaultVariationID: stringPtr("v1"),
	}

	// 1. Setup Mock Evaluator - only returns flag initially
	fetchCount := 0
	evalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		json.NewEncoder(w).Encode([]engine.FlagConfig{mockFlag})
	}))
	defer evalServer.Close()

	// 2. Setup Mock Streamer with delta update broadcast
	deltaChan := make(chan string)
	streamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f, _ := w.(http.Flusher)
		f.Flush()

		for {
			select {
			case msg := <-deltaChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				f.Flush()
			case <-r.Context().Done():
				return
			case <-time.After(5 * time.Second):
				return
			}
		}
	}))
	defer streamServer.Close()

	// 3. Create SDK Client
	client, err := NewClient(Config{
		EvaluatorURL:  evalServer.URL,
		StreamerURL:   streamServer.URL,
		EnvironmentID: "env-1",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 4. Initial check
	if fetchCount != 1 {
		t.Errorf("Expected 1 initial fetch, got %d", fetchCount)
	}

	// 5. Send Delta Update via Stream
	updatedFlag := mockFlag
	updatedFlag.DefaultVariationID = stringPtr("v2") // Change to off
	payload := map[string]interface{}{
		"environment_id": "env-1",
		"data":           updatedFlag,
	}
	data, _ := json.Marshal(payload)
	deltaChan <- string(data)

	// 6. Wait for SDK to sync
	time.Sleep(200 * time.Millisecond)

	// 7. Verify updated evaluation and NO additional fetch
	if client.BoolVariation(flagKey, engine.UserContext{ID: "user-1"}, true) != false {
		t.Errorf("Expected false after delta update")
	}

	if fetchCount != 1 {
		t.Errorf("Expected no additional fetch for delta update, but got %d total fetches", fetchCount)
	}
}

func stringPtr(s string) *string {
	return &s
}
