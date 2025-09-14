#!/usr/bin/env python3

import json
import uuid
import random
import requests
import time
from datetime import datetime, timedelta
from concurrent.futures import ThreadPoolExecutor, as_completed
import argparse
import os
from typing import Dict, List, Any

class SearchServiceDataGenerator:
    def __init__(self, search_service_url: str, internal_api_key: str):
        self.search_service_url = search_service_url
        self.internal_api_key = internal_api_key
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'X-API-Key': internal_api_key
        })
        
        # Data for generating realistic events
        self.event_types = ["concert", "sports", "theater", "conference", "comedy", "festival", "workshop", "exhibition"]
        
        self.cities = [
            "New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia",
            "San Antonio", "San Diego", "Dallas", "San Jose", "Austin", "Jacksonville",
            "Fort Worth", "Columbus", "Charlotte", "San Francisco", "Indianapolis", 
            "Seattle", "Denver", "Boston", "El Paso", "Detroit", "Nashville", 
            "Portland", "Memphis", "Oklahoma City", "Las Vegas", "Louisville", 
            "Baltimore", "Milwaukee", "Miami", "Atlanta", "Tampa", "Orlando"
        ]
        
        self.states = [
            "NY", "CA", "IL", "TX", "AZ", "PA", "TX", "CA", "TX", "CA", "TX", "FL",
            "TX", "OH", "NC", "CA", "IN", "WA", "CO", "MA", "TX", "MI", "TN", "OR",
            "TN", "OK", "NV", "KY", "MD", "WI", "FL", "GA", "FL", "FL"
        ]
        
        self.event_names = {
            "concert": [
                "The Midnight Stars", "Electric Dreams", "Jazz Fusion Collective", 
                "Rock Revolution", "Acoustic Nights", "Symphony of Sound", 
                "Urban Beats", "Classical Crossover", "Indie Underground", "Pop Sensation",
                "Blues Brothers Revival", "Heavy Metal Thunder", "Country Roads Band",
                "Electronic Pulse", "Reggae Vibes", "Hip Hop Nation", "Folk Tales"
            ],
            "sports": [
                "Basketball Championship", "Soccer Finals", "Tennis Tournament", 
                "Baseball Game", "Football Match", "Hockey Playoffs", "Golf Tournament",
                "Swimming Competition", "Track & Field", "Boxing Match", "Wrestling Event",
                "Volleyball Championship", "Marathon Run", "Cycling Race", "Skiing Competition"
            ],
            "theater": [
                "Romeo and Juliet", "The Lion King", "Hamilton", "Phantom of the Opera",
                "Chicago", "Cats", "Les Miserables", "Wicked", "The Book of Mormon",
                "Dear Evan Hansen", "Frozen", "Aladdin", "The Greatest Showman"
            ],
            "conference": [
                "Tech Innovation Summit", "Digital Marketing Conference", "AI & Machine Learning Expo",
                "Startup Pitch Day", "Business Leadership Forum", "Data Science Convention",
                "Cybersecurity Summit", "Cloud Computing Conference", "Mobile App Development",
                "E-commerce Expo", "Blockchain Summit", "IoT Innovation Conference"
            ],
            "comedy": [
                "Stand-Up Comedy Night", "Improv Show", "Comedy Central Live", 
                "Laugh Out Loud", "Comedy Club Special", "Humor Festival"
            ],
            "festival": [
                "Music Festival", "Food & Wine Festival", "Art & Culture Festival",
                "Film Festival", "Beer Festival", "Jazz Festival", "Folk Festival"
            ],
            "workshop": [
                "Photography Workshop", "Cooking Class", "Art Workshop", 
                "Writing Seminar", "Business Skills Workshop", "Tech Bootcamp"
            ],
            "exhibition": [
                "Art Exhibition", "Science Exhibition", "History Exhibition",
                "Photography Exhibition", "Modern Art Showcase", "Cultural Heritage Display"
            ]
        }
        
        self.venue_types = ["Arena", "Theater", "Convention Center", "Stadium", "Hall", 
                           "Auditorium", "Center", "Pavilion", "Coliseum", "Forum", "Club", "Park"]
        
        self.street_names = ["Main St", "Broadway", "First Ave", "Second St", "Park Ave", 
                           "Oak St", "Maple Ave", "Cedar St", "Pine Ave", "Elm St", 
                           "Washington St", "Lincoln Ave", "Jefferson Blvd"]
        
        self.descriptions = [
            "Join us for an unforgettable experience that will leave you wanting more.",
            "Don't miss this incredible opportunity to witness greatness in person.",
            "Experience the magic and excitement of live entertainment at its finest.",
            "A spectacular event featuring world-class performers and production.",
            "Come and be part of something truly special and memorable.",
            "An evening of entertainment that promises to exceed all expectations.",
            "Witness history in the making at this once-in-a-lifetime event.",
            "The ultimate experience for fans and newcomers alike.",
            "Prepare to be amazed by this extraordinary showcase of talent.",
            "An event that brings together the best in entertainment and culture.",
            "Experience the thrill and excitement of live performance.",
            "A celebration of artistry, creativity, and human expression."
        ]
        
        self.base_prices = [25.00, 35.00, 45.00, 55.00, 75.00, 95.00, 125.00, 150.00, 200.00, 250.00, 350.00, 500.00]

    def check_service_health(self) -> bool:
        """Check if the search service is running"""
        try:
            response = requests.get(f"{self.search_service_url}/healthz", timeout=5)
            return response.status_code == 200
        except requests.RequestException:
            return False

    def generate_event_data(self, index: int) -> Dict[str, Any]:
        """Generate realistic event data"""
        event_type = random.choice(self.event_types)
        city = random.choice(self.cities)
        state = random.choice(self.states)
        
        # Generate event name based on type
        if event_type in self.event_names:
            base_name = random.choice(self.event_names[event_type])
            if event_type == "concert":
                event_name = f"{base_name} Live in Concert"
            elif event_type == "sports":
                event_name = f"{city} {base_name}"
            elif event_type == "theater":
                event_name = f"{base_name} - Broadway Musical"
            elif event_type == "conference":
                event_name = f"{base_name} 2024"
            else:
                event_name = base_name
        else:
            event_name = f"{event_type.title()} Event"
        
        # Generate venue
        venue_type = random.choice(self.venue_types)
        venue_name = f"{city} {venue_type}"
        
        # Generate address
        street_number = random.randint(1, 9999)
        street_name = random.choice(self.street_names)
        venue_address = f"{street_number} {street_name}"
        
        # Generate dates (next 6 months)
        days_ahead = random.randint(1, 180)
        start_datetime = datetime.now() + timedelta(days=days_ahead)
        start_datetime = start_datetime.replace(
            hour=random.randint(10, 22),
            minute=random.choice([0, 15, 30, 45]),
            second=0,
            microsecond=0
        )
        
        # End time (1-4 hours later)
        duration_hours = random.randint(1, 4)
        end_datetime = start_datetime + timedelta(hours=duration_hours)
        
        # Generate capacity based on venue type
        if venue_type in ["Stadium", "Arena"]:
            total_capacity = random.randint(20000, 70000)
        elif venue_type in ["Theater", "Auditorium", "Hall"]:
            total_capacity = random.randint(500, 3000)
        elif venue_type in ["Club"]:
            total_capacity = random.randint(100, 800)
        else:
            total_capacity = random.randint(1000, 15000)
        
        # Available seats (80-100% of capacity)
        available_seats = random.randint(int(total_capacity * 0.8), total_capacity)
        
        # Generate pricing
        base_price = random.choice(self.base_prices)
        
        # Generate description
        description = random.choice(self.descriptions)
        
        # Generate timestamps
        created_days_ago = random.randint(1, 60)
        created_at = datetime.now() - timedelta(days=created_days_ago)
        updated_days_ago = random.randint(0, min(created_days_ago, 14))
        updated_at = datetime.now() - timedelta(days=updated_days_ago)
        
        return {
            "event": {
                "event_id": str(uuid.uuid4()),
                "name": event_name,
                "description": description,
                "venue_id": str(uuid.uuid4()),
                "venue_name": venue_name,
                "venue_address": venue_address,
                "venue_city": city,
                "venue_state": state,
                "venue_country": "USA",
                "event_type": event_type,
                "start_datetime": start_datetime.isoformat() + "Z",
                "end_datetime": end_datetime.isoformat() + "Z",
                "base_price": base_price,
                "available_seats": available_seats,
                "total_capacity": total_capacity,
                "status": "published",
                "version": 1,
                "created_at": created_at.isoformat() + "Z",
                "updated_at": updated_at.isoformat() + "Z"
            }
        }

    def create_single_event(self, event_data: Dict[str, Any]) -> tuple[bool, str]:
        """Create a single event via the search service API"""
        try:
            response = self.session.post(
                f"{self.search_service_url}/internal/search/events",
                json=event_data,
                timeout=10
            )
            
            if response.status_code == 200:
                return True, "Success"
            else:
                return False, f"HTTP {response.status_code}: {response.text[:100]}"
                
        except requests.RequestException as e:
            return False, f"Request error: {str(e)[:100]}"

    def create_events_batch(self, events_data: List[Dict[str, Any]], batch_num: int) -> tuple[int, int]:
        """Create a batch of events"""
        success_count = 0
        error_count = 0
        
        print(f"📦 Processing batch {batch_num} ({len(events_data)} events)")
        
        # Use threading for concurrent requests
        with ThreadPoolExecutor(max_workers=10) as executor:
            future_to_event = {
                executor.submit(self.create_single_event, event_data): i 
                for i, event_data in enumerate(events_data)
            }
            
            for future in as_completed(future_to_event):
                event_index = future_to_event[future]
                try:
                    success, message = future.result()
                    if success:
                        success_count += 1
                        if (event_index + 1) % 10 == 0:
                            print(f"  ✅ Event {event_index + 1} created")
                    else:
                        error_count += 1
                        print(f"  ❌ Event {event_index + 1} failed: {message}")
                except Exception as e:
                    error_count += 1
                    print(f"  ❌ Event {event_index + 1} exception: {str(e)[:50]}")
        
        print(f"  📊 Batch {batch_num}: {success_count} success, {error_count} errors")
        return success_count, error_count

    def generate_and_create_events(self, total_events: int, batch_size: int = 50) -> tuple[int, int]:
        """Generate and create all events"""
        print(f"🏗️ Generating {total_events} events in batches of {batch_size}")
        print("")
        
        total_success = 0
        total_errors = 0
        total_batches = (total_events + batch_size - 1) // batch_size
        start_time = time.time()
        
        for batch_num in range(1, total_batches + 1):
            start_index = (batch_num - 1) * batch_size
            end_index = min(start_index + batch_size, total_events)
            current_batch_size = end_index - start_index
            
            # Generate batch data
            batch_events = []
            for i in range(current_batch_size):
                event_data = self.generate_event_data(start_index + i)
                batch_events.append(event_data)
            
            # Create batch
            batch_success, batch_errors = self.create_events_batch(batch_events, batch_num)
            total_success += batch_success
            total_errors += batch_errors
            
            # Progress update
            progress = (batch_num * 100) // total_batches
            elapsed = time.time() - start_time
            rate = total_success / max(elapsed, 1)
            
            print(f"  📈 Progress: {progress}% ({total_success}/{total_events} events, {rate:.1f} events/sec)")
            print("")
            
            # Small delay between batches
            time.sleep(0.1)
        
        return total_success, total_errors

    def test_search_functionality(self):
        """Test the search service with created data"""
        print("🔍 Testing search functionality...")
        print("")
        
        test_cases = [
            ("Basic search", "/api/v1/search?limit=5"),
            ("Search by query", "/api/v1/search?q=concert&limit=3"),
            ("Search by city", "/api/v1/search?city=New York&limit=3"),
            ("Search by type", "/api/v1/search?type=sports&limit=3"),
            ("Filters", "/api/v1/search/filters"),
            ("Trending", "/api/v1/search/trending?limit=5")
        ]
        
        for test_name, endpoint in test_cases:
            try:
                response = requests.get(f"{self.search_service_url}{endpoint}", timeout=10)
                if response.status_code == 200:
                    data = response.json()
                    if "results" in data:
                        print(f"✅ {test_name}: {len(data['results'])} results (total: {data.get('total', 'N/A')})")
                    elif "suggestions" in data:
                        print(f"✅ {test_name}: {len(data['suggestions'])} suggestions")
                    elif "cities" in data:
                        print(f"✅ {test_name}: {len(data['cities'])} cities, {len(data['event_types'])} event types")
                    else:
                        print(f"✅ {test_name}: Success")
                else:
                    print(f"❌ {test_name}: HTTP {response.status_code}")
            except Exception as e:
                print(f"❌ {test_name}: {str(e)[:50]}")
        
        print("")

