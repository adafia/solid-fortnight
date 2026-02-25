package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/adafia/solid-fortnight/internal/storage/store"
)

type FlagsHandler struct {
	flagStore *store.FlagStore
}

func NewFlagsHandler(flagStore *store.FlagStore) *FlagsHandler {
	return &FlagsHandler{flagStore: flagStore}
}

func (h *FlagsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateFlag(w, r)
	case http.MethodGet:
		h.GetFlag(w, r)
	case http.MethodPut:
		h.UpdateFlag(w, r)
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
	id := r.URL.Path[len("/flags/"):]
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
	id := r.URL.Path[len("/flags/"):]
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
	id := r.URL.Path[len("/flags/"):]
	if err := h.flagStore.DeleteFlag(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
