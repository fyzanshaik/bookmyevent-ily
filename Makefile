.PHONY: help build migrate sqlc test docker-up docker-down run kill-services check-env

check-env:
	@if [ ! -f .env ]; then \
		echo "ERROR: .env file not found"; \
		exit 1; \
	fi

kill-services:
	@pkill -f "user-service|event-service|booking-service|search-service" 2>/dev/null || true
	@lsof -ti:8001 | xargs kill -9 2>/dev/null || true
	@lsof -ti:8002 | xargs kill -9 2>/dev/null || true
	@lsof -ti:8003 | xargs kill -9 2>/dev/null || true
	@lsof -ti:8004 | xargs kill -9 2>/dev/null || true

help:
	@echo "Available commands:"
	@echo "  make docker-up                         - Start PostgreSQL database"
	@echo "  make docker-redis-up                   - Start Redis"
	@echo "  make docker-full-up                    - Start all infrastructure (PostgreSQL + Redis + Elasticsearch)"
	@echo "  make docker-down                       - Stop all containers"
	@echo "  make kill-services                     - Stop all running services"
	@echo "  make migrate-up SERVICE=booking        - Run migrations for service (user|event|booking)"
	@echo "  make migrate-down SERVICE=booking      - Rollback migrations"
	@echo "  make sqlc SERVICE=booking              - Generate SQLC for service"
	@echo "  make build SERVICE=booking-service     - Build specific service"
	@echo "  make run SERVICE=booking-service       - Run specific service"
	@echo "  make test                              - Run all tests"
	@echo "  make dev-setup                         - Setup development environment"
	@echo "  make booking-dev-setup                 - Setup booking service development"
	@echo "  make redis-cli                         - Connect to Redis CLI"

docker-up: check-env
	@echo "Starting PostgreSQL database..."
	@export $$(cat .env | grep -v '^#' | xargs) && docker compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5

docker-redis-up: check-env
	@echo "Starting Redis..."
	@export $$(cat .env | grep -v '^#' | xargs) && docker compose up -d redis
	@echo "Waiting for Redis to be ready..."
	@sleep 3

docker-full-up: check-env
	@echo "Starting all infrastructure (PostgreSQL + Redis + Elasticsearch)..."
	@export $$(cat .env | grep -v '^#' | xargs) && docker compose up -d postgres redis elasticsearch
	@echo "Waiting for services to be ready..."
	@sleep 15

docker-down:
	@echo "Stopping all containers..."
	@docker compose down

docker-logs:
	@docker compose logs -f

# Redis utilities
redis-cli:
	@docker exec -it evently_redis redis-cli

redis-monitor:
	@docker exec -it evently_redis redis-cli monitor

# Elasticsearch utilities
elasticsearch-health:
	@curl -s http://localhost:9200/_cluster/health | jq

elasticsearch-indices:
	@curl -s http://localhost:9200/_cat/indices?v

elasticsearch-logs:
	@docker logs evently_elasticsearch --tail 50

# Migration commands
migrate-up: check-env
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@echo "Running migrations for $(SERVICE)-service..."
	@export $$(cat .env | grep -v '^#' | xargs) && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" up

migrate-down: check-env
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@echo "Rolling back migration for $(SERVICE)-service..."
	@export $$(cat .env | grep -v '^#' | xargs) && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" down

migrate-status: check-env
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@echo "Migration status for $(SERVICE)-service..."
	@export $$(cat .env | grep -v '^#' | xargs) && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" status

# SQLC generation
sqlc:
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@echo "Generating SQLC for $(SERVICE)-service..."
	@sqlc generate -f sqlc/$(SERVICE)-service/sqlc.yaml

build:
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<service-name>"; \
		exit 1; \
	fi
	@echo "Building $(SERVICE)..."
	@go build -o bin/$(SERVICE) ./cmd/$(SERVICE)

build-all:
	@echo "Building all services..."
	@mkdir -p bin
	@go build -o bin/user-service ./cmd/user-service
	@go build -o bin/event-service ./cmd/event-service
	@go build -o bin/booking-service ./cmd/booking-service
	@go build -o bin/search-service ./cmd/search-service
	@echo "Built all services successfully"

run: check-env
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<service-name>"; \
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

# Booking service specific development setup
booking-dev-setup:
	@echo "Setting up booking service development environment..."
	@echo "1. Starting all infrastructure (PostgreSQL + Redis)..."
	@make docker-full-up
	@echo "2. Running booking service migrations..."
	@make migrate-up SERVICE=booking
	@echo "3. Generating SQLC code for booking service..."
	@make sqlc SERVICE=booking
	@echo "4. Building booking service..."
	@make build SERVICE=booking-service
	@echo ""
	@echo "✅ Booking service development environment ready!"
	@echo "💡 You can now run: make run SERVICE=booking-service"
	@echo "💡 Redis CLI: make redis-cli"

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