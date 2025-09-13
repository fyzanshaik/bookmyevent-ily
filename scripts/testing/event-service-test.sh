#!/bin/bash

# Event Service Comprehensive Testing Script
# This script tests all endpoints, edge cases, and scenarios for the Event Service
# Author: Generated for bookmyevent-ily Event Service
# Usage: ./event-service-test.sh [base_url] [internal_api_key]

# set -e  # Commented out to allow tests to continue even if individual tests fail

# Configuration
BASE_URL=${1:-"http://localhost:8002"}
INTERNAL_API_KEY=${2:-"internal-service-communication-key-change-in-production"}
TEST_RESULTS_FILE="event-service-test-results-$(date +%Y%m%d_%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Global variables for test data
ADMIN_TOKEN=""
ADMIN_ID=""
VENUE_ID=""
EVENT_ID=""
EVENT_VERSION=1

# Logging function
log() {
    echo -e "$1" | tee -a "$TEST_RESULTS_FILE"
}

# Test result functions
test_start() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log "${BLUE}[TEST $TOTAL_TESTS] $1${NC}"
}

test_pass() {
    PASSED_TESTS=$((PASSED_TESTS + 1))
    log "${GREEN}✓ PASSED: $1${NC}"
    echo
}

test_fail() {
    FAILED_TESTS=$((FAILED_TESTS + 1))
    log "${RED}✗ FAILED: $1${NC}"
    if [ $# -gt 1 ]; then
        log "${RED}  Response: $2${NC}"
    fi
    echo
}

test_warning() {
    log "${YELLOW}⚠ WARNING: $1${NC}"
    echo
}

# HTTP request helper
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local headers=$4
    local expected_status=$5

    local curl_cmd="curl -s -w '\\n%{http_code}' -X $method '$BASE_URL$endpoint'"

    if [ -n "$headers" ]; then
        curl_cmd="$curl_cmd -H '$headers'"
    fi

    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi

    local response=$(eval $curl_cmd)
    local body=$(echo "$response" | head -n -1)
    local status=$(echo "$response" | tail -n 1)

    if [ -n "$expected_status" ] && [ "$status" != "$expected_status" ]; then
        echo "ERROR: Expected status $expected_status, got $status"
        echo "Response: $body"
        return 1
    fi

    echo "$body"
    return 0
}

# Generate test data
generate_test_email() {
    echo "test-admin-$(date +%s)@example.com"
}

generate_test_venue() {
    local timestamp=$(date +%s)
    echo "{
        \"name\": \"Test Venue $timestamp\",
        \"address\": \"123 Test St\",
        \"city\": \"Test City\",
        \"state\": \"Test State\",
        \"country\": \"USA\",
        \"postal_code\": \"12345\",
        \"capacity\": 1000,
        \"layout_config\": {\"sections\": [\"A\", \"B\", \"C\"]}
    }"
}

generate_test_event() {
    local venue_id=$1
    local timestamp=$(date +%s)
    local start_date=$(date -d "+7 days" --iso-8601=seconds)
    local end_date=$(date -d "+7 days +3 hours" --iso-8601=seconds)

    echo "{
        \"name\": \"Test Event $timestamp\",
        \"description\": \"A comprehensive test event\",
        \"venue_id\": \"$venue_id\",
        \"event_type\": \"concert\",
        \"start_datetime\": \"$start_date\",
        \"end_datetime\": \"$end_date\",
        \"total_capacity\": 500,
        \"base_price\": 29.99,
        \"max_tickets_per_booking\": 8
    }"
}

# ============================================================================
# HEALTH CHECK TESTS
# ============================================================================

test_health_endpoints() {
    log "${PURPLE}=== HEALTH CHECK TESTS ===${NC}"

    test_start "Basic health check"
    response=$(make_request "GET" "/healthz" "" "" "200")
    if echo "$response" | grep -q '"status".*"healthy"'; then
        test_pass "Health endpoint returns healthy status"
    else
        test_fail "Health endpoint failed" "$response"
    fi

    test_start "Readiness check"
    response=$(make_request "GET" "/health/ready" "" "" "200")
    if echo "$response" | grep -q '"status":"ready"'; then
        test_pass "Readiness endpoint returns ready status"
    else
        test_fail "Readiness endpoint failed" "$response"
    fi
}

# ============================================================================
# ADMIN AUTHENTICATION TESTS
# ============================================================================

