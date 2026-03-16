package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/adafia/solid-fortnight/apps/analytics/service"
	"github.com/adafia/solid-fortnight/internal/protocol"
	"github.com/adafia/solid-fortnight/internal/storage/db"
	"github.com/adafia/solid-fortnight/internal/storage/store"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestAnalyticsIntegration_RedisStream(t *testing.T) {
	// 1. Setup Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6380" // Default test port from Makefile
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available at %s, skipping integration test: %v", redisAddr, err)
	}
	defer rdb.Close()

	// 2. Setup Processor and Handler
	streamName := "test_analytics_stream"
	rdb.Del(ctx, streamName) // Ensure clean start
	defer rdb.Del(ctx, streamName)

	processor := service.NewRedisStreamProcessor(rdb, streamName)
	handler := NewAnalyticsHandler(processor)

	// 3. Prepare Test Data
	events := []protocol.EvaluationEvent{
		{
			ProjectID:     "proj-integration",
			EnvironmentID: "env-integration",
			FlagKey:       "flag-integration",
			UserID:        "user-integration",
			VariationKey:  "test-variation",
		},
	}
	body, _ := json.Marshal(events)

	// 4. Send Request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/bulk", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 5. Assert Response
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	// 6. Verify Redis Stream Content
	msgs, err := rdb.XRange(ctx, streamName, "-", "+").Result()
	if err != nil {
		t.Fatalf("Failed to read from Redis stream: %v", err)
	}

	if len(msgs) != 1 {
		t.Errorf("Expected 1 message in stream, got %d", len(msgs))
	}

	// Verify payload
	eventJSON := msgs[0].Values["event"].(string)
	var receivedEvent protocol.EvaluationEvent
	if err := json.Unmarshal([]byte(eventJSON), &receivedEvent); err != nil {
		t.Fatalf("Failed to unmarshal event from stream: %v", err)
	}

	if receivedEvent.FlagKey != events[0].FlagKey {
		t.Errorf("Expected flag key %s, got %s", events[0].FlagKey, receivedEvent.FlagKey)
	}
}

func TestAnalyticsWorkerIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Setup Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6380"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}
	defer rdb.Close()

	// 2. Setup PostgreSQL
	pgHost := os.Getenv("POSTGRES_HOST")
	if pgHost == "" { pgHost = "localhost" }
	pgPort := 5433 // Default test port
	pgUser := "testuser"
	pgPass := "testpassword"
	pgDBName := "solid_fortnight_test"

	dsn := db.DSN(pgHost, pgUser, pgPass, pgDBName, pgPort)
	database, err := db.NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available")
	}
	defer database.Close()

	// 3. Ensure tables exist (quick check/create for test)
	setupTestDB(t, database)

	// 4. Setup Worker
	streamName := "worker_test_stream_" + uuid.New().String()
	groupName := "worker_test_group"
	consumerID := "worker_test_consumer"
	
	eventStore := store.NewEvaluationEventStore(database)
	worker := service.NewEvaluationEventWorker(rdb, eventStore, streamName, groupName, consumerID)

	// Start worker in background
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	go func() {
		_ = worker.Start(workerCtx)
	}()

	// 5. Produce an event via Processor
	processor := service.NewRedisStreamProcessor(rdb, streamName)
	projectID := uuid.New().String()
	envID := uuid.New().String()
	
	// Create project and environment to satisfy foreign keys
	_, _ = database.Exec(`INSERT INTO projects (id, name) VALUES ($1, 'test-proj') ON CONFLICT DO NOTHING`, projectID)
	_, _ = database.Exec(`INSERT INTO environments (id, project_id, name, key) VALUES ($1, $2, 'test-env', 'test') ON CONFLICT DO NOTHING`, envID, projectID)

	event := protocol.EvaluationEvent{
		ProjectID:     projectID,
		EnvironmentID: envID,
		FlagKey:       "test-flag",
		UserID:        "user-1",
		VariationKey:  "variation-1",
		EvaluatedAt:   time.Now().Unix(),
	}

	err = processor.Process([]protocol.EvaluationEvent{event})
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	// 6. Verify event is persisted in PostgreSQL
	var count int
	for i := 0; i < 10; i++ {
		err = database.QueryRowContext(ctx, 
			"SELECT COUNT(*) FROM evaluation_events WHERE project_id = $1 AND environment_id = $2 AND flag_key = $3",
			projectID, envID, event.FlagKey).Scan(&count)
		
		if err == nil && count > 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if count == 0 {
		t.Errorf("Event was not persisted to PostgreSQL after timeout")
	}
}

func setupTestDB(t *testing.T, database *sql.DB) {
	// Simple table creation if they don't exist, to avoid dependency on migration files location in tests
	_, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		);
		CREATE TABLE IF NOT EXISTS environments (
			id UUID PRIMARY KEY,
			project_id UUID NOT NULL REFERENCES projects(id),
			name VARCHAR(255) NOT NULL,
			key VARCHAR(255) NOT NULL
		);
		CREATE TABLE IF NOT EXISTS evaluation_events (
			id SERIAL PRIMARY KEY,
			project_id UUID NOT NULL REFERENCES projects(id),
			environment_id UUID NOT NULL REFERENCES environments(id),
			flag_key VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			variation_key VARCHAR(255),
			value JSONB,
			reason VARCHAR(255),
			context JSONB,
			evaluated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Failed to setup test database tables: %v", err)
	}
}
