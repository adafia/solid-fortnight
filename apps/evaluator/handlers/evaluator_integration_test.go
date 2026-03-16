package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/adafia/solid-fortnight/apps/evaluator/handlers"
	"github.com/adafia/solid-fortnight/internal/config"
	"github.com/adafia/solid-fortnight/internal/engine"
	"github.com/adafia/solid-fortnight/internal/storage/db"
	"github.com/adafia/solid-fortnight/internal/storage/store"
	"github.com/google/uuid"
)

var (
	testDB           *sql.DB
	flagStore        *store.FlagStore
	projectStore     *store.ProjectStore
	configStore      *store.FlagConfigStore
	evaluatorHandler http.Handler
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

	// Set up stores and handler
	flagStore = store.NewFlagStore(testDB)
	projectStore = store.NewProjectStore(testDB)
	configStore = store.NewFlagConfigStore(testDB)
	evaluatorEngine := engine.NewEvaluator()
	evaluatorHandler = handlers.NewEvaluatorHandler(evaluatorEngine, flagStore, projectStore, configStore)

	// Run tests
	exitCode := m.Run()

	// Clean up
	truncateTables()

	os.Exit(exitCode)
}

func truncateTables() {
	_, err := testDB.Exec("TRUNCATE TABLE flag_rules, flag_variations, flag_environments, flags, environments, projects RESTART IDENTITY CASCADE")
	if err != nil {
		panic("Failed to truncate tables: " + err.Error())
	}
}

func TestEvaluate_Integration(t *testing.T) {
	truncateTables()

	// Setup Data: Project, Environment, Flag, Variations, Config
	project := &store.Project{Name: "Eval Integration Project"}
	if err := projectStore.CreateProject(project); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	env := &store.Environment{ProjectID: project.ID, Name: "Production", Key: "prod"}
	if err := projectStore.CreateEnvironment(env); err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	userID := uuid.New().String()
	flag := &store.Flag{
		ProjectID: project.ID,
		Key:       "test-flag",
		Name:      "Test Flag",
		Type:      "boolean",
		CreatedBy: userID,
	}
	if err := flagStore.CreateFlag(flag); err != nil {
		t.Fatalf("Failed to create flag: %v", err)
	}

	fe := &store.FlagEnvironment{
		FlagID:        flag.ID,
		EnvironmentID: env.ID,
		Enabled:       true,
		UpdatedBy:     &userID,
	}
	if err := configStore.UpsertFlagEnvironment(fe); err != nil {
		t.Fatalf("Failed to upsert flag environment: %v", err)
	}

	v1 := &store.Variation{
		FlagEnvironmentID: fe.ID,
		Key:               "on",
		Value:             json.RawMessage(`true`),
	}
	if err := configStore.AddVariation(v1); err != nil {
		t.Fatalf("Failed to add variation 1: %v", err)
	}

	v2 := &store.Variation{
		FlagEnvironmentID: fe.ID,
		Key:               "off",
		Value:             json.RawMessage(`false`),
	}
	if err := configStore.AddVariation(v2); err != nil {
		t.Fatalf("Failed to add variation 2: %v", err)
	}

	// Update fe with default variation
	fe.DefaultVariationID = &v1.ID
	if err := configStore.UpsertFlagEnvironment(fe); err != nil {
		t.Fatalf("Failed to update flag environment with default: %v", err)
	}

	t.Run("BasicEvaluation", func(t *testing.T) {
		reqBody := handlers.EvaluationRequest{
			ProjectID:      project.ID,
			EnvironmentKey: env.Key,
			FlagKey:        flag.Key,
			Context: engine.UserContext{
				ID: "user-1",
			},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/evaluate", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		evaluatorHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200; got %d: %s", rr.Code, rr.Body.String())
		}

		var resp handlers.EvaluationResponse
		json.Unmarshal(rr.Body.Bytes(), &resp)

		if string(resp.Value) != "true" {
			t.Errorf("expected true; got %s", string(resp.Value))
		}
		if resp.VariationKey != "on" {
			t.Errorf("expected variation 'on'; got %s", resp.VariationKey)
		}
	})

	t.Run("RuleEvaluation", func(t *testing.T) {
		// Add a rule for beta users
		clauses := []engine.Clause{
			{
				Attribute: "beta",
				Operator:  engine.OperatorEquals,
				Values:    []string{"true"},
			},
		}
		clausesJSON, _ := json.Marshal(clauses)
		
		_, err := testDB.Exec(`
			INSERT INTO flag_rules (flag_environment_id, variation_id, clauses, sort_order)
			VALUES ($1, $2, $3, $4)`,
			fe.ID, v2.ID, clausesJSON, 0)
		if err != nil {
			t.Fatalf("Failed to add rule: %v", err)
		}

		// Evaluate for beta user
		reqBody := handlers.EvaluationRequest{
			ProjectID:      project.ID,
			EnvironmentKey: env.Key,
			FlagKey:        flag.Key,
			Context: engine.UserContext{
				ID: "user-beta",
				Attributes: map[string]interface{}{
					"beta": "true",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/evaluate", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		evaluatorHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200; got %d: %s", rr.Code, rr.Body.String())
		}

		var resp handlers.EvaluationResponse
		json.Unmarshal(rr.Body.Bytes(), &resp)

		if string(resp.Value) != "false" {
			t.Errorf("expected false for beta user; got %s", string(resp.Value))
		}
		if resp.VariationKey != "off" {
			t.Errorf("expected variation 'off'; got %s", resp.VariationKey)
		}
	})
}
