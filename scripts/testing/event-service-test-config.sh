#!/bin/bash

# Event Service Test Configuration and Setup
# This script helps configure and run the Event Service test suite
# Author: Generated for bookmyevent-ily Event Service
# Usage: source ./event-service-test-config.sh

# =============================================================================
# CONFIGURATION VARIABLES
# =============================================================================

# Service Configuration
export EVENT_SERVICE_PORT=${EVENT_SERVICE_PORT:-"8002"}
export EVENT_SERVICE_URL=${EVENT_SERVICE_URL:-"http://localhost:$EVENT_SERVICE_PORT"}

# Database Configuration
export EVENT_SERVICE_DB_URL=${EVENT_SERVICE_DB_URL:-"postgres://postgres:password@localhost:5432/event_service_db?sslmode=disable"}
export EVENT_SERVICE_DB_REPLICA_URL=${EVENT_SERVICE_DB_REPLICA_URL:-""}

# JWT Configuration
export JWT_SECRET=${JWT_SECRET:-"your-super-secret-jwt-key-change-in-production-minimum-32-chars"}
export JWT_ACCESS_TOKEN_DURATION=${JWT_ACCESS_TOKEN_DURATION:-"15m"}
export JWT_REFRESH_TOKEN_DURATION=${JWT_REFRESH_TOKEN_DURATION:-"168h"}

# Internal API Configuration
export INTERNAL_API_KEY=${INTERNAL_API_KEY:-"your-internal-api-key-change-in-production"}

# External Service URLs
export USER_SERVICE_URL=${USER_SERVICE_URL:-"http://localhost:8001"}

# Logging Configuration
export LOG_LEVEL=${LOG_LEVEL:-"info"}
export ENVIRONMENT=${ENVIRONMENT:-"development"}

