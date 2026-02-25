package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adafia/solid-fortnight/apps/management/handlers"
	"github.com/adafia/solid-fortnight/internal/config"
	"github.com/adafia/solid-fortnight/internal/storage/db"
	"github.com/adafia/solid-fortnight/internal/storage/store"
)

func main() {
	// Load configuration
	cfg, err := config.Load("../../deployments/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
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

	// Run migrations
	if err := db.Migrate(database, "../../internal/storage/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Set up stores
	flagStore := store.NewFlagStore(database)

	// Set up handlers
	flagsHandler := handlers.NewFlagsHandler(flagStore)

	// Set up router
	mux := http.NewServeMux()
	mux.Handle("/flags/", flagsHandler)

	// Start server
	port := cfg.Services["management"].Port
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Management service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
