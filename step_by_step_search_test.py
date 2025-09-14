#!/usr/bin/env python3
"""
Step-by-Step Search API Testing Script
Run each test individually to see detailed request/response information

Usage: python3 step_by_step_search_test.py [test_name]

Available tests:
- health: Test health endpoints
- setup: Setup test data (admin + venues + events)
- search_basic: Test basic search functionality
- search_filters: Test all search filters
- search_suggestions: Test suggestions endpoint
- search_metadata: Test filters metadata endpoint
- search_trending: Test trending events
- search_internal: Test internal endpoints
- cleanup: Clean up test data
- all: Run all tests
"""

import requests
import json
import time
import uuid
import sys
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any

# Configuration
BASE_URLS = {
    'user': 'http://localhost:8001',
    'event': 'http://localhost:8002', 
    'search': 'http://localhost:8003',
    'booking': 'http://localhost:8004'
}

INTERNAL_API_KEY = 'internal-service-communication-key-change-in-production'

# Global test state
test_state = {
    'admin_token': None,
    'test_venues': [],
    'test_events': []
}

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

def print_section(title):
    """Print a section header"""
    print(f"\n{Colors.BOLD}{Colors.BLUE}{'='*60}{Colors.END}")
    print(f"{Colors.BOLD}{Colors.BLUE}{title}{Colors.END}")
    print(f"{Colors.BOLD}{Colors.BLUE}{'='*60}{Colors.END}")

def print_request(method, url, headers=None, data=None, params=None):
    """Print request details"""
    print(f"\n{Colors.CYAN}REQUEST:{Colors.END}")
    print(f"{Colors.BOLD}{method} {url}{Colors.END}")
    
    if params:
        print(f"{Colors.YELLOW}Query Params:{Colors.END}")
        for key, value in params.items():
            print(f"  {key}: {value}")
    
    if headers:
        print(f"{Colors.YELLOW}Headers:{Colors.END}")
        for key, value in headers.items():
            if 'token' in key.lower() or 'auth' in key.lower():
                print(f"  {key}: {value[:20]}..." if len(str(value)) > 20 else f"  {key}: {value}")
            else:
                print(f"  {key}: {value}")
    
    if data:
        print(f"{Colors.YELLOW}Body:{Colors.END}")
        print(json.dumps(data, indent=2))

def print_response(response, status_code):
    """Print response details"""
    print(f"\n{Colors.CYAN}RESPONSE:{Colors.END}")
    
    if status_code < 300:
        status_color = Colors.GREEN
    elif status_code < 400:
        status_color = Colors.YELLOW
    else:
        status_color = Colors.RED
        
    print(f"{Colors.BOLD}Status: {status_color}{status_code}{Colors.END}")
    
    if isinstance(response, dict):
        print(f"{Colors.YELLOW}Response Body:{Colors.END}")
        print(json.dumps(response, indent=2, default=str))
    else:
        print(f"{Colors.YELLOW}Response:{Colors.END}")
        print(str(response))

def make_request(method: str, url: str, headers: Dict = None, data: Dict = None, params: Dict = None) -> tuple:
    """Make HTTP request with detailed logging"""
    print_request(method, url, headers, data, params)
    
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
            
        print_response(json_data, response.status_code)
        
        return response.status_code < 400, json_data, response.status_code
    except Exception as e:
        print(f"\n{Colors.RED}ERROR: {str(e)}{Colors.END}")
        return False, str(e), 0

def test_health():
    """Test health endpoints"""
    print_section("Testing Health Endpoints")
    
    print(f"\n{Colors.PURPLE}1. Testing Search Service Health{Colors.END}")
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/healthz")
    
    print(f"\n{Colors.PURPLE}2. Testing Search Service Readiness{Colors.END}")
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/health/ready")
    
    print(f"\n{Colors.GREEN}Health tests completed!{Colors.END}")

