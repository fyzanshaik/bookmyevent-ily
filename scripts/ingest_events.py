#!/usr/bin/env python3
"""
Event Service Data Ingestion Script

This script creates 50-100 diverse events through the event service API.
It handles admin authentication, venue creation, event creation, and publishing.

Requirements:
- Event service running on port 8002
- Database migrations completed
- Internet connection for realistic data generation

Usage:
    python3 scripts/ingest_events.py [--events 75] [--host localhost:8002]
"""

import requests
import json
import random
import time
from datetime import datetime, timedelta
from typing import Dict, List, Optional
import argparse
import sys

# Configuration
DEFAULT_HOST = "localhost:8002"
DEFAULT_EVENT_COUNT = 75

class EventServiceClient:
    def __init__(self, host: str):
        self.base_url = f"http://{host}"
        self.admin_token = None
        self.session = requests.Session()
        
    def register_admin(self) -> Dict:
        """Register a new admin account"""
        admin_data = {
            "email": f"event_admin_{int(time.time())}@bookmyevent.com",
            "password": "AdminPassword123!",
            "name": "Event Data Admin",
            "phone_number": "+1-555-0123",
            "role": "event_manager"
        }
        
        response = self.session.post(
            f"{self.base_url}/api/v1/auth/admin/register",
            json=admin_data
        )
        
        if response.status_code == 201:
            result = response.json()
            self.admin_token = result["access_token"]
            print(f"✅ Admin registered successfully: {admin_data['email']}")
            return result
        else:
            print(f"❌ Admin registration failed: {response.status_code}")
            print(f"Response: {response.text}")
            sys.exit(1)
    
    def create_venue(self, venue_data: Dict) -> Dict:
        """Create a venue"""
        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        response = self.session.post(
            f"{self.base_url}/api/v1/admin/venues",
            json=venue_data,
            headers=headers
        )
        
        if response.status_code == 201:
            return response.json()
        else:
            print(f"❌ Venue creation failed: {response.status_code}")
            print(f"Venue: {venue_data['name']}")
            print(f"Response: {response.text}")
            return None
    
    def create_event(self, event_data: Dict) -> Dict:
        """Create an event"""
        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        response = self.session.post(
            f"{self.base_url}/api/v1/admin/events",
            json=event_data,
            headers=headers
        )
        
        if response.status_code == 201:
            return response.json()
        else:
            print(f"❌ Event creation failed: {response.status_code}")
            print(f"Event: {event_data['name']}")
            print(f"Response: {response.text}")
            return None
    
    def publish_event(self, event_id: str, version: int) -> bool:
        """Publish an event (make it publicly visible)"""
        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        update_data = {
            "status": "published",
            "version": version
        }
        
        response = self.session.put(
            f"{self.base_url}/api/v1/admin/events/{event_id}",
            json=update_data,
            headers=headers
        )
        
        return response.status_code == 200

