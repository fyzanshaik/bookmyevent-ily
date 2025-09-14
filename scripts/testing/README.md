# Search Service Testing Scripts

## 🚀 Quick Testing

After populating the search service with data, use these scripts to test all endpoints:

### Option 1: Bash Script (Simple)
```bash
# Test all endpoints with colored output
./scripts/testing/test-search-endpoints.sh
```

### Option 2: Python Script (Comprehensive)
```bash
# Basic endpoint testing
./scripts/testing/test-search-endpoints.py

# Include performance testing
./scripts/testing/test-search-endpoints.py --performance

# Custom URL and more performance iterations
./scripts/testing/test-search-endpoints.py -u http://localhost:8003 --performance --iterations 20
```

## 📊 What Gets Tested

### Health Endpoints
- `GET /healthz` - Basic health check
- `GET /health/ready` - Readiness with dependencies

### Search Endpoints
- **Basic Search**: `GET /api/v1/search?limit=5`
- **Pagination**: `GET /api/v1/search?page=2&limit=3`
- **Text Search**: `GET /api/v1/search?q=concert`
- **City Filter**: `GET /api/v1/search?city=New York`
- **Type Filter**: `GET /api/v1/search?type=sports`
- **Price Filter**: `GET /api/v1/search?min_price=50&max_price=200`
- **Combined Filters**: `GET /api/v1/search?q=concert&city=New York&min_price=100`

### Special Endpoints
- **Suggestions**: `GET /api/v1/search/suggestions?q=con`
- **Filters**: `GET /api/v1/search/filters`
- **Trending**: `GET /api/v1/search/trending`

## 🔧 Troubleshooting

If tests fail:

1. **Check if search service is running**:
   ```bash
   curl http://localhost:8003/healthz
   ```

2. **Restart search service with fixed code**:
   ```bash
   pkill -f search-service
   make build SERVICE=search-service
   make run SERVICE=search-service
   ```

3. **Check Elasticsearch health**:
   ```bash
   make elasticsearch-health
   ```

4. **Verify data exists**:
   ```bash
   curl http://localhost:9200/events/_count
   ```

## 📈 Expected Results

- ✅ **All endpoints return HTTP 200**
- ✅ **JSON responses are valid**
- ✅ **Search results contain expected fields**
- ✅ **Response times < 200ms**
- ✅ **Facets show proper aggregations**
- ✅ **Suggestions work for autocomplete**

## 🎯 Performance Expectations

With 10,000 events:
- **Search response time**: 50-150ms
- **Concurrent capacity**: 100+ requests/second
- **Index size**: ~5-10MB
- **Memory usage**: ~100-200MB for Elasticsearch

## 🔗 Manual Testing URLs

After running the scripts, you can manually test these URLs:

```
http://localhost:8003/api/v1/search
http://localhost:8003/api/v1/search?q=concert
http://localhost:8003/api/v1/search?city=New%20York
http://localhost:8003/api/v1/search?type=sports&min_price=50
http://localhost:8003/api/v1/search/suggestions?q=con
http://localhost:8003/api/v1/search/filters
http://localhost:8003/api/v1/search/trending
```

## 🎉 Success Criteria

The search service is working correctly if:
- [x] Data generation completed (10,000 events)
- [x] All endpoint tests pass
- [x] Search returns relevant results
- [x] Filters work correctly
- [x] Suggestions provide autocomplete
- [x] Response times are acceptable
- [x] No panic errors in logs