def test_setup():
    """Setup test data"""
    print_section("Setting Up Test Data")
    
    # Step 1: Create admin account
    print(f"\n{Colors.PURPLE}1. Creating Admin Account{Colors.END}")
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
        test_state['admin_token'] = response['access_token']
        print(f"\n{Colors.GREEN}✓ Admin account created successfully!{Colors.END}")
    else:
        print(f"\n{Colors.RED}✗ Failed to create admin account{Colors.END}")
        return
    
    # Step 2: Create test venues
    print(f"\n{Colors.PURPLE}2. Creating Test Venues{Colors.END}")
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
        }
    ]
    
    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {test_state["admin_token"]}'
    }
    
    for i, venue_data in enumerate(venues_data, 1):
        print(f"\n{Colors.YELLOW}Creating Venue {i}: {venue_data['name']}{Colors.END}")
        success, response, status = make_request(
            'POST',
            f"{BASE_URLS['event']}/api/v1/admin/venues",
            headers=headers,
            data=venue_data
        )
        
        if success and isinstance(response, dict) and 'venue_id' in response:
            test_state['test_venues'].append(response)
            print(f"\n{Colors.GREEN}✓ Venue created successfully!{Colors.END}")
        else:
            print(f"\n{Colors.RED}✗ Failed to create venue{Colors.END}")
    
    # Step 3: Create test events
    print(f"\n{Colors.PURPLE}3. Creating Test Events{Colors.END}")
    if len(test_state['test_venues']) < 2:
        print(f"\n{Colors.RED}✗ Not enough venues created. Skipping event creation.{Colors.END}")
        return
    
    base_date = datetime.now() + timedelta(days=30)
    
    events_data = [
        {
            "name": "Jazz Night at Madison Square",
            "description": "An evening of smooth jazz with renowned artists",
            "venue_id": test_state['test_venues'][0]['venue_id'],
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
            "venue_id": test_state['test_venues'][1]['venue_id'],
            "event_type": "concert", 
            "start_datetime": (base_date + timedelta(days=5)).isoformat() + "Z",
            "end_datetime": (base_date + timedelta(days=5, hours=4)).isoformat() + "Z",
            "total_capacity": 15000,
            "base_price": 125.00,
            "max_tickets_per_booking": 6
        },
        {
            "name": "Basketball Championship Game",
            "description": "Exciting basketball playoff game",
            "venue_id": test_state['test_venues'][0]['venue_id'],
            "event_type": "sports",
            "start_datetime": (base_date + timedelta(days=15)).isoformat() + "Z",
            "end_datetime": (base_date + timedelta(days=15, hours=3)).isoformat() + "Z",
            "total_capacity": 18000,
            "base_price": 75.00,
            "max_tickets_per_booking": 10
        }
    ]
    
    for i, event_data in enumerate(events_data, 1):
        print(f"\n{Colors.YELLOW}Creating Event {i}: {event_data['name']}{Colors.END}")
        
        # Create event
        success, response, status = make_request(
            'POST',
            f"{BASE_URLS['event']}/api/v1/admin/events",
            headers=headers,
            data=event_data
        )
        
        if not success or not isinstance(response, dict) or 'event_id' not in response:
            print(f"\n{Colors.RED}✗ Failed to create event{Colors.END}")
            continue
            
        event_id = response['event_id']
        
        # Publish event
        print(f"\n{Colors.YELLOW}Publishing Event: {event_data['name']}{Colors.END}")
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
            test_state['test_events'].append(response)
            print(f"\n{Colors.GREEN}✓ Event created and published successfully!{Colors.END}")
        else:
            print(f"\n{Colors.RED}✗ Failed to publish event{Colors.END}")
    
    print(f"\n{Colors.GREEN}Setup completed! Created {len(test_state['test_venues'])} venues and {len(test_state['test_events'])} events{Colors.END}")
    print(f"{Colors.YELLOW}Waiting 5 seconds for search indexing...{Colors.END}")
    time.sleep(5)

