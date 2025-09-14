#!/usr/bin/env python3

import requests
import json
import sys
from typing import Dict, Any, Optional
import time

class SearchServiceTester:
    def __init__(self, base_url: str = "http://localhost:8003"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.timeout = 10
        
    def test_endpoint(self, name: str, endpoint: str, expected_field: Optional[str] = None) -> Dict[str, Any]:
        """Test a single endpoint and return results"""
        print(f"🔍 Testing: {name}")
        print(f"   URL: {self.base_url}{endpoint}")
        
        try:
            start_time = time.time()
            response = self.session.get(f"{self.base_url}{endpoint}")
            response_time = (time.time() - start_time) * 1000
            
            if response.status_code == 200:
                try:
                    data = response.json()
                    
                    if expected_field and expected_field in data:
                        field_value = data[expected_field]
                        if isinstance(field_value, list):
                            count = len(field_value)
                            print(f"   ✅ SUCCESS - {count} {expected_field} ({response_time:.0f}ms)")
                        else:
                            print(f"   ✅ SUCCESS - {expected_field}: {field_value} ({response_time:.0f}ms)")
                    else:
                        print(f"   ✅ SUCCESS - Valid response ({response_time:.0f}ms)")
                    
                    return {
                        "status": "success",
                        "data": data,
                        "response_time": response_time,
                        "http_code": response.status_code
                    }
                    
                except json.JSONDecodeError:
                    print(f"   ❌ FAILED - Invalid JSON response")
                    return {"status": "failed", "error": "Invalid JSON"}
                    
            else:
                print(f"   ❌ FAILED - HTTP {response.status_code}")
                print(f"   Error: {response.text[:100]}")
                return {"status": "failed", "error": f"HTTP {response.status_code}", "response": response.text[:200]}
                
        except requests.RequestException as e:
            print(f"   ❌ FAILED - Request error: {str(e)[:50]}")
            return {"status": "failed", "error": str(e)}
    
    def check_service_health(self) -> bool:
        """Check if the search service is running"""
        try:
            response = self.session.get(f"{self.base_url}/healthz")
            return response.status_code == 200
        except:
            return False
    
    def run_comprehensive_tests(self):
        """Run all endpoint tests"""
        print("🚀 Search Service Endpoint Testing")
        print("=" * 50)
        print(f"Target URL: {self.base_url}")
        print()
        
        # Check service health
        print("🔍 Checking service health...")
        if not self.check_service_health():
            print(f"❌ Search Service is not running on {self.base_url}")
            print("💡 Start it with: make run SERVICE=search-service")
            return False
        
        print("✅ Search Service is running")
        print()
        
        # Test cases
        test_cases = [
            # Health endpoints
            ("Health Check", "/healthz", "status"),
            ("Readiness Check", "/health/ready", "status"),
            
            # Basic search
            ("Basic Search", "/api/v1/search?limit=5", "results"),
            ("Search with Pagination", "/api/v1/search?page=2&limit=3", "results"),
            
            # Query-based search
            ("Search: Concert", "/api/v1/search?q=concert&limit=3", "results"),
            ("Search: Jazz", "/api/v1/search?q=jazz&limit=3", "results"),
            ("Search: New York", "/api/v1/search?q=New%20York&limit=3", "results"),
            
            # Filter-based search
            ("Filter: City (New York)", "/api/v1/search?city=New%20York&limit=3", "results"),
            ("Filter: City (Los Angeles)", "/api/v1/search?city=Los%20Angeles&limit=3", "results"),
            ("Filter: Type (Concert)", "/api/v1/search?type=concert&limit=3", "results"),
            ("Filter: Type (Sports)", "/api/v1/search?type=sports&limit=3", "results"),
            ("Filter: Price Range", "/api/v1/search?min_price=50&max_price=200&limit=3", "results"),
            
            # Combined filters
            ("Combined: Query + City", "/api/v1/search?q=concert&city=New%20York&limit=3", "results"),
            ("Combined: Type + Price", "/api/v1/search?type=sports&min_price=100&limit=3", "results"),
            
            # Special endpoints
            ("Suggestions: 'con'", "/api/v1/search/suggestions?q=con", "suggestions"),
            ("Suggestions: 'jazz'", "/api/v1/search/suggestions?q=jazz", "suggestions"),
            ("Available Filters", "/api/v1/search/filters", "cities"),
            ("Trending Events", "/api/v1/search/trending?limit=5", "events"),
        ]
        
        results = {}
        successful_tests = 0
        
        for test_name, endpoint, expected_field in test_cases:
            result = self.test_endpoint(test_name, endpoint, expected_field)
            results[test_name] = result
            if result["status"] == "success":
                successful_tests += 1
            print()
        
        # Summary
        total_tests = len(test_cases)
        print("📊 Test Summary")
        print("=" * 30)
        print(f"✅ Successful: {successful_tests}/{total_tests}")
        print(f"❌ Failed: {total_tests - successful_tests}/{total_tests}")
        print()
        
        if successful_tests > 0:
            self.show_sample_data()
        
        return successful_tests == total_tests
    
    def show_sample_data(self):
        """Show sample data from the search service"""
        print("📋 Sample Data Analysis")
        print("=" * 30)
        
        try:
            # Get sample search result
            sample_result = self.test_endpoint("Sample Search", "/api/v1/search?limit=1", None)
            if sample_result["status"] == "success" and "data" in sample_result:
                data = sample_result["data"]
                
                print(f"📈 Total events in index: {data.get('total', 'Unknown')}")
                print(f"🔍 Query time: {data.get('query_time', 'Unknown')}")
                
                if "results" in data and data["results"]:
                    event = data["results"][0]
                    print(f"📍 Sample event: {event.get('name', 'Unknown')}")
                    print(f"🏢 Sample venue: {event.get('venue_name', 'Unknown')} in {event.get('venue_city', 'Unknown')}")
                    print(f"💰 Sample price: ${event.get('base_price', 'Unknown')}")
                
                if "facets" in data:
                    facets = data["facets"]
                    if "cities" in facets:
                        cities_count = len(facets["cities"])
                        print(f"🌆 Available cities: {cities_count}")
                        if cities_count > 0:
                            top_cities = [city["value"] for city in facets["cities"][:5]]
                            print(f"   Top cities: {', '.join(top_cities)}")
                    
                    if "event_types" in facets:
                        types_count = len(facets["event_types"])
                        print(f"🎭 Available event types: {types_count}")
                        if types_count > 0:
                            event_types = [t["value"] for t in facets["event_types"]]
                            print(f"   Types: {', '.join(event_types)}")
                    
                    if "price_range" in facets:
                        price_range = facets["price_range"]
                        print(f"💵 Price range: ${price_range.get('min', 0)} - ${price_range.get('max', 0)}")
            
            print()
            
        except Exception as e:
            print(f"⚠️  Could not analyze sample data: {str(e)[:50]}")
            print()
    
    def performance_test(self, endpoint: str = "/api/v1/search?limit=10", iterations: int = 10):
        """Run a simple performance test"""
        print(f"⚡ Performance Test: {iterations} requests to {endpoint}")
        print("-" * 40)
        
        times = []
        successful = 0
        
        for i in range(iterations):
            try:
                start_time = time.time()
                response = self.session.get(f"{self.base_url}{endpoint}")
                response_time = (time.time() - start_time) * 1000
                
                if response.status_code == 200:
                    times.append(response_time)
                    successful += 1
                    if (i + 1) % 5 == 0:
                        print(f"   Request {i + 1}: {response_time:.0f}ms")
                
            except Exception as e:
                print(f"   Request {i + 1}: Failed - {str(e)[:30]}")
        
        if times:
            avg_time = sum(times) / len(times)
            min_time = min(times)
            max_time = max(times)
            
            print(f"\n📊 Performance Results:")
            print(f"   ✅ Successful requests: {successful}/{iterations}")
            print(f"   ⏱️  Average response time: {avg_time:.0f}ms")
            print(f"   🏃 Fastest response: {min_time:.0f}ms")
            print(f"   🐌 Slowest response: {max_time:.0f}ms")
        
        print()

def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="Test Search Service endpoints")
    parser.add_argument("-u", "--url", default="http://localhost:8003", 
                       help="Search service URL (default: http://localhost:8003)")
    parser.add_argument("-p", "--performance", action="store_true",
                       help="Run performance tests")
    parser.add_argument("--iterations", type=int, default=10,
                       help="Number of performance test iterations (default: 10)")
    
    args = parser.parse_args()
    
    tester = SearchServiceTester(args.url)
    
    # Run comprehensive tests
    success = tester.run_comprehensive_tests()
    
    # Run performance tests if requested
    if args.performance:
        print()
        tester.performance_test(iterations=args.iterations)
    
    # Suggest next steps
    print("🎯 Next Steps:")
    print("=" * 20)
    if success:
        print("✅ All tests passed! The search service is working correctly.")
        print("🔗 Try these URLs in your browser:")
        print(f"   • {args.url}/api/v1/search")
        print(f"   • {args.url}/api/v1/search?q=concert")
        print(f"   • {args.url}/api/v1/search?city=New%20York")
        print(f"   • {args.url}/api/v1/search/suggestions?q=con")
        
        if not args.performance:
            print(f"\n💡 Run performance tests with: {sys.argv[0]} --performance")
    else:
        print("❌ Some tests failed. Check the search service logs.")
        print("🔧 Troubleshooting:")
        print("   • Ensure Elasticsearch is running: make elasticsearch-health")
        print("   • Check search service logs for errors")
        print("   • Verify data was indexed correctly")
    
    return 0 if success else 1

if __name__ == "__main__":
    exit(main())
