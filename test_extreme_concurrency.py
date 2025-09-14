#!/usr/bin/env python3

import requests
import json
import time
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
import random

BASE_URL_USER = "http://localhost:8001"
BASE_URL_EVENT = "http://localhost:8002" 
BASE_URL_BOOKING = "http://localhost:8004"
EVENT_ID = "7204c97d-ae65-4334-86cb-3e834e9b12cf"

def create_user(email, name):
    """Create a user"""
    data = {"email": email, "password": "test123", "name": name}
    response = requests.post(f"{BASE_URL_USER}/api/v1/auth/register", json=data)
    return response.json() if response.status_code in [200, 201] else None

def get_event_availability():
    """Get availability from both services for comparison"""
    event_resp = requests.get(f"{BASE_URL_EVENT}/api/v1/events/{EVENT_ID}/availability")
    booking_resp = requests.get(f"{BASE_URL_BOOKING}/api/v1/bookings/check-availability", 
                              params={"event_id": EVENT_ID, "quantity": 1})
    
    event_data = event_resp.json() if event_resp.status_code == 200 else {}
    booking_data = booking_resp.json() if booking_resp.status_code == 200 else {}
    
    return {
        "event_service": event_data.get("available_seats", "N/A"),
        "booking_service": booking_data.get("available_seats", "N/A"),
        "consistent": event_data.get("available_seats") == booking_data.get("available_seats"),
        "timestamp": time.time()
    }

def rapid_booking_attempt(user_name, token, attempt_id):
    """Single rapid booking attempt"""
    start_time = time.time()
    
    # Phase 1: Reserve
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    idempotency_key = f"{user_name}-extreme-{attempt_id}-{int(time.time()*1000000)}"
    
    reserve_data = {
        "event_id": EVENT_ID,
        "quantity": 1,  # Book 1 seat to maximize concurrency
        "idempotency_key": idempotency_key
    }
    
    try:
        reserve_resp = requests.post(f"{BASE_URL_BOOKING}/api/v1/bookings/reserve", 
                                   headers=headers, json=reserve_data, timeout=5)
        reserve_time = time.time()
        
        result = {
            "user": user_name,
            "attempt_id": attempt_id,
            "idempotency_key": idempotency_key,
            "reserve_status": reserve_resp.status_code,
            "reserve_latency_ms": round((reserve_time - start_time) * 1000, 2),
            "reserve_success": reserve_resp.status_code == 200,
            "confirm_success": False,
            "total_latency_ms": round((reserve_time - start_time) * 1000, 2)
        }
        
        if reserve_resp.status_code == 200:
            reserve_result = reserve_resp.json()
            result["reservation_id"] = reserve_result["reservation_id"]
            result["booking_reference"] = reserve_result["booking_reference"]
            
            # Phase 2: Immediately confirm (simulate fast user)
            confirm_data = {
                "reservation_id": reserve_result["reservation_id"],
                "payment_token": f"mock-{random.randint(1000, 9999)}",
                "payment_method": "credit_card"
            }
            
            confirm_resp = requests.post(f"{BASE_URL_BOOKING}/api/v1/bookings/confirm",
                                       headers=headers, json=confirm_data, timeout=5)
            confirm_time = time.time()
            
            result["confirm_status"] = confirm_resp.status_code
            result["confirm_latency_ms"] = round((confirm_time - reserve_time) * 1000, 2) 
            result["confirm_success"] = confirm_resp.status_code == 200
            result["total_latency_ms"] = round((confirm_time - start_time) * 1000, 2)
            
            if confirm_resp.status_code == 200:
                confirm_result = confirm_resp.json()
                result["final_booking_id"] = confirm_result["booking_id"]
        else:
            result["reserve_error"] = reserve_resp.json() if reserve_resp.text else {"error": "No response"}
            
        return result
        
    except Exception as e:
        return {
            "user": user_name,
            "attempt_id": attempt_id, 
            "error": str(e),
            "reserve_success": False,
            "confirm_success": False
        }

