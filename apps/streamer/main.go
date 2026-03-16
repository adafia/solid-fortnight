package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/adafia/solid-fortnight/internal/config"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	// Map of environment_id -> map of client_channel -> bool
	clients map[string]map[chan string]bool
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[chan string]bool),
	}
}

func (h *Hub) Register(envID string, client chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[envID]; !ok {
		h.clients[envID] = make(map[chan string]bool)
	}
	h.clients[envID][client] = true
	log.Printf("Client registered for environment: %s. Total clients for env: %d", envID, len(h.clients[envID]))
}

func (h *Hub) Unregister(envID string, client chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.clients[envID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, envID)
		}
	}
	log.Printf("Client unregistered from environment: %s", envID)
}

func (h *Hub) Broadcast(envID string, message string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[envID]; ok {
		log.Printf("Broadcasting to %d clients for environment: %s", len(clients), envID)
		for client := range clients {
			select {
			case client <- message:
			default:
				// Skip client if channel is full or slow
			}
		}
	}
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "../../deployments/config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Storage.Redis.Addr,
		Password: cfg.Storage.Redis.Password,
		DB:       cfg.Storage.Redis.DB,
	})

	hub := NewHub()

	// Subscribe to Redis for environment updates
	go func() {
		ctx := context.Background()
		pubsub := rdb.Subscribe(ctx, "environment_updates")
		defer pubsub.Close()

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("Redis subscribe error: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			var payload struct {
				EnvironmentID string `json:"environment_id"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				log.Printf("Failed to unmarshal Redis message: %v", err)
				continue
			}

			log.Printf("Received update for environment: %s", payload.EnvironmentID)
			hub.Broadcast(payload.EnvironmentID, "update")
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		envID := r.URL.Query().Get("environment_id")
		if envID == "" {
			http.Error(w, "Missing environment_id parameter", http.StatusBadRequest)
			return
		}

		// SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		clientChan := make(chan string, 10)
		hub.Register(envID, clientChan)

		// Create a separate goroutine for heartbeat to keep connection alive
		heartbeatStop := make(chan struct{})
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					fmt.Fprintf(w, ": keep-alive\n\n")
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
				case <-heartbeatStop:
					return
				}
			}
		}()

		// Clean up on disconnect
		defer func() {
			close(heartbeatStop)
			hub.Unregister(envID, clientChan)
			close(clientChan)
		}()

		// Wait for messages to broadcast
		for {
			select {
			case msg := <-clientChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			case <-r.Context().Done():
				return
			}
		}
	})

	port := cfg.Services["streamer"].Port
	if port == 0 {
		port = 8084
	}
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Streamer service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
