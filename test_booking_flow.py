#!/usr/bin/env python3

import requests
import json
import uuid
from datetime import datetime, timedelta

BASE_URLS = {
    'user': 'http://localhost:8001',
    'event': 'http://localhost:8002', 
    'booking': 'http://localhost:8004'
}

def create_user(email, password, first_name, last_name):
    url = f"{BASE_URLS['user']}/api/v1/auth/register"
    data = {
        "email": email,
        "password": password,
        "name": f"{first_name} {last_name}"
    }
    response = requests.post(url, json=data)
    print(f"Registration response status: {response.status_code}")
    print(f"Registration response: {response.text}")
    return response.json() if response.status_code == 201 else None

def login_user(email, password):
    url = f"{BASE_URLS['user']}/api/v1/auth/login"
    data = {"email": email, "password": password}
    response = requests.post(url, json=data)
    return response.json() if response.status_code == 200 else None

def create_admin(email, password, name):
    url = f"{BASE_URLS['event']}/api/v1/auth/admin/register"
    data = {
        "email": email,
        "password": password,
        "name": name
    }
    response = requests.post(url, json=data)
    print(f"Admin registration response status: {response.status_code}")
    print(f"Admin registration response: {response.text}")
    return response.json() if response.status_code == 201 else None

def login_admin(email, password):
    url = f"{BASE_URLS['event']}/api/v1/auth/admin/login"
    data = {"email": email, "password": password}
    response = requests.post(url, json=data)
    print(f"Admin login response status: {response.status_code}")
    print(f"Admin login response: {response.text}")
    return response.json() if response.status_code == 200 else None

def create_admin_event(token, name, total_seats=100, base_price=29.99):
    url = f"{BASE_URLS['event']}/api/v1/admin/events"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    venue_id = str(uuid.uuid4())
    start_date = (datetime.utcnow() + timedelta(days=30)).strftime("%Y-%m-%dT%H:%M:%SZ")
    end_date = (datetime.utcnow() + timedelta(days=30, hours=3)).strftime("%Y-%m-%dT%H:%M:%SZ")
    
    data = {
        "name": name,
        "description": f"Test event: {name}",
        "event_type": "conference",
        "venue_id": venue_id,
        "start_date": start_date,
        "end_date": end_date,
        "total_seats": total_seats,
        "available_seats": total_seats,
        "base_price": base_price,
        "max_tickets_per_booking": 8,
        "status": "published"
    }
    response = requests.post(url, json=data, headers=headers)
    print(f"Event creation response status: {response.status_code}")
    print(f"Event creation response: {response.text}")
    return response.json() if response.status_code == 201 else None

def get_events(token):
    url = f"{BASE_URLS['event']}/api/v1/events"
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.get(url, headers=headers)
    return response.json()

def check_availability(event_id, quantity=2):
    url = f"{BASE_URLS['booking']}/api/v1/bookings/check-availability"
    params = {"event_id": event_id, "quantity": quantity}
    response = requests.get(url, params=params)
    return response.json()

def reserve_seats(token, event_id, quantity=2):
    url = f"{BASE_URLS['booking']}/api/v1/bookings/reserve"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    data = {
        "event_id": event_id,
        "quantity": quantity,
        "idempotency_key": str(uuid.uuid4())
    }
    response = requests.post(url, json=data, headers=headers)
    return response.json(), response.status_code

def confirm_booking(token, reservation_id):
    url = f"{BASE_URLS['booking']}/api/v1/bookings/confirm"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    data = {
        "reservation_id": reservation_id,
        "payment_token": f"mock_token_{uuid.uuid4()}",
        "payment_method": "credit_card"
    }
    response = requests.post(url, json=data, headers=headers)
    return response.json(), response.status_code

def main():
    print("=== Booking Service Flow Test ===")
    
    print("\n1. Creating regular user...")
    claude_user = create_user("claude@example.com", "password123", "Claude", "AI")
    print(f"Claude user creation: {claude_user}")
    
    print("\n2. Creating admin user...")
    admin_user = create_admin("claudeadmin@example.com", "admin123", "Claude Admin")
    print(f"Admin user creation: {admin_user}")
    
    print("\n3. Logging in users...")
    claude_login = login_user("claude@example.com", "password123")
    admin_login = login_admin("claudeadmin@example.com", "admin123")
    print(f"Claude login: {claude_login}")
    print(f"Admin login: {admin_login}")
    
    if not claude_login or 'access_token' not in (claude_login or {}):
        print("Failed to login Claude user")
        return
    
    if not admin_login or 'access_token' not in (admin_login or {}):
        print("Failed to login admin user")
        return
    
    claude_token = claude_login.get('access_token')
    admin_token = admin_login.get('access_token')
    
    print(f"Claude Token: {claude_token[:30]}...")
    print(f"Admin Token: {admin_token[:30]}...")
    
    print("\n4. Creating test events...")
    events = [
        create_admin_event(admin_token, "Tech Conference 2024", 200, 99.99),
        create_admin_event(admin_token, "Music Concert", 150, 49.99),
        create_admin_event(admin_token, "Comedy Show", 80, 25.99)
    ]
    
    for i, event in enumerate(events, 1):
        print(f"Event {i}: {event}")
    
    print("\n5. Getting available events...")
    available_events = get_events(claude_token)
    print(f"Available events: {available_events}")
    
    if not available_events.get('events'):
        print("No events found!")
        return
    
    event_id = available_events['events'][0]['event_id']
    print(f"\n6. Testing booking flow for event: {event_id}")
    
    print("\n7. Checking availability...")
    availability = check_availability(event_id, 2)
    print(f"Availability: {availability}")
    
    print("\n8. Reserving seats (Phase 1)...")
    reservation, status = reserve_seats(claude_token, event_id, 2)
    print(f"Reservation (Status {status}): {reservation}")
    
    if status != 200 or 'reservation_id' not in reservation:
        print("Failed to reserve seats!")
        return
    
    reservation_id = reservation['reservation_id']
    
    print("\n9. Confirming booking (Phase 2)...")
    confirmation, status = confirm_booking(claude_token, reservation_id)
    print(f"Confirmation (Status {status}): {confirmation}")
    
    if status == 200:
        print("\n✅ SUCCESS: Complete booking flow worked!")
        print(f"Booking ID: {confirmation.get('booking_id')}")
        print(f"Ticket URL: {confirmation.get('ticket_url')}")
    else:
        print(f"\n❌ FAILED: Booking confirmation failed with status {status}")

if __name__ == "__main__":
    main()