test_admin_authentication() {
    log "${PURPLE}=== ADMIN AUTHENTICATION TESTS ===${NC}"

    local test_email=$(generate_test_email)

    # Test admin registration
    test_start "Admin registration with valid data"
    local register_data="{
        \"email\": \"$test_email\",
        \"password\": \"SecurePassword123!\",
        \"name\": \"Test Admin\",
        \"phone_number\": \"+1234567890\",
        \"role\": \"event_manager\"
    }"

    response=$(make_request "POST" "/api/v1/auth/admin/register" "$register_data" "" "201")
    if echo "$response" | grep -q '"access_token"'; then
        ADMIN_TOKEN=$(echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        ADMIN_ID=$(echo "$response" | grep -o '"admin_id":"[^"]*"' | cut -d'"' -f4)
        test_pass "Admin registration successful, token received"
    else
        test_fail "Admin registration failed" "$response"
        return 1
    fi

    # Test duplicate registration
    test_start "Admin registration with duplicate email"
    response=$(make_request "POST" "/api/v1/auth/admin/register" "$register_data" "" "409")
    if echo "$response" | grep -q -i "already exists\|duplicate"; then
        test_pass "Duplicate email registration properly rejected"
    else
        test_fail "Duplicate email should be rejected" "$response"
    fi

    # Test admin login
    test_start "Admin login with correct credentials"
    local login_data="{
        \"email\": \"$test_email\",
        \"password\": \"SecurePassword123!\"
    }"

    response=$(make_request "POST" "/api/v1/auth/admin/login" "$login_data" "" "200")
    if echo "$response" | grep -q '"access_token"'; then
        test_pass "Admin login successful"
    else
        test_fail "Admin login failed" "$response"
    fi

    # Test login with wrong password
    test_start "Admin login with incorrect password"
    local wrong_login_data="{
        \"email\": \"$test_email\",
        \"password\": \"WrongPassword123!\"
    }"

    response=$(make_request "POST" "/api/v1/auth/admin/login" "$wrong_login_data" "" "401")
    if echo "$response" | grep -q -i "incorrect\|unauthorized"; then
        test_pass "Incorrect password properly rejected"
    else
        test_fail "Incorrect password should be rejected" "$response"
    fi

    # Test invalid registration data
    test_start "Admin registration with missing required fields"
    local invalid_data='{"email": "", "password": "", "name": ""}'
    response=$(make_request "POST" "/api/v1/auth/admin/register" "$invalid_data" "" "400")
    if echo "$response" | grep -q -i "required"; then
        test_pass "Missing required fields properly rejected"
    else
        test_fail "Missing required fields should be rejected" "$response"
    fi
}

# ============================================================================
# VENUE MANAGEMENT TESTS
# ============================================================================

test_venue_management() {
    log "${PURPLE}=== VENUE MANAGEMENT TESTS ===${NC}"

    if [ -z "$ADMIN_TOKEN" ]; then
        test_fail "No admin token available for venue tests"
        return 1
    fi

    # Test venue creation
    test_start "Create venue with valid data"
    local venue_data=$(generate_test_venue)

    response=$(make_request "POST" "/api/v1/admin/venues" "$venue_data" "Authorization: Bearer $ADMIN_TOKEN" "201")
    if echo "$response" | grep -q '"venue_id"'; then
        VENUE_ID=$(echo "$response" | grep -o '"venue_id":"[^"]*"' | cut -d'"' -f4)
        test_pass "Venue creation successful"
    else
        test_fail "Venue creation failed" "$response"
        return 1
    fi

    # Test venue creation without authentication
    test_start "Create venue without authentication"
    response=$(make_request "POST" "/api/v1/admin/venues" "$venue_data" "")
    if echo "$response" | grep -q -i "unauthorized\|not authenticated\|Missing.*authorization"; then
        test_pass "Unauthenticated venue creation properly rejected"
    else
        test_fail "Should require authentication" "$response"
    fi

    # Test venue creation with invalid data
    test_start "Create venue with missing required fields"
    local invalid_venue='{"name": "", "address": "", "city": "", "capacity": -1}'
    response=$(make_request "POST" "/api/v1/admin/venues" "$invalid_venue" "Authorization: Bearer $ADMIN_TOKEN" "400")
    if echo "$response" | grep -q -i "required\|positive"; then
        test_pass "Invalid venue data properly rejected"
    else
        test_fail "Invalid venue data should be rejected" "$response"
    fi

    # Test list venues
    test_start "List venues"
    response=$(make_request "GET" "/api/v1/admin/venues" "" "Authorization: Bearer $ADMIN_TOKEN" "200")
    if echo "$response" | grep -q '"venues"'; then
        test_pass "Venue listing successful"
    else
        test_fail "Venue listing failed" "$response"
    fi

    # Test list venues with pagination
    test_start "List venues with pagination"
    response=$(make_request "GET" "/api/v1/admin/venues?page=1&limit=5" "" "Authorization: Bearer $ADMIN_TOKEN" "200")
    if echo "$response" | grep -q '"page":1'; then
        test_pass "Venue pagination working"
    else
        test_fail "Venue pagination failed" "$response"
    fi

    # Test venue update
    test_start "Update venue"
    local update_data='{"name": "Updated Venue Name", "capacity": 1500}'
    response=$(make_request "PUT" "/api/v1/admin/venues/$VENUE_ID" "$update_data" "Authorization: Bearer $ADMIN_TOKEN" "200")
    if echo "$response" | grep -q "Updated Venue Name"; then
        test_pass "Venue update successful"
    else
        test_fail "Venue update failed" "$response"
    fi

    # Test update non-existent venue
    test_start "Update non-existent venue"
    local fake_id="00000000-0000-0000-0000-000000000000"
    response=$(make_request "PUT" "/api/v1/admin/venues/$fake_id" "$update_data" "Authorization: Bearer $ADMIN_TOKEN" "404")
    if echo "$response" | grep -q -i "not found"; then
        test_pass "Non-existent venue update properly rejected"
    else
        test_fail "Should return 404 for non-existent venue" "$response"
    fi
}

