# Search Service Testing Guide

## 🚀 Quick Start

### 1. Environment Setup

Add these to your `.env` file:
```env
# Search Service Configuration
SEARCH_SERVICE_PORT=8003
ELASTICSEARCH_URL=http://localhost:9200
ELASTICSEARCH_INDEX_NAME=events
REDIS_URL=redis://localhost:6380
EVENT_SERVICE_URL=http://localhost:8002
INTERNAL_API_KEY=your-internal-api-key
```

### 2. Start Infrastructure

```bash
# Start all infrastructure (PostgreSQL + Redis + Elasticsearch)
make docker-full-up

# Wait for services to be ready (especially Elasticsearch takes ~30 seconds)
sleep 30

# Check Elasticsearch health
make elasticsearch-health
```

### 3. Build and Run Search Service

```bash
# Build the service
make build SERVICE=search-service

# Run the service
make run SERVICE=search-service
```

The service will start on port 8003 and automatically create the Elasticsearch index.

### 4. Generate Test Data

We have two data generators available:

#### Option A: Python Generator (Recommended)
```bash
# Generate 10,000 events (default)
./scripts/testing/search-service-data-generator.py

# Generate 1,000 events for quick testing
./scripts/testing/search-service-data-generator.py -n 1000

# Custom configuration
./scripts/testing/search-service-data-generator.py -u http://localhost:8003 -k your-api-key -n 5000 -b 100
```

#### Option B: Bash Generator
```bash
# Generate 10,000 events
./scripts/testing/search-service-data-generator.sh

# Generate 1,000 events
./scripts/testing/search-service-data-generator.sh -n 1000
```

## 📊 Generated Data Characteristics

The data generator creates realistic events with:

### Event Types
- **Concerts**: "The Midnight Stars Live in Concert", "Jazz Fusion Collective Live in Concert"
- **Sports**: "New York Basketball Championship", "Chicago Soccer Finals"
- **Theater**: "Hamilton - Broadway Musical", "The Lion King - Broadway Musical"
- **Conferences**: "Tech Innovation Summit 2024", "AI & Machine Learning Expo 2024"
- **Comedy**: "Stand-Up Comedy Night", "Improv Show"
- **Festivals**: "Music Festival", "Food & Wine Festival"
- **Workshops**: "Photography Workshop", "Cooking Class"
- **Exhibitions**: "Art Exhibition", "Science Exhibition"

### Geographic Distribution
- **34 Major US Cities**: New York, Los Angeles, Chicago, Houston, etc.
- **Realistic Venues**: "New York Arena", "Chicago Theater", "Houston Stadium"
- **Proper Addresses**: "1234 Broadway", "5678 Main St"

### Realistic Pricing
- **Concert/Festival**: $95-500
- **Sports**: $45-350
- **Theater**: $75-250
- **Conference**: $125-500
- **Comedy/Workshop**: $25-95

### Time Distribution
- **Next 6 months**: Events spread across upcoming dates
- **Realistic times**: 10 AM - 10 PM start times
- **Duration**: 1-4 hours per event

### Capacity Ranges
- **Stadiums/Arenas**: 20,000-70,000 capacity
- **Theaters/Auditoriums**: 500-3,000 capacity
- **Clubs**: 100-800 capacity
- **Other venues**: 1,000-15,000 capacity

## 🔍 Testing Search Functionality

### Basic Search Endpoints

```bash
# Health check
curl http://localhost:8003/healthz

# Basic search (returns 20 events by default)
curl "http://localhost:8003/api/v1/search"

# Search with query
curl "http://localhost:8003/api/v1/search?q=concert"
curl "http://localhost:8003/api/v1/search?q=jazz"
curl "http://localhost:8003/api/v1/search?q=New York"

# Filter by city
curl "http://localhost:8003/api/v1/search?city=New York"
curl "http://localhost:8003/api/v1/search?city=Los Angeles"

# Filter by event type
curl "http://localhost:8003/api/v1/search?type=concert"
curl "http://localhost:8003/api/v1/search?type=sports"

# Price filtering
curl "http://localhost:8003/api/v1/search?min_price=50&max_price=200"

# Date filtering
curl "http://localhost:8003/api/v1/search?date_from=2024-01-01&date_to=2024-06-30"

# Combined filters
curl "http://localhost:8003/api/v1/search?q=concert&city=New York&min_price=100"

# Pagination
curl "http://localhost:8003/api/v1/search?page=2&limit=10"
```

### Advanced Features

