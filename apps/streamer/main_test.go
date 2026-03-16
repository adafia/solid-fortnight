package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestStreamerIntegration(t *testing.T) {
	// 1. Setup Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6380" // Default test port
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available at %s, skipping integration test: %v", redisAddr, err)
	}
	defer rdb.Close()

	// 2. Setup Hub and Subscriber
	hub := NewHub()
	
	stopSub := make(chan struct{})
	go func() {
		pubsub := rdb.Subscribe(ctx, "environment_updates")
		defer pubsub.Close()
		
		// Wait for subscription to be active
		_, err := pubsub.Receive(ctx)
		if err != nil {
			return
		}

		ch := pubsub.Channel()
		for {
			select {
			case msg := <-ch:
				var payload struct {
					EnvironmentID string `json:"environment_id"`
				}
				if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
					continue
				}
				hub.Broadcast(payload.EnvironmentID, "update")
			case <-stopSub:
				return
			}
		}
	}()
	defer close(stopSub)

	// 3. Setup Handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		envID := r.URL.Query().Get("environment_id")
		if envID == "" {
			http.Error(w, "Missing environment_id parameter", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		
		f, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}
		f.Flush() // Flush headers to signal connection established

		clientChan := make(chan string, 10)
		hub.Register(envID, clientChan)
		defer hub.Unregister(envID, clientChan)

		for {
			select {
			case msg := <-clientChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				f.Flush()
			case <-r.Context().Done():
				return
			case <-time.After(5 * time.Second):
				return
			}
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// 4. Connect to SSE
	envID := "test-env-123"
	
	// Start publish in background
	go func() {
		// Wait for client registration to settle
		time.Sleep(500 * time.Millisecond)
		payload := map[string]string{"environment_id": envID}
		data, _ := json.Marshal(payload)
		rdb.Publish(ctx, "environment_updates", data)
	}()

	resp, err := http.Get(server.URL + "/stream?environment_id=" + envID)
	if err != nil {
		t.Fatalf("Failed to connect to stream: %v", err)
	}
	defer resp.Body.Close()

	// 6. Assert SSE output
	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read from stream: %v", err)
	}

	if !strings.Contains(line, "data: update") {
		t.Errorf("Expected 'data: update', got %q", line)
	}
}
