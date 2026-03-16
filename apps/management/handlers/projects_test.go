package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adafia/solid-fortnight/internal/storage/store"
)

func TestCRUD_Projects(t *testing.T) {
	truncateTables()

	var createdProject store.Project

	// Create
	t.Run("CreateProject", func(t *testing.T) {
		project := store.Project{
			Name: "Test Project CRUD",
		}
		body, _ := json.Marshal(project)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		projectsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %d; got %d", http.StatusCreated, rr.Code)
		}

		json.Unmarshal(rr.Body.Bytes(), &createdProject)
		if createdProject.ID == "" {
			t.Error("expected project to have an ID")
		}
		if createdProject.Name != project.Name {
			t.Errorf("expected name %s; got %s", project.Name, createdProject.Name)
		}
	})

	// Get
	t.Run("GetProject", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s", createdProject.ID), nil)
		rr := httptest.NewRecorder()

		projectsHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
		}

		var fetchedProject store.Project
		json.Unmarshal(rr.Body.Bytes(), &fetchedProject)
		if fetchedProject.ID != createdProject.ID {
			t.Errorf("expected project ID %s; got %s", createdProject.ID, fetchedProject.ID)
		}
	})
}
