package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adafia/solid-fortnight/internal/storage/store"
	"github.com/google/uuid"
)

func TestCRUD_Environments(t *testing.T) {
	truncateTables()
	
	// Create project
	var project store.Project
	body, _ := json.Marshal(store.Project{Name: "Project Env Test"})
	req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	projectsHandler.ServeHTTP(rr, req)
	json.Unmarshal(rr.Body.Bytes(), &project)

	var createdEnv store.Environment

	// Create Environment
	t.Run("CreateEnvironment", func(t *testing.T) {
		env := store.Environment{
			Name:      "Production",
			Key:       "prod",
			SortOrder: 1,
		}
		body, _ := json.Marshal(env)
		path := fmt.Sprintf("/projects/%s/environments", project.ID)
		req, _ := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		environmentsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %d; got %d", http.StatusCreated, rr.Code)
		}

		json.Unmarshal(rr.Body.Bytes(), &createdEnv)
		if createdEnv.ID == "" {
			t.Error("expected environment to have an ID")
		}
		if createdEnv.Name != env.Name {
			t.Errorf("expected name %s; got %s", env.Name, createdEnv.Name)
		}
	})

	// Get Environments
	t.Run("GetEnvironments", func(t *testing.T) {
		path := fmt.Sprintf("/projects/%s/environments", project.ID)
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		environmentsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
		}

		var envs []store.Environment
		json.Unmarshal(rr.Body.Bytes(), &envs)
		if len(envs) == 0 {
			t.Error("expected at least one environment")
		}
	})
}

func TestFlagEnvironmentConfigs(t *testing.T) {
	truncateTables()
	
	// Setup: Project, Environment, Flag
	var project store.Project
	body, _ := json.Marshal(store.Project{Name: "Flag Config Test"})
	req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	projectsHandler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to create project: %d %s", rr.Code, rr.Body.String())
	}
	json.Unmarshal(rr.Body.Bytes(), &project)

	var env store.Environment
	body, _ = json.Marshal(store.Environment{Name: "Staging", Key: "staging"})
	path := fmt.Sprintf("/projects/%s/environments", project.ID)
	req, _ = http.NewRequest(http.MethodPost, path, bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	environmentsHandler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to create environment: %d %s", rr.Code, rr.Body.String())
	}
	json.Unmarshal(rr.Body.Bytes(), &env)

	var flag store.Flag
	body, _ = json.Marshal(store.Flag{
		ProjectID: project.ID,
		Key:       "test-flag",
		Name:      "Test Flag",
		Type:      "boolean",
		CreatedBy: uuid.NewString(),
	})
	req, _ = http.NewRequest(http.MethodPost, "/flags", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	flagsHandler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to create flag: %d %s", rr.Code, rr.Body.String())
	}
	json.Unmarshal(rr.Body.Bytes(), &flag)

	// Upsert Flag Environment
	t.Run("UpsertFlagEnvironment", func(t *testing.T) {
		fe := store.FlagEnvironment{
			Enabled: true,
		}
		body, _ := json.Marshal(fe)
		path := fmt.Sprintf("/flags/%s/environments/%s", flag.ID, env.ID)
		req, _ := http.NewRequest(http.MethodPut, path, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		flagsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d; got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var updatedFE store.FlagEnvironment
		json.Unmarshal(rr.Body.Bytes(), &updatedFE)
		if !updatedFE.Enabled {
			t.Error("expected enabled to be true")
		}
	})

	// Get Flag Environment
	t.Run("GetFlagEnvironment", func(t *testing.T) {
		path := fmt.Sprintf("/flags/%s/environments/%s", flag.ID, env.ID)
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		flagsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
		}

		var fetchedFE store.FlagEnvironment
		json.Unmarshal(rr.Body.Bytes(), &fetchedFE)
		if fetchedFE.FlagID != flag.ID {
			t.Errorf("expected flag ID %s; got %s", flag.ID, fetchedFE.FlagID)
		}
	})
}