# Test Configuration
export TEST_TIMEOUT=${TEST_TIMEOUT:-"30"}
export TEST_RETRY_COUNT=${TEST_RETRY_COUNT:-"3"}
export TEST_PARALLEL_REQUESTS=${TEST_PARALLEL_REQUESTS:-"10"}
export TEST_CLEANUP=${TEST_CLEANUP:-"true"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if a service is running on a port
check_port() {
    local port=$1
    if command_exists nc; then
        nc -z localhost "$port" 2>/dev/null
    elif command_exists telnet; then
        timeout 1 telnet localhost "$port" 2>/dev/null | grep -q "Connected"
    else
        # Fallback: try to curl the port
        curl -s --connect-timeout 1 "localhost:$port" >/dev/null 2>&1
    fi
}

# Wait for a service to be ready
wait_for_service() {
    local url=$1
    local timeout=${2:-30}
    local counter=0

    log_info "Waiting for service at $url to be ready..."

    while [ $counter -lt $timeout ]; do
        if curl -s --fail "$url/healthz" >/dev/null 2>&1; then
            log_success "Service is ready!"
            return 0
        fi

        echo -n "."
        sleep 1
        counter=$((counter + 1))
    done

    echo
    log_error "Service did not become ready within $timeout seconds"
    return 1
}

# =============================================================================
# DEPENDENCY CHECKS
# =============================================================================

check_dependencies() {
    log_step "Checking dependencies..."

    local missing_deps=()

    # Required dependencies
    if ! command_exists curl; then
        missing_deps+=("curl")
    fi

    if ! command_exists go; then
        missing_deps+=("go")
    fi

    if ! command_exists docker; then
        missing_deps+=("docker")
    fi

    # Optional but recommended dependencies
    if ! command_exists jq; then
        log_warning "jq not found - JSON responses will be less readable"
    fi

    if ! command_exists bc; then
        log_warning "bc not found - performance calculations may not work"
    fi

    if ! command_exists nc && ! command_exists telnet; then
        log_warning "Neither nc nor telnet found - port checking will be limited"
    fi

    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_info "Please install the missing dependencies and try again."
        return 1
    fi

    log_success "All required dependencies are available"
    return 0
}

# =============================================================================
# DATABASE SETUP
# =============================================================================

setup_database() {
    log_step "Setting up database..."

    # Check if PostgreSQL is running
    if ! check_port 5432; then
        log_info "Starting PostgreSQL with Docker..."
        if ! docker compose up -d postgres; then
            log_error "Failed to start PostgreSQL"
            return 1
        fi

        # Wait for PostgreSQL to be ready
        log_info "Waiting for PostgreSQL to be ready..."
        sleep 10
    else
        log_info "PostgreSQL is already running"
    fi

    # Run migrations
    log_info "Running database migrations..."
    if ! make migrate-up SERVICE=event; then
        log_error "Failed to run database migrations"
        return 1
    fi

    log_success "Database setup completed"
    return 0
}

# =============================================================================
# SERVICE MANAGEMENT
# =============================================================================

build_service() {
    log_step "Building Event Service..."

    if ! make build SERVICE=event-service; then
        log_error "Failed to build Event Service"
        return 1
    fi

    log_success "Event Service built successfully"
    return 0
}

start_service() {
    log_step "Starting Event Service..."

    # Check if service is already running
    if check_port "$EVENT_SERVICE_PORT"; then
        log_warning "Service is already running on port $EVENT_SERVICE_PORT"
        return 0
    fi

    # Start the service in background
    nohup make run SERVICE=event-service > event-service.log 2>&1 &
    local service_pid=$!

    # Wait for service to be ready
    if wait_for_service "$EVENT_SERVICE_URL" 30; then
        log_success "Event Service started successfully (PID: $service_pid)"
        echo "$service_pid" > event-service.pid
        return 0
    else
        log_error "Failed to start Event Service"
        return 1
    fi
}

stop_service() {
    log_step "Stopping Event Service..."

    if [ -f event-service.pid ]; then
        local pid=$(cat event-service.pid)
        if kill "$pid" 2>/dev/null; then
            log_success "Event Service stopped (PID: $pid)"
        else
            log_warning "Failed to stop service with PID $pid (may already be stopped)"
        fi
        rm -f event-service.pid
    else
        log_warning "No PID file found - service may not be running"
    fi

    # Also try to kill any remaining processes
    pkill -f "event-service" 2>/dev/null || true
}

restart_service() {
    stop_service
    sleep 2
    start_service
}

# =============================================================================
# TEST EXECUTION FUNCTIONS
# =============================================================================

run_quick_tests() {
    log_step "Running quick smoke tests..."

    # Just run health checks and basic authentication
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    "$script_dir/event-service-test.sh" "$EVENT_SERVICE_URL" "$INTERNAL_API_KEY" | grep -E "(PASSED|FAILED|ERROR|SUCCESS|WARNING)" | head -20
}

run_full_tests() {
    log_step "Running comprehensive test suite..."

    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    "$script_dir/event-service-test.sh" "$EVENT_SERVICE_URL" "$INTERNAL_API_KEY"
}

run_performance_tests() {
    log_step "Running performance tests..."

    # Run multiple concurrent requests to test performance
    log_info "Testing concurrent load..."

    local start_time=$(date +%s)
    for i in $(seq 1 "$TEST_PARALLEL_REQUESTS"); do
        curl -s "$EVENT_SERVICE_URL/healthz" > /dev/null &
    done
    wait
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    log_info "Completed $TEST_PARALLEL_REQUESTS concurrent requests in ${duration}s"

    # Test individual endpoint performance
    log_info "Testing individual endpoint performance..."
    time curl -s "$EVENT_SERVICE_URL/api/v1/events" > /dev/null
}

# =============================================================================
# ENVIRONMENT SETUP
# =============================================================================

setup_test_environment() {
    log_step "Setting up test environment..."

    # Create necessary directories
    mkdir -p logs
    mkdir -p test-results

    # Export all environment variables
    export EVENT_SERVICE_PORT EVENT_SERVICE_URL EVENT_SERVICE_DB_URL
    export JWT_SECRET JWT_ACCESS_TOKEN_DURATION JWT_REFRESH_TOKEN_DURATION
    export INTERNAL_API_KEY USER_SERVICE_URL
    export LOG_LEVEL ENVIRONMENT

    log_success "Test environment configured"

    # Print configuration summary
    echo
    log_info "Configuration Summary:"
    echo "  Event Service URL: $EVENT_SERVICE_URL"
    echo "  Database URL: ${EVENT_SERVICE_DB_URL%\?*}... (credentials hidden)"
    echo "  JWT Secret: ${JWT_SECRET:0:10}... (truncated)"
    echo "  Internal API Key: ${INTERNAL_API_KEY:0:10}... (truncated)"
    echo "  Log Level: $LOG_LEVEL"
    echo "  Environment: $ENVIRONMENT"
    echo
}

# =============================================================================
# MAIN SETUP FUNCTION
# =============================================================================

setup_all() {
    log_step "Setting up complete Event Service test environment..."

    # Check dependencies
    if ! check_dependencies; then
        return 1
    fi

    # Setup environment
    setup_test_environment

    # Setup database
    if ! setup_database; then
        return 1
    fi

    # Build service
    if ! build_service; then
        return 1
    fi

    # Start service
    if ! start_service; then
        return 1
    fi

    log_success "Complete setup finished successfully!"
    log_info "You can now run tests with: run_full_tests"
    return 0
}

# =============================================================================
# CLEANUP FUNCTIONS
# =============================================================================

cleanup_test_data() {
    log_step "Cleaning up test data..."

    # This would typically connect to the database and clean up test records
    # For now, we'll just log what should be done
    log_info "Test data cleanup would remove:"
    log_info "  - Test admin accounts"
    log_info "  - Test venues"
    log_info "  - Test events"
    log_info "  - Test tokens"

    log_warning "Automated cleanup not implemented - manual cleanup may be required"
}

cleanup_all() {
    log_step "Performing complete cleanup..."

    stop_service
    cleanup_test_data

    # Clean up log files
    rm -f event-service.log
    rm -f event-service-test-results-*.log

    log_success "Cleanup completed"
}

# =============================================================================
# HELP AND USAGE
# =============================================================================

show_help() {
    echo "Event Service Test Configuration and Setup"
    echo
    echo "This script provides functions to set up and run tests for the Event Service."
    echo
    echo "Usage:"
    echo "  source ./event-service-test-config.sh"
    echo "  setup_all                    # Complete setup"
    echo "  run_full_tests               # Run all tests"
    echo
    echo "Available Functions:"
    echo
    echo "Setup Functions:"
    echo "  check_dependencies           # Check required tools"
    echo "  setup_database              # Set up PostgreSQL and run migrations"
    echo "  build_service               # Build the Event Service"
    echo "  setup_test_environment      # Configure environment variables"
    echo "  setup_all                   # Complete setup (recommended)"
    echo
    echo "Service Management:"
    echo "  start_service               # Start the Event Service"
    echo "  stop_service                # Stop the Event Service"
    echo "  restart_service             # Restart the Event Service"
    echo
    echo "Test Execution:"
    echo "  run_quick_tests             # Run smoke tests only"
    echo "  run_full_tests              # Run comprehensive test suite"
    echo "  run_performance_tests       # Run performance tests"
    echo
    echo "Cleanup:"
    echo "  cleanup_test_data           # Clean up test data"
    echo "  cleanup_all                 # Complete cleanup"
    echo
    echo "Environment Variables:"
    echo "  EVENT_SERVICE_PORT          # Service port (default: 8002)"
    echo "  EVENT_SERVICE_URL           # Service URL"
    echo "  EVENT_SERVICE_DB_URL        # Database connection string"
    echo "  JWT_SECRET                  # JWT signing secret"
    echo "  INTERNAL_API_KEY            # Internal API key"
    echo "  LOG_LEVEL                   # Log level (default: info)"
    echo
    echo "Examples:"
    echo "  # Quick setup and test"
    echo "  source ./event-service-test-config.sh"
    echo "  setup_all && run_quick_tests"
    echo
    echo "  # Full test suite"
    echo "  source ./event-service-test-config.sh"
    echo "  setup_all && run_full_tests"
    echo
    echo "  # Custom configuration"
    echo "  export EVENT_SERVICE_PORT=8003"
    echo "  export JWT_SECRET='my-custom-secret'"
    echo "  source ./event-service-test-config.sh"
    echo "  setup_all"
    echo
}

# =============================================================================
# INITIALIZATION
# =============================================================================

# Show help if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    show_help
    exit 0
fi

# If sourced, show a brief welcome message
log_info "Event Service test configuration loaded"
log_info "Run 'show_help' for usage information"
log_info "Run 'setup_all' to set up the complete test environment"
