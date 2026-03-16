package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adafia/solid-fortnight/apps/evaluator/handlers"
	"github.com/adafia/solid-fortnight/internal/config"
	"github.com/adafia/solid-fortnight/internal/engine"
	"github.com/adafia/solid-fortnight/internal/storage/db"
	"github.com/adafia/solid-fortnight/internal/storage/store"
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

	// Connect to the database
	dsn := db.DSN(
		cfg.Storage.Postgres.Host,
		cfg.Storage.Postgres.User,
		cfg.Storage.Postgres.Password,
		cfg.Storage.Postgres.DBName,
		cfg.Storage.Postgres.Port,
	)
	database, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Run migrations (Evaluator service might not need it, but for now we'll match)
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "internal/storage/migrations"
	}
	if err := db.Migrate(database, migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations from %s: %v", migrationsPath, err)
	}

	// Set up stores
	flagStore := store.NewFlagStore(database)
	projectStore := store.NewProjectStore(database)
	configStore := store.NewFlagConfigStore(database)

	// Set up evaluator engine
	evaluatorEngine := engine.NewEvaluator()

	// Set up handler
	evalHandler := handlers.NewEvaluatorHandler(evaluatorEngine, flagStore, projectStore, configStore)

	// Set up router
	mux := http.NewServeMux()
	mux.Handle("/api/v1/evaluate", evalHandler)

	// Start server
	port := cfg.Services["evaluator"].Port
	if port == 0 {
		port = 8081 // Default for evaluator if not specified
	}
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Evaluator service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