```bash
# Autocomplete suggestions
curl "http://localhost:8003/api/v1/search/suggestions?q=con"
curl "http://localhost:8003/api/v1/search/suggestions?q=jazz"

# Available filters
curl "http://localhost:8003/api/v1/search/filters"

# Trending events
curl "http://localhost:8003/api/v1/search/trending"
curl "http://localhost:8003/api/v1/search/trending?limit=20"
```

## 🧪 Expected Test Results

### Search Response Format
```json
{
  "results": [
    {
      "event_id": "uuid",
      "name": "The Midnight Stars Live in Concert",
      "description": "Join us for an unforgettable experience...",
      "venue_name": "New York Arena",
      "venue_city": "New York",
      "venue_address": "1234 Broadway",
      "event_type": "concert",
      "start_datetime": "2024-03-15T20:00:00Z",
      "end_datetime": "2024-03-15T23:00:00Z",
      "base_price": 125.00,
      "available_seats": 15000,
      "status": "published",
      "score": 1.0
    }
  ],
  "total": 1250,
  "page": 1,
  "limit": 20,
  "query_time": "45ms",
  "facets": {
    "cities": [
      {"value": "New York", "count": 320},
      {"value": "Los Angeles", "count": 285}
    ],
    "event_types": [
      {"value": "concert", "count": 1800},
      {"value": "sports", "count": 1500}
    ],
    "price_range": {
      "min": 25.0,
      "max": 500.0
    }
  }
}
```

### Performance Expectations
- **Search Response Time**: <200ms for most queries
- **Index Creation**: ~30 seconds for 10,000 events
- **Data Generation**: ~50-100 events/second
- **Concurrent Search**: Should handle 100+ concurrent requests

## 🐛 Troubleshooting

### Common Issues

1. **Elasticsearch not starting**
   ```bash
   # Check Elasticsearch logs
   make elasticsearch-logs
   
   # Check if port 9200 is available
   lsof -i :9200
   ```

2. **Search service fails to start**
   ```bash
   # Check if Elasticsearch is healthy
   curl http://localhost:9200/_cluster/health
   
   # Check if Redis is running
   redis-cli ping
   ```

3. **Data generator fails**
   ```bash
   # Check if search service is running
   curl http://localhost:8003/healthz
   
   # Verify INTERNAL_API_KEY is set correctly
   echo $INTERNAL_API_KEY
   ```

4. **No search results**
   ```bash
   # Check if index exists
   curl http://localhost:9200/_cat/indices
   
   # Check index document count
   curl http://localhost:9200/events/_count
   ```

### Useful Debug Commands

```bash
# Check Elasticsearch cluster health
make elasticsearch-health

# List all indices
make elasticsearch-indices

# View Elasticsearch logs
make elasticsearch-logs

# Connect to Redis
make redis-cli

# Check search service logs
# (when running with make run SERVICE=search-service)
```

## 📈 Performance Testing

### Load Testing Search Endpoints
```bash
# Install Apache Bench for load testing
sudo apt-get install apache2-utils  # Ubuntu/Debian
# or
brew install httpie  # macOS

# Test basic search performance
ab -n 1000 -c 10 "http://localhost:8003/api/v1/search"

# Test search with query
ab -n 500 -c 5 "http://localhost:8003/api/v1/search?q=concert"
```

### Expected Performance
- **10,000 events**: Search should complete in <100ms
- **100,000 events**: Search should complete in <200ms
- **Concurrent users**: Should handle 100+ concurrent searches

## 🎯 Success Criteria

✅ **Infrastructure**: Elasticsearch + Redis running healthy  
✅ **Service**: Search service starts without errors  
✅ **Data**: 10,000+ events indexed successfully  
✅ **Search**: Basic text search returns relevant results  
✅ **Filters**: City, type, price, date filters work correctly  
✅ **Facets**: Aggregations return proper counts  
✅ **Performance**: Search responds in <200ms  
✅ **Suggestions**: Autocomplete returns relevant suggestions  
✅ **Pagination**: Page navigation works correctly  

## 🔗 Integration with Event Service

The search service is designed to receive data from the Event Service via CDC:

```bash
# These internal endpoints are available for CDC integration:
# POST /internal/search/events - Index a single event
# DELETE /internal/search/events/{id} - Delete an event
# POST /internal/search/resync - Full resync from Event Service
```

Once the Event Service CDC integration is implemented, events will be automatically indexed when created/updated in the Event Service.

## 🎉 Next Steps

After successful testing:
1. **Event Service Integration**: Implement CDC calls from Event Service
2. **Production Optimization**: Tune Elasticsearch settings for production
3. **Monitoring**: Add metrics and alerting
4. **Caching**: Optimize Redis caching strategies
5. **Advanced Search**: Add geo-location search, more sophisticated ranking
