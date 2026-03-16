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
	case http.MethodGet:
		h.GetProject(w, r)
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

func (h *ProjectsHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	parts := SplitPath(r.URL.Path)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[1]
	project, err := h.projectStore.GetProject(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}
