.PHONY: help build migrate sqlc test docker-up docker-down run

# Help
help:
	@echo "Available commands:"
	@echo "  make docker-up                         - Start PostgreSQL database"
	@echo "  make docker-down                       - Stop all containers"
	@echo "  make migrate-up SERVICE=user           - Run migrations for service"
	@echo "  make migrate-down SERVICE=user         - Rollback migrations"
	@echo "  make sqlc SERVICE=user                 - Generate SQLC for service"
	@echo "  make build SERVICE=user-service        - Build specific service"
	@echo "  make run SERVICE=user-service          - Run specific service"
	@echo "  make test                              - Run all tests"
	@echo "  make dev-setup                         - Setup development environment"

# Docker commands
docker-up:
	@echo "Starting PostgreSQL database..."
	@docker compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5

docker-down:
	@echo "Stopping all containers..."
	@docker compose down

docker-logs:
	@docker compose logs -f

# Migration commands
migrate-up:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@echo "Running migrations for $(SERVICE)-service..."
	@set -a && source .env && set +a && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" up

migrate-down:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@echo "Rolling back migration for $(SERVICE)-service..."
	@set -a && source .env && set +a && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" down

migrate-status:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@echo "Migration status for $(SERVICE)-service..."
	@set -a && source .env && set +a && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" status

# SQLC generation
sqlc:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@echo "Generating SQLC for $(SERVICE)-service..."
	@sqlc generate -f sqlc/$(SERVICE)-service/sqlc.yaml

# Build commands
build:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<service-name>"; \
		exit 1; \
	fi
	@echo "Building $(SERVICE)..."
	@go build -o bin/$(SERVICE) ./cmd/$(SERVICE)

build-all:
	@echo "Building all services..."
	@mkdir -p bin
	@go build -o bin/user-service ./cmd/user-service
	@echo "Built all services successfully"

# Run commands
run:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<service-name>"; \
		exit 1; \
	fi
	@echo "Running $(SERVICE)..."
	@export $$(cat .env | grep -v '^#' | xargs) && go run ./cmd/$(SERVICE)/main.go

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "1. Starting PostgreSQL database..."
	@make docker-up
	@echo "2. Waiting for database to initialize..."
	@sleep 10
	@echo "3. Running user service migrations..."
	@make migrate-up SERVICE=user
	@echo "4. Generating SQLC code..."
	@make sqlc SERVICE=user
	@echo ""
	@echo "✅ Development environment ready!"
	@echo "💡 You can now run: make run SERVICE=user-service"

# Testing
test:
	@go test -v ./...

test-service:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<user|event|booking|search|analytics>"; \
		exit 1; \
	fi
	@go test -v ./services/$(SERVICE)/...

# Go module management
tidy:
	@go mod tidy

# Clean
clean:
	@rm -rf bin/
	@echo "Cleaned build artifacts"

# Generate everything
generate: sqlc

# Install dependencies (for SQLC and Goose)
install-tools:
	@echo "Installing development tools..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "Tools installed successfully"