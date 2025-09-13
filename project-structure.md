# Evently Monorepo Structure - Refined Architecture

## Project Structure

```
evently/
├── go.mod                           # Single module for entire project
├── go.sum
├── Makefile
├── docker-compose.yml
├── .env.example                     # All service configurations
├── .env                            # Actual environment file (gitignored)
├── README.md
│
├── cmd/                            # Service entry points
│   ├── user-service/
│   │   └── main.go                # Entry point
│   ├── event-service/
│   │   └── main.go
│   ├── search-service/
│   │   └── main.go
│   ├── booking-service/
│   │   └── main.go
│   └── analytics-service/
│       └── main.go
│
├── internal/                       # Shared packages across all services
│   ├── config/
│   │   └── config.go              # Load env vars for each service
│   ├── database/                  # Database connection management
│   │   ├── postgres.go            # PostgreSQL connection factory
│   │   ├── redis.go               # Redis client factory
│   │   └── elasticsearch.go       # Elasticsearch client factory
│   ├── repository/                # ALL database queries for ALL services
│   │   ├── users/                 # User service queries
│   │   │   ├── querier.go         # SQLC generated interface
│   │   │   ├── users.sql.go       # SQLC generated
│   │   │   ├── tokens.sql.go      # SQLC generated
│   │   │   └── models.go          # SQLC generated models
│   │   ├── events/                # Event service queries
│   │   │   ├── querier.go
│   │   │   ├── events.sql.go
│   │   │   ├── venues.sql.go
│   │   │   ├── admins.sql.go
│   │   │   └── models.go
│   │   ├── bookings/              # Booking service queries
│   │   │   ├── querier.go
│   │   │   ├── bookings.sql.go
│   │   │   ├── payments.sql.go
│   │   │   ├── waitlist.sql.go
│   │   │   └── models.go
│   │   └── search/                # Elasticsearch queries
│   │       └── search.go          # ES query builders
│   ├── auth/                      # Shared auth utilities
│   │   ├── jwt.go                 # JWT generation/validation
│   │   ├── middleware.go          # Auth middleware
│   │   ├── password.go            # Password hashing
│   │   └── service_auth.go        # Inter-service authentication
│   ├── utils/
│   │   ├── response.go            # Standard HTTP responses
│   │   ├── errors.go              # Error types
│   │   ├── validator.go           # Input validation
│   │   └── idempotency.go         # Idempotency key handling
│   ├── logger/
│   │   └── logger.go
│   └── metrics/
│       └── metrics.go
│
├── services/                       # Service-specific logic
│   ├── user/
│   │   ├── server.go              # HTTP server setup & routes
│   │   ├── handlers.go            # All HTTP handlers
│   │   ├── internal_handlers.go   # Inter-service endpoints
│   │   └── service.go             # Business logic
│   ├── event/
│   │   ├── server.go
│   │   ├── handlers.go
│   │   ├── internal_handlers.go
│   │   └── service.go
│   ├── search/
│   │   ├── server.go
│   │   ├── handlers.go
│   │   └── service.go
│   ├── booking/
│   │   ├── server.go
│   │   ├── handlers.go
│   │   ├── internal_handlers.go
│   │   └── service.go
│   └── analytics/
│       ├── server.go
│       ├── handlers.go
│       └── service.go
│
├── migrations/                     # Goose migrations per service
│   ├── user-service/
│   │   ├── 00001_create_users_table.sql
│   │   ├── 00002_create_refresh_tokens_table.sql
│   │   └── 00003_add_user_indexes.sql
│   ├── event-service/
│   │   ├── 00001_create_venues_table.sql
│   │   ├── 00002_create_events_table.sql
│   │   ├── 00003_create_admins_table.sql
│   │   └── 00004_add_event_indexes.sql
│   └── booking-service/
│       ├── 00001_create_bookings_table.sql
│       ├── 00002_create_payments_table.sql
│       ├── 00003_create_waitlist_table.sql
│       └── 00004_add_booking_indexes.sql
│
├── sqlc/                          # SQLC configuration and queries
│   ├── user-service/
│   │   ├── sqlc.yaml
│   │   ├── schema.sql             # Combined schema for SQLC
│   │   └── queries/
│   │       ├── users.sql
│   │       └── tokens.sql
│   ├── event-service/
│   │   ├── sqlc.yaml
│   │   ├── schema.sql
│   │   └── queries/
│   │       ├── events.sql
│   │       ├── venues.sql
│   │       └── admins.sql
│   └── booking-service/
│       ├── sqlc.yaml
│       ├── schema.sql
│       └── queries/
│           ├── bookings.sql
│           ├── payments.sql
│           └── waitlist.sql
│
├── scripts/
│   ├── migrate.sh                 # Service-specific migration runner
│   ├── build.sh                   # Build individual services
│   └── setup-dbs.sql              # Create all databases
│
└── deployments/
    └── docker/
        ├── user-service.Dockerfile
        ├── event-service.Dockerfile
        └── ...
```

