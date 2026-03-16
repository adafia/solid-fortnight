package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/adafia/solid-fortnight/internal/protocol"
)

// EvaluationEventStore handles database operations for evaluation events.
type EvaluationEventStore struct {
	db *sql.DB
}

// NewEvaluationEventStore creates a new EvaluationEventStore.
func NewEvaluationEventStore(db *sql.DB) *EvaluationEventStore {
	return &EvaluationEventStore{db: db}
}

// SaveEvaluationEvents performs a batch insert of evaluation events into PostgreSQL.
func (s *EvaluationEventStore) SaveEvaluationEvents(ctx context.Context, events []protocol.EvaluationEvent) error {
	if len(events) == 0 {
		return nil
	}

	query := `
		INSERT INTO evaluation_events (
			project_id, environment_id, flag_key, user_id, 
			variation_key, value, reason, context, evaluated_at
		) VALUES 
	`
	vals := []interface{}{}

	for i, e := range events {
		p := i * 9
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d),", p+1, p+2, p+3, p+4, p+5, p+6, p+7, p+8, p+9)
		
		evaluatedAt := time.Unix(e.EvaluatedAt, 0)
		vals = append(vals, 
			e.ProjectID, e.EnvironmentID, e.FlagKey, e.UserID, 
			e.VariationKey, e.Value, e.Reason, e.Context, evaluatedAt,
		)
	}

	// Trim the last comma and add it back or use a different joining strategy
	query = query[:len(query)-1]

	_, err := s.db.ExecContext(ctx, query, vals...)
	if err != nil {
		return fmt.Errorf("failed to batch insert evaluation events: %w", err)
	}

	return nil
}
