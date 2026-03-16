package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/adafia/solid-fortnight/internal/storage/store"
)

type EnvironmentsHandler struct {
	projectStore *store.ProjectStore
}

func NewEnvironmentsHandler(projectStore *store.ProjectStore) *EnvironmentsHandler {
	return &EnvironmentsHandler{projectStore: projectStore}
}

func (h *EnvironmentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := SplitPath(r.URL.Path)
	// Expecting path format: /projects/{id}/environments
	if len(parts) < 3 || parts[2] != "environments" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.CreateEnvironment(w, r)
	case http.MethodGet:
		h.GetEnvironments(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *EnvironmentsHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	// Path format: /projects/{id}/environments
	parts := SplitPath(r.URL.Path)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	projectID := parts[1]
	
	var env store.Environment
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	env.ProjectID = projectID

	if err := h.projectStore.CreateEnvironment(&env); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(env)
}

func (h *EnvironmentsHandler) GetEnvironments(w http.ResponseWriter, r *http.Request) {
	// Path format: /projects/{id}/environments
	parts := SplitPath(r.URL.Path)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	projectID := parts[1]
	
	environments, err := h.projectStore.GetEnvironments(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(environments)
}