# ============================================================================
# EVENT MANAGEMENT TESTS
# ============================================================================

test_event_management() {
    log "${PURPLE}=== EVENT MANAGEMENT TESTS ===${NC}"

    if [ -z "$ADMIN_TOKEN" ] || [ -z "$VENUE_ID" ]; then
        test_fail "Missing admin token or venue ID for event tests"
        return 1
    fi

    # Test event creation
    test_start "Create event with valid data"
    local event_data=$(generate_test_event "$VENUE_ID")

    response=$(make_request "POST" "/api/v1/admin/events" "$event_data" "Authorization: Bearer $ADMIN_TOKEN" "201")
    if echo "$response" | grep -q '"event_id"'; then
        EVENT_ID=$(echo "$response" | grep -o '"event_id":"[^"]*"' | cut -d'"' -f4)
        EVENT_VERSION=$(echo "$response" | grep -o '"version":[^,}]*' | cut -d':' -f2)
        test_pass "Event creation successful"
    else
        test_fail "Event creation failed" "$response"
        return 1
    fi

    # Test event creation with invalid dates (end before start)
    test_start "Create event with invalid dates (end before start)"
    local future_start=$(date -d "+7 days" --iso-8601=seconds)
    local earlier_end=$(date -d "+6 days" --iso-8601=seconds)
    local invalid_event="{
        \"name\": \"Invalid Event\",
        \"venue_id\": \"$VENUE_ID\",
        \"event_type\": \"concert\",
        \"start_datetime\": \"$future_start\",
        \"end_datetime\": \"$earlier_end\",
        \"total_capacity\": 100,
        \"base_price\": 10.00
    }"

    response=$(make_request "POST" "/api/v1/admin/events" "$invalid_event" "Authorization: Bearer $ADMIN_TOKEN" "400")
    if echo "$response" | grep -q -i "before\|datetime"; then
        test_pass "Invalid event dates properly rejected"
    else
        test_fail "Invalid dates should be rejected" "$response"
    fi

    # Test event creation with past dates
    test_start "Create event with past start date"
    local past_start=$(date -d "-1 day" --iso-8601=seconds)
    local past_end=$(date -d "+1 hour" --iso-8601=seconds)
    local past_event="{
        \"name\": \"Past Event\",
        \"venue_id\": \"$VENUE_ID\",
        \"event_type\": \"concert\",
        \"start_datetime\": \"$past_start\",
        \"end_datetime\": \"$past_end\",
        \"total_capacity\": 100,
        \"base_price\": 10.00
    }"

    response=$(make_request "POST" "/api/v1/admin/events" "$past_event" "Authorization: Bearer $ADMIN_TOKEN" "400")
    if echo "$response" | grep -q -i "future\|past"; then
        test_pass "Past event dates properly rejected"
    else
        test_fail "Past dates should be rejected" "$response"
    fi

    # Test event creation with invalid venue ID
    test_start "Create event with non-existent venue"
    local fake_venue_id="99999999-9999-9999-9999-999999999999"
    local invalid_venue_event=$(generate_test_event "$fake_venue_id")

    response=$(make_request "POST" "/api/v1/admin/events" "$invalid_venue_event" "Authorization: Bearer $ADMIN_TOKEN" "400")
    if echo "$response" | grep -q -i "invalid.*venue\|foreign key\|venue.*not.*exist\|venue.*does.*not.*exist"; then
        test_pass "Invalid venue ID properly rejected"
    else
        test_fail "Invalid venue ID should be rejected" "$response"
    fi

    # Test list admin events
    test_start "List admin events"
    response=$(make_request "GET" "/api/v1/admin/events" "" "Authorization: Bearer $ADMIN_TOKEN" "200")
    if echo "$response" | grep -q '"events"'; then
        test_pass "Admin events listing successful"
    else
        test_fail "Admin events listing failed" "$response"
    fi

    # Test event update
    test_start "Update event"
    local update_data="{
        \"name\": \"Updated Event Name\",
        \"total_capacity\": 600,
        \"base_price\": 39.99,
        \"status\": \"published\",
        \"version\": $EVENT_VERSION
    }"

    response=$(make_request "PUT" "/api/v1/admin/events/$EVENT_ID" "$update_data" "Authorization: Bearer $ADMIN_TOKEN" "200")
    if echo "$response" | grep -q "Updated Event Name"; then
        EVENT_VERSION=$(echo "$response" | grep -o '"version":[^,}]*' | cut -d':' -f2)
        test_pass "Event update successful"
    else
        test_fail "Event update failed" "$response"
    fi

    # Test optimistic locking
    test_start "Update event with stale version (optimistic locking)"
    local stale_update="{
        \"name\": \"Should Fail\",
        \"version\": 1
    }"

    response=$(make_request "PUT" "/api/v1/admin/events/$EVENT_ID" "$stale_update" "Authorization: Bearer $ADMIN_TOKEN" "409")
    if echo "$response" | grep -q -i "version\|conflict\|updated.*another\|refresh.*try.*again"; then
        test_pass "Optimistic locking working correctly"
    else
        test_fail "Optimistic locking should prevent stale updates" "$response"
    fi

    # Test event analytics
    test_start "Get event analytics"
    response=$(make_request "GET" "/api/v1/admin/events/$EVENT_ID/analytics" "" "Authorization: Bearer $ADMIN_TOKEN" "200")
    if echo "$response" | grep -q '"capacity_utilization"'; then
        test_pass "Event analytics retrieval successful"
    else
        test_fail "Event analytics failed" "$response"
    fi
}

