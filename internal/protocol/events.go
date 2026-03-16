package protocol

import "encoding/json"

type EvaluationEvent struct {
	ProjectID      string          `json:"project_id"`
	EnvironmentID  string          `json:"environment_id"`
	FlagKey        string          `json:"flag_key"`
	UserID         string          `json:"user_id"`
	VariationKey   string          `json:"variation_key"`
	Value          json.RawMessage `json:"value"`
	Reason         string          `json:"reason"`
	Context        json.RawMessage `json:"context"`
	EvaluatedAt    int64           `json:"evaluated_at"` // Unix timestamp
}
