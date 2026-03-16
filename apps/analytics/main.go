package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adafia/solid-fortnight/apps/analytics/handlers"
	"github.com/adafia/solid-fortnight/apps/analytics/service"
	"github.com/adafia/solid-fortnight/internal/config"
	"github.com/adafia/solid-fortnight/internal/storage/db"
	"github.com/adafia/solid-fortnight/internal/storage/store"
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

	// Set up PostgreSQL
	dsn := db.DSN(
		cfg.Storage.Postgres.Host,
		cfg.Storage.Postgres.User,
		cfg.Storage.Postgres.Password,
		cfg.Storage.Postgres.DBName,
		cfg.Storage.Postgres.Port,
	)
	database, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer database.Close()

	// Set up processor
	streamName := "evaluation_events_stream"
	processor := service.NewRedisStreamProcessor(rdb, streamName)

	// Set up worker
	eventStore := store.NewEvaluationEventStore(database)
	worker := service.NewEvaluationEventWorker(rdb, eventStore, streamName, "analytics_consumer_group", "consumer-1")

	// Start worker in a separate goroutine
	go func() {
		if err := worker.Start(context.Background()); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()

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