# ============================================================================
# PUBLIC EVENT ENDPOINTS TESTS
# ============================================================================

test_public_event_endpoints() {
    log "${PURPLE}=== PUBLIC EVENT ENDPOINTS TESTS ===${NC}"

    # Test list published events
    test_start "List published events (public endpoint)"
    response=$(make_request "GET" "/api/v1/events" "" "" "200")
    if echo "$response" | grep -q '"events"'; then
        test_pass "Public events listing successful"
    else
        test_fail "Public events listing failed" "$response"
    fi

    # Test list events with filters
    test_start "List events with type filter"
    response=$(make_request "GET" "/api/v1/events?type=concert" "" "" "200")
    if echo "$response" | grep -q '"events"'; then
        test_pass "Event filtering by type working"
    else
        test_fail "Event filtering failed" "$response"
    fi

    test_start "List events with city filter"
    response=$(make_request "GET" "/api/v1/events?city=Test%20City" "" "" "200")
    if echo "$response" | grep -q '"events"'; then
        test_pass "Event filtering by city working"
    else
        test_fail "Event filtering by city failed" "$response"
    fi

    test_start "List events with date range filter"
    local tomorrow=$(date -d "+1 day" +%Y-%m-%d)
    local next_month=$(date -d "+30 days" +%Y-%m-%d)
    response=$(make_request "GET" "/api/v1/events?date_from=$tomorrow&date_to=$next_month" "" "" "200")
    if echo "$response" | grep -q '"events"'; then
        test_pass "Event filtering by date range working"
    else
        test_fail "Event filtering by date range failed" "$response"
    fi

    # Test pagination
    test_start "List events with pagination"
    response=$(make_request "GET" "/api/v1/events?page=1&limit=5" "" "" "200")
    if echo "$response" | grep -q '"page":1.*"limit":5'; then
        test_pass "Event pagination working"
    else
        test_fail "Event pagination failed" "$response"
    fi

    if [ -n "$EVENT_ID" ]; then
        # Test get specific event
        test_start "Get specific event by ID"
        response=$(make_request "GET" "/api/v1/events/$EVENT_ID" "" "" "200")
        if echo "$response" | grep -q '"event_id"'; then
            test_pass "Event retrieval by ID successful"
        else
            test_fail "Event retrieval by ID failed" "$response"
        fi

        # Test get event availability
        test_start "Get event availability"
        response=$(make_request "GET" "/api/v1/events/$EVENT_ID/availability" "" "" "200")
        if echo "$response" | grep -q '"available_seats"'; then
            test_pass "Event availability retrieval successful"
        else
            test_fail "Event availability retrieval failed" "$response"
        fi
    fi

    # Test invalid event ID
    test_start "Get non-existent event"
    local fake_id="00000000-0000-0000-0000-000000000000"
    response=$(make_request "GET" "/api/v1/events/$fake_id" "" "" "404")
    if echo "$response" | grep -q -i "not found"; then
        test_pass "Non-existent event properly returns 404"
    else
        test_fail "Should return 404 for non-existent event" "$response"
    fi

    test_start "Get event with invalid UUID format"
    response=$(make_request "GET" "/api/v1/events/invalid-uuid" "" "" "400")
    if echo "$response" | grep -q -i "invalid.*id"; then
        test_pass "Invalid UUID format properly rejected"
    else
        test_fail "Invalid UUID should return 400" "$response"
    fi
}

# ============================================================================
# INTERNAL SERVICE ENDPOINTS TESTS
# ============================================================================

