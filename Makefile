.PHONY: help build migrate sqlc test run kill-services check-env migrate-up-all build-all

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
	@echo ""
	@echo "🏗️  Build & Development:"
	@echo "  make build SERVICE=booking-service     - Build specific service"
	@echo "  make build-all                         - Build all services"
	@echo "  make run SERVICE=booking-service       - Run specific service"
	@echo "  make kill-services                     - Stop all running services"
	@echo "  make clean                             - Clean build artifacts"
	@echo "  make tidy                              - Tidy Go modules"
	@echo ""
	@echo "🗄️  Database & Migrations:"
	@echo "  make migrate-up SERVICE=booking        - Run migrations for service"
	@echo "  make migrate-down SERVICE=booking      - Rollback migrations"
	@echo "  make migrate-status SERVICE=booking    - Check migration status"
	@echo "  make migrate-up-all                    - Run all migrations"
	@echo ""
	@echo "🔧 Code Generation:"
	@echo "  make sqlc SERVICE=booking              - Generate SQLC for service"
	@echo "  make sqlc-all                          - Generate SQLC for all services"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test                              - Run all tests"
	@echo "  make test-service SERVICE=booking      - Run tests for specific service"
	@echo ""
	@echo "🚀 Production Deployment:"
	@echo "  make deploy-production                 - Deploy all services"
	@echo "  make setup-db                          - Run database migrations"
	@echo "  make seed-data-production              - Create test data"
	@echo "  make deploy-full                       - Complete deployment"
	@echo "  make reset-data                        - Clean all data"
	@echo "  make status-production                 - Check deployment status"
	@echo "  make stop-production                   - Stop deployment"

migrate-up: check-env
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@export $$(cat .env | grep -v '^#' | xargs) && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" up

migrate-down: check-env
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@export $$(cat .env | grep -v '^#' | xargs) && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" down

migrate-status: check-env
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@export $$(cat .env | grep -v '^#' | xargs) && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" status

migrate-up-all:
	@make migrate-up SERVICE=user
	@make migrate-up SERVICE=event
	@make migrate-up SERVICE=booking

sqlc:
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<user|event|booking>"; \
		exit 1; \
	fi
	@sqlc generate -f sqlc/$(SERVICE)-service/sqlc.yaml

sqlc-all:
	@make sqlc SERVICE=user
	@make sqlc SERVICE=event
	@make sqlc SERVICE=booking

build:
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<service-name>"; \
		exit 1; \
	fi
	@go build -o bin/$(SERVICE) ./cmd/$(SERVICE)

build-all:
	@mkdir -p bin
	@go build -o bin/user-service ./cmd/user-service
	@go build -o bin/event-service ./cmd/event-service
	@go build -o bin/booking-service ./cmd/booking-service
	@go build -o bin/search-service ./cmd/search-service

run: check-env
	@if [ -z "$(SERVICE)" ]; then \
		echo "ERROR: Please specify SERVICE=<service-name>"; \
		exit 1; \
	fi
	@export $$(cat .env | grep -v '^#' | xargs) && go run ./cmd/$(SERVICE)/main.go

test:
	@go test -v ./...

test-service:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<user|event|booking|search>"; \
		exit 1; \
	fi
	@go test -v ./services/$(SERVICE)/...

tidy:
	@go mod tidy

clean:
	@rm -rf bin/

deploy-production:
	@if [ ! -f .env.production ]; then echo "ERROR: .env.production not found"; exit 1; fi
	@COMPOSE_FILE=docker-compose.yml env $$(cat .env.production | grep -v '^#' | xargs) docker compose up -d --build
	@sleep 60
	@echo "Deployment ready. Next: make setup-db && make seed-data-production"

setup-db:
	@COMPOSE_FILE=docker-compose.yml env $$(cat .env.production | grep -v '^#' | xargs) $(MAKE) migrate-up-all

reset-data:
	@docker compose exec postgres psql -U postgres -d users_db -c "TRUNCATE TABLE users, refresh_tokens RESTART IDENTITY CASCADE;" || true
	@docker compose exec postgres psql -U postgres -d events_db -c "TRUNCATE TABLE venues, events, admins, admin_refresh_tokens RESTART IDENTITY CASCADE;" || true
	@docker compose exec postgres psql -U postgres -d bookings_db -c "TRUNCATE TABLE bookings, payments, waitlist, booking_seats RESTART IDENTITY CASCADE;" || true
	@docker compose exec redis redis-cli FLUSHALL || true
	@curl -s -X POST "http://localhost:9200/events/_delete_by_query?conflicts=proceed" -H 'Content-Type: application/json' -d'{"query": {"match_all": {}}}' > /dev/null || true

seed-data-production:
	@sleep 10
	@COMPOSE_FILE=docker-compose.yml env $$(cat .env.production | grep -v '^#' | xargs) docker compose run --rm init-container

status-production:
	@docker compose ps
	@curl -s http://localhost/health || echo "Gateway: FAILED"

logs-production:
	@docker compose logs -f --tail=100

stop-production:
	@docker compose down

destroy-production:
	@docker compose down --volumes --remove-orphans
	@docker system prune -f

deploy-full:
	@$(MAKE) deploy-production
	@$(MAKE) setup-db
	@$(MAKE) seed-data-production