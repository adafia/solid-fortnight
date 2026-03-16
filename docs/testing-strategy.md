# Testing Strategy

This document outlines the testing strategy for the `solid-fortnight` project, ensuring reliability across the management, evaluator, streamer, and analytics services.

## 1. Overview

The project uses a multi-layered testing approach to ensure reliability and maintainability:

* **Unit Tests:** Test individual components (e.g., `internal/engine`, configuration loading, utility functions) in isolation.
* **Integration Tests:** Verify that different parts of the system work together, specifically handlers and storage layers interacting with real PostgreSQL and Redis instances.
* **Benchmarks:** Measure and ensure the performance of critical components like the flag evaluation engine.

## 2. Environment Isolation

To prevent tests from interfering with development data, a dedicated test environment is used via Docker Compose.

### Development Environment

* **PostgreSQL:** `localhost:5432` (DB: `solid_fortnight`)
* **Redis:** `localhost:6379`
* **Command:** `make start-db`

### Testing Environment

* **PostgreSQL:** `localhost:5433` (DB: `solid_fortnight_test`)
* **Redis:** `localhost:6380`
* **Command:** `make test`

## 3. Testing Patterns

### Integration Testing (Synchronous)

Used for the Management API. These tests ensure that HTTP requests correctly modify the PostgreSQL state.

* **Workflow:** Setup DB -> Run Migration -> Execute Request -> Verify DB State -> Truncate Tables.

### Asynchronous Testing (Worker Pattern)

Used for the Analytics service. Testing the flow from the Ingestion API to Redis Streams and finally to PostgreSQL.

* **Pattern:** "Poll with Timeout". The test produces an event and then polls the database in a loop (with a timeout) until the expected record appears.
* **Example:** See `apps/analytics/handlers/analytics_integration_test.go`.

### Performance Benchmarking

Critical for the `internal/engine` package to ensure flag evaluation remains in the sub-millisecond range.

* **Command:** `go test -bench=. ./internal/engine`

## 4. Mocking Strategy

* **Prefer Real Infrastructure:** For integration tests, always use the real PostgreSQL/Redis containers provided by the test environment.
* **Use Mocks for Isolation:** Use interfaces and mocks (e.g., `mockProcessor` in Analytics) when testing high-level handlers to avoid heavy setup for simple logic checks.

## 5. Running Tests

### All Tests

```bash
make test
```

### Specific Service Tests

```bash
# Start test containers
make test-db-up

# Run specific tests (e.g., Evaluator)
go test -v ./apps/evaluator/handlers

# Run benchmarks
go test -bench=. ./internal/engine
```

## 6. Adding New Tests

* **Unit Tests:** Create `*_test.go` files in the same package.
* **Integration Tests:** Use the `TestMain` pattern or ensure containers are up.
* **Cleanup:** Always use `truncateTables()` or unique identifiers (UUIDs) for test data to avoid cross-test interference.