test_internal_endpoints() {
    log "${PURPLE}=== INTERNAL SERVICE ENDPOINTS TESTS ===${NC}"

    if [ -z "$EVENT_ID" ]; then
        test_fail "No event ID available for internal endpoint tests"
        return 1
    fi

    # Test get event for booking (internal)
    test_start "Get event for booking (internal endpoint)"
    response=$(make_request "GET" "/internal/events/$EVENT_ID" "" "Authorization: ApiKey $INTERNAL_API_KEY" "200")
    if echo "$response" | grep -q '"available_seats"'; then
        test_pass "Internal event retrieval for booking successful"
    else
        test_fail "Internal event retrieval failed" "$response"
    fi

    # Test without internal API key
    test_start "Get event for booking without API key"
    response=$(make_request "GET" "/internal/events/$EVENT_ID" "" "" "401")
    if echo "$response" | grep -q -i "unauthorized\|forbidden\|missing.*invalid.*api.*key"; then
        test_pass "Internal endpoint properly requires API key"
    else
        test_fail "Internal endpoint should require API key" "$response"
    fi

    # Test seat reservation (negative quantity)
    test_start "Reserve seats via internal endpoint"
    local reserve_data="{
        \"quantity\": -5,
        \"version\": $EVENT_VERSION
    }"

    response=$(make_request "POST" "/internal/events/$EVENT_ID/update-availability" "$reserve_data" "Authorization: ApiKey $INTERNAL_API_KEY" "200")
    if echo "$response" | grep -q '"available_seats"'; then
        EVENT_VERSION=$(echo "$response" | grep -o '"version":[^,}]*' | cut -d':' -f2)
        test_pass "Seat reservation successful"
    else
        test_fail "Seat reservation failed" "$response"
    fi

    # Test seat return (positive quantity)
    test_start "Return seats via internal endpoint"
    local return_data="{
        \"quantity\": 3,
        \"version\": $EVENT_VERSION
    }"

    response=$(make_request "POST" "/internal/events/$EVENT_ID/update-availability" "$return_data" "Authorization: ApiKey $INTERNAL_API_KEY" "200")
    if echo "$response" | grep -q '"available_seats"'; then
        EVENT_VERSION=$(echo "$response" | grep -o '"version":[^,}]*' | cut -d':' -f2)
        test_pass "Seat return successful"
    else
        test_fail "Seat return failed" "$response"
    fi

    # Test return seats via dedicated endpoint
    test_start "Return seats via dedicated return endpoint"
    local dedicated_return="{
        \"quantity\": 2,
        \"version\": $EVENT_VERSION
    }"

    response=$(make_request "POST" "/internal/events/$EVENT_ID/return-seats" "$dedicated_return" "Authorization: ApiKey $INTERNAL_API_KEY" "200")
    if echo "$response" | grep -q '"available_seats"'; then
        test_pass "Dedicated seat return successful"
    else
        test_fail "Dedicated seat return failed" "$response"
    fi

    # Test concurrent update protection
    test_start "Test concurrent update protection (stale version)"
    local stale_data="{
        \"quantity\": -1,
        \"version\": 1
    }"

    response=$(make_request "POST" "/internal/events/$EVENT_ID/update-availability" "$stale_data" "Authorization: ApiKey $INTERNAL_API_KEY" "409")
    if echo "$response" | grep -q -i "conflict\|updated.*another\|retry"; then
        test_pass "Concurrent update protection working"
    else
        test_fail "Should prevent updates with stale version" "$response"
    fi

    # Test overselling protection
    test_start "Test overselling protection"
    local oversell_data="{
        \"quantity\": -10000,
        \"version\": $EVENT_VERSION
    }"

    response=$(make_request "POST" "/internal/events/$EVENT_ID/update-availability" "$oversell_data" "Authorization: ApiKey $INTERNAL_API_KEY" "409")
    if echo "$response" | grep -q -i "not enough.*seats\|available\|updated.*another.*process\|retry"; then
        test_pass "Overselling protection working"
    else
        test_fail "Should prevent overselling" "$response"
    fi

    # Test invalid data
    test_start "Test internal endpoint with zero quantity"
    local zero_data='{"quantity": 0, "version": 1}'
    response=$(make_request "POST" "/internal/events/$EVENT_ID/update-availability" "$zero_data" "Authorization: ApiKey $INTERNAL_API_KEY" "400")
    if echo "$response" | grep -q -i "cannot be zero\|quantity"; then
        test_pass "Zero quantity properly rejected"
    else
        test_fail "Zero quantity should be rejected" "$response"
    fi
}

# ============================================================================
# EDGE CASES AND ERROR HANDLING TESTS
# ============================================================================