## Key Files Implementation

### 1. Environment File (`.env.example`)

```bash
# Service Ports
USER_SERVICE_PORT=8001
EVENT_SERVICE_PORT=8002
SEARCH_SERVICE_PORT=8003
BOOKING_SERVICE_PORT=8004
ANALYTICS_SERVICE_PORT=8005

# Database URLs - Master
USER_SERVICE_DB_URL=postgresql://postgres:postgres@localhost:5432/users_db?sslmode=disable
EVENT_SERVICE_DB_URL=postgresql://postgres:postgres@localhost:5432/events_db?sslmode=disable
BOOKING_SERVICE_DB_URL=postgresql://postgres:postgres@localhost:5432/bookings_db?sslmode=disable

# Database URLs - Read Replicas
USER_SERVICE_DB_REPLICA_URL=postgresql://postgres:postgres@localhost:5433/users_db?sslmode=disable
EVENT_SERVICE_DB_REPLICA_URL=postgresql://postgres:postgres@localhost:5433/events_db?sslmode=disable
BOOKING_SERVICE_DB_REPLICA_URL=postgresql://postgres:postgres@localhost:5433/bookings_db?sslmode=disable

# Shared Infrastructure
REDIS_URL=redis://localhost:6379
ELASTICSEARCH_URL=http://localhost:9200

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_ACCESS_TOKEN_DURATION=15m
JWT_REFRESH_TOKEN_DURATION=168h

# Inter-Service Authentication
INTERNAL_API_KEY=internal-service-communication-key

# Service URLs (for inter-service communication)
USER_SERVICE_URL=http://localhost:8001
EVENT_SERVICE_URL=http://localhost:8002
SEARCH_SERVICE_URL=http://localhost:8003
BOOKING_SERVICE_URL=http://localhost:8004

# Application Settings
LOG_LEVEL=info
ENVIRONMENT=development
```

### 2. Main Entry Point (`cmd/user-service/main.go`)

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/yourusername/evently/internal/config"
    "github.com/yourusername/evently/internal/database"
    "github.com/yourusername/evently/internal/logger"
    "github.com/yourusername/evently/internal/repository/users"
    "github.com/yourusername/evently/services/user"
)

