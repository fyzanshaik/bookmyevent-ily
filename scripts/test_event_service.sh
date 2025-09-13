#!/bin/bash

# Comprehensive Test Script for Event Service

# Exit on any error
set -e

# --- Configuration ---
BASE_URL="http://localhost:8002"
INTERNAL_API_KEY="internal-service-communication-key-change-in-production"
ADMIN_EMAIL="test-admin-$(date +%s)@example.com"
ADMIN_PASSWORD="password123"

# --- Helper Functions ---
log() {
    echo "[INFO] $1"
}

run_test() {
    local test_name=$1
    local command=$2
    log "Running test: $test_name"
    eval $command
    log "Test PASSED: $test_name"
}

# --- Pre-flight Checks ---
check_jq() {
    if ! command -v jq &> /dev/null
    then
        echo "jq could not be found. Please install jq to run this script."
        exit 1
    fi
}

# --- Test Cases ---

# 1. Health Check
test_health_check() {
    # Retry mechanism for health check
    for i in {1..5}; do
        if curl -sS --fail "$BASE_URL/healthz" | grep "healthy"; then
            return
        fi
        sleep 1
    done
    echo "Health check failed after 5 attempts."
    exit 1
}

# 2. Admin Registration
test_admin_registration() {
    REGISTRATION_RESPONSE=$(curl -sS -X POST "$BASE_URL/api/v1/auth/admin/register" \
        -H "Content-Type: application/json" \
        -d '{ \
            "email": "'$ADMIN_EMAIL'", \
            "password": "'$ADMIN_PASSWORD'", \
            "name": "Test Admin" \
        }')
    
    ADMIN_TOKEN=$(echo "$REGISTRATION_RESPONSE" | jq -r '.access_token')
    if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" == "null" ]; then
        echo "Admin registration failed or token not found."
        exit 1
    fi
}

# 3. Admin Login
test_admin_login() {
    LOGIN_RESPONSE=$(curl -sS -X POST "$BASE_URL/api/v1/auth/admin/login" \
        -H "Content-Type: application/json" \
        -d '{ \
            "email": "'$ADMIN_EMAIL'", \
            "password": "'$ADMIN_PASSWORD'" \
        }')

    ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
    if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" == "null" ]; then
        echo "Admin login failed or token not found."
        exit 1
    fi
}