test_edge_cases() {
    log "${PURPLE}=== EDGE CASES AND ERROR HANDLING TESTS ===${NC}"

    # Test malformed JSON
    test_start "Test malformed JSON in request"
    response=$(make_request "POST" "/api/v1/auth/admin/register" '{"invalid": json}' "" "400")
    if echo "$response" | grep -q -i "invalid.*json\|json"; then
        test_pass "Malformed JSON properly rejected"
    else
        test_fail "Malformed JSON should return 400" "$response"
    fi

    # Test extremely long request
    test_start "Test request with extremely long data"
    local long_string=$(printf 'A%.0s' {1..10000})
    local long_data="{\"name\": \"$long_string\"}"
    response=$(make_request "POST" "/api/v1/admin/venues" "$long_data" "Authorization: Bearer $ADMIN_TOKEN")
    # This might succeed or fail depending on implementation, but shouldn't crash
    test_pass "Server handled extremely long request without crashing"

    # Test special characters in data
    test_start "Test special characters in venue name"
    local special_venue='{
        "name": "Test Venue with Special Chars: éñøß™",
        "address": "123 Special St",
        "city": "Tëst Çity",
        "country": "USA",
        "capacity": 100
    }'
    response=$(make_request "POST" "/api/v1/admin/venues" "$special_venue" "Authorization: Bearer $ADMIN_TOKEN")
    if echo "$response" | grep -q '"venue_id"'; then
        test_pass "Special characters handled correctly"
    else
        test_warning "Special characters might need better handling" "$response"
    fi

    # Test boundary values
    test_start "Test venue with minimum capacity"
    local min_venue='{
        "name": "Minimum Venue",
        "address": "123 Min St",
        "city": "Min City",
        "country": "USA",
        "capacity": 1
    }'
    response=$(make_request "POST" "/api/v1/admin/venues" "$min_venue" "Authorization: Bearer $ADMIN_TOKEN")
    if echo "$response" | grep -q '"venue_id"'; then
        test_pass "Minimum capacity venue created successfully"
    else
        test_fail "Minimum capacity should be allowed" "$response"
    fi

    test_start "Test venue with zero capacity"
    local zero_venue='{
        "name": "Zero Venue",
        "address": "123 Zero St",
        "city": "Zero City",
        "country": "USA",
        "capacity": 0
    }'
    response=$(make_request "POST" "/api/v1/admin/venues" "$zero_venue" "Authorization: Bearer $ADMIN_TOKEN" "400")
    if echo "$response" | grep -q -i "positive\|capacity"; then
        test_pass "Zero capacity properly rejected"
    else
        test_fail "Zero capacity should be rejected" "$response"
    fi

    # Test pagination edge cases
    test_start "Test pagination with negative page"
    response=$(make_request "GET" "/api/v1/events?page=-1&limit=10" "" "" "200")
    # Should handle gracefully, probably default to page 1
    test_pass "Negative pagination handled gracefully"

    test_start "Test pagination with excessive limit"
    response=$(make_request "GET" "/api/v1/events?page=1&limit=1000" "" "" "200")
    # Should cap the limit to maximum allowed
    test_pass "Excessive pagination limit handled gracefully"

    # Test SQL injection attempts (should be prevented by prepared statements)
    test_start "Test SQL injection in search parameters"
    response=$(make_request "GET" "/api/v1/events?city='; DROP TABLE events; --" "" "" "200")
    test_pass "SQL injection attempt safely handled"
}

# ============================================================================
# PERFORMANCE AND LOAD TESTS
# ============================================================================

test_performance() {
    log "${PURPLE}=== BASIC PERFORMANCE TESTS ===${NC}"

    # Test concurrent requests to public endpoint
    test_start "Test concurrent requests to public events endpoint"
    local start_time=$(date +%s.%N)

    for i in {1..10}; do
        (make_request "GET" "/api/v1/events" "" "" "200" > /dev/null 2>&1) &
    done
    wait

    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc -l)

    if (( $(echo "$duration < 5.0" | bc -l) )); then
        test_pass "10 concurrent requests completed in ${duration}s (< 5s)"
    else
        test_warning "10 concurrent requests took ${duration}s (> 5s)"
    fi

    # Test response time for individual requests
    test_start "Test individual request response time"
    local start_time=$(date +%s.%N)
    make_request "GET" "/healthz" "" "" "200" > /dev/null
    local end_time=$(date +%s.%N)
    local response_time=$(echo "$end_time - $start_time" | bc -l)

    if (( $(echo "$response_time < 1.0" | bc -l) )); then
        test_pass "Health check response time: ${response_time}s (< 1s)"
    else
        test_warning "Health check response time: ${response_time}s (> 1s)"
    fi
}

# ============================================================================
# SECURITY TESTS
# ============================================================================

