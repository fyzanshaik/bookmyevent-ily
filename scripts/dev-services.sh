#!/bin/bash

# Development Services Management Script
# This script helps manage all services for the bookmyevent-ily project
# Usage: ./dev-services.sh [command]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICES=("user-service" "event-service" "booking-service")
SERVICE_PORTS=("8001" "8002" "8004")

log() {
    echo -e "${1}"
}

# Function to kill all services
kill_services() {
    log "${YELLOW}Stopping all services...${NC}"
    
    # Kill by process name
    pkill -f "user-service|event-service|booking-service" 2>/dev/null || true
    
    # Kill by port
    for port in "${SERVICE_PORTS[@]}"; do
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
    done
    
    # Wait a moment for processes to die
    sleep 2
    
    log "${GREEN}✓ All services stopped${NC}"
}

# Function to stop infrastructure
stop_infrastructure() {
    log "${YELLOW}Stopping infrastructure (PostgreSQL, Redis)...${NC}"
    cd "$PROJECT_ROOT"
    make docker-down
    log "${GREEN}✓ Infrastructure stopped${NC}"
}

# Function to start infrastructure
start_infrastructure() {
    log "${BLUE}Starting infrastructure (PostgreSQL + Redis)...${NC}"
    cd "$PROJECT_ROOT"
    make docker-full-up
    log "${GREEN}✓ Infrastructure started${NC}"
}

# Function to run migrations
run_migrations() {
    log "${BLUE}Running migrations...${NC}"
    cd "$PROJECT_ROOT"
    
    for service in "user" "event" "booking"; do
        log "  Running ${service} migrations..."
        make migrate-up SERVICE=$service
    done
    
    log "${GREEN}✓ All migrations completed${NC}"
}

# Function to build all services
build_services() {
    log "${BLUE}Building all services...${NC}"
    cd "$PROJECT_ROOT"
    make build-all
    log "${GREEN}✓ All services built${NC}"
}

# Function to start a single service with env
start_service() {
    local service_name=$1
    local log_prefix="[${service_name}]"
    
    log "${BLUE}Starting ${service_name}...${NC}"
    cd "$PROJECT_ROOT"
    
    # Export environment and run service in background
    export $(cat .env | grep -v '^#' | xargs)
    ./bin/${service_name} &
    
    local service_pid=$!
    
    # Wait a moment for service to start
    sleep 3
    
    # Check if service is still running
    if kill -0 $service_pid 2>/dev/null; then
        log "${GREEN}✓ ${service_name} started successfully (PID: $service_pid)${NC}"
    else
        log "${RED}✗ ${service_name} failed to start${NC}"
        return 1
    fi
}

# Function to start all services
start_all_services() {
    log "${BLUE}Starting all services...${NC}"
    cd "$PROJECT_ROOT"
    
    # Start services in dependency order: user -> event -> booking
    for service in "${SERVICES[@]}"; do
        start_service "$service"
        sleep 2  # Give each service time to fully start before starting the next
    done
    
    log "${GREEN}✓ All services started${NC}"
}

# Function to check service health
check_service_health() {
    local service_name=$1
    local port=$2
    local health_endpoint="http://localhost:${port}/healthz"
    
    log "  Checking ${service_name} health..."
    
    local response=$(curl -s "$health_endpoint" 2>/dev/null || echo "ERROR")
    if echo "$response" | grep -q '"status".*"healthy"'; then
        log "${GREEN}    ✓ ${service_name} is healthy${NC}"
        return 0
    else
        log "${RED}    ✗ ${service_name} is not responding${NC}"
        return 1
    fi
}

# Function to check all services health
check_all_health() {
    log "${BLUE}Checking all services health...${NC}"
    
    local all_healthy=true
    for i in "${!SERVICES[@]}"; do
        if ! check_service_health "${SERVICES[$i]}" "${SERVICE_PORTS[$i]}"; then
            all_healthy=false
        fi
    done
    
    if $all_healthy; then
        log "${GREEN}✓ All services are healthy${NC}"
        return 0
    else
        log "${RED}✗ Some services are not healthy${NC}"
        return 1
    fi
}