def generate_venues() -> List[Dict]:
    """Generate diverse venue data"""
    venues = [
        # Concert Halls & Arenas
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
            "name": "Red Rocks Amphitheatre",
            "address": "18300 W Alameda Pkwy",
            "city": "Morrison",
            "state": "CO",
            "country": "USA",
            "postal_code": "80465",
            "capacity": 9525
        },
        {
            "name": "Hollywood Bowl",
            "address": "2301 Highland Ave",
            "city": "Los Angeles",
            "state": "CA",
            "country": "USA",
            "postal_code": "90068",
            "capacity": 17500
        },
        {
            "name": "Royal Albert Hall",
            "address": "Kensington Gore",
            "city": "London",
            "state": "",
            "country": "UK",
            "postal_code": "SW7 2AP",
            "capacity": 5272
        },
        {
            "name": "Sydney Opera House",
            "address": "Bennelong Point",
            "city": "Sydney",
            "state": "NSW",
            "country": "Australia",
            "postal_code": "2000",
            "capacity": 2679
        },
        
        # Sports Venues
        {
            "name": "Wembley Stadium",
            "address": "Wembley",
            "city": "London",
            "state": "",
            "country": "UK",
            "postal_code": "HA9 0WS",
            "capacity": 90000
        },
        {
            "name": "MetLife Stadium",
            "address": "1 MetLife Stadium Dr",
            "city": "East Rutherford",
            "state": "NJ",
            "country": "USA",
            "postal_code": "07073",
            "capacity": 82500
        },
        
        # Conference Centers
        {
            "name": "Moscone Center",
            "address": "747 Howard St",
            "city": "San Francisco",
            "state": "CA",
            "country": "USA",
            "postal_code": "94103",
            "capacity": 15000
        },
        {
            "name": "Las Vegas Convention Center",
            "address": "3150 Paradise Rd",
            "city": "Las Vegas",
            "state": "NV",
            "country": "USA",
            "postal_code": "89109",
            "capacity": 25000
        },
        
        # Theaters & Smaller Venues
        {
            "name": "The Apollo Theater",
            "address": "253 W 125th St",
            "city": "New York",
            "state": "NY",
            "country": "USA",
            "postal_code": "10027",
            "capacity": 1506
        },
        {
            "name": "The Troubadour",
            "address": "9081 Santa Monica Blvd",
            "city": "West Hollywood",
            "state": "CA",
            "country": "USA",
            "postal_code": "90069",
            "capacity": 500
        },
        {
            "name": "Blue Note",
            "address": "131 W 3rd St",
            "city": "New York",
            "state": "NY",
            "country": "USA",
            "postal_code": "10012",
            "capacity": 300
        },
        
        # Festival Grounds
        {
            "name": "Coachella Valley Music Festival Grounds",
            "address": "81800 51st Ave",
            "city": "Indio",
            "state": "CA",
            "country": "USA",
            "postal_code": "92201",
            "capacity": 125000
        },
        {
            "name": "Austin City Limits Music Festival Grounds",
            "address": "2100 Barton Springs Rd",
            "city": "Austin",
            "state": "TX",
            "country": "USA",
            "postal_code": "78746",
            "capacity": 75000
        },
        
        # International Venues
        {
            "name": "Berghain",
            "address": "Am Wriezener Bahnhof",
            "city": "Berlin",
            "state": "",
            "country": "Germany",
            "postal_code": "10243",
            "capacity": 1500
        },
        {
            "name": "Tokyo Dome",
            "address": "1-3-61 Koraku",
            "city": "Tokyo",
            "state": "",
            "country": "Japan",
            "postal_code": "112-0004",
            "capacity": 55000
        }
    ]
    
    return venues

