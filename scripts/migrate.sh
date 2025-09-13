#!/bin/bash

# Load environment variables
set -a
source .env
set +a

# Function to run migrations for a specific service
migrate_service() {
    local service=$1
    local action=$2
    local db_url_var="${service^^}_SERVICE_DB_URL"
    local db_url=${!db_url_var}

    if [ -z "$db_url" ]; then
        echo "Error: Database URL not found for $service service"
        echo "Expected environment variable: $db_url_var"
        exit 1
    fi

    echo "Running $action migrations for $service service..."
    echo "Database URL: $db_url"
    goose -dir "migrations/${service}-service" postgres "$db_url" $action
}

# Check arguments
if [ $# -ne 2 ]; then
    echo "Usage: ./migrate.sh <service|all> <up|down|status>"
    echo "Example: ./migrate.sh user up"
    echo "Example: ./migrate.sh all status"
    exit 1
fi

SERVICE=$1
ACTION=$2

# Validate action
if [[ ! "$ACTION" =~ ^(up|down|status|reset)$ ]]; then
    echo "Invalid action. Use: up, down, status, or reset"
    exit 1
fi

# Run migrations
if [ "$SERVICE" == "all" ]; then
    for svc in user; do  # Only user service for now
        echo "--- Processing $svc service ---"
        migrate_service $svc $ACTION
        echo ""
    done
else
    if [[ ! "$SERVICE" =~ ^(user)$ ]]; then  # Only user service for now
        echo "Invalid service. Currently supported: user"
        echo "Future services: event, booking"
        exit 1
    fi
    migrate_service $SERVICE $ACTION
fi

echo "Migration operation complete!"