# 4. Create Venue
test_create_venue() {
    VENUE_RESPONSE=$(curl -sS -X POST "$BASE_URL/api/v1/admin/venues" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{ \
            "name": "Test Venue", \
            "address": "123 Test St", \
            "city": "Testville", \
            "state": "TS", \
            "country": "USA", \
            "postal_code": "12345", \
            "capacity": 500 \
        }')
    
    VENUE_ID=$(echo "$VENUE_RESPONSE" | jq -r '.venue_id')
    if [ -z "$VENUE_ID" ] || [ "$VENUE_ID" == "null" ]; then
        echo "Create venue failed."
        exit 1
    fi
}

# 5. List Venues
test_list_venues() {
    curl -sS --fail "$BASE_URL/api/v1/admin/venues" \
        -H "Authorization: Bearer $ADMIN_TOKEN" | jq -e '.venues | length > 0'
}

# 6. Update Venue
test_update_venue() {
    curl -sS -X PUT "$BASE_URL/api/v1/admin/venues/$VENUE_ID" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{ \
            "name": "Updated Test Venue" \
        }' | jq -e '.name == "Updated Test Venue"'
}

# 7. Create Event
test_create_event() {
    EVENT_RESPONSE=$(curl -sS -X POST "$BASE_URL/api/v1/admin/events" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{ \
            "name": "Test Event", \
            "description": "A great event", \
            "venue_id": "'$VENUE_ID'", \
            "event_type": "concert", \
            "start_datetime": "2025-12-01T20:00:00Z", \
            "end_datetime": "2025-12-01T23:00:00Z", \
            "total_capacity": 100, \
            "base_price": 50.00 \
        }')

    EVENT_ID=$(echo "$EVENT_RESPONSE" | jq -r '.event_id')
    EVENT_VERSION=$(echo "$EVENT_RESPONSE" | jq -r '.version')
    if [ -z "$EVENT_ID" ] || [ "$EVENT_ID" == "null" ]; then
        echo "Create event failed."
        exit 1
    fi
}

# 8. List Admin Events
test_list_admin_events() {
    curl -sS --fail "$BASE_URL/api/v1/admin/events" \
        -H "Authorization: Bearer $ADMIN_TOKEN" | jq -e '.events | length > 0'
}

# 9. Update Event
test_update_event() {
    UPDATE_RESPONSE=$(curl -sS -X PUT "$BASE_URL/api/v1/admin/events/$EVENT_ID" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{ \
            "name": "Updated Test Event", \
            "version": '$EVENT_VERSION'
        }')
    
    echo "$UPDATE_RESPONSE" | jq -e '.name == "Updated Test Event"'
    EVENT_VERSION=$(echo "$UPDATE_RESPONSE" | jq -r '.version')
}

# 10. Concurrency Test (Optimistic Locking)
test_concurrency() {
    # Try to update with the old version, expecting failure
    curl -sS -X PUT "$BASE_URL/api/v1/admin/events/$EVENT_ID" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{ \
            "name": "Concurrency Fail Test", \
            "version": 1 
        }' | jq -e '.error.code == "CONFLICT"'
}

# 11. List Public Events
test_list_public_events() {
    # First publish the event
    curl -sS -X PUT "$BASE_URL/api/v1/admin/events/$EVENT_ID" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{ \
            "status": "published", \
            "version": '$EVENT_VERSION'
        }' > /dev/null
    
    curl -sS --fail "$BASE_URL/api/v1/events" | jq -e '.events | length > 0'
}

# 12. Get Public Event by ID
test_get_public_event_by_id() {
    curl -sS --fail "$BASE_URL/api/v1/events/$EVENT_ID" | jq -e '.event_id == "'$EVENT_ID'"'
}

# 13. Internal API - Update Availability
test_internal_update_availability() {
    GET_EVENT_RESPONSE=$(curl -sS -X GET "$BASE_URL/internal/events/$EVENT_ID" -H "X-API-Key: $INTERNAL_API_KEY")
    INTERNAL_EVENT_VERSION=$(echo "$GET_EVENT_RESPONSE" | jq -r '.version')

    curl -sS -X POST "$BASE_URL/internal/events/$EVENT_ID/update-availability" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $INTERNAL_API_KEY" \
        -d '{ \
            "quantity": -10, \
            "version": '$INTERNAL_EVENT_VERSION'
        }') | jq -e '.available_seats == 90'
}

# 14. Delete Event
test_delete_event() {
    GET_EVENT_RESPONSE=$(curl -sS -X GET "$BASE_URL/api/v1/events/$EVENT_ID" -H "Authorization: Bearer $ADMIN_TOKEN")
    DELETE_EVENT_VERSION=$(echo "$GET_EVENT_RESPONSE" | jq -r '.version')

    curl -sS -X DELETE "$BASE_URL/api/v1/admin/events/$EVENT_ID" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{ \
            "version": '$DELETE_EVENT_VERSION'
        }') | jq -e '.message == "Event cancelled successfully"'
}

# 15. Delete Venue
test_delete_venue() {
    curl -sS -X DELETE "$BASE_URL/api/v1/admin/venues/$VENUE_ID" \
        -H "Authorization: Bearer $ADMIN_TOKEN" | jq -e '.message == "Venue deleted successfully"'
}


# --- Main Execution ---
main() {
    check_jq
    run_test "Health Check" test_health_check
    run_test "Admin Registration" test_admin_registration
    run_test "Admin Login" test_admin_login
    run_test "Create Venue" test_create_venue
    run_test "List Venues" test_list_venues
    run_test "Update Venue" test_update_venue
    run_test "Create Event" test_create_event
    run_test "List Admin Events" test_list_admin_events
    run_test "Update Event" test_update_event
    run_test "Concurrency (Optimistic Locking)" test_concurrency
    run_test "List Public Events" test_list_public_events
    run_test "Get Public Event by ID" test_get_public_event_by_id
    run_test "Internal API - Update Availability" test_internal_update_availability
    run_test "Delete Event" test_delete_event
    run_test "Delete Venue" test_delete_venue
    
    log "All tests completed successfully!"
}

main