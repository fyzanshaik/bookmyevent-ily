#!/usr/bin/env python3

import requests
import json
import time
import threading
import random
from concurrent.futures import ThreadPoolExecutor, as_completed

# User tokens from concurrent user creation
USERS = [
    {"name": "User 1", "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJldmVudGx5Iiwic3ViIjoiNmZhZmNlMjYtNWNhOC00MWM5LThhYTMtNDBhNGNmYjBhOTQyIiwiZXhwIjoxNzU3ODY1OTI2LCJpYXQiOjE3NTc4NjUwMjZ9.3xgxsky5uxC54QlBaWXJXyFQRjFgtdBvv7tjUgqIKXY"},
    {"name": "User 2", "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJldmVudGx5Iiwic3ViIjoiYzBiZmMwOTgtMjI5OS00MTkzLWJkNzUtMTk5OWQzYzNiNjU3IiwiZXhwIjoxNzU3ODY1OTI2LCJpYXQiOjE3NTc4NjUwMjZ9.p3v_znhcW9HiWtngLjHvDEN6261hg1IFCDtunK_HDlA"},
    {"name": "User 3", "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJldmVudGx5Iiwic3ViIjoiZWY0ZDcxMWUtMTM1YS00MWU3LTgwMzktNGYyYWIyZDFhMmNjIiwiZXhwIjoxNzU3ODY1OTI2LCJpYXQiOjE3NTc4NjUwMjZ9.LcsLWlgmNwyJ1kFJvudaBN_COYPB8xWfSb7WIDBUx6E"}
]

EVENT_ID = "7204c97d-ae65-4334-86cb-3e834e9b12cf"  # Rolling Stones event
BASE_URL = "http://localhost:8004"

def check_availability(event_id, quantity):
    """Check current availability"""
    response = requests.get(f"{BASE_URL}/api/v1/bookings/check-availability", 
                          params={"event_id": event_id, "quantity": quantity})
    return response.json() if response.status_code == 200 else None

def reserve_seats(user_name, token, event_id, quantity, attempt_num):
    """Attempt to reserve seats - returns detailed results"""
    start_time = time.time()
    
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    idempotency_key = f"{user_name}-attempt-{attempt_num}-{int(time.time()*1000)}"
    
    data = {
        "event_id": event_id,
        "quantity": quantity,
        "idempotency_key": idempotency_key
    }
    
    try:
        response = requests.post(f"{BASE_URL}/api/v1/bookings/reserve", 
                               headers=headers, json=data, timeout=10)
        end_time = time.time()
        
        return {
            "user": user_name,
            "attempt": attempt_num,
            "status_code": response.status_code,
            "success": response.status_code == 200,
            "response": response.json(),
            "latency_ms": round((end_time - start_time) * 1000, 2),
            "idempotency_key": idempotency_key
        }
    except Exception as e:
        return {
            "user": user_name, 
            "attempt": attempt_num,
            "status_code": None,
            "success": False,
            "response": {"error": str(e)},
            "latency_ms": None,
            "idempotency_key": idempotency_key
        }