# Function to show service status
show_status() {
    log "${CYAN}Service Status:${NC}"
    echo
    
    # Check infrastructure
    log "${BLUE}Infrastructure:${NC}"
    if docker ps | grep -q "evently_postgres"; then
        log "${GREEN}  ✓ PostgreSQL: Running${NC}"
    else
        log "${RED}  ✗ PostgreSQL: Not running${NC}"
    fi
    
    if docker ps | grep -q "evently_redis"; then
        log "${GREEN}  ✓ Redis: Running${NC}"
    else
        log "${RED}  ✗ Redis: Not running${NC}"
    fi
    
    echo
    
    # Check services
    log "${BLUE}Services:${NC}"
    for i in "${!SERVICES[@]}"; do
        local service="${SERVICES[$i]}"
        local port="${SERVICE_PORTS[$i]}"
        
        if lsof -i:$port >/dev/null 2>&1; then
            log "${GREEN}  ✓ ${service}: Running on port ${port}${NC}"
        else
            log "${RED}  ✗ ${service}: Not running${NC}"
        fi
    done
    
    echo
}

# Function to run booking service tests
run_booking_tests() {
    log "${PURPLE}Running booking service tests...${NC}"
    
    # Check if booking service test script exists
    local test_script="$PROJECT_ROOT/scripts/testing/booking-service-test.sh"
    if [ ! -f "$test_script" ]; then
        log "${RED}✗ Booking service test script not found at: $test_script${NC}"
        return 1
    fi
    
    # Make sure it's executable
    chmod +x "$test_script"
    
    # Run the tests
    cd "$PROJECT_ROOT"
    "$test_script"
}

# Function to setup everything and run tests
full_setup_and_test() {
    log "${CYAN}========================================${NC}"
    log "${CYAN}Full Setup and Test Execution${NC}"
    log "${CYAN}========================================${NC}"
    echo
    
    # 1. Stop everything
    kill_services
    stop_infrastructure
    
    # 2. Start infrastructure
    start_infrastructure
    
    # 3. Run migrations
    run_migrations
    
    # 4. Build services
    build_services
    
    # 5. Start all services
    start_all_services
    
    # 6. Wait for services to stabilize
    log "${BLUE}Waiting for services to stabilize...${NC}"
    sleep 10
    
    # 7. Check health
    if check_all_health; then
        log "${GREEN}✓ All services are ready${NC}"
    else
        log "${RED}✗ Some services are not ready - tests may fail${NC}"
    fi
    
    echo
    
    # 8. Run tests
    run_booking_tests
}

# Function to quick start (assumes infrastructure is already running)
quick_start() {
    log "${CYAN}Quick Start (assuming infrastructure is running)${NC}"
    echo
    
    # Kill existing services
    kill_services
    
    # Build and start services
    build_services
    start_all_services
    
    # Wait and check health
    sleep 5
    check_all_health
}

# Function to show help
show_help() {
    echo "Development Services Management Script"
    echo
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  start-infra     - Start PostgreSQL and Redis containers"
    echo "  stop-infra      - Stop PostgreSQL and Redis containers"
    echo "  migrate         - Run database migrations for all services"
    echo "  build           - Build all services"
    echo "  start-services  - Start all services (user, event, booking)"
    echo "  stop-services   - Stop all running services"
    echo "  status          - Show status of infrastructure and services"
    echo "  health          - Check health of all services"
    echo "  test            - Run booking service tests"
    echo "  full-setup      - Complete setup: infra + migrate + build + start + test"
    echo "  quick-start     - Quick start services (assumes infra is running)"
    echo "  restart         - Stop and start all services"
    echo "  clean           - Stop everything (services + infrastructure)"
    echo "  help            - Show this help message"
    echo
    echo "Examples:"
    echo "  $0 full-setup   # Complete setup and test"
    echo "  $0 quick-start  # Just restart services"
    echo "  $0 status       # Check what's running"
    echo "  $0 clean        # Stop everything"
}

# Main script logic
main() {
    local command=${1:-"help"}
    
    case $command in
        "start-infra")
            start_infrastructure
            ;;
        "stop-infra")
            stop_infrastructure
            ;;
        "migrate")
            run_migrations
            ;;
        "build")
            build_services
            ;;
        "start-services")
            start_all_services
            ;;
        "stop-services")
            kill_services
            ;;
        "status")
            show_status
            ;;
        "health")
            check_all_health
            ;;
        "test")
            run_booking_tests
            ;;
        "full-setup")
            full_setup_and_test
            ;;
        "quick-start")
            quick_start
            ;;
        "restart")
            kill_services
            sleep 2
            start_all_services
            ;;
        "clean")
            kill_services
            stop_infrastructure
            log "${GREEN}✓ Everything stopped${NC}"
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log "${RED}Unknown command: $command${NC}"
            echo
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"