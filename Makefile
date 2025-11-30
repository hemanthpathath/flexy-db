# flex-db Makefile
#
# Development and testing workflow for the Python JSON-RPC backend
#
# Usage:
#   make setup-dev   - Build and start the development environment
#   make test-all    - Run all tests in isolation and tear down
#   make stop        - Stop all running containers
#   make clean       - Remove all containers, volumes, and images
#

.PHONY: setup-dev test-all stop clean help build logs status

# Default target
help:
	@echo "flex-db Development and Testing Workflow"
	@echo ""
	@echo "Usage:"
	@echo "  make setup-dev   - Build and start the development environment"
	@echo "  make test-all    - Run all tests in isolation and tear down"
	@echo "  make stop        - Stop all running containers"
	@echo "  make clean       - Remove all containers, volumes, and images"
	@echo "  make build       - Build all Docker images"
	@echo "  make logs        - Show logs from running containers"
	@echo "  make status      - Show status of running containers"
	@echo ""
	@echo "Development Environment:"
	@echo "  - PostgreSQL:     localhost:5432"
	@echo "  - JSON-RPC API:   http://localhost:5000/jsonrpc"
	@echo "  - OpenRPC Spec:   http://localhost:5000/openrpc.json"
	@echo "  - Health Check:   http://localhost:5000/health"

# Build all Docker images
build:
	@echo "Building Docker images..."
	docker compose --profile dev build

# Setup development environment
setup-dev:
	@echo "=========================================="
	@echo "Setting up development environment..."
	@echo "=========================================="
	@echo ""
	docker compose --profile dev up --build -d
	@echo ""
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@echo ""
	@echo "=========================================="
	@echo "Development environment is ready!"
	@echo "=========================================="
	@echo ""
	@echo "Services available:"
	@echo "  - PostgreSQL:      localhost:5432"
	@echo "  - JSON-RPC API:    http://localhost:5000/jsonrpc"
	@echo "  - OpenRPC Spec:    http://localhost:5000/openrpc.json"
	@echo "  - Health Check:    http://localhost:5000/health"
	@echo ""
	@echo "To view logs: make logs"
	@echo "To stop:      make stop"

# Run all tests in isolation
test-all:
	@echo "=========================================="
	@echo "Running all tests in isolation..."
	@echo "=========================================="
	@echo ""
	@echo "Starting test environment..."
	docker compose --profile test up --build --abort-on-container-exit --exit-code-from test-runner
	@TEST_EXIT_CODE=$$?; \
	echo ""; \
	echo "Tearing down test environment..."; \
	docker compose --profile test down -v; \
	echo ""; \
	if [ $$TEST_EXIT_CODE -eq 0 ]; then \
		echo "=========================================="; \
		echo "All tests passed!"; \
		echo "=========================================="; \
	else \
		echo "=========================================="; \
		echo "Tests failed with exit code: $$TEST_EXIT_CODE"; \
		echo "=========================================="; \
		exit $$TEST_EXIT_CODE; \
	fi

# Stop all running containers
stop:
	@echo "Stopping all containers..."
	docker compose --profile dev down
	docker compose --profile test down
	@echo "All containers stopped."

# Clean up everything (containers, volumes, images)
clean:
	@echo "Cleaning up..."
	docker compose --profile dev down -v --rmi local
	docker compose --profile test down -v --rmi local
	@echo "Cleanup complete."

# Show logs from running containers
logs:
	docker compose --profile dev logs -f

# Show status of running containers
status:
	@echo "Container status:"
	@docker compose --profile dev ps
	@echo ""
	@echo "Test container status:"
	@docker compose --profile test ps