def extreme_concurrency_test():
    """Test extreme concurrency scenarios"""
    print("🚀 EXTREME CONCURRENCY TEST - BOOKING SYSTEM STRESS TEST")
    print("=" * 70)
    
    # Get baseline
    baseline = get_event_availability()
    print(f"Baseline - Event Service: {baseline['event_service']}, Booking Service: {baseline['booking_service']}")
    print(f"Services consistent: {baseline['consistent']}")
    
    # Create army of users
    print(f"\n📥 Creating 25 concurrent users...")
    users = []
    
    with ThreadPoolExecutor(max_workers=10) as executor:
        user_futures = []
        for i in range(25):
            future = executor.submit(create_user, f"extreme{i}@test.com", f"Extreme User {i}")
            user_futures.append(future)
        
        for future in as_completed(user_futures):
            user = future.result()
            if user:
                users.append(user)
    
    print(f"✅ Created {len(users)} users successfully")
    
    if len(users) < 10:
        print("❌ Not enough users created, aborting test")
        return
    
    # WAVE 1: 15 users try to book simultaneously 
    print(f"\n⚡ WAVE 1: 15 users booking simultaneously")
    print("-" * 50)
    
    wave1_start = time.time()
    with ThreadPoolExecutor(max_workers=15) as executor:
        futures = []
        for i, user in enumerate(users[:15]):
            future = executor.submit(rapid_booking_attempt, user["name"], user["access_token"], f"wave1-{i}")
            futures.append(future)
        
        wave1_results = []
        for future in as_completed(futures):
            result = future.result()
            wave1_results.append(result)
    
    wave1_duration = time.time() - wave1_start
    
    # Analyze Wave 1
    wave1_reserve_success = sum(1 for r in wave1_results if r.get("reserve_success", False))
    wave1_confirm_success = sum(1 for r in wave1_results if r.get("confirm_success", False))
    
    print(f"📊 Wave 1 Results ({wave1_duration:.2f}s):")
    print(f"   Reserve Success: {wave1_reserve_success}/15 ({wave1_reserve_success/15*100:.1f}%)")
    print(f"   Confirm Success: {wave1_confirm_success}/15 ({wave1_confirm_success/15*100:.1f}%)")
    
    # Check consistency
    post_wave1 = get_event_availability()
    expected_consumed = wave1_confirm_success
    actual_consumed = baseline['event_service'] - post_wave1['event_service'] if isinstance(baseline['event_service'], int) and isinstance(post_wave1['event_service'], int) else 0
    
    print(f"   Expected seats consumed: {expected_consumed}")
    print(f"   Actual seats consumed: {actual_consumed}")
    print(f"   ✅ Data consistency: {'PASS' if expected_consumed == actual_consumed else 'FAIL'}")
    print(f"   Services in sync: {'YES' if post_wave1['consistent'] else 'NO'}")
    
    # WAVE 2: Remaining 10 users (simulate users arriving later)
    print(f"\n⚡ WAVE 2: 10 more users booking (after {wave1_duration:.1f}s delay)")
    print("-" * 50)
    
    wave2_start = time.time()
    with ThreadPoolExecutor(max_workers=10) as executor:
        futures = []
        for i, user in enumerate(users[15:]):
            future = executor.submit(rapid_booking_attempt, user["name"], user["access_token"], f"wave2-{i}")
            futures.append(future)
        
        wave2_results = []
        for future in as_completed(futures):
            result = future.result()
            wave2_results.append(result)
    
    wave2_duration = time.time() - wave2_start
    
    # Analyze Wave 2
    wave2_reserve_success = sum(1 for r in wave2_results if r.get("reserve_success", False))
    wave2_confirm_success = sum(1 for r in wave2_results if r.get("confirm_success", False))
    
    print(f"📊 Wave 2 Results ({wave2_duration:.2f}s):")
    print(f"   Reserve Success: {wave2_reserve_success}/10 ({wave2_reserve_success/10*100:.1f}%)")
    print(f"   Confirm Success: {wave2_confirm_success}/10 ({wave2_confirm_success/10*100:.1f}%)")
    
    # Final consistency check
    final_state = get_event_availability()
    total_expected_consumed = wave1_confirm_success + wave2_confirm_success
    total_actual_consumed = baseline['event_service'] - final_state['event_service'] if isinstance(baseline['event_service'], int) and isinstance(final_state['event_service'], int) else 0
    
    print(f"\n🎯 FINAL SYSTEM STATE:")
    print(f"   Total bookings expected: {total_expected_consumed}")
    print(f"   Total seats consumed: {total_actual_consumed}")
    print(f"   Final Event Service seats: {final_state['event_service']}")
    print(f"   Final Booking Service seats: {final_state['booking_service']}")
    print(f"   Services in perfect sync: {'✅ YES' if final_state['consistent'] else '❌ NO'}")
    print(f"   Data integrity: {'✅ PERFECT' if total_expected_consumed == total_actual_consumed else '❌ INCONSISTENT'}")
    
    # Performance summary
    all_results = wave1_results + wave2_results
    successful_results = [r for r in all_results if r.get("total_latency_ms")]
    
    if successful_results:
        avg_latency = sum(r["total_latency_ms"] for r in successful_results) / len(successful_results)
        max_latency = max(r["total_latency_ms"] for r in successful_results)
        min_latency = min(r["total_latency_ms"] for r in successful_results)
        
        print(f"\n⚡ PERFORMANCE METRICS:")
        print(f"   Average end-to-end latency: {avg_latency:.2f}ms")
        print(f"   Max latency: {max_latency:.2f}ms")
        print(f"   Min latency: {min_latency:.2f}ms")
        print(f"   Total successful bookings: {total_expected_consumed}")
        print(f"   Overall success rate: {total_expected_consumed/25*100:.1f}%")
    
    return {
        "total_attempts": 25,
        "total_success": total_expected_consumed,
        "data_consistent": total_expected_consumed == total_actual_consumed,
        "services_in_sync": final_state['consistent']
    }

if __name__ == "__main__":
    result = extreme_concurrency_test()
    
    print(f"\n🏆 TEST VERDICT:")
    if result["data_consistent"] and result["services_in_sync"]:
        print("✅ EXCELLENT: System handled extreme concurrency perfectly!")
        print("✅ Data consistency maintained across all services")
        print("✅ No race conditions or overselling detected")
    else:
        print("❌ ISSUES DETECTED:")
        if not result["data_consistent"]:
            print("❌ Data consistency failure")
        if not result["services_in_sync"]:
            print("❌ Service synchronization failure")