def test_search_basic():
    """Test basic search functionality"""
    print_section("Testing Basic Search Functionality")
    
    print(f"\n{Colors.PURPLE}1. Search without parameters (should return all events){Colors.END}")
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search")
    
    print(f"\n{Colors.PURPLE}2. Search with text query - 'jazz'{Colors.END}")
    params = {'q': 'jazz'}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}3. Search with text query - 'concert'{Colors.END}")
    params = {'q': 'concert'}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}4. Search with pagination{Colors.END}")
    params = {'page': 1, 'limit': 2}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)

def test_search_filters():
    """Test all search filters"""
    print_section("Testing Search Filters")
    
    print(f"\n{Colors.PURPLE}1. City filter - New York{Colors.END}")
    params = {'city': 'New York'}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}2. City filter - Los Angeles{Colors.END}")
    params = {'city': 'Los Angeles'}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}3. Event type filter - concert{Colors.END}")
    params = {'type': 'concert'}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}4. Event type filter - sports{Colors.END}")
    params = {'type': 'sports'}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}5. Price range filter - $50 to $100{Colors.END}")
    params = {'min_price': 50, 'max_price': 100}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}6. Price range filter - minimum $100{Colors.END}")
    params = {'min_price': 100}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}7. Date range filter{Colors.END}")
    future_date = (datetime.now() + timedelta(days=35)).isoformat() + "Z"
    far_future = (datetime.now() + timedelta(days=50)).isoformat() + "Z"
    params = {'date_from': future_date, 'date_to': far_future}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)
    
    print(f"\n{Colors.PURPLE}8. Combined filters{Colors.END}")
    params = {
        'q': 'music',
        'type': 'concert', 
        'city': 'New York',
        'min_price': 50
    }
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search", params=params)

def test_search_suggestions():
    """Test search suggestions"""
    print_section("Testing Search Suggestions")
    
    print(f"\n{Colors.PURPLE}1. Search suggestions for 'jazz'{Colors.END}")
    params = {'q': 'jazz', 'limit': 5}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/suggestions", params=params)
    
    print(f"\n{Colors.PURPLE}2. Search suggestions for 'concert'{Colors.END}")
    params = {'q': 'concert', 'limit': 3}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/suggestions", params=params)
    
    print(f"\n{Colors.PURPLE}3. Search suggestions with empty query (should fail){Colors.END}")
    params = {'q': ''}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/suggestions", params=params)

def test_search_metadata():
    """Test search metadata endpoints"""
    print_section("Testing Search Metadata")
    
    print(f"\n{Colors.PURPLE}1. Get available filters metadata{Colors.END}")
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/filters")

def test_search_trending():
    """Test trending events"""
    print_section("Testing Trending Events")
    
    print(f"\n{Colors.PURPLE}1. Get trending events{Colors.END}")
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/trending")
    
    print(f"\n{Colors.PURPLE}2. Get trending events with limit{Colors.END}")
    params = {'limit': 5}
    success, response, status = make_request('GET', f"{BASE_URLS['search']}/api/v1/search/trending", params=params)

def test_search_internal():
    """Test internal search endpoints"""
    print_section("Testing Internal Search Endpoints")
    
    headers = {
        'Content-Type': 'application/json',
        'X-API-Key': INTERNAL_API_KEY
    }
    
    print(f"\n{Colors.PURPLE}1. Full resync{Colors.END}")
    resync_data = {"force_reindex": False}
    success, response, status = make_request(
        'POST',
        f"{BASE_URLS['search']}/internal/search/resync",
        headers=headers,
        data=resync_data
    )
    
    if test_state['test_events']:
        print(f"\n{Colors.PURPLE}2. Manual event indexing{Colors.END}")
        test_event = test_state['test_events'][0]
        
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

