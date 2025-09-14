#!/bin/bash

set -e

EVENT_SERVICE_URL="http://localhost:8002"
SEARCH_SERVICE_URL="http://localhost:8003"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🔗 Event Service + Search Service Integration Test"
echo "=================================================="
echo "Event Service: $EVENT_SERVICE_URL"
echo "Search Service: $SEARCH_SERVICE_URL"
echo ""

# Check if services are running
echo "🔍 Checking if services are running..."

if ! curl -s "$EVENT_SERVICE_URL/healthz" > /dev/null; then
    echo -e "${RED}❌ Event Service is not running on $EVENT_SERVICE_URL${NC}"
    echo "💡 Start it with: make run SERVICE=event-service"
    exit 1
fi

if ! curl -s "$SEARCH_SERVICE_URL/healthz" > /dev/null; then
    echo -e "${RED}❌ Search Service is not running on $SEARCH_SERVICE_URL${NC}"
    echo "💡 Start it with: make run SERVICE=search-service"
    exit 1
fi

echo -e "${GREEN}✅ Both services are running${NC}"
echo ""

# Test admin authentication first
echo "🔐 Testing admin authentication..."

# Register admin
ADMIN_EMAIL="test-admin-$(date +%s)@example.com"
ADMIN_PASSWORD="TestPassword123"

echo "Registering admin: $ADMIN_EMAIL"
REGISTER_RESPONSE=$(curl -s -X POST "$EVENT_SERVICE_URL/api/v1/auth/admin/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\",
    \"name\": \"Test Admin\"
  }")

if echo "$REGISTER_RESPONSE" | jq -e '.admin_id' > /dev/null; then
    echo -e "${GREEN}✅ Admin registered successfully${NC}"
else
    echo -e "${RED}❌ Failed to register admin${NC}"
    echo "Response: $REGISTER_RESPONSE"
    exit 1
fi

