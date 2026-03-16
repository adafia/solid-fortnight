package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adafia/solid-fortnight/internal/engine"
	"github.com/adafia/solid-fortnight/internal/storage/store"
)

// EvaluationRequest represents the request for flag evaluation.
type EvaluationRequest struct {
	ProjectID      string                 `json:"project_id"`
	EnvironmentKey string                 `json:"environment_key"`
	FlagKey        string                 `json:"flag_key"`
	Context        engine.UserContext     `json:"context"`
}

// EvaluationResponse represents the response for flag evaluation.
type EvaluationResponse struct {
	Value        json.RawMessage `json:"value"`
	VariationKey string          `json:"variation_key"`
	Reason       string          `json:"reason"`
}

// EvaluatorHandler handles evaluation requests.
type EvaluatorHandler struct {
	evaluator    *engine.Evaluator
	flagStore    *store.FlagStore
	projectStore *store.ProjectStore
	configStore  *store.FlagConfigStore
}

// NewEvaluatorHandler creates a new EvaluatorHandler.
func NewEvaluatorHandler(
	evaluator *engine.Evaluator,
	flagStore *store.FlagStore,
	projectStore *store.ProjectStore,
	configStore *store.FlagConfigStore,
) *EvaluatorHandler {
	return &EvaluatorHandler{
		evaluator:    evaluator,
		flagStore:    flagStore,
		projectStore: projectStore,
		configStore:  configStore,
	}
}

func (h *EvaluatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.GetFlags(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Get Flag
	flag, err := h.flagStore.GetFlagByKey(req.ProjectID, req.FlagKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get flag: %v", err), http.StatusInternalServerError)
		return
	}
	if flag == nil {
		http.Error(w, "Flag not found", http.StatusNotFound)
		return
	}

	// 2. Get Environment
	env, err := h.projectStore.GetEnvironmentByKey(req.ProjectID, req.EnvironmentKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get environment: %v", err), http.StatusInternalServerError)
		return
	}
	if env == nil {
		http.Error(w, "Environment not found", http.StatusNotFound)
		return
	}

	// 3. Get Flag Config
	config, err := h.configStore.GetFlagEnvironment(flag.ID, env.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get flag configuration: %v", err), http.StatusInternalServerError)
		return
	}
	if config == nil {
		http.Error(w, "Flag configuration not found for this environment", http.StatusNotFound)
		return
	}

	// 4. Map to Engine Config
	engineConfig := h.mapToEngineConfig(flag.Key, config)

	// 5. Evaluate
	result, err := h.evaluator.Evaluate(engineConfig, req.Context)
	if err != nil {
		http.Error(w, fmt.Sprintf("Evaluation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 6. Return Response
	resp := EvaluationResponse{
		Value:        result.Value,
		VariationKey: result.VariationKey,
		Reason:       result.Reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *EvaluatorHandler) GetFlags(w http.ResponseWriter, r *http.Request) {
	envID := r.URL.Query().Get("environment_id")
	if envID == "" {
		http.Error(w, "Missing environment_id parameter", http.StatusBadRequest)
		return
	}

	configs, err := h.configStore.GetFlagsForEnvironment(envID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get flag configurations: %v", err), http.StatusInternalServerError)
		return
	}

	// Map to Engine Configs
	engineConfigs := make([]engine.FlagConfig, len(configs))
	for i, fe := range configs {
		engineConfigs[i] = h.mapToEngineConfig(fe.FlagKey, &fe)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engineConfigs)
}

func (h *EvaluatorHandler) mapToEngineConfig(flagKey string, fe *store.FlagEnvironment) engine.FlagConfig {
	variations := make([]engine.Variation, len(fe.Variations))
	for i, v := range fe.Variations {
		variations[i] = engine.Variation{
			ID:    v.ID,
			Key:   v.Key,
			Value: v.Value,
		}
	}

	rules := make([]engine.Rule, len(fe.Rules))
	for i, r := range fe.Rules {
		var clauses []engine.Clause
		json.Unmarshal(r.Clauses, &clauses)
		rules[i] = engine.Rule{
			ID:          r.ID,
			VariationID: r.VariationID,
			Clauses:     clauses,
		}
	}

	return engine.FlagConfig{
		ID:                 fe.FlagID,
		Key:                flagKey,
		Enabled:            fe.Enabled,
		DefaultVariationID: fe.DefaultVariationID,
		RolloutVariationID: fe.RolloutVariationID,
		RolloutPercentage:  fe.RolloutPercentage,
		Variations:         variations,
		Rules:              rules,
	}
}
