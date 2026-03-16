package engine

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// UserContext represents the context of a user for flag evaluation.
type UserContext struct {
	ID         string                 `json:"id"`
	Attributes map[string]interface{} `json:"attributes"`
}

// Variation represents a flag variation.
type Variation struct {
	ID    string          `json:"id"`
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// FlagConfig represents the configuration of a flag in a specific environment.
type FlagConfig struct {
	ID                 string      `json:"id"`
	Key                string      `json:"key"`
	Enabled            bool        `json:"enabled"`
	DefaultVariationID *string     `json:"default_variation_id"`
	RolloutVariationID *string     `json:"rollout_variation_id"`
	Variations         []Variation `json:"variations"`
	RolloutPercentage  *float64    `json:"rollout_percentage"` // 0.0 to 100.0
	Rules              []Rule      `json:"rules"`
}

// Rule represents a targeting rule.
type Rule struct {
	ID          string   `json:"id"`
	VariationID string   `json:"variation_id"`
	Clauses     []Clause `json:"clauses"`
}

// Clause represents a single condition within a rule.
type Clause struct {
	Attribute string   `json:"attribute"`
	Operator  Operator `json:"operator"`
	Values    []string `json:"values"`
}

// Operator defines the type of comparison for a clause.
type Operator string

const (
	OperatorEquals      Operator = "EQUALS"
	OperatorNotEquals   Operator = "NOT_EQUALS"
	OperatorIn          Operator = "IN"
	OperatorNotIn       Operator = "NOT_IN"
	OperatorContains    Operator = "CONTAINS"
	OperatorNotContains Operator = "NOT_CONTAINS"
	OperatorStartsWith  Operator = "STARTS_WITH"
	OperatorEndsWith    Operator = "ENDS_WITH"
)

// EvaluationResult represents the result of a flag evaluation.
type EvaluationResult struct {
	Value          json.RawMessage `json:"value"`
	VariationKey   string          `json:"variation_key"`
	VariationID    string          `json:"variation_id"`
	Reason         string          `json:"reason"`
}

// Evaluator is responsible for evaluating flags.
type Evaluator struct{}

// NewEvaluator creates a new Evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate evaluates a flag for a given user context.
func (e *Evaluator) Evaluate(config FlagConfig, context UserContext) (*EvaluationResult, error) {
	if !config.Enabled {
		return e.defaultResult(config, "flag disabled"), nil
	}

	// 1. Targeting Rules
	for _, rule := range config.Rules {
		if e.matchesRule(rule, context) {
			return e.getVariation(config, rule.VariationID, fmt.Sprintf("rule match: %s", rule.ID))
		}
	}

	// 2. Percentage Rollout (Consistent Hashing)
	if config.RolloutPercentage != nil && *config.RolloutPercentage < 100.0 {
		if e.isUserInRollout(config.ID, context.ID, *config.RolloutPercentage) {
			if config.RolloutVariationID != nil {
				return e.getVariation(config, *config.RolloutVariationID, "user in rollout")
			}
		} else {
			return e.defaultResult(config, "user not in rollout"), nil
		}
	}

	// 3. Default Variation
	return e.defaultResult(config, "default variation"), nil
}

func (e *Evaluator) matchesRule(rule Rule, context UserContext) bool {
	if len(rule.Clauses) == 0 {
		return false
	}
	for _, clause := range rule.Clauses {
		if !e.matchesClause(clause, context) {
			return false
		}
	}
	return true
}

func (e *Evaluator) matchesClause(clause Clause, context UserContext) bool {
	attrValue, ok := context.Attributes[clause.Attribute]
	if !ok {
		// Special case: "id" attribute maps to context.ID if not in attributes
		if clause.Attribute == "id" {
			attrValue = context.ID
		} else {
			return false
		}
	}

	strVal := fmt.Sprintf("%v", attrValue)

	switch clause.Operator {
	case OperatorEquals, OperatorIn:
		for _, v := range clause.Values {
			if strVal == v {
				return true
			}
		}
	case OperatorNotEquals, OperatorNotIn:
		for _, v := range clause.Values {
			if strVal == v {
				return false
			}
		}
		return true
	case OperatorContains:
		for _, v := range clause.Values {
			if strings.Contains(strVal, v) {
				return true
			}
		}
	case OperatorNotContains:
		for _, v := range clause.Values {
			if strings.Contains(strVal, v) {
				return false
			}
		}
		return true
	case OperatorStartsWith:
		for _, v := range clause.Values {
			if strings.HasPrefix(strVal, v) {
				return true
			}
		}
	case OperatorEndsWith:
		for _, v := range clause.Values {
			if strings.HasSuffix(strVal, v) {
				return true
			}
		}
	}

	return false
}

func (e *Evaluator) getVariation(config FlagConfig, variationID, reason string) (*EvaluationResult, error) {
	for _, v := range config.Variations {
		if v.ID == variationID {
			return &EvaluationResult{
				Value:        v.Value,
				VariationKey: v.Key,
				VariationID:  v.ID,
				Reason:       reason,
			}, nil
		}
	}
	return nil, fmt.Errorf("variation %s not found for flag %s", variationID, config.Key)
}

func (e *Evaluator) defaultResult(config FlagConfig, reason string) *EvaluationResult {
	// If disabled or not in rollout, we might still want to return a specific "off" variation
	// For now, we'll return the default variation or the first one if it's disabled.
	if config.DefaultVariationID != nil {
		for _, v := range config.Variations {
			if v.ID == *config.DefaultVariationID {
				return &EvaluationResult{
					Value:        v.Value,
					VariationKey: v.Key,
					VariationID:  v.ID,
					Reason:       reason,
				}
			}
		}
	}
	return &EvaluationResult{
		Reason: reason,
	}
}

// isUserInRollout determines if a user is included in a percentage rollout.
// It uses MD5 hashing for consistency across different SDKs/languages.
func (e *Evaluator) isUserInRollout(flagID, userID string, percentage float64) bool {
	if userID == "" {
		// If no user ID is provided, we can't do consistent hashing.
		// For safety, we'll exclude them from rollouts or use a random one (less ideal).
		userID = uuid.New().String()
	}

	hashKey := fmt.Sprintf("%s:%s", flagID, userID)
	hash := md5.Sum([]byte(hashKey))
	
	// Use the first 8 bytes of the hash to get a uint64
	hashUint := binary.BigEndian.Uint64(hash[:8])
	
	// Map the hash to a value between 0 and 100
	userValue := float64(hashUint%10000) / 100.0
	
	return userValue <= percentage
}
