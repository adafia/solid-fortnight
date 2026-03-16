package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/adafia/solid-fortnight/internal/storage/store"
)

type FlagsHandler struct {
	flagStore   *store.FlagStore
	configStore *store.FlagConfigStore
}

func NewFlagsHandler(flagStore *store.FlagStore, configStore *store.FlagConfigStore) *FlagsHandler {
	return &FlagsHandler{flagStore: flagStore, configStore: configStore}
}

func (h *FlagsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Simple path routing
	// /flags/
	// /flags/{id}
	// /flags/{id}/environments/{envId}
	// /flags/{id}/environments/{envId}/variations

	parts := SplitPath(r.URL.Path)
	
	switch r.Method {
	case http.MethodPost:
		if len(parts) >= 5 && parts[2] == "environments" && parts[4] == "variations" {
			h.AddVariation(w, r)
		} else {
			h.CreateFlag(w, r)
		}
	case http.MethodGet:
		if len(parts) >= 4 && parts[2] == "environments" {
			h.GetFlagEnvironment(w, r)
		} else {
			h.GetFlag(w, r)
		}
	case http.MethodPut:
		if len(parts) >= 4 && parts[2] == "environments" {
			h.UpsertFlagEnvironment(w, r)
		} else {
			h.UpdateFlag(w, r)
		}
	case http.MethodDelete:
		h.DeleteFlag(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *FlagsHandler) CreateFlag(w http.ResponseWriter, r *http.Request) {
	var flag store.Flag
	if err := json.NewDecoder(r.Body).Decode(&flag); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.flagStore.CreateFlag(&flag); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(flag)
}

func (h *FlagsHandler) GetFlag(w http.ResponseWriter, r *http.Request) {
	parts := SplitPath(r.URL.Path)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[1]
	flag, err := h.flagStore.GetFlag(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if flag == nil {
		http.Error(w, "Flag not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flag)
}

func (h *FlagsHandler) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	parts := SplitPath(r.URL.Path)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[1]
	var flag store.Flag
	if err := json.NewDecoder(r.Body).Decode(&flag); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	flag.ID = id

	if err := h.flagStore.UpdateFlag(&flag); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flag)
}

func (h *FlagsHandler) DeleteFlag(w http.ResponseWriter, r *http.Request) {
	parts := SplitPath(r.URL.Path)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[1]
	if err := h.flagStore.DeleteFlag(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FlagsHandler) GetFlagEnvironment(w http.ResponseWriter, r *http.Request) {
	// Path: /flags/{id}/environments/{envId}
	parts := SplitPath(r.URL.Path)
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	flagID := parts[1]
	envID := parts[3]

	fe, err := h.configStore.GetFlagEnvironment(flagID, envID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fe == nil {
		http.Error(w, "Flag environment config not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fe)
}

func (h *FlagsHandler) UpsertFlagEnvironment(w http.ResponseWriter, r *http.Request) {
	// Path: /flags/{id}/environments/{envId}
	parts := SplitPath(r.URL.Path)
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	flagID := parts[1]
	envID := parts[3]

	var fe store.FlagEnvironment
	if err := json.NewDecoder(r.Body).Decode(&fe); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fe.FlagID = flagID
	fe.EnvironmentID = envID

	if err := h.configStore.UpsertFlagEnvironment(&fe); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fe)
}

func (h *FlagsHandler) AddVariation(w http.ResponseWriter, r *http.Request) {
	// Path: /flags/{id}/environments/{envId}/variations
	parts := SplitPath(r.URL.Path)
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	// We need the flag_environment_id, but the path has flagID and envID.
	// We might need to look it up first or use a different endpoint.
	// For simplicity, let's assume the body contains the flag_environment_id or we fetch it.
	
	var v store.Variation
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.configStore.AddVariation(&v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}