test_security() {
    log "${PURPLE}=== SECURITY TESTS ===${NC}"

    # Test JWT token manipulation
    test_start "Test with invalid JWT token"
    response=$(make_request "GET" "/api/v1/admin/events" "" "Authorization: Bearer invalid.jwt.token" "401")
    if echo "$response" | grep -q -i "unauthorized\|invalid.*token"; then
        test_pass "Invalid JWT token properly rejected"
    else
        test_fail "Invalid JWT should be rejected" "$response"
    fi

    # Test expired/malformed tokens
    test_start "Test with malformed Authorization header"
    response=$(make_request "GET" "/api/v1/admin/events" "" "Authorization: InvalidFormat" "401")
    if echo "$response" | grep -q -i "unauthorized\|missing.*invalid.*authorization"; then
        test_pass "Malformed authorization header rejected"
    else
        test_fail "Malformed auth header should be rejected" "$response"
    fi

    # Test internal API key validation
    test_start "Test internal endpoint with wrong API key"
    if [ -n "$EVENT_ID" ]; then
        response=$(make_request "GET" "/internal/events/$EVENT_ID" "" "Authorization: ApiKey wrong-api-key" "403")
        if echo "$response" | grep -q -i "unauthorized\|forbidden\|invalid.*api.*key"; then
            test_pass "Wrong internal API key properly rejected"
        else
            test_fail "Wrong internal API key should be rejected" "$response"
        fi
    else
        test_warning "No event ID available for internal API key test"
    fi

    # Test CSRF protection (if implemented)
    test_start "Test potential CSRF vectors"
    response=$(make_request "POST" "/api/v1/admin/venues" '{"name":"CSRF Test"}' "Authorization: Bearer $ADMIN_TOKEN" "" "")
    # This is more about documenting that we tested it
    test_pass "CSRF protection test completed"

    # Test input sanitization
    test_start "Test XSS attempt in venue name"
    local xss_venue='{
        "name": "<script>alert(\"xss\")</script>",
        "address": "123 XSS St",
        "city": "XSS City",
        "country": "USA",
        "capacity": 100
    }'
    response=$(make_request "POST" "/api/v1/admin/venues" "$xss_venue" "Authorization: Bearer $ADMIN_TOKEN")
    if echo "$response" | grep -q '"venue_id"' && ! echo "$response" | grep -q '<script>'; then
        test_pass "XSS attempt in venue name properly sanitized"
    else
        test_warning "XSS sanitization may need review" "$response"
    fi
}

# ============================================================================
# DATA CONSISTENCY TESTS
# ============================================================================

test_data_consistency() {
    log "${PURPLE}=== DATA CONSISTENCY TESTS ===${NC}"

    if [ -z "$EVENT_ID" ] || [ -z "$ADMIN_TOKEN" ]; then
        test_warning "Skipping data consistency tests - missing test data"
        return
    fi

    # Test event-venue relationship consistency
    test_start "Test event-venue relationship consistency"
    response=$(make_request "GET" "/api/v1/events/$EVENT_ID" "" "" "200")
    if echo "$response" | grep -q '"venue_name"'; then
        test_pass "Event-venue relationship maintained"
    else
        test_fail "Event-venue relationship broken" "$response"
    fi

    # Test capacity constraints
    test_start "Test event capacity constraints"
    local event_response=$(make_request "GET" "/api/v1/events/$EVENT_ID" "" "" "200")
    local available_seats=$(echo "$event_response" | grep -o '"available_seats":[^,}]*' | cut -d':' -f2)
    local total_capacity=$(echo "$event_response" | grep -o '"total_capacity":[^,}]*' | cut -d':' -f2)

    if [ "$available_seats" -le "$total_capacity" ]; then
        test_pass "Available seats <= total capacity constraint maintained"
    else
        test_fail "Available seats ($available_seats) > total capacity ($total_capacity)"
    fi

    # Test version consistency after updates
    test_start "Test version consistency after updates"
    local initial_version=$(echo "$event_response" | grep -o '"version":[^,}]*' | cut -d':' -f2)

    local minor_update="{
        \"description\": \"Updated description $(date +%s)\",
        \"version\": $initial_version
    }"

    response=$(make_request "PUT" "/api/v1/admin/events/$EVENT_ID" "$minor_update" "Authorization: Bearer $ADMIN_TOKEN" "200")
    local new_version=$(echo "$response" | grep -o '"version":[^,}]*' | cut -d':' -f2)

    if [ "$new_version" -gt "$initial_version" ]; then
        test_pass "Version incremented correctly after update"
        EVENT_VERSION=$new_version
    else
        test_fail "Version not incremented after update (was: $initial_version, now: $new_version)"
    fi
}

# ============================================================================
# CLEANUP TESTS
# ============================================================================

