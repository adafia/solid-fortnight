package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adafia/solid-fortnight/apps/management/handlers"
	"github.com/adafia/solid-fortnight/internal/config"
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

	// Run migrations
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

	// Set up handlers
	flagsHandler := handlers.NewFlagsHandler(flagStore, configStore)
	projectsHandler := handlers.NewProjectsHandler(projectStore)
	environmentsHandler := handlers.NewEnvironmentsHandler(projectStore)

	// Set up router
	mux := http.NewServeMux()
	mux.Handle("/flags/", flagsHandler)
	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		parts := handlers.SplitPath(r.URL.Path)
		if len(parts) >= 3 && parts[2] == "environments" {
			environmentsHandler.ServeHTTP(w, r)
			return
		}
		projectsHandler.ServeHTTP(w, r)
	})

	// Start server
	port := cfg.Services["management"].Port
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Management service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
