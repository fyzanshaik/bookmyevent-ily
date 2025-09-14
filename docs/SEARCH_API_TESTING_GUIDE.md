# Search API Testing Guide

This guide provides comprehensive testing for the BookMyEvent Search API with focus on all endpoints and filters.

## 🚀 Quick Start

### Prerequisites
1. All services running:
   ```bash
   make docker-full-up
   make run SERVICE=user-service &
   make run SERVICE=event-service &  
   make run SERVICE=search-service &
   make run SERVICE=booking-service &
   ```

2. Python 3 with requests library:
   ```bash
   pip3 install requests
   ```

### Testing Options

#### 1. Quick Shell-Based Testing (Immediate Results)
```bash
./quick_search_test.sh
```
- Uses curl for fast testing
- Shows immediate request/response
- Tests all major endpoints
- No test data setup required

#### 2. Step-by-Step Python Testing (Detailed)
```bash
# Test individual components
python3 step_by_step_search_test.py health
python3 step_by_step_search_test.py setup
python3 step_by_step_search_test.py search_basic
python3 step_by_step_search_test.py search_filters
python3 step_by_step_search_test.py cleanup

# Or run all tests
python3 step_by_step_search_test.py all
```

#### 3. Comprehensive Automated Testing
```bash
python3 comprehensive_search_api_test.py
```
- Full automated test suite
- Creates and cleans up test data
- Comprehensive validation
- Detailed reporting

## 📋 Test Coverage

### Health & Status Endpoints
- `GET /healthz` - Basic health check
- `GET /health/ready` - Readiness with dependencies

### Public Search Endpoints
- `GET /api/v1/search` - Main search with all filters
- `GET /api/v1/search/suggestions` - Autocomplete suggestions  
- `GET /api/v1/search/filters` - Available filter metadata
- `GET /api/v1/search/trending` - Trending events

### Internal Endpoints (API Key Required)
- `POST /internal/search/events` - Index event manually
- `DELETE /internal/search/events/{id}` - Remove event from index
- `POST /internal/search/resync` - Full resynchronization

## 🔍 Search Filters Tested

### Text Search
- Query parameter: `q=jazz`
- Multi-field search across name, description, venue
- Relevance scoring

### Location Filters
- City filter: `city=New York`
- State/Country filtering via venue data

### Event Type Filters
- Type filter: `type=concert`
- Supported types: concert, sports, theater, comedy, festival

### Price Range Filters
- Minimum price: `min_price=50`
- Maximum price: `max_price=150`
- Combined range: `min_price=50&max_price=150`

### Date Range Filters
- Start date: `date_from=2024-12-01T00:00:00Z`
- End date: `date_to=2024-12-31T23:59:59Z`
- Combined date range

### Pagination
- Page number: `page=2`
- Results per page: `limit=10`
- Default: page=1, limit=20, max=100

### Combined Filters
Tests complex queries with multiple filters:
```
/api/v1/search?q=music&type=concert&city=New%20York&min_price=50&max_price=150&page=1&limit=10
```

## 🧪 Test Data Creation

The testing scripts create realistic test data via Event API:

### Test Venues
- Madison Square Garden (New York)
- Hollywood Bowl (Los Angeles) 
- Red Rocks Amphitheatre (Morrison)

### Test Events
- **Concerts**: Jazz Night, Rock Concert, Classical Symphony
- **Sports**: Basketball Championship, Soccer World Cup
- **Theater**: Broadway Musical Show
- **Comedy**: Stand-up Comedy Night

### Event Properties Tested
- Various price ranges ($35 - $150)
- Different capacities (500 - 18,000)
- Multiple cities and venues
- Future date ranges
- Published status for search indexing

## 📊 Expected Results

### Successful Responses
```json
{
  "results": [
    {
      "event_id": "uuid",
      "name": "Event Name",
      "description": "Event description",
      "venue_name": "Venue Name",
      "venue_city": "City",
      "event_type": "concert",
      "start_datetime": "2024-12-15T20:00:00Z",
      "base_price": 85.50,
      "available_seats": 500,
      "status": "published",
      "score": 8.5
    }
  ],
  "total": 10,
  "page": 1,
  "limit": 20,
  "query_time": "45ms",
  "facets": {
    "cities": [{"value": "New York", "count": 5}],
    "event_types": [{"value": "concert", "count": 3}],
    "price_range": {"min": 35.0, "max": 150.0}
  }
}
```

### Error Responses
- `400 Bad Request` - Invalid parameters
- `500 Internal Server Error` - Elasticsearch/Redis issues
- `401 Unauthorized` - Missing API key for internal endpoints

## 🔧 Troubleshooting

### Common Issues

#### 1. Search Service Not Running
```
Error: Cannot connect to search service
Solution: make run SERVICE=search-service
```

