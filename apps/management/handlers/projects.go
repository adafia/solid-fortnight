package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/adafia/solid-fortnight/internal/storage/store"
)

type ProjectsHandler struct {
	projectStore *store.ProjectStore
}

func NewProjectsHandler(projectStore *store.ProjectStore) *ProjectsHandler {
	return &ProjectsHandler{projectStore: projectStore}
}

func (h *ProjectsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProjectsHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var project store.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.projectStore.CreateProject(&project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}