def generate_events(venues: List[Dict]) -> List[Dict]:
    """Generate diverse event data using the created venues"""
    
    # Event templates with realistic data
    event_templates = [
        # Music Events
        {
            "name": "The Rolling Stones World Tour 2025",
            "description": "Legendary rock band returns with their biggest hits and new material",
            "event_type": "concert",
            "base_price": 150.0,
            "max_tickets_per_booking": 8,
            "suitable_venues": ["Madison Square Garden", "Wembley Stadium", "MetLife Stadium", "Tokyo Dome"]
        },
        {
            "name": "Taylor Swift: Eras Tour",
            "description": "Pop sensation performing songs from every era of her career",
            "event_type": "concert",
            "base_price": 200.0,
            "max_tickets_per_booking": 6,
            "suitable_venues": ["Madison Square Garden", "Wembley Stadium", "MetLife Stadium"]
        },
        {
            "name": "Jazz at the Blue Note",
            "description": "Intimate jazz performance featuring world-renowned artists",
            "event_type": "concert",
            "base_price": 85.0,
            "max_tickets_per_booking": 4,
            "suitable_venues": ["Blue Note", "The Apollo Theater"]
        },
        {
            "name": "Electronic Music Festival",
            "description": "Three-day electronic music festival with top DJs from around the world",
            "event_type": "festival",
            "base_price": 300.0,
            "max_tickets_per_booking": 10,
            "suitable_venues": ["Coachella Valley Music Festival Grounds", "Austin City Limits Music Festival Grounds"]
        },
        {
            "name": "Classical Symphony Orchestra",
            "description": "World-class orchestra performing Beethoven's complete symphonies",
            "event_type": "concert",
            "base_price": 120.0,
            "max_tickets_per_booking": 6,
            "suitable_venues": ["Royal Albert Hall", "Sydney Opera House", "Hollywood Bowl"]
        },
        
        # Sports Events
        {
            "name": "World Cup Final 2026",
            "description": "The ultimate football showdown - World Cup Final",
            "event_type": "sports",
            "base_price": 500.0,
            "max_tickets_per_booking": 4,
            "suitable_venues": ["Wembley Stadium", "MetLife Stadium"]
        },
        {
            "name": "NBA Finals Game 7",
            "description": "Winner-takes-all championship game",
            "event_type": "sports",
            "base_price": 400.0,
            "max_tickets_per_booking": 8,
            "suitable_venues": ["Madison Square Garden"]
        },
        {
            "name": "Champions League Final",
            "description": "Europe's premier club football competition final",
            "event_type": "sports",
            "base_price": 350.0,
            "max_tickets_per_booking": 6,
            "suitable_venues": ["Wembley Stadium"]
        },
        
        # Tech Conferences
        {
            "name": "TechCrunch Disrupt 2025",
            "description": "The world's leading tech conference featuring startup pitches and industry leaders",
            "event_type": "conference",
            "base_price": 1200.0,
            "max_tickets_per_booking": 5,
            "suitable_venues": ["Moscone Center", "Las Vegas Convention Center"]
        },
        {
            "name": "AI & Machine Learning Summit",
            "description": "Deep dive into the latest AI technologies and applications",
            "event_type": "conference",
            "base_price": 800.0,
            "max_tickets_per_booking": 3,
            "suitable_venues": ["Moscone Center", "Las Vegas Convention Center"]
        },
        {
            "name": "DevOps World Conference",
            "description": "Best practices in software development and operations",
            "event_type": "conference",
            "base_price": 600.0,
            "max_tickets_per_booking": 4,
            "suitable_venues": ["Moscone Center", "Las Vegas Convention Center"]
        },
        
        # Theater & Arts
        {
            "name": "Hamilton: An American Musical",
            "description": "The revolutionary musical about Alexander Hamilton",
            "event_type": "theater",
            "base_price": 180.0,
            "max_tickets_per_booking": 8,
            "suitable_venues": ["The Apollo Theater"]
        },
        {
            "name": "Shakespeare in the Park",
            "description": "Classic Shakespearean plays performed in outdoor venues",
            "event_type": "theater",
            "base_price": 50.0,
            "max_tickets_per_booking": 6,
            "suitable_venues": ["Hollywood Bowl", "Red Rocks Amphitheatre"]
        },
        
        # Comedy
        {
            "name": "Dave Chappelle: Stand-Up Special",
            "description": "Comedy legend performing his latest material",
            "event_type": "comedy",
            "base_price": 95.0,
            "max_tickets_per_booking": 6,
            "suitable_venues": ["The Apollo Theater", "The Troubadour"]
        },
        {
            "name": "Comedy Central Roast",
            "description": "Celebrity roast with top comedians",
            "event_type": "comedy",
            "base_price": 75.0,
            "max_tickets_per_booking": 4,
            "suitable_venues": ["The Apollo Theater"]
        },
        
        # Cultural Events
        {
            "name": "International Food Festival",
            "description": "Taste cuisines from around the world with live cooking demonstrations",
            "event_type": "festival",
            "base_price": 45.0,
            "max_tickets_per_booking": 10,
            "suitable_venues": ["Austin City Limits Music Festival Grounds", "Coachella Valley Music Festival Grounds"]
        },
        {
            "name": "Art Basel Contemporary Art Fair",
            "description": "Premier international art fair showcasing contemporary works",
            "event_type": "exhibition",
            "base_price": 65.0,
            "max_tickets_per_booking": 4,
            "suitable_venues": ["Las Vegas Convention Center", "Moscone Center"]
        }
    ]
    
    # Create venue lookup
    venue_lookup = {venue["name"]: venue for venue in venues}
    
    events = []
    used_combinations = set()
    
    # Generate events by combining templates with venues
    for template in event_templates:
        for venue_name in template["suitable_venues"]:
            if venue_name in venue_lookup:
                venue = venue_lookup[venue_name]
                
                # Create multiple time slots for popular events
                time_slots = generate_time_slots()
                
                for i, (start_time, end_time) in enumerate(time_slots[:3]):  # Max 3 shows per template-venue combo
                    # Create unique event name
                    event_name = template["name"]
                    if i > 0:
                        event_name += f" - Show {i+1}"
                    
                    # Avoid duplicates
                    combo_key = (event_name, venue_name, start_time.date())
                    if combo_key in used_combinations:
                        continue
                    used_combinations.add(combo_key)
                    
                    # Calculate capacity (use percentage of venue capacity)
                    capacity_percentage = random.uniform(0.6, 1.0)  # 60-100% of venue capacity
                    event_capacity = int(venue["capacity"] * capacity_percentage)
                    
                    # Add some price variation
                    price_variation = random.uniform(0.8, 1.3)
                    final_price = template["base_price"] * price_variation
                    
                    event = {
                        "name": event_name,
                        "description": template["description"],
                        "venue_id": venue["venue_id"],  # Will be set after venue creation
                        "event_type": template["event_type"],
                        "start_datetime": start_time.isoformat() + "Z",
                        "end_datetime": end_time.isoformat() + "Z",
                        "total_capacity": event_capacity,
                        "base_price": round(final_price, 2),
                        "max_tickets_per_booking": template["max_tickets_per_booking"]
                    }
                    
                    events.append(event)
    
    # Shuffle to randomize order
    random.shuffle(events)
    
    return events

