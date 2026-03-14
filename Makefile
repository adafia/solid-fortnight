.PHONY: all start-db stop-db run-app test-api-create test-api-get test-api-update test-api-delete

# ====================================================================================
#  DATABASE
# ====================================================================================

start-db:
	@echo "Starting PostgreSQL database..."
	@docker-compose -f deployments/docker-compose.yml up -d postgres

stop-db:
	@echo "Stopping PostgreSQL database..."
	@docker-compose -f deployments/docker-compose.yml stop postgres

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
	@go run apps/management/main.go

# ====================================================================================
#  API TESTS
# ====================================================================================

# Use the Bruno collection in the /bruno directory to test the API.
