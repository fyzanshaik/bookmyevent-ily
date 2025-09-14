#!/usr/bin/env python3
"""
Comprehensive Search API Testing Suite for BookMyEvent
Focuses on Search Service API with thorough testing of all filters and endpoints

This script will:
1. Create test data via Event API (admin authentication)
2. Test all Search API endpoints with various filters
3. Verify search functionality with different parameters
4. Test internal search endpoints (indexing, deletion)
5. Clean up test data after testing

Run with: python3 comprehensive_search_api_test.py
"""

import requests
import json
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
import sys
import os
from urllib.parse import urlencode

# Configuration
BASE_URLS = {
    'user': 'http://localhost:8001',
    'event': 'http://localhost:8002', 
    'search': 'http://localhost:8003',
    'booking': 'http://localhost:8004'
}

INTERNAL_API_KEY = 'internal-service-communication-key-change-in-production'

class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    PURPLE = '\033[95m'
    CYAN = '\033[96m'
    WHITE = '\033[97m'
    BOLD = '\033[1m'
    END = '\033[0m'

class TestResults:
    def __init__(self):
        self.total = 0
        self.passed = 0
        self.failed = 0
        self.results = []

    def add_result(self, test_name: str, passed: bool, details: str = ""):
        self.total += 1
        if passed:
            self.passed += 1
            status = f"{Colors.GREEN}✓ PASS{Colors.END}"
        else:
            self.failed += 1
            status = f"{Colors.RED}✗ FAIL{Colors.END}"
        
        result = f"{status} {test_name}"
        if details:
            result += f" - {details}"
        
        print(result)
        self.results.append((test_name, passed, details))

    def print_summary(self):
        print(f"\n{Colors.BOLD}{'='*60}{Colors.END}")
        print(f"{Colors.BOLD}TEST SUMMARY{Colors.END}")
        print(f"{Colors.BOLD}{'='*60}{Colors.END}")
        print(f"Total Tests: {self.total}")
        print(f"{Colors.GREEN}Passed: {self.passed}{Colors.END}")
        print(f"{Colors.RED}Failed: {self.failed}{Colors.END}")
        
        if self.failed > 0:
            print(f"\n{Colors.RED}FAILED TESTS:{Colors.END}")
            for name, passed, details in self.results:
                if not passed:
                    print(f"  - {name}: {details}")

def make_request(method: str, url: str, headers: Dict = None, data: Dict = None, params: Dict = None) -> tuple:
    """Make HTTP request and return (success, response, status_code)"""
    try:
        if headers is None:
            headers = {'Content-Type': 'application/json'}
        
        kwargs = {
            'headers': headers,
            'timeout': 30
        }
        
        if data:
            kwargs['json'] = data
        
        if params:
            kwargs['params'] = params
            
        response = requests.request(method, url, **kwargs)
        
        try:
            json_data = response.json()
        except:
            json_data = response.text
            
        return response.status_code < 400, json_data, response.status_code
    except Exception as e:
        return False, str(e), 0