def generate_time_slots() -> List[tuple]:
    """Generate realistic event time slots"""
    base_date = datetime.now() + timedelta(days=30)  # Start 30 days from now
    time_slots = []
    
    # Generate various time slots over the next 6 months
    for week in range(26):  # 26 weeks = ~6 months
        event_date = base_date + timedelta(weeks=week)
        
        # Weekend events (Friday, Saturday, Sunday)
        for day_offset in [4, 5, 6]:  # Fri, Sat, Sun
            event_day = event_date + timedelta(days=(day_offset - event_date.weekday()) % 7)
            
            # Different time slots
            time_options = [
                (19, 0, 22, 0),   # 7 PM - 10 PM
                (20, 0, 23, 0),   # 8 PM - 11 PM
                (14, 0, 17, 0),   # 2 PM - 5 PM (matinee)
                (15, 30, 18, 30), # 3:30 PM - 6:30 PM
            ]
            
            for start_hour, start_min, end_hour, end_min in time_options:
                start_time = event_day.replace(hour=start_hour, minute=start_min, second=0, microsecond=0)
                end_time = event_day.replace(hour=end_hour, minute=end_min, second=0, microsecond=0)
                
                # Skip past dates
                if start_time > datetime.now():
                    time_slots.append((start_time, end_time))
    
    return time_slots

