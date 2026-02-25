.PHONY: all start-db stop-db run-app test-api-create test-api-get test-api-update test-api-delete

# ====================================================================================
#  DATABASE
# ====================================================================================

start-db:
	@echo "Starting PostgreSQL database..."
	@docker-compose -f deployments/docker-compose.yml up -d

stop-db:
	@echo "Stopping PostgreSQL database..."
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

# Usage: make test-api-create project_id=<your-project-id>
test-api-create:
	@echo "Creating a new flag..."
	@./scripts/create_flag.sh $(project_id)

# Usage: make test-api-get flag_id=<your-flag-id>
test-api-get:
	@echo "Getting flag with ID $(flag_id)..."
	@./scripts/get_flag.sh $(flag_id)

# Usage: make test-api-update project_id=<your-project-id> flag_id=<your-flag-id>
test-api-update:
	@echo "Updating flag with ID $(flag_id)..."
	@./scripts/update_flag.sh $(project_id) $(flag_id)

# Usage: make test-api-delete flag_id=<your-flag-id>
test-api-delete:
	@echo "Deleting flag with ID $(flag_id)..."
	@./scripts/delete_flag.sh $(flag_id)