def test_cleanup():
    """Clean up test data"""
    print_section("Cleaning Up Test Data")
    
    if not test_state['admin_token']:
        print(f"{Colors.RED}No admin token available for cleanup{Colors.END}")
        return
        
    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {test_state["admin_token"]}'
    }
    
    # Delete test events
    print(f"\n{Colors.PURPLE}1. Deleting Test Events{Colors.END}")
    for i, event in enumerate(test_state['test_events'], 1):
        print(f"\n{Colors.YELLOW}Deleting Event {i}: {event.get('name', event['event_id'])}{Colors.END}")
        success, response, status = make_request(
            'DELETE',
            f"{BASE_URLS['event']}/api/v1/admin/events/{event['event_id']}",
            headers=headers
        )
        
    # Delete test venues  
    print(f"\n{Colors.PURPLE}2. Deleting Test Venues{Colors.END}")
    for i, venue in enumerate(test_state['test_venues'], 1):
        print(f"\n{Colors.YELLOW}Deleting Venue {i}: {venue.get('name', venue['venue_id'])}{Colors.END}")
        success, response, status = make_request(
            'DELETE',
            f"{BASE_URLS['event']}/api/v1/admin/venues/{venue['venue_id']}",
            headers=headers
        )
    
    # Clear state
    test_state['admin_token'] = None
    test_state['test_venues'] = []
    test_state['test_events'] = []
    
    print(f"\n{Colors.GREEN}Cleanup completed!{Colors.END}")

def main():
    """Main function"""
    if len(sys.argv) < 2:
        print(f"""
{Colors.BOLD}{Colors.PURPLE}BookMyEvent - Step-by-Step Search API Testing{Colors.END}

{Colors.BOLD}Usage:{Colors.END} python3 step_by_step_search_test.py [test_name]

{Colors.BOLD}Available tests:{Colors.END}
  {Colors.CYAN}health{Colors.END}           - Test health endpoints
  {Colors.CYAN}setup{Colors.END}            - Setup test data (admin + venues + events)
  {Colors.CYAN}search_basic{Colors.END}     - Test basic search functionality
  {Colors.CYAN}search_filters{Colors.END}   - Test all search filters
  {Colors.CYAN}search_suggestions{Colors.END} - Test suggestions endpoint
  {Colors.CYAN}search_metadata{Colors.END}  - Test filters metadata endpoint
  {Colors.CYAN}search_trending{Colors.END}  - Test trending events
  {Colors.CYAN}search_internal{Colors.END}  - Test internal endpoints
  {Colors.CYAN}cleanup{Colors.END}          - Clean up test data
  {Colors.CYAN}all{Colors.END}              - Run all tests

{Colors.BOLD}Examples:{Colors.END}
  python3 step_by_step_search_test.py health
  python3 step_by_step_search_test.py setup
  python3 step_by_step_search_test.py search_basic
  python3 step_by_step_search_test.py all
        """)
        return
    
    test_name = sys.argv[1].lower()
    
    # Check if search service is running
    try:
        response = requests.get(f"{BASE_URLS['search']}/healthz", timeout=5)
        if response.status_code != 200:
            print(f"{Colors.RED}Search service is not running on {BASE_URLS['search']}{Colors.END}")
            return
    except Exception as e:
        print(f"{Colors.RED}Cannot connect to search service: {str(e)}{Colors.END}")
        print(f"{Colors.YELLOW}Make sure to start the search service: make run SERVICE=search-service{Colors.END}")
        return
    
    # Run specific test
    tests = {
        'health': test_health,
        'setup': test_setup,
        'search_basic': test_search_basic,
        'search_filters': test_search_filters,
        'search_suggestions': test_search_suggestions,
        'search_metadata': test_search_metadata,
        'search_trending': test_search_trending,
        'search_internal': test_search_internal,
        'cleanup': test_cleanup
    }
    
    if test_name == 'all':
        print(f"{Colors.BOLD}{Colors.PURPLE}Running All Search API Tests{Colors.END}")
        for test_func in tests.values():
            test_func()
            print(f"\n{Colors.YELLOW}Press Enter to continue to next test...{Colors.END}")
            input()
    elif test_name in tests:
        tests[test_name]()
    else:
        print(f"{Colors.RED}Unknown test: {test_name}{Colors.END}")
        print(f"Available tests: {', '.join(tests.keys())}, all")

if __name__ == "__main__":
    main()
