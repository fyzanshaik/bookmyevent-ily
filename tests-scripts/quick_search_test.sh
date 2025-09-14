#!/bin/bash

# Quick Search API Testing Script
# Simple curl-based tests for immediate verification

BASE_URL="http://localhost:8003"
EVENT_URL="http://localhost:8002"
INTERNAL_KEY="internal-service-communication-key-change-in-production"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${PURPLE}========================================${NC}"
echo -e "${PURPLE}Quick Search API Testing${NC}"
echo -e "${PURPLE}========================================${NC}"

# Function to make a request and show result
test_endpoint() {
    local name="$1"
    local method="$2"
    local url="$3"
    local headers="$4"
    local data="$5"
    
    echo -e "\n${CYAN}Testing: $name${NC}"
    echo -e "${YELLOW}Request: $method $url${NC}"
    
    if [ ! -z "$data" ]; then
        echo -e "${YELLOW}Data: $data${NC}"
    fi
    
    if [ ! -z "$headers" ]; then
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" "$url" -H "$headers" -d "$data")
    else
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" "$url")
    fi
    
    http_code=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    body=$(echo "$response" | grep -v "HTTP_STATUS:")
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ Status: $http_code${NC}"
    else
        echo -e "${RED}✗ Status: $http_code${NC}"
    fi
    
    echo -e "${BLUE}Response:${NC}"
    echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
    echo -e "${PURPLE}----------------------------------------${NC}"
}

# Check if search service is running
echo -e "\n${CYAN}Checking if Search Service is running...${NC}"
if ! curl -s "$BASE_URL/healthz" > /dev/null; then
    echo -e "${RED}❌ Search service is not running on $BASE_URL${NC}"
    echo -e "${YELLOW}Please start the search service first:${NC}"
    echo -e "${YELLOW}make run SERVICE=search-service${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Search service is running${NC}"

# Test 1: Health Check
test_endpoint "Health Check" "GET" "$BASE_URL/healthz"

# Test 2: Readiness Check
test_endpoint "Readiness Check" "GET" "$BASE_URL/health/ready"

# Test 3: Basic Search (no parameters)
test_endpoint "Basic Search - No Parameters" "GET" "$BASE_URL/api/v1/search"

# Test 4: Search with query
test_endpoint "Search with Query" "GET" "$BASE_URL/api/v1/search?q=concert"

# Test 5: Search with filters
test_endpoint "Search with City Filter" "GET" "$BASE_URL/api/v1/search?city=New%20York"

# Test 6: Search with event type filter
test_endpoint "Search with Type Filter" "GET" "$BASE_URL/api/v1/search?type=concert"

# Test 7: Search with price range
test_endpoint "Search with Price Range" "GET" "$BASE_URL/api/v1/search?min_price=50&max_price=150"

# Test 8: Search with pagination
test_endpoint "Search with Pagination" "GET" "$BASE_URL/api/v1/search?page=1&limit=5"

# Test 9: Combined filters
test_endpoint "Search with Combined Filters" "GET" "$BASE_URL/api/v1/search?q=music&type=concert&min_price=50"

# Test 10: Search suggestions
test_endpoint "Search Suggestions" "GET" "$BASE_URL/api/v1/search/suggestions?q=jazz&limit=5"

# Test 11: Search suggestions with empty query (should fail)
test_endpoint "Search Suggestions - Empty Query" "GET" "$BASE_URL/api/v1/search/suggestions?q="

# Test 12: Get filters metadata
test_endpoint "Get Filters Metadata" "GET" "$BASE_URL/api/v1/search/filters"

# Test 13: Get trending events
test_endpoint "Get Trending Events" "GET" "$BASE_URL/api/v1/search/trending"

# Test 14: Get trending events with limit
test_endpoint "Get Trending Events with Limit" "GET" "$BASE_URL/api/v1/search/trending?limit=3"

# Test 15: Internal endpoint - Full resync (requires API key)
echo -e "\n${CYAN}Testing Internal Endpoints${NC}"
test_endpoint "Internal - Full Resync" "POST" "$BASE_URL/internal/search/resync" "Content-Type: application/json" '{"force_reindex": false}'

echo -e "\n${PURPLE}========================================${NC}"
echo -e "${PURPLE}Quick Search API Testing Completed${NC}"
echo -e "${PURPLE}========================================${NC}"

echo -e "\n${YELLOW}For more detailed testing, use:${NC}"
echo -e "${CYAN}python3 step_by_step_search_test.py all${NC}"
echo -e "${CYAN}python3 comprehensive_search_api_test.py${NC}"
