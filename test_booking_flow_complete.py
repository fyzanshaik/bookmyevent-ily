#!/usr/bin/env python3

import requests
import json
import time
import random
import string
from concurrent.futures import ThreadPoolExecutor

BASE_URL_USER = "http://localhost:8001"
BASE_URL_EVENT = "http://localhost:8002"
BASE_URL_BOOKING = "http://localhost:8004"

def random_string(length=8):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def create_user(email, name):
    data = {
        "email": email,
        "password": "testpass123",
        "name": name
    }
    response = requests.post(f"{BASE_URL_USER}/api/v1/auth/register", json=data)
    return response.json() if response.status_code == 201 else None

def get_events():
    response = requests.get(f"{BASE_URL_EVENT}/api/v1/events")
    if response.status_code == 200:
        events = response.json().get("events", [])
        published_events = [e for e in events if e.get("status") == "published"]
        if published_events:
            return published_events
    
    # Fallback to known test event from previous runs
    test_event = {
        "event_id": "92935cce-79cd-4380-ba62-f203c8dc7392",
        "name": "Test Event (Fallback)",
        "status": "published"
    }
    
    # Verify this event exists by checking availability
    avail = check_availability(test_event["event_id"], 1)
    if avail and avail.get("available"):
        return [test_event]
    
    return []

def check_availability(event_id, quantity):
    response = requests.get(f"{BASE_URL_BOOKING}/api/v1/bookings/check-availability", 
                          params={"event_id": event_id, "quantity": quantity})
    return response.json() if response.status_code == 200 else None

def reserve_seats(token, event_id, quantity, idempotency_key):
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    data = {
        "event_id": event_id,
        "quantity": quantity,
        "idempotency_key": idempotency_key
    }
    response = requests.post(f"{BASE_URL_BOOKING}/api/v1/bookings/reserve", 
                           headers=headers, json=data)
    return response.json(), response.status_code

def confirm_booking(token, reservation_id):
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    data = {
        "reservation_id": reservation_id,
        "payment_token": f"mock-token-{random_string()}",
        "payment_method": "credit_card"
    }
    response = requests.post(f"{BASE_URL_BOOKING}/api/v1/bookings/confirm", 
                           headers=headers, json=data)
    return response.json(), response.status_code

def get_booking_details(token, booking_id):
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.get(f"{BASE_URL_BOOKING}/api/v1/bookings/{booking_id}", 
                          headers=headers)
    return response.json(), response.status_code

def cancel_booking(token, booking_id):
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.delete(f"{BASE_URL_BOOKING}/api/v1/bookings/{booking_id}", 
                             headers=headers)
    return response.json(), response.status_code

def user_booking_flow(user_data, event_id, quantity):
    results = {"user": user_data["name"], "steps": []}
    
    idempotency_key = f"{user_data['name']}-{random_string()}-{int(time.time())}"
    
    print(f"\n=== {user_data['name']} Booking Flow ===")
    
    avail = check_availability(event_id, quantity)
    results["steps"].append({"step": "check_availability", "success": avail is not None, "data": avail})
    print(f"✓ Availability: {avail['available_seats']} seats available")
    
    reserve_result, status = reserve_seats(user_data["access_token"], event_id, quantity, idempotency_key)
    reserve_success = status == 200 and "reservation_id" in reserve_result
    results["steps"].append({"step": "reserve_seats", "success": reserve_success, "status": status, "data": reserve_result})
    
    if reserve_success:
        print(f"✓ Reserved: {reserve_result['booking_reference']} - ${reserve_result['total_amount']}")
        reservation_id = reserve_result["reservation_id"]
        
        confirm_result, status = confirm_booking(user_data["access_token"], reservation_id)
        confirm_success = status == 200 and confirm_result.get("status") == "confirmed"
        results["steps"].append({"step": "confirm_booking", "success": confirm_success, "status": status, "data": confirm_result})
        
        if confirm_success:
            print(f"✓ Confirmed: {confirm_result['booking_reference']} - Ticket: {confirm_result['ticket_url']}")
            booking_id = confirm_result["booking_id"]
            
            details_result, status = get_booking_details(user_data["access_token"], booking_id)
            details_success = status == 200 and details_result.get("status") == "confirmed"
            results["steps"].append({"step": "get_details", "success": details_success, "status": status, "data": details_result})
            print(f"✓ Details: Status={details_result.get('status')}, Amount=${details_result.get('total_amount')}")
            
            cancel_result, status = cancel_booking(user_data["access_token"], booking_id)
            cancel_success = status == 200 and "cancelled" in cancel_result.get("message", "")
            results["steps"].append({"step": "cancel_booking", "success": cancel_success, "status": status, "data": cancel_result})
            print(f"✓ Cancelled: Refund=${cancel_result.get('refund_amount')}")
            
            final_details, status = get_booking_details(user_data["access_token"], booking_id)
            verify_cancelled = status == 200 and final_details.get("status") == "cancelled"
            results["steps"].append({"step": "verify_cancelled", "success": verify_cancelled, "status": status, "data": final_details})
            print(f"✓ Verified: Status={final_details.get('status')}")
            
        else:
            print(f"✗ Confirm failed: {confirm_result}")
    else:
        print(f"✗ Reserve failed: {reserve_result}")
    
    return results

def main():
    print("🚀 BookMyEvent Complete Flow Test")
    print("=" * 50)
    
    user1_email = f"batman-{random_string()}@example.com"
    user2_email = f"robin-{random_string()}@example.com"
    
    user1 = create_user(user1_email, "Batman")
    user2 = create_user(user2_email, "Robin")
    
    if not user1 or not user2:
        print("❌ Failed to create users")
        return
    
    print(f"✓ Created users: {user1['name']} & {user2['name']}")
    
    events = get_events()
    if not events:
        print("❌ No events available")
        return
    
    test_event = events[0]
    event_id = test_event["event_id"]
    print(f"✓ Using event: {test_event['name']} (ID: {event_id})")
    
    print("\n🎯 Testing Concurrent Booking Flow")
    with ThreadPoolExecutor(max_workers=2) as executor:
        future1 = executor.submit(user_booking_flow, user1, event_id, 2)
        future2 = executor.submit(user_booking_flow, user2, event_id, 2)
        
        results1 = future1.result()
        results2 = future2.result()
    
    print("\n📊 FINAL RESULTS")
    print("=" * 50)
    
    for user_result in [results1, results2]:
        user = user_result["user"]
        success_count = sum(1 for step in user_result["steps"] if step["success"])
        total_steps = len(user_result["steps"])
        print(f"{user}: {success_count}/{total_steps} steps successful")
        
        for step in user_result["steps"]:
            status = "✓" if step["success"] else "✗"
            print(f"  {status} {step['step']}")
    
    print("\n🎉 Test Complete!")

if __name__ == "__main__":
    main()