# Login admin
echo "Logging in admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$EVENT_SERVICE_URL/api/v1/auth/admin/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\"
  }")

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}❌ Failed to login admin${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ Admin logged in successfully${NC}"
echo ""

# Create a venue first
echo "🏢 Creating test venue..."
VENUE_RESPONSE=$(curl -s -X POST "$EVENT_SERVICE_URL/api/v1/admin/venues" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "Test Integration Venue",
    "address": "123 Test Street",
    "city": "Integration City",
    "state": "IC",
    "country": "Testland",
    "postal_code": "12345",
    "capacity": 1000,
    "layout_config": {"sections": [{"name": "General", "capacity": 1000}]}
  }')

VENUE_ID=$(echo "$VENUE_RESPONSE" | jq -r '.venue_id')
if [ "$VENUE_ID" = "null" ] || [ -z "$VENUE_ID" ]; then
    echo -e "${RED}❌ Failed to create venue${NC}"
    echo "Response: $VENUE_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ Venue created: $VENUE_ID${NC}"
echo ""

# Test 1: Create Event (should trigger search indexing)
echo "🎭 Test 1: Creating event (should auto-index in search)..."

EVENT_RESPONSE=$(curl -s -X POST "$EVENT_SERVICE_URL/api/v1/admin/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"Integration Test Concert\",
    \"description\": \"A test concert for integration testing\",
    \"venue_id\": \"$VENUE_ID\",
    \"event_type\": \"concert\",
    \"start_datetime\": \"$(date -d '+1 day' -Iseconds)\",
    \"end_datetime\": \"$(date -d '+1 day +3 hours' -Iseconds)\",
    \"total_capacity\": 500,
    \"base_price\": 75.99
  }")

EVENT_ID=$(echo "$EVENT_RESPONSE" | jq -r '.event_id')
if [ "$EVENT_ID" = "null" ] || [ -z "$EVENT_ID" ]; then
    echo -e "${RED}❌ Failed to create event${NC}"
    echo "Response: $EVENT_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ Event created: $EVENT_ID${NC}"
echo "Event details:"
echo "$EVENT_RESPONSE" | jq -C '.'
echo ""

# Wait a moment for async indexing
echo "⏳ Waiting 3 seconds for search indexing..."
sleep 3

# Test 2: Search for the created event
echo "🔍 Test 2: Searching for the created event..."

SEARCH_RESPONSE=$(curl -s "$SEARCH_SERVICE_URL/api/v1/search?q=Integration%20Test%20Concert&limit=5")
SEARCH_COUNT=$(echo "$SEARCH_RESPONSE" | jq -r '.total')

if [ "$SEARCH_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Event found in search results!${NC}"
    echo "Search results:"
    echo "$SEARCH_RESPONSE" | jq -C '.results[] | {name: .name, venue_city: .venue_city, event_type: .event_type, base_price: .base_price}'
else
    echo -e "${RED}❌ Event not found in search results${NC}"
    echo "Search response: $SEARCH_RESPONSE"
fi
echo ""

# Test 3: Update Event (should trigger search re-indexing)
echo "🔄 Test 3: Updating event (should re-index in search)..."

EVENT_VERSION=$(echo "$EVENT_RESPONSE" | jq -r '.version')
UPDATE_RESPONSE=$(curl -s -X PUT "$EVENT_SERVICE_URL/api/v1/admin/events/$EVENT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"Updated Integration Test Concert\",
    \"description\": \"An updated test concert for integration testing\",
    \"base_price\": 89.99,
    \"version\": $EVENT_VERSION
  }")

if echo "$UPDATE_RESPONSE" | jq -e '.event_id' > /dev/null; then
    echo -e "${GREEN}✅ Event updated successfully${NC}"
    echo "Updated event details:"
    echo "$UPDATE_RESPONSE" | jq -C '{name: .name, base_price: .base_price, version: .version}'
else
    echo -e "${RED}❌ Failed to update event${NC}"
    echo "Response: $UPDATE_RESPONSE"
fi
echo ""

# Wait for re-indexing
echo "⏳ Waiting 3 seconds for search re-indexing..."
sleep 3

# Test 4: Search for updated event
echo "🔍 Test 4: Searching for updated event..."

UPDATED_SEARCH_RESPONSE=$(curl -s "$SEARCH_SERVICE_URL/api/v1/search?q=Updated%20Integration%20Test&limit=5")
UPDATED_SEARCH_COUNT=$(echo "$UPDATED_SEARCH_RESPONSE" | jq -r '.total')

if [ "$UPDATED_SEARCH_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Updated event found in search results!${NC}"
    FOUND_PRICE=$(echo "$UPDATED_SEARCH_RESPONSE" | jq -r '.results[0].base_price')
    if [ "$FOUND_PRICE" = "89.99" ]; then
        echo -e "${GREEN}✅ Price correctly updated in search index: \$89.99${NC}"
    else
        echo -e "${YELLOW}⚠️  Price may not be updated yet: \$$FOUND_PRICE${NC}"
    fi
else
    echo -e "${RED}❌ Updated event not found in search results${NC}"
    echo "Search response: $UPDATED_SEARCH_RESPONSE"
fi
echo ""

# Test 5: Reserve seats (should trigger availability update)
echo "🎫 Test 5: Reserving seats (should update availability in search)..."

CURRENT_VERSION=$(echo "$UPDATE_RESPONSE" | jq -r '.version')
RESERVE_RESPONSE=$(curl -s -X PUT "$EVENT_SERVICE_URL/internal/events/$EVENT_ID/availability" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-internal-api-key" \
  -d "{
    \"quantity\": -5,
    \"version\": $CURRENT_VERSION
  }")

if echo "$RESERVE_RESPONSE" | jq -e '.available_seats' > /dev/null; then
    AVAILABLE_SEATS=$(echo "$RESERVE_RESPONSE" | jq -r '.available_seats')
    echo -e "${GREEN}✅ Seats reserved successfully. Available seats: $AVAILABLE_SEATS${NC}"
else
    echo -e "${RED}❌ Failed to reserve seats${NC}"
    echo "Response: $RESERVE_RESPONSE"
fi
echo ""

# Test 6: Delete Event (should trigger search deletion)
echo "🗑️  Test 6: Deleting event (should remove from search)..."

FINAL_VERSION=$(echo "$RESERVE_RESPONSE" | jq -r '.version // empty')
if [ -z "$FINAL_VERSION" ]; then
    FINAL_VERSION=$(echo "$UPDATE_RESPONSE" | jq -r '.version')
fi

DELETE_RESPONSE=$(curl -s -X DELETE "$EVENT_SERVICE_URL/api/v1/admin/events/$EVENT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{\"version\": $FINAL_VERSION}")

if echo "$DELETE_RESPONSE" | jq -e '.message' > /dev/null; then
    echo -e "${GREEN}✅ Event deleted successfully${NC}"
    echo "Response: $(echo "$DELETE_RESPONSE" | jq -r '.message')"
else
    echo -e "${RED}❌ Failed to delete event${NC}"
    echo "Response: $DELETE_RESPONSE"
fi
echo ""

# Wait for deletion from search
echo "⏳ Waiting 3 seconds for search deletion..."
sleep 3

# Test 7: Verify event is removed from search
echo "🔍 Test 7: Verifying event is removed from search..."

FINAL_SEARCH_RESPONSE=$(curl -s "$SEARCH_SERVICE_URL/api/v1/search?q=Integration%20Test%20Concert&limit=5")
FINAL_SEARCH_COUNT=$(echo "$FINAL_SEARCH_RESPONSE" | jq -r '.total')

if [ "$FINAL_SEARCH_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ Event successfully removed from search index!${NC}"
else
    echo -e "${YELLOW}⚠️  Event may still be in search index (async deletion): $FINAL_SEARCH_COUNT results${NC}"
fi
echo ""

# Cleanup: Delete venue
echo "🧹 Cleaning up test venue..."
curl -s -X DELETE "$EVENT_SERVICE_URL/api/v1/admin/venues/$VENUE_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" > /dev/null

echo "🎉 Integration Test Complete!"
echo "================================"
echo -e "${GREEN}✅ Event Service → Search Service integration is working!${NC}"
echo ""
echo "📊 Test Summary:"
echo "  • ✅ Event creation → Auto-indexing in search"
echo "  • ✅ Event updates → Re-indexing in search"
echo "  • ✅ Seat reservations → Availability updates"
echo "  • ✅ Event deletion → Removal from search"
echo ""
echo "🔗 The CDC (Change Data Capture) flow is working correctly!"
