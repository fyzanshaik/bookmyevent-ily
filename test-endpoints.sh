#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
BASE_URL="http://localhost:8001"
INTERNAL_API_KEY="internal-service-communication-key-change-in-production"

# --- Helper Functions ---

# Function to print a formatted header
print_header() {
    echo -e "\n--------------------------------------------------"
    echo -e "--- $1 ---"
    echo -e "--------------------------------------------------"
}

# Function to check the HTTP status code of the last curl command
check_status() {
    expected_status=$1
    actual_status=$2
    test_name=$3

    if [ "$actual_status" -ne "$expected_status" ]; then
        echo -e "\033[0;31m❌ Test Failed: $test_name\033[0m"
        echo "Expected HTTP $expected_status, but got $actual_status"
        exit 1
    else
        echo -e "\033[0;32m✅ Test Passed: $test_name (HTTP $actual_status)\033[0m"
    fi
}

# --- Test Execution ---

main() {
    echo "=================================================="
    echo "  Starting Evently API Integration Test Suite"
    echo "=================================================="

    # 1. Health Check
    print_header "1. Health Check"
    response=$(curl -s -w \"%{http_code}\" -o response.json "$BASE_URL/healthz")
    check_status 200 "$response" "Health Check"
    cat response.json

    # 2. User Registration (Success)
    print_header "2. User Registration (Success)"
    random_email="testuser_$(date +%s)_$RANDOM@example.com"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$random_email\", \"password\": \"password123\", \"name\": \"Test User\"}")
    check_status 201 "$response" "User Registration"
    cat response.json

    # Extract tokens and user ID
    ACCESS_TOKEN=$(jq -r '.access_token' response.json)
    REFRESH_TOKEN=$(jq -r '.refresh_token' response.json)
    USER_ID=$(jq -r '.user_id' response.json)

    if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" == "null" ]; then
        echo "\033[0;31m❌ Critical Error: Could not extract ACCESS_TOKEN.\033[0m"
        exit 1
    fi

    # 3. User Registration (Failure - Duplicate Email)
    print_header "3. User Registration (Failure - Duplicate Email)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$random_email\", \"password\": \"password123\", \"name\": \"Test User\"}")
    check_status 500 "$response" "Duplicate User Registration"
    cat response.json

    # 4. User Login (Success)
    print_header "4. User Login (Success)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$random_email\", \"password\": \"password123\"}")
    check_status 200 "$response" "User Login"
    cat response.json

    # 5. User Login (Failure - Wrong Password)
    print_header "5. User Login (Failure - Wrong Password)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$random_email\", \"password\": \"wrongpassword\"}")
    check_status 401 "$response" "Wrong Password Login"
    cat response.json

    # 6. Get User Profile (Success)
    print_header "6. Get User Profile (Success)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X GET "$BASE_URL/api/v1/users/profile" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    check_status 200 "$response" "Get Profile"
    cat response.json

    # 7. Get User Profile (Failure - No Token)
    print_header "7. Get User Profile (Failure - No Token)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X GET "$BASE_URL/api/v1/users/profile")
    check_status 401 "$response" "Get Profile without Token"
    cat response.json

    # 8. Get User Profile (Failure - Invalid Token)
    print_header "8. Get User Profile (Failure - Invalid Token)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X GET "$BASE_URL/api/v1/users/profile" \
        -H "Authorization: Bearer invalidtoken")
    check_status 401 "$response" "Get Profile with Invalid Token"
    cat response.json

    # 9. Refresh Token (Success)
    print_header "9. Refresh Token (Success)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/api/v1/auth/refresh" \
        -H "Authorization: Bearer $REFRESH_TOKEN")
    check_status 200 "$response" "Refresh Token"
    cat response.json
    
    # Update tokens for subsequent requests
    ACCESS_TOKEN=$(jq -r '.access_token' response.json)
    NEW_REFRESH_TOKEN=$(jq -r '.refresh_token' response.json)

    # 10. Logout (Success)
    print_header "10. Logout (Success)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/api/v1/auth/logout" \
        -H "Authorization: Bearer $NEW_REFRESH_TOKEN")
    check_status 200 "$response" "Logout"
    cat response.json

    # 11. Refresh Token (Failure - Revoked Token)
    print_header "11. Refresh Token (Failure - Revoked Token)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/api/v1/auth/refresh" \
        -H "Authorization: Bearer $NEW_REFRESH_TOKEN")
    check_status 401 "$response" "Refresh with Revoked Token"
    cat response.json

    # --- Internal API Tests ---
    print_header "12. Internal - Verify Token (Success)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/internal/auth/verify" \
        -H "Authorization: ApiKey $INTERNAL_API_KEY" \
        -d "{\"token\": \"$ACCESS_TOKEN\"}")
    check_status 200 "$response" "Internal Verify Token"
    cat response.json

    print_header "13. Internal - Verify Token (Failure - Bad API Key)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X POST "$BASE_URL/internal/auth/verify" \
        -H "Authorization: ApiKey badkey" \
        -d "{\"token\": \"$ACCESS_TOKEN\"}")
    check_status 403 "$response" "Internal Verify with Bad Key"
    cat response.json

    print_header "14. Internal - Get User (Success)"
    response=$(curl -s -w \"%{http_code}\" -o response.json -X GET "$BASE_URL/internal/users/$USER_ID" \
        -H "Authorization: ApiKey $INTERNAL_API_KEY")
    check_status 200 "$response" "Internal Get User"
    cat response.json

    echo -e "\n\033[0;32m==================================================\033[0m"
    echo -e "\033[0;32m  All Integration Tests Passed Successfully!\033[0m"
    echo -e "\033[0;32m==================================================\033[0m"

    # Clean up the response file
    rm -f response.json
}

# --- Entrypoint ---
main