def main():
    parser = argparse.ArgumentParser(description="Generate test data for Search Service")
    parser.add_argument("-u", "--url", default="http://localhost:8003", 
                       help="Search service URL (default: http://localhost:8003)")
    parser.add_argument("-k", "--key", help="Internal API key")
    parser.add_argument("-n", "--count", type=int, default=10000,
                       help="Number of events to create (default: 10000)")
    parser.add_argument("-b", "--batch", type=int, default=50,
                       help="Batch size (default: 50)")
    
    args = parser.parse_args()
    
    # Get API key from argument, environment, or .env file
    api_key = args.key
    if not api_key:
        api_key = os.getenv("INTERNAL_API_KEY")
    
    if not api_key and os.path.exists(".env"):
        with open(".env", "r") as f:
            for line in f:
                if line.startswith("INTERNAL_API_KEY="):
                    api_key = line.split("=", 1)[1].strip().strip('"')
                    break
    
    if not api_key:
        print("❌ Internal API key is required. Set INTERNAL_API_KEY environment variable or use -k option.")
        return 1
    
    print("🚀 Search Service Data Generator")
    print("=" * 40)
    print(f"Target URL: {args.url}")
    print(f"Events to create: {args.count}")
    print(f"Batch size: {args.batch}")
    print("")
    
    generator = SearchServiceDataGenerator(args.url, api_key)
    
    # Check service health
    print("🔍 Checking if Search Service is running...")
    if not generator.check_service_health():
        print(f"❌ Search Service is not running on {args.url}")
        print("💡 Start it with: make run SERVICE=search-service")
        return 1
    
    print("✅ Search Service is running")
    print("")
    
    # Generate and create events
    start_time = time.time()
    total_success, total_errors = generator.generate_and_create_events(args.count, args.batch)
    end_time = time.time()
    
    duration = end_time - start_time
    
    print("🎉 Data generation completed!")
    print("=" * 40)
    print(f"✅ Total events created: {total_success}/{args.count}")
    print(f"❌ Total errors: {total_errors}")
    print(f"⏱️  Time taken: {duration:.1f}s")
    print(f"📊 Average rate: {total_success / max(duration, 1):.1f} events/second")
    print("")
    
    if total_success > 0:
        generator.test_search_functionality()
        
        print("🎯 Search service is ready for testing!")
        print("Try these endpoints:")
        print(f"  • GET {args.url}/api/v1/search")
        print(f"  • GET {args.url}/api/v1/search?q=concert")
        print(f"  • GET {args.url}/api/v1/search?city=New York")
        print(f"  • GET {args.url}/api/v1/search/suggestions?q=con")
        print(f"  • GET {args.url}/api/v1/search/trending")
        return 0
    else:
        print("❌ No events were created successfully. Check the search service logs.")
        return 1

if __name__ == "__main__":
    exit(main())
