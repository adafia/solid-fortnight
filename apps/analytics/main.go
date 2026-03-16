package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adafia/solid-fortnight/apps/analytics/handlers"
	"github.com/adafia/solid-fortnight/apps/analytics/service"
	"github.com/adafia/solid-fortnight/internal/config"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "deployments/config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
	}

	// Set up Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Storage.Redis.Addr,
		Password: cfg.Storage.Redis.Password,
		DB:       cfg.Storage.Redis.DB,
	})

	// Set up processor
	streamName := "evaluation_events_stream"
	processor := service.NewRedisStreamProcessor(rdb, streamName)

	// Set up handlers
	analyticsHandler := handlers.NewAnalyticsHandler(processor)

	// Set up router
	mux := http.NewServeMux()
	mux.Handle("/api/v1/events/", analyticsHandler)

	// Start server
	port := cfg.Services["analytics"].Port
	if port == 0 {
		port = 8081
	}
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Analytics service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
