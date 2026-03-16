package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/adafia/solid-fortnight/apps/analytics/service"
	"github.com/adafia/solid-fortnight/internal/protocol"
	"github.com/redis/go-redis/v9"
)

func TestAnalyticsIntegration_RedisStream(t *testing.T) {
	// 1. Setup Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6380" // Default test port from Makefile
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available at %s, skipping integration test: %v", redisAddr, err)
	}
	defer rdb.Close()

	// 2. Setup Processor and Handler
	streamName := "test_analytics_stream"
	rdb.Del(ctx, streamName) // Ensure clean start
	defer rdb.Del(ctx, streamName)

	processor := service.NewRedisStreamProcessor(rdb, streamName)
	handler := NewAnalyticsHandler(processor)

	// 3. Prepare Test Data
	events := []protocol.EvaluationEvent{
		{
			ProjectID:     "proj-integration",
			EnvironmentID: "env-integration",
			FlagKey:       "flag-integration",
			UserID:        "user-integration",
			VariationKey:  "test-variation",
		},
	}
	body, _ := json.Marshal(events)

	// 4. Send Request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/bulk", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 5. Assert Response
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	// 6. Verify Redis Stream Content
	msgs, err := rdb.XRange(ctx, streamName, "-", "+").Result()
	if err != nil {
		t.Fatalf("Failed to read from Redis stream: %v", err)
	}

	if len(msgs) != 1 {
		t.Errorf("Expected 1 message in stream, got %d", len(msgs))
	}

	// Verify payload
	eventJSON := msgs[0].Values["event"].(string)
	var receivedEvent protocol.EvaluationEvent
	if err := json.Unmarshal([]byte(eventJSON), &receivedEvent); err != nil {
		t.Fatalf("Failed to unmarshal event from stream: %v", err)
	}

	if receivedEvent.FlagKey != events[0].FlagKey {
		t.Errorf("Expected flag key %s, got %s", events[0].FlagKey, receivedEvent.FlagKey)
	}
}
