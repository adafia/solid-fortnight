.PHONY: all start-db stop-db test-db-up test-db-down test start-all stop-all run-app test-api-create test-api-get test-api-update test-api-delete

# ====================================================================================
#  DATABASE
# ====================================================================================

start-db:
	@echo "Starting PostgreSQL database..."
	@docker-compose -f deployments/docker-compose.yml up -d postgres

stop-db:
	@echo "Stopping PostgreSQL database..."
	@docker-compose -f deployments/docker-compose.yml stop postgres

test-db-up:
	@echo "Starting PostgreSQL test database..."
	@docker-compose -f deployments/docker-compose.test.yml up -d postgres-test
	@echo "Waiting for test database to be ready..."
	@for i in {1..20}; do \
		if docker inspect -f '{{.State.Health.Status}}' deployments-postgres-test-1 | grep -q "healthy"; then \
			echo "Test database is ready!"; \
			exit 0; \
		fi; \
		echo "Waiting... ($$i/20)"; \
		sleep 1; \
	done; \
	echo "Error: Test database failed to become healthy."; \
	exit 1

test-db-down:
	@echo "Stopping PostgreSQL test database..."
	@docker-compose -f deployments/docker-compose.test.yml down

test: test-db-up
	@echo "Running tests..."
	@POSTGRES_HOST=localhost POSTGRES_PORT=5433 DB_NAME=solid_fortnight_test DB_USER=testuser DB_PASSWORD=testpassword go test -v ./apps/management/handlers ./apps/evaluator/handlers ./internal/config ./internal/engine
	@$(MAKE) test-db-down

start-all:
	@echo "Starting the entire application with Docker Compose..."
	@docker-compose -f deployments/docker-compose.yml up -d --build

stop-all:
	@echo "Stopping the entire application with Docker Compose..."
	@docker-compose -f deployments/docker-compose.yml down

# ====================================================================================
#  APPLICATION
# ====================================================================================

run-app:
	@echo "Running the management service..."
	@POSTGRES_HOST=localhost POSTGRES_PORT=5432 DB_NAME=solid_fortnight DB_USER=postgres DB_PASSWORD=password go run apps/management/main.go

# ====================================================================================
#  API TESTS
# ====================================================================================

# Use the Bruno collection in the /bruno directory to test the API.