func main() {
    // Load service-specific configuration
    cfg := config.LoadUserServiceConfig()

    // Initialize logger
    logger := logger.New(cfg.LogLevel)

    // Connect to master database (for writes)
    masterDB, err := database.NewPostgresConnection(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect to master database: %v", err)
    }
    defer masterDB.Close()

    // Connect to replica database (for reads)
    replicaDB, err := database.NewPostgresConnection(cfg.DatabaseReplicaURL)
    if err != nil {
        logger.Warn("Failed to connect to replica, falling back to master")
        replicaDB = masterDB
    } else {
        defer replicaDB.Close()
    }

    // Initialize Redis
    redisClient := database.NewRedisClient(cfg.RedisURL)
    defer redisClient.Close()

    // Initialize repository with database connections
    userRepo := users.New(masterDB)
    userRepoRead := users.New(replicaDB)

    // Create service with dependencies
    userService := user.NewService(userRepo, userRepoRead, redisClient, cfg)

    // Create and start server
    server := user.NewServer(userService, cfg, logger)

    // Start server in goroutine
    go func() {
        logger.Info("Starting User Service", "port", cfg.Port)
        if err := server.Start(); err != nil {
            logger.Fatal("Failed to start server", "error", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Graceful shutdown
    logger.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        logger.Error("Server forced to shutdown", "error", err)
    }

    logger.Info("Server exited")
}
```

### 3. Server Setup (`services/user/server.go`)

```go
package user

import (
    "context"
    "net/http"
    "time"

    "github.com/yourusername/evently/internal/auth"
    "github.com/yourusername/evently/internal/config"
    "github.com/yourusername/evently/internal/logger"
)

type Server struct {
    service    *Service
    httpServer *http.Server
    config     *config.UserServiceConfig
    logger     *logger.Logger
}

func NewServer(service *Service, cfg *config.UserServiceConfig, logger *logger.Logger) *Server {
    server := &Server{
        service: service,
        config:  cfg,
        logger:  logger,
    }

    // Setup routes
    mux := http.NewServeMux()
    server.setupRoutes(mux)

    // Create HTTP server
    server.httpServer = &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    return server
}

func (s *Server) setupRoutes(mux *http.ServeMux) {
    // Public endpoints
    mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
    mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
    mux.HandleFunc("POST /api/v1/auth/refresh", s.handleRefresh)

    // Protected endpoints (with auth middleware)
    mux.HandleFunc("GET /api/v1/users/profile", auth.RequireAuth(s.handleGetProfile))
    mux.HandleFunc("PUT /api/v1/users/profile", auth.RequireAuth(s.handleUpdateProfile))
    mux.HandleFunc("GET /api/v1/users/bookings", auth.RequireAuth(s.handleGetBookings))

    // Internal endpoints (for inter-service communication)
    mux.HandleFunc("POST /internal/auth/verify", auth.RequireInternalAuth(s.handleInternalVerify))
    mux.HandleFunc("GET /internal/users/{userId}", auth.RequireInternalAuth(s.handleInternalGetUser))

    // Health check
    mux.HandleFunc("GET /health", s.handleHealth)
    mux.HandleFunc("GET /health/ready", s.handleReadiness)
}

func (s *Server) Start() error {
    return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.httpServer.Shutdown(ctx)
}
```

### 4. Handlers (`services/user/handlers.go`)

```go
package user

import (
    "encoding/json"
    "net/http"

    "github.com/yourusername/evently/internal/utils"
)

// Public handlers
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate input
    if err := req.Validate(); err != nil {
        utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Call service
    resp, err := s.service.Register(r.Context(), req)
    if err != nil {
        s.logger.Error("Registration failed", "error", err)
        utils.HandleServiceError(w, err)
        return
    }

    utils.JSONResponse(w, resp, http.StatusCreated)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    resp, err := s.service.Login(r.Context(), req)
    if err != nil {
        utils.HandleServiceError(w, err)
        return
    }

    utils.JSONResponse(w, resp, http.StatusOK)
}

// Health checks
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    utils.JSONResponse(w, map[string]string{"status": "healthy"}, http.StatusOK)
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
    // Check database connectivity
    if err := s.service.CheckHealth(r.Context()); err != nil {
        utils.ErrorResponse(w, "Service not ready", http.StatusServiceUnavailable)
        return
    }
    utils.JSONResponse(w, map[string]string{"status": "ready"}, http.StatusOK)
}
```

### 5. Internal Handlers (`services/user/internal_handlers.go`)

```go
package user

import (
    "encoding/json"
    "net/http"

    "github.com/yourusername/evently/internal/utils"
)

// Internal endpoints for inter-service communication
func (s *Server) handleInternalVerify(w http.ResponseWriter, r *http.Request) {
    var req VerifyTokenRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.ErrorResponse(w, "Invalid request", http.StatusBadRequest)
        return
    }

    claims, err := s.service.VerifyToken(r.Context(), req.Token)
    if err != nil {
        utils.ErrorResponse(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    utils.JSONResponse(w, claims, http.StatusOK)
}

func (s *Server) handleInternalGetUser(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("userId")

    user, err := s.service.GetUserByID(r.Context(), userID)
    if err != nil {
        utils.HandleServiceError(w, err)
        return
    }

    utils.JSONResponse(w, user, http.StatusOK)
}
```

### 6. Config Loader (`internal/config/config.go`)

```go
package config

import (
    "os"
    "time"
)

type UserServiceConfig struct {
    Port                string
    DatabaseURL         string
    DatabaseReplicaURL  string
    RedisURL           string
    JWTSecret          string
    JWTAccessDuration  time.Duration
    JWTRefreshDuration time.Duration
    InternalAPIKey     string
    LogLevel           string
}

func LoadUserServiceConfig() *UserServiceConfig {
    return &UserServiceConfig{
        Port:                getEnv("USER_SERVICE_PORT", "8001"),
        DatabaseURL:         getEnvRequired("USER_SERVICE_DB_URL"),
        DatabaseReplicaURL:  getEnv("USER_SERVICE_DB_REPLICA_URL", ""),
        RedisURL:           getEnvRequired("REDIS_URL"),
        JWTSecret:          getEnvRequired("JWT_SECRET"),
        JWTAccessDuration:  getDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute),
        JWTRefreshDuration: getDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
        InternalAPIKey:     getEnvRequired("INTERNAL_API_KEY"),
        LogLevel:          getEnv("LOG_LEVEL", "info"),
    }
}

type EventServiceConfig struct {
    Port                string
    DatabaseURL         string
    DatabaseReplicaURL  string
    RedisURL           string
    InternalAPIKey     string
    UserServiceURL     string
    LogLevel           string
}

func LoadEventServiceConfig() *EventServiceConfig {
    return &EventServiceConfig{
        Port:                getEnv("EVENT_SERVICE_PORT", "8002"),
        DatabaseURL:         getEnvRequired("EVENT_SERVICE_DB_URL"),
        DatabaseReplicaURL:  getEnv("EVENT_SERVICE_DB_REPLICA_URL", ""),
        RedisURL:           getEnvRequired("REDIS_URL"),
        InternalAPIKey:     getEnvRequired("INTERNAL_API_KEY"),
        UserServiceURL:     getEnvRequired("USER_SERVICE_URL"),
        LogLevel:          getEnv("LOG_LEVEL", "info"),
    }
}

// Helper functions
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvRequired(key string) string {
    value := os.Getenv(key)
    if value == "" {
        panic("Required environment variable not set: " + key)
    }
    return value
}
```

### 7. SQLC Configuration (`sqlc/user-service/sqlc.yaml`)

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/*.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "users"
        out: "../../internal/repository/users"
        emit_json_tags: true
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
```

### 8. Makefile

```makefile
.PHONY: help build migrate sqlc test

# Help
help:
    @echo "Available commands:"
    @echo "  make build SERVICE=user-service    - Build specific service"
    @echo "  make build-all                      - Build all services"
    @echo "  make migrate-up SERVICE=user        - Run migrations for service"
    @echo "  make migrate-down SERVICE=user      - Rollback migrations"
    @echo "  make sqlc SERVICE=user              - Generate SQLC for service"
    @echo "  make test                           - Run all tests"
    @echo "  make run SERVICE=user-service       - Run specific service"

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
    @go build -o bin/user-service ./cmd/user-service
    @go build -o bin/event-service ./cmd/event-service
    @go build -o bin/search-service ./cmd/search-service
    @go build -o bin/booking-service ./cmd/booking-service
    @go build -o bin/analytics-service ./cmd/analytics-service

# Migration commands
migrate-up:
    @if [ -z "$(SERVICE)" ]; then \
        echo "Please specify SERVICE=<user|event|booking>"; \
        exit 1; \
    fi
    @echo "Running migrations for $(SERVICE)-service..."
    @source .env && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" up

migrate-down:
    @if [ -z "$(SERVICE)" ]; then \
        echo "Please specify SERVICE=<user|event|booking>"; \
        exit 1; \
    fi
    @echo "Rolling back migration for $(SERVICE)-service..."
    @source .env && goose -dir migrations/$(SERVICE)-service postgres "$${$(shell echo $(SERVICE) | tr '[:lower:]' '[:upper:]')_SERVICE_DB_URL}" down

migrate-all-up:
    @echo "Running all migrations..."
    @make migrate-up SERVICE=user
    @make migrate-up SERVICE=event
    @make migrate-up SERVICE=booking

# SQLC generation
sqlc:
    @if [ -z "$(SERVICE)" ]; then \
        echo "Please specify SERVICE=<user|event|booking>"; \
        exit 1; \
    fi
    @echo "Generating SQLC for $(SERVICE)-service..."
    @sqlc generate -f sqlc/$(SERVICE)-service/sqlc.yaml

sqlc-all:
    @echo "Generating all SQLC code..."
    @make sqlc SERVICE=user
    @make sqlc SERVICE=event
    @make sqlc SERVICE=booking

# Run commands
run:
    @if [ -z "$(SERVICE)" ]; then \
        echo "Please specify SERVICE=<service-name>"; \
        exit 1; \
    fi
    @echo "Running $(SERVICE)..."
    @source .env && go run ./cmd/$(SERVICE)/main.go

# Docker commands
docker-up:
    docker-compose up -d

docker-down:
    docker-compose down

docker-logs:
    docker-compose logs -f $(SERVICE)

# Development setup
dev-setup:
    @echo "Setting up development environment..."
    @cp .env.example .env
    @echo "Please update .env with your configuration"
    @make docker-up
    @sleep 5
    @make migrate-all-up
    @make sqlc-all
    @echo "Development environment ready!"

# Testing
test:
    @go test -v ./...

test-service:
    @if [ -z "$(SERVICE)" ]; then \
        echo "Please specify SERVICE=<user|event|booking|search|analytics>"; \
        exit 1; \
    fi
    @go test -v ./services/$(SERVICE)/...

# Clean
clean:
    @rm -rf bin/
    @echo "Cleaned build artifacts"
```

### 9. Migration Script (`scripts/migrate.sh`)

```bash
#!/bin/bash

# Load environment variables
source .env

# Function to run migrations for a specific service
migrate_service() {
    local service=$1
    local action=$2
    local db_url_var="${service^^}_SERVICE_DB_URL"
    local db_url=${!db_url_var}

    if [ -z "$db_url" ]; then
        echo "Error: Database URL not found for $service service"
        exit 1
    fi

    echo "Running $action migrations for $service service..."
    goose -dir "migrations/${service}-service" postgres "$db_url" $action
}

# Check arguments
if [ $# -ne 2 ]; then
    echo "Usage: ./migrate.sh <service|all> <up|down|status>"
    echo "Example: ./migrate.sh user up"
    echo "Example: ./migrate.sh all up"
    exit 1
fi

SERVICE=$1
ACTION=$2

# Validate action
if [[ ! "$ACTION" =~ ^(up|down|status)$ ]]; then
    echo "Invalid action. Use: up, down, or status"
    exit 1
fi

# Run migrations
if [ "$SERVICE" == "all" ]; then
    for svc in user event booking; do
        migrate_service $svc $ACTION
    done
else
    if [[ ! "$SERVICE" =~ ^(user|event|booking)$ ]]; then
        echo "Invalid service. Use: user, event, booking, or all"
        exit 1
    fi
    migrate_service $SERVICE $ACTION
fi

echo "Migration complete!"
```

## Key Benefits of This Structure

1. **Database Isolation with Shared Code:**
   - Each service connects only to its designated database
   - All query logic is centralized in `internal/repository/`
   - Type-safe database access through SQLC

2. **Service Independence:**
   - Each service can be built and deployed independently
   - Clear separation of concerns
   - Independent migration management

3. **Shared Utilities:**
   - JWT/Auth functions available to all services
   - Common database connection logic
   - Standardized error handling and responses

4. **Inter-Service Communication:**
   - Dedicated internal endpoints with authentication
   - Clear service boundaries
   - Service discovery through environment variables

5. **Development Workflow:**
   - Service-specific migrations
   - Independent build targets
   - Easy local development with docker-compose

This structure provides the isolation you need while maintaining code reusability and consistency across services.