class SearchAPITester:
    def __init__(self):
        self.results = TestResults()
        self.admin_token = None
        self.test_venues = []
        self.test_events = []
        
    def run_all_tests(self):
        """Run comprehensive search API tests"""
        print(f"{Colors.BOLD}{Colors.BLUE}Starting Comprehensive Search API Testing{Colors.END}")
        print(f"{Colors.BOLD}{'='*60}{Colors.END}\n")
        
        # Step 1: Health checks
        self.test_health_endpoints()
        
        # Step 2: Setup test data
        if not self.setup_test_data():
            print(f"{Colors.RED}Failed to setup test data. Aborting tests.{Colors.END}")
            return
            
        # Step 3: Wait for search indexing
        print(f"{Colors.YELLOW}Waiting 5 seconds for search indexing...{Colors.END}")
        time.sleep(5)
        
        # Step 4: Test search endpoints
        self.test_basic_search()
        self.test_search_filters()
        self.test_search_suggestions()
        self.test_search_metadata()
        self.test_trending_events()
        
        # Step 5: Test internal endpoints
        self.test_internal_endpoints()
        
        # Step 6: Cleanup
        self.cleanup_test_data()
        
        # Step 7: Print results
        self.results.print_summary()
        
        return self.results.failed == 0

    def test_health_endpoints(self):
        """Test health and readiness endpoints"""
        print(f"{Colors.CYAN}Testing Health Endpoints{Colors.END}")
        
        # Test basic health
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/healthz")
        self.results.add_result(
            "Search Service Health Check",
            success and status == 200,
            f"Status: {status}, Response: {response}"
        )
        
        # Test readiness
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/health/ready")
        self.results.add_result(
            "Search Service Readiness Check", 
            success and status == 200,
            f"Status: {status}, Elasticsearch: {response.get('elasticsearch') if isinstance(response, dict) else 'unknown'}"
        )

    def setup_test_data(self) -> bool:
        """Setup test data via Event API"""
        print(f"{Colors.CYAN}Setting up Test Data via Event API{Colors.END}")
        
        # Step 1: Create admin account
        if not self.create_admin_account():
            return False
            
        # Step 2: Create test venues
        if not self.create_test_venues():
            return False
            
        # Step 3: Create diverse test events
        if not self.create_test_events():
            return False
            
        return True

    def create_admin_account(self) -> bool:
        """Create admin account for testing"""
        admin_data = {
            "email": f"test-admin-{uuid.uuid4()}@test.com",
            "password": "testpass123",
            "name": "Test Admin",
            "phone_number": "+1234567890",
            "role": "event_manager"
        }
        
        success, response, status = make_request(
            'POST',
            f"{BASE_URLS['event']}/api/v1/auth/admin/register",
            data=admin_data
        )
        
        if success and isinstance(response, dict) and 'access_token' in response:
            self.admin_token = response['access_token']
            self.results.add_result("Admin Account Creation", True, "Admin registered successfully")
            return True
        else:
            self.results.add_result("Admin Account Creation", False, f"Status: {status}, Response: {response}")
            return False

    def create_test_venues(self) -> bool:
        """Create test venues for events"""
        venues_data = [
            {
                "name": "Madison Square Garden",
                "address": "4 Pennsylvania Plaza",
                "city": "New York",
                "state": "NY", 
                "country": "USA",
                "postal_code": "10001",
                "capacity": 20000
            },
            {
                "name": "Hollywood Bowl",
                "address": "2301 N Highland Ave",
                "city": "Los Angeles",
                "state": "CA",
                "country": "USA", 
                "postal_code": "90068",
                "capacity": 17500
            },
            {
                "name": "Red Rocks Amphitheatre",
                "address": "18300 W Alameda Pkwy",
                "city": "Morrison",
                "state": "CO",
                "country": "USA",
                "postal_code": "80465", 
                "capacity": 9525
            }
        ]
        
        headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {self.admin_token}'
        }
        
        for venue_data in venues_data:
            success, response, status = make_request(
                'POST',
                f"{BASE_URLS['event']}/api/v1/admin/venues",
                headers=headers,
                data=venue_data
            )
            
            if success and isinstance(response, dict) and 'venue_id' in response:
                self.test_venues.append(response)
                self.results.add_result(f"Venue Creation: {venue_data['name']}", True)
            else:
                self.results.add_result(f"Venue Creation: {venue_data['name']}", False, f"Status: {status}")
                return False
                
        return len(self.test_venues) == len(venues_data)

    def create_test_events(self) -> bool:
        """Create diverse test events for comprehensive search testing"""
        base_date = datetime.now() + timedelta(days=30)
        
        events_data = [
            # Concert events
            {
                "name": "Jazz Night at Madison Square",
                "description": "An evening of smooth jazz with renowned artists",
                "venue_id": self.test_venues[0]['venue_id'],
                "event_type": "concert",
                "start_datetime": (base_date + timedelta(days=1)).isoformat() + "Z",
                "end_datetime": (base_date + timedelta(days=1, hours=3)).isoformat() + "Z", 
                "total_capacity": 500,
                "base_price": 85.50,
                "max_tickets_per_booking": 8
            },
            {
                "name": "Rock Concert Extravaganza",
                "description": "High-energy rock music festival",
                "venue_id": self.test_venues[1]['venue_id'],
                "event_type": "concert", 
                "start_datetime": (base_date + timedelta(days=5)).isoformat() + "Z",
                "end_datetime": (base_date + timedelta(days=5, hours=4)).isoformat() + "Z",
                "total_capacity": 15000,
                "base_price": 125.00,
                "max_tickets_per_booking": 6
            },
            {
                "name": "Classical Music Symphony",
                "description": "Beautiful classical music performance",
                "venue_id": self.test_venues[2]['venue_id'],
                "event_type": "concert",
                "start_datetime": (base_date + timedelta(days=10)).isoformat() + "Z", 
                "end_datetime": (base_date + timedelta(days=10, hours=2)).isoformat() + "Z",
                "total_capacity": 2000,
                "base_price": 45.00,
                "max_tickets_per_booking": 4
            },
            # Sports events
            {
                "name": "Basketball Championship Game",
                "description": "Exciting basketball playoff game",
                "venue_id": self.test_venues[0]['venue_id'],
                "event_type": "sports",
                "start_datetime": (base_date + timedelta(days=15)).isoformat() + "Z",
                "end_datetime": (base_date + timedelta(days=15, hours=3)).isoformat() + "Z",
                "total_capacity": 18000,
                "base_price": 75.00,
                "max_tickets_per_booking": 10
            },
            {
                "name": "Soccer World Cup Match",
                "description": "International soccer championship",
                "venue_id": self.test_venues[1]['venue_id'], 
                "event_type": "sports",
                "start_datetime": (base_date + timedelta(days=20)).isoformat() + "Z",
                "end_datetime": (base_date + timedelta(days=20, hours=2)).isoformat() + "Z",
                "total_capacity": 12000,
                "base_price": 150.00,
                "max_tickets_per_booking": 8
            },
            # Theater events
            {
                "name": "Broadway Musical Show",
                "description": "Award-winning Broadway musical",
                "venue_id": self.test_venues[2]['venue_id'],
                "event_type": "theater", 
                "start_datetime": (base_date + timedelta(days=25)).isoformat() + "Z",
                "end_datetime": (base_date + timedelta(days=25, hours=3)).isoformat() + "Z",
                "total_capacity": 1500,
                "base_price": 95.00,
                "max_tickets_per_booking": 6
            },
            # Comedy events
            {
                "name": "Stand-up Comedy Night",
                "description": "Hilarious stand-up comedy show",
                "venue_id": self.test_venues[0]['venue_id'],
                "event_type": "comedy",
                "start_datetime": (base_date + timedelta(days=8)).isoformat() + "Z",
                "end_datetime": (base_date + timedelta(days=8, hours=2)).isoformat() + "Z", 
                "total_capacity": 800,
                "base_price": 35.00,
                "max_tickets_per_booking": 4
            }
        ]
        
        headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {self.admin_token}'
        }
        
        # Create and publish events
        for event_data in events_data:
            # Create event
            success, response, status = make_request(
                'POST',
                f"{BASE_URLS['event']}/api/v1/admin/events",
                headers=headers,
                data=event_data
            )
            
            if not success or not isinstance(response, dict) or 'event_id' not in response:
                self.results.add_result(f"Event Creation: {event_data['name']}", False, f"Status: {status}")
                continue
                
            event_id = response['event_id']
            
            # Publish event
            publish_data = {
                "status": "published",
                "version": response.get('version', 1)
            }
            
            success, pub_response, pub_status = make_request(
                'PUT',
                f"{BASE_URLS['event']}/api/v1/admin/events/{event_id}",
                headers=headers,
                data=publish_data
            )
            
            if success:
                self.test_events.append(response)
                self.results.add_result(f"Event Creation & Publishing: {event_data['name']}", True)
            else:
                self.results.add_result(f"Event Publishing: {event_data['name']}", False, f"Status: {pub_status}")
                
        return len(self.test_events) >= 5  # At least 5 events created successfully

    def test_basic_search(self):
        """Test basic search functionality"""
        print(f"\n{Colors.CYAN}Testing Basic Search Functionality{Colors.END}")
        
        # Test 1: Search without parameters (should return all events)
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search")
        
        is_valid = (success and status == 200 and 
                   isinstance(response, dict) and 
                   'results' in response and
                   isinstance(response['results'], list))
        
        self.results.add_result(
            "Basic Search - No Parameters",
            is_valid,
            f"Status: {status}, Results count: {len(response.get('results', [])) if isinstance(response, dict) else 'N/A'}"
        )
        
        # Test 2: Search with text query
        params = {'q': 'jazz'}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
        
        jazz_results = 0
        if success and isinstance(response, dict) and 'results' in response:
            jazz_results = len([r for r in response['results'] if 'jazz' in r.get('name', '').lower()])
            
        self.results.add_result(
            "Text Search - 'jazz'",
            success and status == 200 and jazz_results > 0,
            f"Status: {status}, Jazz events found: {jazz_results}"
        )
        
        # Test 3: Search with pagination
        params = {'page': 1, 'limit': 3}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
        
        is_paginated = (success and isinstance(response, dict) and 
                       response.get('page') == 1 and 
                       response.get('limit') == 3 and
                       len(response.get('results', [])) <= 3)
        
        self.results.add_result(
            "Search Pagination",
            is_paginated,
            f"Status: {status}, Page: {response.get('page') if isinstance(response, dict) else 'N/A'}, Results: {len(response.get('results', [])) if isinstance(response, dict) else 'N/A'}"
        )

    def test_search_filters(self):
        """Test all search filters comprehensively"""
        print(f"\n{Colors.CYAN}Testing Search Filters{Colors.END}")
        
        # Test 1: City filter
        params = {'city': 'New York'}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
        
        ny_events = 0
        if success and isinstance(response, dict) and 'results' in response:
            ny_events = len([r for r in response['results'] if r.get('venue_city') == 'New York'])
            
        self.results.add_result(
            "City Filter - New York",
            success and status == 200,
            f"Status: {status}, NY events: {ny_events}"
        )
        
        # Test 2: Event type filter
        params = {'type': 'concert'}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
        
        concert_events = 0
        if success and isinstance(response, dict) and 'results' in response:
            concert_events = len([r for r in response['results'] if r.get('event_type') == 'concert'])
            
        self.results.add_result(
            "Event Type Filter - Concert",
            success and status == 200,
            f"Status: {status}, Concert events: {concert_events}"
        )
        
        # Test 3: Price range filter
        params = {'min_price': 50, 'max_price': 100}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
        
        price_filtered = 0
        if success and isinstance(response, dict) and 'results' in response:
            price_filtered = len([r for r in response['results'] 
                                if 50 <= r.get('base_price', 0) <= 100])
            
        self.results.add_result(
            "Price Range Filter ($50-$100)",
            success and status == 200,
            f"Status: {status}, Events in range: {price_filtered}"
        )
        
        # Test 4: Date range filter
        future_date = (datetime.now() + timedelta(days=35)).isoformat() + "Z"
        far_future = (datetime.now() + timedelta(days=50)).isoformat() + "Z"
        
        params = {'date_from': future_date, 'date_to': far_future}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
        
        self.results.add_result(
            "Date Range Filter",
            success and status == 200,
            f"Status: {status}, Events in date range: {len(response.get('results', [])) if isinstance(response, dict) else 'N/A'}"
        )
        
        # Test 5: Combined filters
        params = {
            'q': 'music',
            'type': 'concert', 
            'city': 'Los Angeles',
            'min_price': 100
        }
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
        
        self.results.add_result(
            "Combined Filters (text + type + city + price)",
            success and status == 200,
            f"Status: {status}, Filtered results: {len(response.get('results', [])) if isinstance(response, dict) else 'N/A'}"
        )

    def test_search_suggestions(self):
        """Test search suggestions endpoint"""
        print(f"\n{Colors.CYAN}Testing Search Suggestions{Colors.END}")
        
        # Test 1: Valid suggestion request
        params = {'q': 'jazz', 'limit': 5}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/suggestions", params=params)
        
        has_suggestions = (success and isinstance(response, dict) and 
                          'suggestions' in response and
                          isinstance(response['suggestions'], list))
        
        self.results.add_result(
            "Search Suggestions - 'jazz'",
            has_suggestions,
            f"Status: {status}, Suggestions count: {len(response.get('suggestions', [])) if isinstance(response, dict) else 'N/A'}"
        )
        
        # Test 2: Empty query (should fail)
        params = {'q': ''}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/suggestions", params=params)
        
        self.results.add_result(
            "Search Suggestions - Empty Query",
            not success or status == 400,
            f"Status: {status} (should be 400 for empty query)"
        )
        
        # Test 3: Limit parameter
        params = {'q': 'concert', 'limit': 3}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/suggestions", params=params)
        
        limited_suggestions = (success and isinstance(response, dict) and 
                             len(response.get('suggestions', [])) <= 3)
        
        self.results.add_result(
            "Search Suggestions - Limit Parameter",
            limited_suggestions,
            f"Status: {status}, Suggestions returned: {len(response.get('suggestions', [])) if isinstance(response, dict) else 'N/A'}"
        )

    def test_search_metadata(self):
        """Test search filters metadata endpoint"""
        print(f"\n{Colors.CYAN}Testing Search Metadata{Colors.END}")
        
        # Test filters endpoint
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/filters")
        
        has_metadata = (success and isinstance(response, dict) and
                       'cities' in response and
                       'event_types' in response and
                       'price_range' in response and
                       isinstance(response['cities'], list) and
                       isinstance(response['event_types'], list))
        
        cities_count = len(response.get('cities', [])) if isinstance(response, dict) else 0
        types_count = len(response.get('event_types', [])) if isinstance(response, dict) else 0
        
        self.results.add_result(
            "Search Filters Metadata",
            has_metadata,
            f"Status: {status}, Cities: {cities_count}, Event types: {types_count}"
        )

    def test_trending_events(self):
        """Test trending events endpoint"""
        print(f"\n{Colors.CYAN}Testing Trending Events{Colors.END}")
        
        # Test 1: Basic trending request
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/trending")
        
        has_trending = (success and isinstance(response, dict) and
                       'events' in response and
                       isinstance(response['events'], list))
        
        self.results.add_result(
            "Trending Events - Basic",
            has_trending,
            f"Status: {status}, Trending events: {len(response.get('events', [])) if isinstance(response, dict) else 'N/A'}"
        )
        
        # Test 2: Trending with limit
        params = {'limit': 5}
        success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/trending", params=params)
        
        limited_trending = (success and isinstance(response, dict) and
                           len(response.get('events', [])) <= 5)
        
        self.results.add_result(
            "Trending Events - With Limit",
            limited_trending,
            f"Status: {status}, Limited trending events: {len(response.get('events', [])) if isinstance(response, dict) else 'N/A'}"
        )

    def test_internal_endpoints(self):
        """Test internal search endpoints"""
        print(f"\n{Colors.CYAN}Testing Internal Search Endpoints{Colors.END}")
        
        headers = {
            'Content-Type': 'application/json',
            'X-API-Key': INTERNAL_API_KEY
        }
        
        # Test 1: Full resync
        resync_data = {"force_reindex": False}
        success, response, status = make_request(
            'POST',
            f"{BASE_URLS['search']}/internal/search/resync",
            headers=headers,
            data=resync_data
        )
        
        resync_success = (success and isinstance(response, dict) and
                         'events_indexed' in response)
        
        self.results.add_result(
            "Internal - Full Resync",
            resync_success,
            f"Status: {status}, Events indexed: {response.get('events_indexed') if isinstance(response, dict) else 'N/A'}"
        )
        
        # Test 2: Manual event indexing (if we have test events)
        if self.test_events:
            test_event = self.test_events[0]
            
            # Create event document for indexing
            event_doc = {
                "event": {
                    "event_id": test_event['event_id'],
                    "name": test_event['name'],
                    "description": test_event.get('description', ''),
                    "venue_id": test_event['venue_id'],
                    "venue_name": "Test Venue",
                    "venue_city": "Test City",
                    "venue_country": "USA",
                    "event_type": test_event['event_type'],
                    "start_datetime": test_event['start_datetime'],
                    "end_datetime": test_event['end_datetime'],
                    "base_price": test_event['base_price'],
                    "available_seats": test_event['total_capacity'],
                    "total_capacity": test_event['total_capacity'],
                    "status": "published",
                    "version": test_event.get('version', 1),
                    "created_at": datetime.now().isoformat() + "Z",
                    "updated_at": datetime.now().isoformat() + "Z"
                }
            }
            
            success, response, status = make_request(
                'POST',
                f"{BASE_URLS['search']}/internal/search/events",
                headers=headers,
                data=event_doc
            )
            
            self.results.add_result(
                "Internal - Manual Event Indexing",
                success and status == 200,
                f"Status: {status}, Response: {response.get('status') if isinstance(response, dict) else 'N/A'}"
            )

    def cleanup_test_data(self):
        """Clean up test data"""
        print(f"\n{Colors.CYAN}Cleaning up Test Data{Colors.END}")
        
        if not self.admin_token:
            return
            
        headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {self.admin_token}'
        }
        
        # Delete test events
        deleted_events = 0
        for event in self.test_events:
            success, _, status = make_request(
                'DELETE',
                f"{BASE_URLS['event']}/api/v1/admin/events/{event['event_id']}",
                headers=headers
            )
            if success:
                deleted_events += 1
                
        # Delete test venues  
        deleted_venues = 0
        for venue in self.test_venues:
            success, _, status = make_request(
                'DELETE',
                f"{BASE_URLS['event']}/api/v1/admin/venues/{venue['venue_id']}",
                headers=headers
            )
            if success:
                deleted_venues += 1
                
        self.results.add_result(
            "Cleanup - Test Data",
            True,
            f"Deleted {deleted_events} events and {deleted_venues} venues"
        )

