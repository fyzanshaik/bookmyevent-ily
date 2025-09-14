# Development Commands Reference

## Quick Start Commands

### 1. Database Setup

```bash
# Start PostgreSQL with Docker
docker run --name evently-postgres \
  -e POSTGRES_DB=evently_dev \
  -e POSTGRES_USER=evently_user \
  -e POSTGRES_PASSWORD=evently_pass \
  -p 5433:5432 \
  -d postgres:15

# Connect to PostgreSQL
psql -h localhost -p 5433 -U evently_user -d evently_dev

# Or using Docker exec
docker exec -it evently-postgres psql -U evently_user -d evently_dev
```

### 2. Database Migrations

```bash
# Install goose (migration tool)
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run User Service migrations
goose -dir migrations/user-service postgres "postgres://evently_user:evently_pass@localhost:5433/users_db" up

# Run Event Service migrations
goose -dir migrations/event-service postgres "postgres://evently_user:evently_pass@localhost:5433/events_db" up

# Check migration status
goose -dir migrations/event-service postgres "postgres://evently_user:evently_pass@localhost:5433/events_db" status
```

### 3. SQLC Code Generation

```bash
# Install SQLC
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate User Service queries
cd sqlc/user-service && sqlc generate

# Generate Event Service queries
cd sqlc/event-service && sqlc generate
```

### 4. Build & Run Services

```bash
# Build specific service
make build SERVICE=user-service
make build SERVICE=event-service

# Run specific service
make run SERVICE=user-service    # Port 8001
make run SERVICE=event-service   # Port 8002

# Build all services
make build-all

# Clean build artifacts
make clean
```

### 5. Development Tools

```bash
# Format Go code
gofmt -w .

# Run linter
golangci-lint run

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Install dependencies
go mod tidy
go mod download
```

## Service-Specific Commands

### User Service (Port 8001)

```bash
# Test user registration
curl -X POST http://localhost:8001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"password123","name":"Test User"}'

# Test user login
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"password123"}'
```

### Event Service (Port 8002)

```bash
# Test admin registration
curl -X POST http://localhost:8002/api/v1/auth/admin/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"password123","name":"Test Admin"}'

# Test admin login
curl -X POST http://localhost:8002/api/v1/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"password123"}'

# Health check
curl http://localhost:8002/healthz
```

## Database Management

### PostgreSQL Commands

```bash
# Create databases manually
psql -h localhost -p 5433 -U evently_user -c "CREATE DATABASE users_db;"
psql -h localhost -p 5433 -U evently_user -c "CREATE DATABASE events_db;"
psql -h localhost -p 5433 -U evently_user -c "CREATE DATABASE bookings_db;"

# List databases
psql -h localhost -p 5433 -U evently_user -l

# Connect to specific database
psql -h localhost -p 5433 -U evently_user -d events_db

# Useful SQL commands
\dt         # List tables
\d events   # Describe events table
\q          # Quit
```

### Common Database Queries

```sql
-- Check event service tables
\c events_db
SELECT * FROM venues LIMIT 5;
SELECT * FROM events LIMIT 5;
SELECT * FROM admins LIMIT 5;

-- Check user service tables
\c users_db
SELECT * FROM users LIMIT 5;
SELECT * FROM refresh_tokens LIMIT 5;

-- Monitor event availability
SELECT name, available_seats, total_capacity, status, version FROM events WHERE status = 'published';
```

## Docker Commands

### Full Docker Setup (Future)

```bash
# Build services with Docker
docker build -t evently-user-service -f docker/user-service/Dockerfile .
docker build -t evently-event-service -f docker/event-service/Dockerfile .

# Docker Compose (when available)
docker-compose up -d
docker-compose logs user-service
docker-compose logs event-service
```

### Docker PostgreSQL Management

```bash
# Stop PostgreSQL container
docker stop evently-postgres

# Start existing PostgreSQL container
docker start evently-postgres

# Remove PostgreSQL container (WARNING: loses data)
docker rm -f evently-postgres

# View PostgreSQL logs
docker logs evently-postgres

# Backup database
docker exec evently-postgres pg_dump -U evently_user events_db > events_backup.sql
```

## Environment Variables

### Required .env File

```bash
# Copy example environment file
cp .env.example .env

# Edit with your values
DATABASE_URL="postgres://evently_user:evently_pass@localhost:5433"
USER_SERVICE_PORT=8001
EVENT_SERVICE_PORT=8002
JWT_SECRET="your-super-secret-jwt-key-here"
INTERNAL_API_KEY="your-internal-service-api-key"
```

### Service-Specific Environment