def concurrent_booking_test():
    """Test concurrent booking attempts on same event"""
    print("🎯 CONCURRENT BOOKING TEST")
    print("=" * 50)
    
    # Check initial availability
    initial_avail = check_availability(EVENT_ID, 2)
    if not initial_avail:
        print("❌ Cannot check availability")
        return
        
    print(f"Initial available seats: {initial_avail['available_seats']}")
    print(f"Max per booking: {initial_avail['max_per_booking']}")
    print(f"Base price: ${initial_avail['base_price']}")
    
    # Test 1: 3 users try to book 2 seats each simultaneously
    print("\n🔥 TEST 1: 3 users booking 2 seats each SIMULTANEOUSLY")
    print("-" * 50)
    
    with ThreadPoolExecutor(max_workers=3) as executor:
        futures = []
        
        # Start all requests at exactly the same time
        for i, user in enumerate(USERS):
            future = executor.submit(reserve_seats, user["name"], user["token"], EVENT_ID, 2, 1)
            futures.append(future)
        
        results = []
        for future in as_completed(futures):
            result = future.result()
            results.append(result)
    
    # Sort results by completion order
    results.sort(key=lambda x: x.get("latency_ms", float('inf')) if x.get("latency_ms") else float('inf'))
    
    successful_bookings = []
    failed_bookings = []
    
    print("Results (in completion order):")
    for result in results:
        status = "✅ SUCCESS" if result["success"] else "❌ FAILED"
        latency = f"{result['latency_ms']}ms" if result['latency_ms'] else "N/A"
        
        print(f"  {status} - {result['user']} (Status: {result['status_code']}, Latency: {latency})")
        
        if result["success"]:
            successful_bookings.append(result)
            booking_ref = result["response"].get("booking_reference", "N/A")
            expires_at = result["response"].get("expires_at", "N/A") 
            print(f"    Booking: {booking_ref}, Expires: {expires_at}")
        else:
            failed_bookings.append(result)
            error_msg = result["response"].get("error", "Unknown error")
            print(f"    Error: {error_msg}")
    
    print(f"\n📊 Summary: {len(successful_bookings)} successful, {len(failed_bookings)} failed")
    
    # Check final availability after concurrent attempts
    final_avail = check_availability(EVENT_ID, 2)
    if final_avail:
        seats_consumed = initial_avail["available_seats"] - final_avail["available_seats"] 
        print(f"Seats consumed: {seats_consumed} (Expected: {len(successful_bookings) * 2})")
        if seats_consumed == len(successful_bookings) * 2:
            print("✅ Seat count perfectly consistent!")
        else:
            print("❌ Seat count inconsistency detected!")
    
    return successful_bookings, failed_bookings

def high_concurrency_test():
    """Test with many simultaneous users (race condition stress test)"""
    print("\n\n🚀 HIGH CONCURRENCY STRESS TEST")
    print("=" * 50)
    print("Creating 10 rapid concurrent booking attempts...")
    
    # Create more users for stress testing
    stress_users = []
    for i in range(10):
        user_data = {
            "email": f"stress{i}@test.com",
            "password": "test123",
            "name": f"Stress User {i}"
        }
        
        response = requests.post("http://localhost:8001/api/v1/auth/register", json=user_data)
        if response.status_code in [200, 201]:
            user_info = response.json()
            stress_users.append({
                "name": user_info["name"],
                "token": user_info["access_token"]
            })
    
    print(f"Created {len(stress_users)} stress test users")
    
    # All users try to book 1 seat each at the exact same time
    with ThreadPoolExecutor(max_workers=10) as executor:
        futures = []
        
        for i, user in enumerate(stress_users):
            future = executor.submit(reserve_seats, user["name"], user["token"], EVENT_ID, 1, i+1)
            futures.append(future)
        
        results = []
        for future in as_completed(futures):
            result = future.result()
            results.append(result)
    
    successful = [r for r in results if r["success"]]
    failed = [r for r in results if not r["success"]]
    
    print(f"\n📊 High Concurrency Results:")
    print(f"  ✅ Successful: {len(successful)}")
    print(f"  ❌ Failed: {len(failed)}")
    print(f"  📈 Success Rate: {len(successful)/len(results)*100:.1f}%")
    
    # Show latency distribution
    latencies = [r["latency_ms"] for r in results if r["latency_ms"]]
    if latencies:
        print(f"  ⚡ Average Latency: {sum(latencies)/len(latencies):.2f}ms")
        print(f"  ⚡ Max Latency: {max(latencies):.2f}ms")
        print(f"  ⚡ Min Latency: {min(latencies):.2f}ms")

if __name__ == "__main__":
    concurrent_booking_test()
    high_concurrency_test()