def main():
    parser = argparse.ArgumentParser(description="Ingest events into the event service")
    parser.add_argument("--events", type=int, default=DEFAULT_EVENT_COUNT,
                       help=f"Number of events to create (default: {DEFAULT_EVENT_COUNT})")
    parser.add_argument("--host", default=DEFAULT_HOST,
                       help=f"Event service host:port (default: {DEFAULT_HOST})")
    parser.add_argument("--dry-run", action="store_true",
                       help="Show what would be created without actually creating")
    
    args = parser.parse_args()
    
    print(f"🚀 Event Service Data Ingestion Script")
    print(f"📊 Target: {args.events} events on {args.host}")
    print(f"🔄 Mode: {'DRY RUN' if args.dry_run else 'LIVE'}")
    print("=" * 60)
    
    if args.dry_run:
        venues = generate_venues()
        events = generate_events(venues)
        
        print(f"📍 Would create {len(venues)} venues:")
        for venue in venues[:5]:  # Show first 5
            print(f"   • {venue['name']} ({venue['city']}, {venue['country']}) - {venue['capacity']:,} capacity")
        if len(venues) > 5:
            print(f"   ... and {len(venues) - 5} more venues")
        
        print(f"\n🎪 Would create {len(events)} events (showing first 10):")
        for event in events[:10]:
            print(f"   • {event['name']} - {event['event_type']} - ${event['base_price']}")
        if len(events) > 10:
            print(f"   ... and {len(events) - 10} more events")
        
        print(f"\n📈 Would then publish all {len(events)} events")
        return
    
    # Initialize client
    client = EventServiceClient(args.host)
    
    # Step 1: Register admin
    print("👤 Registering admin account...")
    admin_info = client.register_admin()
    
    # Step 2: Create venues
    print("\n📍 Creating venues...")
    venues_data = generate_venues()
    created_venues = []
    
    for i, venue_data in enumerate(venues_data):
        print(f"   Creating venue {i+1}/{len(venues_data)}: {venue_data['name']}")
        created_venue = client.create_venue(venue_data)
        if created_venue:
            venue_data["venue_id"] = created_venue["venue_id"]
            created_venues.append(venue_data)
            print(f"   ✅ Created: {created_venue['venue_id']}")
        else:
            print(f"   ❌ Failed to create venue: {venue_data['name']}")
        
        time.sleep(0.1)  # Small delay to avoid overwhelming the server
    
    print(f"✅ Successfully created {len(created_venues)}/{len(venues_data)} venues")
    
    # Step 3: Generate and create events
    print(f"\n🎪 Generating events...")
    events_data = generate_events(created_venues)
    
    # Limit to requested number
    events_data = events_data[:args.events]
    
    print(f"📊 Creating {len(events_data)} events...")
    created_events = []
    
    for i, event_data in enumerate(events_data):
        print(f"   Creating event {i+1}/{len(events_data)}: {event_data['name']}")
        created_event = client.create_event(event_data)
        if created_event:
            created_events.append(created_event)
            print(f"   ✅ Created: {created_event['event_id']}")
        else:
            print(f"   ❌ Failed to create event: {event_data['name']}")
        
        time.sleep(0.1)  # Small delay
    
    print(f"✅ Successfully created {len(created_events)}/{len(events_data)} events")
    
    # Step 4: Publish events
    print(f"\n📢 Publishing events...")
    published_count = 0
    
    for i, event in enumerate(created_events):
        print(f"   Publishing event {i+1}/{len(created_events)}: {event['name']}")
        if client.publish_event(event["event_id"], event["version"]):
            published_count += 1
            print(f"   ✅ Published")
        else:
            print(f"   ❌ Failed to publish")
        
        time.sleep(0.1)  # Small delay
    
    print(f"✅ Successfully published {published_count}/{len(created_events)} events")
    
    # Summary
    print("\n" + "=" * 60)
    print("📊 INGESTION SUMMARY")
    print("=" * 60)
    print(f"👤 Admin Account: {admin_info['email']}")
    print(f"📍 Venues Created: {len(created_venues)}")
    print(f"🎪 Events Created: {len(created_events)}")
    print(f"📢 Events Published: {published_count}")
    print(f"🌐 Event Service: {client.base_url}")
    print("=" * 60)
    
    if published_count > 0:
        print(f"✅ SUCCESS! {published_count} events are now live and bookable!")
        print(f"🔗 Check them out at: {client.base_url}/api/v1/events")
    else:
        print("❌ No events were successfully published.")

if __name__ == "__main__":
    main()