```bash
# User Service
USER_SERVICE_DB_URL="postgres://evently_user:evently_pass@localhost:5433/users_db"
USER_SERVICE_PORT=8001
JWT_SECRET="your-jwt-secret"

# Event Service
EVENT_SERVICE_DB_URL="postgres://evently_user:evently_pass@localhost:5433/events_db"
EVENT_SERVICE_PORT=8002
INTERNAL_API_KEY="your-internal-api-key"
```

## Troubleshooting Commands

### Check Service Status

```bash
# Check if services are running
lsof -i :8001  # User Service
lsof -i :8002  # Event Service
lsof -i :5433  # PostgreSQL

# Kill processes on ports
kill $(lsof -ti:8001)
kill $(lsof -ti:8002)
```

### Database Connection Issues

```bash
# Test database connection
pg_isready -h localhost -p 5433

# Check PostgreSQL logs
docker logs evently-postgres --tail 50

# Reset PostgreSQL password
docker exec -it evently-postgres psql -U postgres
ALTER USER evently_user PASSWORD 'new_password';
```

### Build Issues

```bash
# Clean Go module cache
go clean -modcache

# Rebuild with verbose output
go build -v ./cmd/event-service

# Check Go environment
go env
```

## Monitoring & Debugging

### Service Logs

```bash
# Run services with debug logging
LOG_LEVEL=debug make run SERVICE=event-service

# Follow logs
tail -f logs/event-service.log
```

### Database Monitoring

```bash
# Monitor database connections
SELECT * FROM pg_stat_activity;

# Check table sizes
SELECT schemaname,tablename,pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Makefile Targets Available

```bash
make help           # Show all available targets
make build-all      # Build all services
make test          # Run all tests
make fmt           # Format code
make lint          # Run linter
make clean         # Clean build artifacts
make deps          # Install dependencies
```

## API Testing

### Using curl

```bash
# Test with authentication
TOKEN=$(curl -s -X POST http://localhost:8002/api/v1/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"password123"}' | jq -r .access_token)

curl -H "Authorization: Bearer $TOKEN" http://localhost:8002/api/v1/admin/events
```

### Using HTTPie (Alternative)

```bash
# Install HTTPie
pip install httpie

# Test API
http POST localhost:8002/api/v1/auth/admin/register email=admin@test.com password=password123 name="Test Admin"
```

Stop Everything:
pkill -f "user-service|event-service|booking-service"
make docker-down

Start Everything:
make docker-full-up
make migrate-up SERVICE=user
make migrate-up SERVICE=event
make migrate-up SERVICE=booking
make build-all

# Start services with env

export $(cat .env | grep -v '^#' | xargs) && ./bin/user-service &
export $(cat .env | grep -v '^#' | xargs) && ./bin/event-service &
export $(cat .env | grep -v '^#' | xargs) && ./bin/booking-service &

# Test

./scripts/testing/booking-service-test.sh

./scripts/dev-services.sh full-setup

2. Quick Commands for Development

Stop Everything:
./scripts/dev-services.sh clean

Quick Start (when infra is already running):
./scripts/dev-services.sh quick-start

Check Status:
./scripts/dev-services.sh status

Just Run Tests:
./scripts/dev-services.sh test

3. Individual Commands

Infrastructure Management:
./scripts/dev-services.sh start-infra # PostgreSQL + Redis
./scripts/dev-services.sh stop-infra # Stop containers

Service Management:
./scripts/dev-services.sh build # Build all services
./scripts/dev-services.sh start-services # Start user → event → booking
./scripts/dev-services.sh stop-services # Kill all services
./scripts/dev-services.sh restart # Restart services only

Database:
./scripts/dev-services.sh migrate # Run all migrations

Health Checks:
./scripts/dev-services.sh health # Check all services

For Initial Setup:
./scripts/dev-services.sh full-setup

For Daily Development:

# When you make changes to code

./scripts/dev-services.sh quick-start

# When you want to test

./scripts/dev-services.sh test

# When you're done

./scripts/dev-services.sh clean

📋 What the Script Does

full-setup Command Flow:

1. Stop all running services and infrastructure
2. Start PostgreSQL + Redis containers
3. Run migrations for user, event, booking services
4. Build all service binaries
5. Start services in correct order with env variables:
   - export $(cat .env | grep -v '^#' | xargs) && ./bin/user-service &
   - export $(cat .env | grep -v '^#' | xargs) && ./bin/event-service &
   - export $(cat .env | grep -v '^#' | xargs) && ./bin/booking-service &

6. Wait for services to stabilize
7. Check health of all services
8. Run the comprehensive booking service test suite