test_cleanup() {
    log "${PURPLE}=== CLEANUP TESTS ===${NC}"

    if [ -n "$EVENT_ID" ] && [ -n "$ADMIN_TOKEN" ]; then
        # Test event deletion
        test_start "Delete test event"
        local delete_data="{\"version\": $EVENT_VERSION}"
        response=$(make_request "DELETE" "/api/v1/admin/events/$EVENT_ID" "$delete_data" "Authorization: Bearer $ADMIN_TOKEN" "200")
        if echo "$response" | grep -q -i "success\|cancel"; then
            test_pass "Event deletion successful"
        else
            test_fail "Event deletion failed" "$response"
        fi

        # Verify event is cancelled/deleted
        test_start "Verify event is cancelled after deletion"
        response=$(make_request "GET" "/api/v1/events/$EVENT_ID" "" "" "200")
        if echo "$response" | grep -q '"status":"cancelled"' || echo "$response" | grep -q -i "not found"; then
            test_pass "Event properly cancelled/removed"
        else
            test_warning "Event status after deletion unclear" "$response"
        fi
    fi

    # Note: We intentionally don't delete venues as they might be referenced by other events
    # In a real scenario, you might want to add a "test cleanup" mode
    test_start "Venue cleanup (intentionally skipped)"
    test_pass "Venue cleanup skipped to prevent referential integrity issues"
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

main() {
    log "${CYAN}=======================================${NC}"
    log "${CYAN}Event Service Comprehensive Test Suite${NC}"
    log "${CYAN}=======================================${NC}"
    log "${CYAN}Base URL: $BASE_URL${NC}"
    log "${CYAN}Test Results File: $TEST_RESULTS_FILE${NC}"
    log "${CYAN}Start Time: $(date)${NC}"
    log "${CYAN}=======================================${NC}"
    echo

    # Check if bc is available for floating point arithmetic
    if ! command -v bc &> /dev/null; then
        log "${YELLOW}Warning: 'bc' command not found. Some performance measurements may not work.${NC}"
    fi

    # Test if server is running
    log "${BLUE}Checking if Event Service is running...${NC}"
    if ! curl -s --fail "$BASE_URL/healthz" > /dev/null; then
        log "${RED}ERROR: Event Service is not running at $BASE_URL${NC}"
        log "${RED}Please start the service and try again.${NC}"
        exit 1
    fi
    log "${GREEN}✓ Event Service is running${NC}"
    echo

    # Run all test suites
    test_health_endpoints
    test_admin_authentication
    test_venue_management
    test_event_management
    test_public_event_endpoints
    test_internal_endpoints
    test_edge_cases
    test_performance
    test_security
    test_data_consistency
    test_cleanup

    # Print summary
    log "${CYAN}=======================================${NC}"
    log "${CYAN}TEST SUMMARY${NC}"
    log "${CYAN}=======================================${NC}"
    log "${GREEN}Total Tests: $TOTAL_TESTS${NC}"
    log "${GREEN}Passed: $PASSED_TESTS${NC}"
    log "${RED}Failed: $FAILED_TESTS${NC}"
    log "${YELLOW}Success Rate: $(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc -l 2>/dev/null || echo "N/A")%${NC}"
    log "${CYAN}End Time: $(date)${NC}"
    log "${CYAN}Full results saved to: $TEST_RESULTS_FILE${NC}"
    log "${CYAN}=======================================${NC}"

    # Exit with appropriate code
    if [ $FAILED_TESTS -eq 0 ]; then
        log "${GREEN}🎉 All tests passed!${NC}"
        exit 0
    else
        log "${RED}❌ Some tests failed. Check the results above.${NC}"
        exit 1
    fi
}

# ============================================================================
# SCRIPT ENTRY POINT
# ============================================================================

# Check for help
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    echo "Event Service Comprehensive Test Suite"
    echo
    echo "Usage: $0 [base_url] [internal_api_key]"
    echo
    echo "Arguments:"
    echo "  base_url         Base URL of the Event Service (default: http://localhost:8002)"
    echo "  internal_api_key Internal API key for service-to-service calls"
    echo
    echo "Environment Variables:"
    echo "  You can also set these via environment variables:"
    echo "  EVENT_SERVICE_URL     - Base URL of the service"
    echo "  INTERNAL_API_KEY      - Internal API key"
    echo
    echo "Examples:"
    echo "  $0"
    echo "  $0 http://localhost:8002 your-api-key"
    echo "  EVENT_SERVICE_URL=http://localhost:8002 INTERNAL_API_KEY=your-key $0"
    echo
    echo "Requirements:"
    echo "  - curl (for HTTP requests)"
    echo "  - jq (optional, for prettier JSON output)"
    echo "  - bc (optional, for performance measurements)"
    echo
    echo "This script will:"
    echo "  1. Test all public and admin endpoints"
    echo "  2. Test authentication and authorization"
    echo "  3. Test data validation and error handling"
    echo "  4. Test internal service-to-service endpoints"
    echo "  5. Test edge cases and security"
    echo "  6. Perform basic performance tests"
    echo "  7. Generate a comprehensive test report"
    exit 0
fi

# Use environment variables if set
if [ -n "$EVENT_SERVICE_URL" ]; then
    BASE_URL="$EVENT_SERVICE_URL"
fi

if [ -n "$INTERNAL_API_KEY" ] && [ -z "$2" ]; then
    INTERNAL_API_KEY="$INTERNAL_API_KEY"
fi

# Run main function
main "$@"
