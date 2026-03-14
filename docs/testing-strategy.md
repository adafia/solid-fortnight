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

Integration tests for the management service (located in `apps/management/handlers/`) follow this lifecycle:

1. **Orchestration:** The `make test` command starts a dedicated PostgreSQL container using `deployments/docker-compose.test.yml`.
2. **Health Check:** The Makefile waits for the database to be "healthy" (via `pg_isready`) before starting the tests.
3. **Setup (`TestMain`):**
    * Loads configuration from `deployments/config.yaml`.
    * Overrides connection details using environment variables (`POSTGRES_PORT=5433`, etc.).
    * Runs database migrations to ensure the test schema is up-to-date.
4. **Execution:** Each test run truncates relevant tables to ensure a clean state for every test case.
5. **Cleanup:** After the tests finish, the test database container and network are automatically stopped and removed.

## 4. Running Tests

### All Tests

To run all tests (unit and integration) across the workspace:

```bash
make test
```

### Specific Module Tests

If you want to run tests for a specific module without the full `make` orchestration, you must ensure the test database is running and provide the environment variables:

```bash
# 1. Start the test database
make test-db-up

# 2. Run your specific tests
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
DB_NAME=solid_fortnight_test
DB_USER=testuser
DB_PASSWORD=testpassword
go test -v ./apps/management/handlers/...

# 3. Stop the test database
make test-db-down
```

## 5. Adding New Tests

* **Unit Tests:** Create `*_test.go` files in the same package as the code being tested.
* **Integration Tests:** Add tests to `apps/management/handlers/` or create new handler test files. Use the existing `TestMain` pattern to leverage the database orchestration.
* **Data Cleanup:** Always use `truncateTables()` or a similar mechanism in your tests to avoid side effects between test runs.
