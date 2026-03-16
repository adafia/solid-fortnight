# Testing Strategy

This document outlines the testing strategy for the `solid-fortnight` project, focusing on the management service and its integration with the database.

## 1. Overview

The project uses a multi-layered testing approach to ensure reliability and maintainability:

* **Unit Tests:** Test individual components (e.g., configuration loading, utility functions) in isolation without external dependencies.
* **Integration Tests:** Verify that different parts of the system work together, specifically handlers and storage layers interacting with a real PostgreSQL database.

## 2. Environment Isolation

To prevent tests from interfering with development data, a dedicated test environment is used.

### Development Environment

* **Database Port:** `5432`
* **Database Name:** `solid_fortnight`
* **Command to start:** `make start-db`

### Testing Environment

* **Database Port:** `5433`
* **Database Name:** `solid_fortnight_test`
* **Command to start/run:** `make test`

## 3. Integration Testing Workflow

Integration tests for the management service (located in `apps/management/handlers/`) and the streamer service (`apps/streamer/`) follow this lifecycle:

1. **Orchestration:** The `make test` command starts dedicated PostgreSQL and Redis containers using `deployments/docker-compose.test.yml`.
2. **Health Check:** The Makefile waits for both databases to be "healthy" before starting the tests.
3. **Setup:**
    * **Management/Evaluator:** Loads configuration, overrides connection details, and runs migrations.
    * **Streamer:** Connects to the test Redis instance and verifies Pub/Sub broadcasting.
4. **Execution:** Each test run ensures a clean state (truncating tables for Postgres).
5. **Cleanup:** Containers are automatically stopped and removed.

## 4. Running Tests

### All Tests

```bash
make test
```

### Specific Module Tests

```bash
# 1. Start the test databases
make test-db-up

# 2. Run your specific tests (e.g., Streamer)
REDIS_ADDR=localhost:6380 go test -v ./apps/streamer

# 3. Stop the test databases
make test-db-down
```

## 5. Adding New Tests

* **Unit Tests:** Create `*_test.go` files in the same package as the code being tested.
* **Integration Tests:** Add tests to `apps/management/handlers/` or create new handler test files. Use the existing `TestMain` pattern to leverage the database orchestration.
* **Data Cleanup:** Always use `truncateTables()` or a similar mechanism in your tests to avoid side effects between test runs.
