package engine

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestEvaluate_Basic(t *testing.T) {
	evaluator := NewEvaluator()

	flagID := uuid.New().String()
	variationID := uuid.New().String()
	variationValue := json.RawMessage(`{"color": "blue"}`)

	config := FlagConfig{
		ID:      flagID,
		Key:     "test-flag",
		Enabled: true,
		DefaultVariationID: &variationID,
		Variations: []Variation{
			{
				ID:    variationID,
				Key:   "variation-1",
				Value: variationValue,
			},
		},
	}

	context := UserContext{
		ID: "user-1",
	}

	result, err := evaluator.Evaluate(config, context)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if result.VariationID != variationID {
		t.Errorf("expected variation %s, got %s", variationID, result.VariationID)
	}

	if string(result.Value) != string(variationValue) {
		t.Errorf("expected value %s, got %s", string(variationValue), string(result.Value))
	}
}

func TestEvaluate_Disabled(t *testing.T) {
	evaluator := NewEvaluator()

	flagID := uuid.New().String()
	variationID := uuid.New().String()

	config := FlagConfig{
		ID:      flagID,
		Key:     "test-flag",
		Enabled: false,
		DefaultVariationID: &variationID,
		Variations: []Variation{
			{
				ID:    variationID,
				Key:   "variation-1",
				Value: json.RawMessage(`true`),
			},
		},
	}

	context := UserContext{
		ID: "user-1",
	}

	result, err := evaluator.Evaluate(config, context)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if result.Reason != "flag disabled" {
		t.Errorf("expected reason 'flag disabled', got %s", result.Reason)
	}
}

func TestEvaluate_PercentageRollout(t *testing.T) {
	evaluator := NewEvaluator()

	flagID := "flag-1"
	varID_A := "v-a"
	varID_B := "v-b"
	rollout := 50.0

	config := FlagConfig{
		ID:      flagID,
		Key:     "rollout-flag",
		Enabled: true,
		DefaultVariationID: &varID_A,
		RolloutVariationID: &varID_B,
		Variations: []Variation{
			{ID: varID_A, Key: "a", Value: json.RawMessage(`"A"`)},
			{ID: varID_B, Key: "b", Value: json.RawMessage(`"B"`)},
		},
		RolloutPercentage: &rollout,
	}

	// Test a sample of users to check consistency and approximate distribution
	countA := 0
	countB := 0
	total := 1000

	for i := 0; i < total; i++ {
		userID := fmt.Sprintf("user-%d", i)
		result, _ := evaluator.Evaluate(config, UserContext{ID: userID})
		
		if result.VariationID == varID_A {
			countA++
		} else {
			countB++
		}
	}

	if countA < 400 || countA > 600 {
		t.Errorf("Expected roughly 50%% distribution, got A: %d, B: %d", countA, countB)
	}
}

func TestEvaluate_Rules(t *testing.T) {
	evaluator := NewEvaluator()

	flagID := "flag-1"
	varID_Default := "v-default"
	varID_Premium := "v-premium"

	config := FlagConfig{
		ID:      flagID,
		Key:     "premium-flag",
		Enabled: true,
		DefaultVariationID: &varID_Default,
		Variations: []Variation{
			{ID: varID_Default, Key: "default", Value: json.RawMessage(`"regular"`)},
			{ID: varID_Premium, Key: "premium", Value: json.RawMessage(`"premium"`)},
		},
		Rules: []Rule{
			{
				ID:          "premium-users-rule",
				VariationID: varID_Premium,
				Clauses: []Clause{
					{
						Attribute: "plan",
						Operator:  OperatorEquals,
						Values:    []string{"premium"},
					},
				},
			},
		},
	}

	// Test Premium User
	premiumResult, _ := evaluator.Evaluate(config, UserContext{
		ID: "user-premium",
		Attributes: map[string]interface{}{
			"plan": "premium",
		},
	})
	if premiumResult.VariationID != varID_Premium {
		t.Errorf("expected premium variation, got %s", premiumResult.VariationID)
	}

	// Test Regular User
	regularResult, _ := evaluator.Evaluate(config, UserContext{
		ID: "user-regular",
		Attributes: map[string]interface{}{
			"plan": "free",
		},
	})
	if regularResult.VariationID != varID_Default {
		t.Errorf("expected default variation, got %s", regularResult.VariationID)
	}
}
