package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/adafia/solid-fortnight/apps/management/handlers"
	"github.com/adafia/solid-fortnight/internal/config"
	"github.com/adafia/solid-fortnight/internal/storage/db"
	"github.com/adafia/solid-fortnight/internal/storage/store"
	"github.com/google/uuid"
)

var (
	testDB          *sql.DB
	flagStore       *store.FlagStore
	projectStore    *store.ProjectStore
	configStore     *store.FlagConfigStore
	flagsHandler       http.Handler
	projectsHandler    http.Handler
	environmentsHandler http.Handler
	projectID          string
)

func TestMain(m *testing.M) {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "../../../deployments/config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		panic("Failed to load configuration from " + configPath + ": " + err.Error())
	}

	// Connect to the database
	dsn := db.DSN(
		cfg.Storage.Postgres.Host,
		cfg.Storage.Postgres.User,
		cfg.Storage.Postgres.Password,
		cfg.Storage.Postgres.DBName,
		cfg.Storage.Postgres.Port,
	)
	testDB, err = db.NewDB(dsn)
	if err != nil {
		panic("Failed to connect to the database: " + err.Error())
	}

	// Run migrations
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../../../internal/storage/migrations"
	}
	if err := db.Migrate(testDB, migrationsPath); err != nil {
		panic("Failed to run migrations from " + migrationsPath + ": " + err.Error())
	}

	// Set up stores and handlers
	flagStore = store.NewFlagStore(testDB)
	projectStore = store.NewProjectStore(testDB)
	configStore = store.NewFlagConfigStore(testDB)
	flagsHandler = handlers.NewFlagsHandler(flagStore, configStore, nil)
	projectsHandler = handlers.NewProjectsHandler(projectStore)
	environmentsHandler = handlers.NewEnvironmentsHandler(projectStore)

	// Create a project for the tests
	truncateTables()
	_, err = testDB.Exec("INSERT INTO projects (name) VALUES ('test-project')")
	if err != nil {
		panic(err)
	}
	err = testDB.QueryRow("SELECT id FROM projects WHERE name = 'test-project'").Scan(&projectID)
	if err != nil {
		panic(err)
	}

	// Run tests
	exitCode := m.Run()

	// Clean up
	truncateTables()

	os.Exit(exitCode)
}

func truncateTables() {
	_, err := testDB.Exec("TRUNCATE TABLE flags, environments, projects RESTART IDENTITY CASCADE")
	if err != nil {
		panic("Failed to truncate tables: " + err.Error())
	}
}

func TestCRUD_Flags(t *testing.T) {
	truncateTables()
	// Need to recreate the project for each test run
	_, err := testDB.Exec("INSERT INTO projects (name) VALUES ('test-project')")
	if err != nil {
		panic(err)
	}
	err = testDB.QueryRow("SELECT id FROM projects WHERE name = 'test-project'").Scan(&projectID)
	if err != nil {
		panic(err)
	}

	var createdFlag store.Flag

	// Create
	t.Run("CreateFlag", func(t *testing.T) {
		flag := &store.Flag{
			ProjectID:   projectID,
			Key:         "new-flag",
			Name:        "New Flag",
			Description: "A new feature flag",
			Type:        "boolean",
			CreatedBy:   uuid.NewString(),
		}
		body, _ := json.Marshal(flag)

		req, _ := http.NewRequest(http.MethodPost, "/flags", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		flagsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %d; got %d", http.StatusCreated, rr.Code)
		}

		json.Unmarshal(rr.Body.Bytes(), &createdFlag)

		if createdFlag.ID == "" {
			t.Error("expected flag to have an ID")
		}
		if createdFlag.Name != flag.Name {
			t.Errorf("expected name %s; got %s", flag.Name, createdFlag.Name)
		}
	})

	// Get
	t.Run("GetFlag", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/flags/%s", createdFlag.ID), nil)
		rr := httptest.NewRecorder()

		flagsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
		}

		var fetchedFlag store.Flag
		json.Unmarshal(rr.Body.Bytes(), &fetchedFlag)

		if fetchedFlag.ID != createdFlag.ID {
			t.Errorf("expected flag ID %s; got %s", createdFlag.ID, fetchedFlag.ID)
		}
	})

	// Update
	t.Run("UpdateFlag", func(t *testing.T) {
		updatedFlag := createdFlag
		updatedFlag.Name = "Updated Flag Name"
		body, _ := json.Marshal(updatedFlag)

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/flags/%s", updatedFlag.ID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		flagsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
		}

		var returnedFlag store.Flag
		json.Unmarshal(rr.Body.Bytes(), &returnedFlag)

		if returnedFlag.Name != updatedFlag.Name {
			t.Errorf("expected updated name %s; got %s", updatedFlag.Name, returnedFlag.Name)
		}
	})

	// Delete
	t.Run("DeleteFlag", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/flags/%s", createdFlag.ID), nil)
		rr := httptest.NewRecorder()

		flagsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status %d; got %d", http.StatusNoContent, rr.Code)
		}

		// Verify it's gone
		req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/flags/%s", createdFlag.ID), nil)
		rr = httptest.NewRecorder()
		flagsHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status %d; got %d", http.StatusNotFound, rr.Code)
		}
	})
}
