.PHONY: all start-db stop-db test-db-up test-db-down test start-all stop-all run-app test-api-create test-api-get test-api-update test-api-delete setup-colima

# ====================================================================================
#  SETUP
# ====================================================================================

setup-colima:
	@echo "Configuring Colima with 4 CPUs and 8GiB RAM for better build performance..."
	@colima stop || true
	@colima start --cpu 4 --memory 8

clean-docker:
	@echo "Cleaning up unused Docker resources..."
	@docker system prune -af
	@docker builder prune -af

# ====================================================================================
#  DATABASE
# ====================================================================================

start-db:
	@echo "Starting PostgreSQL and Redis..."
	@docker-compose -f deployments/docker-compose.yml up -d postgres redis

stop-db:
	@echo "Stopping PostgreSQL and Redis..."
	@docker-compose -f deployments/docker-compose.yml stop postgres redis

test-db-up:
	@echo "Starting PostgreSQL and Redis test databases..."
	@docker-compose -f deployments/docker-compose.test.yml up -d postgres-test redis-test
	@echo "Waiting for test databases to be ready..."
	@for i in {1..20}; do \
		if docker inspect -f '{{.State.Health.Status}}' deployments-postgres-test-1 | grep -q "healthy" && \
		   docker inspect -f '{{.State.Health.Status}}' deployments-redis-test-1 | grep -q "healthy"; then \
			echo "Test databases are ready!"; \
			exit 0; \
		fi; \
		echo "Waiting... ($$i/20)"; \
		sleep 1; \
	done; \
	echo "Error: Test databases failed to become healthy."; \
	exit 1

test-db-down:
	@echo "Stopping PostgreSQL and Redis test databases..."
	@docker-compose -f deployments/docker-compose.test.yml down

test: test-db-up
	@echo "Running tests..."
	@POSTGRES_HOST=localhost POSTGRES_PORT=5433 DB_NAME=solid_fortnight_test DB_USER=testuser DB_PASSWORD=testpassword REDIS_ADDR=localhost:6380 go test -v ./apps/management/handlers ./apps/evaluator/handlers ./apps/streamer ./apps/analytics/handlers ./apps/gateway/... ./internal/config ./internal/engine
	@$(MAKE) test-db-down

test-e2e:
	@echo "Running E2E integration tests..."
	@$(MAKE) start-all
	@echo "Waiting for services to be ready..."
	@sleep 10
	@cd cmd/dashboard && bunx playwright test

test-ui:
	@echo "Running fast UI unit tests..."
	@cd cmd/dashboard && bun run test --run

start-all:
	@echo "Starting the entire application with Docker Compose..."
	@docker-compose -f deployments/docker-compose.yml up -d

stop-all:
	@echo "Stopping the entire application with Docker Compose..."
	@docker-compose -f deployments/docker-compose.yml down

# ====================================================================================
#  APPLICATION
# ====================================================================================

run-app:
	@echo "Running the management service..."
	@POSTGRES_HOST=localhost POSTGRES_PORT=5432 DB_NAME=solid_fortnight DB_USER=postgres DB_PASSWORD=password REDIS_ADDR=localhost:6379 go run apps/management/main.go

run-evaluator:
	@echo "Running the evaluator service..."
	@POSTGRES_HOST=localhost POSTGRES_PORT=5432 DB_NAME=solid_fortnight DB_USER=postgres DB_PASSWORD=password go run apps/evaluator/main.go

run-streamer:
	@echo "Running the streamer service..."
	@REDIS_ADDR=localhost:6379 go run apps/streamer/main.go

run-analytics:
	@echo "Running the analytics service..."
	@REDIS_ADDR=localhost:6379 go run apps/analytics/main.go

run-gateway:
	@echo "Running the API Gateway..."
	@go run apps/gateway/main.go

start-dashboard:
	@echo "Starting the Admin Dashboard..."
	@cd cmd/dashboard && bun run dev

# ====================================================================================
#  API TESTS
# ====================================================================================

# Use the Bruno collection in the /bruno directory to test the API.