#### 2. Elasticsearch Not Available
```
Error: Elasticsearch health check failed
Solution: make docker-full-up (starts Elasticsearch)
```

#### 3. No Search Results
```
Issue: Empty results array
Cause: No events indexed yet
Solution: 
1. Run setup: python3 step_by_step_search_test.py setup
2. Wait for indexing: sleep 5
3. Try search again
```

#### 4. Internal API Authentication
```
Error: 401 Unauthorized for internal endpoints
Solution: Check INTERNAL_API_KEY in test scripts matches service config
```

### Debug Commands

#### Check Elasticsearch Health
```bash
curl -s http://localhost:9200/_cluster/health | jq
curl -s http://localhost:9200/_cat/indices?v
```

#### Check Redis Connection
```bash
docker exec -it evently_redis redis-cli ping
```

#### Check Service Logs
```bash
# Check search service logs
make run SERVICE=search-service  # Shows logs in console

# Check Elasticsearch logs  
docker logs evently_elasticsearch --tail 50
```

#### Manual Search Test
```bash
# Basic search
curl -s "http://localhost:8003/api/v1/search" | jq

# Search with filters
curl -s "http://localhost:8003/api/v1/search?q=concert&city=New%20York" | jq

# Check suggestions
curl -s "http://localhost:8003/api/v1/search/suggestions?q=jazz" | jq
```

## 📈 Performance Testing

### Expected Response Times
- Basic search: < 100ms
- Filtered search: < 150ms  
- Suggestions: < 50ms
- Trending events: < 200ms

### Load Testing
```bash
# Simple load test with curl
for i in {1..10}; do
  time curl -s "http://localhost:8003/api/v1/search?q=concert" > /dev/null &
done
wait
```

## 🎯 Test Scenarios

### Scenario 1: User Searching for Events
1. User visits search page
2. Gets filter options: `GET /api/v1/search/filters`
3. Performs basic search: `GET /api/v1/search?q=music`
4. Refines with filters: `GET /api/v1/search?q=music&city=New%20York&type=concert`
5. Gets suggestions while typing: `GET /api/v1/search/suggestions?q=jaz`

### Scenario 2: Event Discovery
1. User browses trending: `GET /api/v1/search/trending`
2. Filters by location: `GET /api/v1/search?city=Los%20Angeles`
3. Filters by price range: `GET /api/v1/search?city=Los%20Angeles&min_price=50&max_price=100`
4. Paginates through results: `GET /api/v1/search?city=Los%20Angeles&page=2`

### Scenario 3: Admin Event Management
1. Admin creates event via Event API
2. Event auto-indexes in Search Service
3. Verify event appears in search: `GET /api/v1/search?q=new_event_name`
4. Admin updates event
5. Verify updated data in search results

## 🔄 CI/CD Integration

### Automated Testing Script
```bash
#!/bin/bash
# ci_search_test.sh

set -e

echo "Starting Search API Tests..."

# Start services
make docker-full-up
make run SERVICE=search-service &
SEARCH_PID=$!

# Wait for services
sleep 10

# Run tests
python3 comprehensive_search_api_test.py

# Cleanup
kill $SEARCH_PID
make docker-down

echo "Search API Tests Completed Successfully!"
```

### Test Exit Codes
- `0` - All tests passed
- `1` - Some tests failed
- `2` - Service unavailable

## 📝 Manual Testing Checklist

- [ ] Health endpoints respond correctly
- [ ] Basic search returns results
- [ ] Text search filters correctly
- [ ] City filter works
- [ ] Event type filter works  
- [ ] Price range filter works
- [ ] Date range filter works
- [ ] Combined filters work
- [ ] Pagination works correctly
- [ ] Suggestions endpoint works
- [ ] Empty query suggestions fail appropriately
- [ ] Filters metadata endpoint works
- [ ] Trending events endpoint works
- [ ] Internal resync works with API key
- [ ] Internal indexing works with API key
- [ ] Error handling works correctly
- [ ] Response times are acceptable
- [ ] Facets data is correct
- [ ] Cache headers are present
- [ ] Search relevance scoring works

## 🎉 Success Criteria

### Functional Requirements
✅ All public endpoints return 200 OK  
✅ All filters work correctly  
✅ Search results are relevant and accurate  
✅ Pagination works properly  
✅ Error handling is appropriate  
✅ Internal endpoints require authentication  

### Performance Requirements  
✅ Search response time < 200ms  
✅ Suggestions response time < 100ms  
✅ System handles concurrent requests  
✅ Elasticsearch indexing works correctly  

### Data Quality Requirements
✅ Search results match filter criteria  
✅ Facets reflect actual data  
✅ Event data is complete and accurate  
✅ Cache invalidation works properly
