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
	@echo "  - JSON-RPC API:   http://localhost:8080/jsonrpc"
	@echo "  - OpenRPC Spec:   http://localhost:8080/openrpc.json"
	@echo "  - Health Check:   http://localhost:8080/health"

# Build all Docker images
build:
	@echo "Building Docker images..."
	@cp .env.local .env.compose
	@export $$(cat .env.local | grep -v '^#' | xargs) && docker compose -p flex-db-dev build
	@rm -f .env.compose

# Setup development environment
setup-dev:
	@echo "=========================================="
	@echo "Setting up development environment..."
	@echo "=========================================="
	@echo ""
	@cp .env.local .env.compose
	@export $$(cat .env.local | grep -v '^#' | xargs) && docker compose -p flex-db-dev up --build -d
	@rm -f .env.compose
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
	@echo "  - JSON-RPC API:    http://localhost:8080/jsonrpc"
	@echo "  - OpenRPC Spec:    http://localhost:8080/openrpc.json"
	@echo "  - Health Check:    http://localhost:8080/health"
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
	@cp .env.test .env.compose
	@export $$(cat .env.test | grep -v '^#' | xargs) && docker compose -p flex-db-test up --build --abort-on-container-exit --exit-code-from backend; \
	TEST_EXIT_CODE=$$?; \
	echo ""; \
	echo "Tearing down test environment..."; \
	export $$(cat .env.test | grep -v '^#' | xargs) && docker compose -p flex-db-test down -v; \
	rm -f .env.compose; \
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
	@if [ -f .env.local ]; then cp .env.local .env.compose && export $$(cat .env.local | grep -v '^#' | xargs) && docker compose -p flex-db-dev down && rm -f .env.compose; fi
	@if [ -f .env.test ]; then cp .env.test .env.compose && export $$(cat .env.test | grep -v '^#' | xargs) && docker compose -p flex-db-test down && rm -f .env.compose; fi
	@echo "All containers stopped."

# Clean up everything (containers, volumes, images)
clean:
	@echo "Cleaning up..."
	@if [ -f .env.local ]; then cp .env.local .env.compose && export $$(cat .env.local | grep -v '^#' | xargs) && docker compose -p flex-db-dev down -v --rmi local && rm -f .env.compose; fi
	@if [ -f .env.test ]; then cp .env.test .env.compose && export $$(cat .env.test | grep -v '^#' | xargs) && docker compose -p flex-db-test down -v --rmi local && rm -f .env.compose; fi
	@echo "Cleanup complete."

# Show logs from running containers
logs:
	@if [ -f .env.compose ]; then \
		export $$(cat .env.compose | grep -v '^#' | xargs) && docker compose -p flex-db-dev logs -f; \
	else \
		cp .env.local .env.compose && export $$(cat .env.local | grep -v '^#' | xargs) && docker compose -p flex-db-dev logs -f && rm -f .env.compose; \
	fi

# Show status of running containers
status:
	@echo "Development container status:"
	@cp .env.local .env.compose 2>/dev/null || true
	@export $$(cat .env.local 2>/dev/null | grep -v '^#' | xargs) && docker compose -p flex-db-dev ps 2>/dev/null || echo "  (no dev containers running)"
	@echo ""
	@echo "Test container status:"
	@cp .env.test .env.compose 2>/dev/null || true
	@export $$(cat .env.test 2>/dev/null | grep -v '^#' | xargs) && docker compose -p flex-db-test ps 2>/dev/null || echo "  (no test containers running)"
	@rm -f .env.compose
