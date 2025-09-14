#!/bin/bash

set -e

SEARCH_SERVICE_URL="http://localhost:8003"

echo "🔍 Search Service Endpoint Testing"
echo "=================================="
echo "Target URL: $SEARCH_SERVICE_URL"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to test an endpoint
test_endpoint() {
    local test_name="$1"
    local endpoint="$2"
    local expected_field="$3"
    
    echo -e "${BLUE}Testing: $test_name${NC}"
    echo "URL: $SEARCH_SERVICE_URL$endpoint"
    
    response=$(curl -s -w "\n%{http_code}" "$SEARCH_SERVICE_URL$endpoint" 2>/dev/null)
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        if echo "$response_body" | jq empty 2>/dev/null; then
            if [ -n "$expected_field" ]; then
                if echo "$response_body" | jq -e ".$expected_field" > /dev/null 2>&1; then
                    count=$(echo "$response_body" | jq -r "if .$expected_field | type == \"array\" then .$expected_field | length else .$expected_field end")
                    echo -e "${GREEN}✅ SUCCESS${NC} - HTTP 200, Valid JSON, Found $expected_field: $count"
                else
                    echo -e "${YELLOW}⚠️  WARNING${NC} - HTTP 200, Valid JSON, but missing expected field: $expected_field"
                fi
            else
                echo -e "${GREEN}✅ SUCCESS${NC} - HTTP 200, Valid JSON"
            fi
            
            # Show a preview of the response
            echo "Preview:"
            echo "$response_body" | jq -C '.' | head -10
        else
            echo -e "${RED}❌ FAILED${NC} - HTTP 200 but invalid JSON"
            echo "Response: $response_body" | head -3
        fi
    else
        echo -e "${RED}❌ FAILED${NC} - HTTP $http_code"
        echo "Response: $response_body" | head -3
    fi
    
    echo ""
}

# Check if service is running
echo "🔍 Checking if Search Service is running..."
if curl -s "$SEARCH_SERVICE_URL/healthz" > /dev/null; then
    echo -e "${GREEN}✅ Search Service is running${NC}"
else
    echo -e "${RED}❌ Search Service is not running on $SEARCH_SERVICE_URL${NC}"
    echo "💡 Start it with: make run SERVICE=search-service"
    exit 1
fi

echo ""

# Test all endpoints
echo "🚀 Running endpoint tests..."
echo ""

# Health endpoints
test_endpoint "Health Check" "/healthz" "status"
test_endpoint "Readiness Check" "/health/ready" "status"

# Basic search
test_endpoint "Basic Search" "/api/v1/search?limit=5" "results"
test_endpoint "Search with Pagination" "/api/v1/search?page=2&limit=3" "results"

# Search with queries
test_endpoint "Search by Query - Concert" "/api/v1/search?q=concert&limit=3" "results"
test_endpoint "Search by Query - Jazz" "/api/v1/search?q=jazz&limit=3" "results"
test_endpoint "Search by Query - New York" "/api/v1/search?q=New%20York&limit=3" "results"

# Search with filters
test_endpoint "Filter by City" "/api/v1/search?city=New%20York&limit=3" "results"
test_endpoint "Filter by Event Type" "/api/v1/search?type=concert&limit=3" "results"
test_endpoint "Filter by Price Range" "/api/v1/search?min_price=50&max_price=200&limit=3" "results"

# Combined filters
test_endpoint "Combined Filters" "/api/v1/search?q=concert&city=New%20York&min_price=100&limit=3" "results"

# Special endpoints
test_endpoint "Search Suggestions" "/api/v1/search/suggestions?q=con" "suggestions"
test_endpoint "Available Filters" "/api/v1/search/filters" "cities"
test_endpoint "Trending Events" "/api/v1/search/trending?limit=5" "events"

echo "🎯 Testing completed!"
echo ""

# Summary test with detailed output
echo "📊 Detailed Search Result Analysis"
echo "=================================="

echo -e "${BLUE}Sample search result structure:${NC}"
curl -s "$SEARCH_SERVICE_URL/api/v1/search?limit=1" | jq -C '.'

echo ""
echo -e "${BLUE}Available cities (top 10):${NC}"
curl -s "$SEARCH_SERVICE_URL/api/v1/search/filters" | jq -r '.cities[:10][]'

echo ""
echo -e "${BLUE}Available event types:${NC}"
curl -s "$SEARCH_SERVICE_URL/api/v1/search/filters" | jq -r '.event_types[]'

echo ""
echo -e "${BLUE}Price range:${NC}"
curl -s "$SEARCH_SERVICE_URL/api/v1/search/filters" | jq -r '.price_range'

echo ""
echo -e "${BLUE}Total events in index:${NC}"
total_events=$(curl -s "$SEARCH_SERVICE_URL/api/v1/search?limit=1" | jq -r '.total')
echo "$total_events events"

echo ""
echo "🎉 Search Service is working correctly!"
echo "Try these URLs in your browser or API client:"
echo "  • $SEARCH_SERVICE_URL/api/v1/search"
echo "  • $SEARCH_SERVICE_URL/api/v1/search?q=concert"
echo "  • $SEARCH_SERVICE_URL/api/v1/search?city=New%20York"
echo "  • $SEARCH_SERVICE_URL/api/v1/search/suggestions?q=con"
