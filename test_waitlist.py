#!/usr/bin/env python3

import requests
import json
import time
import threading
from concurrent.futures import ThreadPoolExecutor

BASE_URL_USER = "http://localhost:8001"
BASE_URL_BOOKING = "http://localhost:8004"
EVENT_ID = "7204c97d-ae65-4334-86cb-3e834e9b12cf"  # Rolling Stones event

def create_user(email, name):
    """Create a new user"""
    data = {"email": email, "password": "test123", "name": name}
    response = requests.post(f"{BASE_URL_USER}/api/v1/auth/register", json=data)
    return response.json() if response.status_code in [200, 201] else None

def check_availability(event_id, quantity):
    """Check availability"""
    response = requests.get(f"{BASE_URL_BOOKING}/api/v1/bookings/check-availability", 
                          params={"event_id": event_id, "quantity": quantity})
    return response.json() if response.status_code == 200 else None

def reserve_seats(token, event_id, quantity, idempotency_key):
    """Reserve seats"""
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    data = {"event_id": event_id, "quantity": quantity, "idempotency_key": idempotency_key}
    
    response = requests.post(f"{BASE_URL_BOOKING}/api/v1/bookings/reserve", 
                           headers=headers, json=data)
    return response.json(), response.status_code

def join_waitlist(token, event_id, quantity):
    """Join waitlist when event is sold out"""
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    data = {"event_id": event_id, "quantity": quantity}
    
    response = requests.post(f"{BASE_URL_BOOKING}/api/v1/waitlist/join", 
                           headers=headers, json=data)
    return response.json(), response.status_code

def get_waitlist_position(token, event_id):
    """Get current waitlist position"""
    headers = {"Authorization": f"Bearer {token}"}
    
    response = requests.get(f"{BASE_URL_BOOKING}/api/v1/waitlist/position", 
                          headers=headers, params={"event_id": event_id})
    return response.json(), response.status_code

def leave_waitlist(token, event_id):
    """Leave waitlist"""
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    data = {"event_id": event_id}
    
    response = requests.delete(f"{BASE_URL_BOOKING}/api/v1/waitlist/leave", 
                             headers=headers, json=data)
    return response.json(), response.status_code

def test_waitlist_functionality():
    """Test waitlist functionality"""
    print("🎯 WAITLIST FUNCTIONALITY TEST")
    print("=" * 50)
    
    # Check current availability
    avail = check_availability(EVENT_ID, 1)
    if not avail:
        print("❌ Cannot check availability")
        return
        
    print(f"Current available seats: {avail['available_seats']}")
    
    # Create multiple users to test waitlist
    users = []
    for i in range(5):
        user = create_user(f"waitlist{i}@test.com", f"Waitlist User {i}")
        if user:
            users.append(user)
    
    print(f"Created {len(users)} users for waitlist testing")
    
    # Test 1: Try to book when seats are still available
    print("\n📊 Test 1: Normal booking when seats available")
    if avail['available_seats'] > 0:
        user1 = users[0]
        result, status = reserve_seats(user1['access_token'], EVENT_ID, 1, f"waitlist-test-1")
        
        if status == 200:
            print(f"✅ {user1['name']}: Successfully reserved seat")
            print(f"   Booking: {result['booking_reference']}")
        else:
            print(f"❌ {user1['name']}: Failed to reserve - {result}")
    
    # Test 2: Try to book more seats than available (should trigger waitlist suggestion)
    print("\n📊 Test 2: Attempt to book more seats than available")
    user2 = users[1] 
    large_quantity = avail['available_seats'] + 100  # Request way more than available
    
    result, status = reserve_seats(user2['access_token'], EVENT_ID, large_quantity, f"waitlist-test-2")
    print(f"{user2['name']} trying to book {large_quantity} seats:")
    print(f"  Status: {status}")
    print(f"  Response: {result}")
    
    # Test 3: Join waitlist
    print("\n📊 Test 3: Join waitlist")
    waitlist_results = []
    
    for i, user in enumerate(users[2:]):  # Use remaining users
        result, status = join_waitlist(user['access_token'], EVENT_ID, 2)
        waitlist_results.append((user, result, status))
        
        if status == 200:
            print(f"✅ {user['name']}: Joined waitlist at position {result.get('position', 'N/A')}")
        else:
            print(f"❌ {user['name']}: Failed to join waitlist - {result}")
    
    # Test 4: Check waitlist positions
    print("\n📊 Test 4: Check waitlist positions")
    for user, _, status in waitlist_results:
        if status == 200:
            position_result, pos_status = get_waitlist_position(user['access_token'], EVENT_ID)
            
            if pos_status == 200:
                print(f"👥 {user['name']}: Position {position_result.get('position', 'N/A')}, "
                      f"Total waiting: {position_result.get('total_waiting', 'N/A')}")
            else:
                print(f"❌ {user['name']}: Could not get position - {position_result}")
    
    # Test 5: Leave waitlist
    print("\n📊 Test 5: Leave waitlist")
    if waitlist_results:
        user_to_leave = waitlist_results[0][0]  # First user in waitlist
        result, status = leave_waitlist(user_to_leave['access_token'], EVENT_ID)
        
        if status == 200:
            print(f"✅ {user_to_leave['name']}: Successfully left waitlist")
            
            # Check if positions updated for remaining users
            print("   Checking if positions updated for remaining users:")
            for user, _, prev_status in waitlist_results[1:]:
                if prev_status == 200:
                    position_result, pos_status = get_waitlist_position(user['access_token'], EVENT_ID)
                    if pos_status == 200:
                        print(f"   👥 {user['name']}: New position {position_result.get('position', 'N/A')}")
        else:
            print(f"❌ {user_to_leave['name']}: Failed to leave waitlist - {result}")

if __name__ == "__main__":
    test_waitlist_functionality()