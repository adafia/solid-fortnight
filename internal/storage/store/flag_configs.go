package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

type FlagEnvironment struct {
	ID                 string      `json:"id"`
	FlagID             string      `json:"flag_id"`
	EnvironmentID      string      `json:"environment_id"`
	Enabled            bool        `json:"enabled"`
	DefaultVariationID *string     `json:"default_variation_id"`
	RolloutPercentage  *float64    `json:"rollout_percentage"`
	RolloutVariationID *string     `json:"rollout_variation_id"`
	Version            int         `json:"version"`
	UpdatedAt          time.Time   `json:"updated_at"`
	UpdatedBy          *string     `json:"updated_by"`
	Variations         []Variation `json:"variations,omitempty"`
	Rules              []FlagRule  `json:"rules,omitempty"`
}

type Variation struct {
	ID                string          `json:"id"`
	FlagEnvironmentID string          `json:"flag_environment_id"`
	Key               string          `json:"key"`
	Value             json.RawMessage `json:"value"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
}

type FlagRule struct {
	ID                string          `json:"id"`
	FlagEnvironmentID string          `json:"flag_environment_id"`
	VariationID       string          `json:"variation_id"`
	Clauses           json.RawMessage `json:"clauses"` // JSON array
	SortOrder         int             `json:"sort_order"`
	CreatedAt         time.Time       `json:"created_at"`
}

type FlagConfigStore struct {
	db *sql.DB
}

func NewFlagConfigStore(db *sql.DB) *FlagConfigStore {
	return &FlagConfigStore{db: db}
}

// GetFlagEnvironment retrieves the configuration for a specific flag and environment.
func (s *FlagConfigStore) GetFlagEnvironment(flagID, environmentID string) (*FlagEnvironment, error) {
	fe := &FlagEnvironment{}
	query := `
		SELECT id, flag_id, environment_id, enabled, default_variation_id, version, updated_at, updated_by, rollout_percentage, rollout_variation_id
		FROM flag_environments
		WHERE flag_id = $1 AND environment_id = $2`

	err := s.db.QueryRow(query, flagID, environmentID).Scan(
		&fe.ID,
		&fe.FlagID,
		&fe.EnvironmentID,
		&fe.Enabled,
		&fe.DefaultVariationID,
		&fe.Version,
		&fe.UpdatedAt,
		&fe.UpdatedBy,
		&fe.RolloutPercentage,
		&fe.RolloutVariationID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Fetch variations
	variations, err := s.GetVariations(fe.ID)
	if err != nil {
		return nil, err
	}
	fe.Variations = variations

	// Fetch rules
	rules, err := s.GetRules(fe.ID)
	if err != nil {
		return nil, err
	}
	fe.Rules = rules

	return fe, nil
}

// GetVariations retrieves all variations for a flag environment.
func (s *FlagConfigStore) GetVariations(flagEnvironmentID string) ([]Variation, error) {
	query := `
		SELECT id, flag_environment_id, key, value, name, description
		FROM flag_variations
		WHERE flag_environment_id = $1`

	rows, err := s.db.Query(query, flagEnvironmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variations []Variation
	for rows.Next() {
		var v Variation
		err := rows.Scan(
			&v.ID,
			&v.FlagEnvironmentID,
			&v.Key,
			&v.Value,
			&v.Name,
			&v.Description,
		)
		if err != nil {
			return nil, err
		}
		variations = append(variations, v)
	}
	return variations, nil
}

// GetRules retrieves all rules for a flag environment.
func (s *FlagConfigStore) GetRules(flagEnvironmentID string) ([]FlagRule, error) {
	query := `
		SELECT id, flag_environment_id, variation_id, clauses, sort_order, created_at
		FROM flag_rules
		WHERE flag_environment_id = $1
		ORDER BY sort_order ASC`

	rows, err := s.db.Query(query, flagEnvironmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []FlagRule
	for rows.Next() {
		var r FlagRule
		err := rows.Scan(
			&r.ID,
			&r.FlagEnvironmentID,
			&r.VariationID,
			&r.Clauses,
			&r.SortOrder,
			&r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

// UpsertFlagEnvironment creates or updates the flag environment configuration.
func (s *FlagConfigStore) UpsertFlagEnvironment(fe *FlagEnvironment) error {
	query := `
		INSERT INTO flag_environments (flag_id, environment_id, enabled, default_variation_id, updated_by, rollout_percentage, rollout_variation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (flag_id, environment_id) DO UPDATE
		SET enabled = EXCLUDED.enabled,
		    default_variation_id = EXCLUDED.default_variation_id,
		    updated_by = EXCLUDED.updated_by,
		    rollout_percentage = EXCLUDED.rollout_percentage,
		    rollout_variation_id = EXCLUDED.rollout_variation_id,
		    version = flag_environments.version + 1,
		    updated_at = CURRENT_TIMESTAMP
		RETURNING id, version, updated_at`

	err := s.db.QueryRow(
		query,
		fe.FlagID,
		fe.EnvironmentID,
		fe.Enabled,
		fe.DefaultVariationID,
		fe.UpdatedBy,
		fe.RolloutPercentage,
		fe.RolloutVariationID,
	).Scan(&fe.ID, &fe.Version, &fe.UpdatedAt)
	return err
}

// AddVariation adds a new variation to a flag environment.
func (s *FlagConfigStore) AddVariation(v *Variation) error {
	query := `
		INSERT INTO flag_variations (flag_environment_id, key, value, name, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := s.db.QueryRow(
		query,
		v.FlagEnvironmentID,
		v.Key,
		v.Value,
		v.Name,
		v.Description,
	).Scan(&v.ID)
	return err
}