def main():
    """Main test execution"""
    print(f"{Colors.BOLD}{Colors.PURPLE}BookMyEvent - Comprehensive Search API Testing Suite{Colors.END}")
    print(f"{Colors.BOLD}{'='*70}{Colors.END}")
    
    # Check if services are running
    print(f"{Colors.YELLOW}Checking service availability...{Colors.END}")
    
    services_ok = True
    for service, url in BASE_URLS.items():
        try:
            response = requests.get(f"{url}/healthz", timeout=5)
            if response.status_code == 200:
                print(f"{Colors.GREEN}✓{Colors.END} {service.title()} Service: {url}")
            else:
                print(f"{Colors.RED}✗{Colors.END} {service.title()} Service: {url} (Status: {response.status_code})")
                services_ok = False
        except Exception as e:
            print(f"{Colors.RED}✗{Colors.END} {service.title()} Service: {url} (Error: {str(e)})")
            services_ok = False
    
    if not services_ok:
        print(f"\n{Colors.RED}Some services are not available. Please ensure all services are running.{Colors.END}")
        print(f"{Colors.YELLOW}Run: make docker-full-up && make run SERVICE=user-service & make run SERVICE=event-service & make run SERVICE=search-service{Colors.END}")
        return False
    
    print(f"\n{Colors.GREEN}All services are available!{Colors.END}")
    
    # Run tests
    tester = SearchAPITester()
    success = tester.run_all_tests()
    
    if success:
        print(f"\n{Colors.GREEN}{Colors.BOLD}🎉 ALL TESTS PASSED! Search API is working correctly.{Colors.END}")
        return True
    else:
        print(f"\n{Colors.RED}{Colors.BOLD}❌ SOME TESTS FAILED. Check the results above.{Colors.END}")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
