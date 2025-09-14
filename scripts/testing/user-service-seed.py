#!/usr/bin/env python3
"""
User Service Seed Script
========================

Creates test users for the event booking platform.
Generates various user types: regular users, premium users, and test scenarios.

Usage:
    python user-service-seed.py

Environment:
    USER_SERVICE_URL: Default http://localhost:8001
"""

import json
import requests
import sys
from typing import Dict, List, Optional, Tuple
import time

# Configuration
USER_SERVICE_URL = "http://localhost:8001"
INTERNAL_API_KEY = "internal-service-communication-key-change-in-production"

class UserServiceClient:
    def __init__(self, base_url: str = USER_SERVICE_URL):
        self.base_url = base_url
        self.session = requests.Session()
        
    def health_check(self) -> bool:
        """Check if User Service is healthy"""
        try:
            response = self.session.get(f"{self.base_url}/healthz")
            return response.status_code == 200
        except:
            return False
    
    def register_user(self, email: str, password: str, name: str, phone_number: str = None) -> Optional[Dict]:
        """Register a new user"""
        payload = {
            "email": email,
            "password": password,
            "name": name
        }
        if phone_number:
            payload["phone_number"] = phone_number
            
        try:
            response = self.session.post(
                f"{self.base_url}/api/v1/auth/register",
                json=payload,
                headers={"Content-Type": "application/json"}
            )
            if response.status_code == 200:
                return response.json()
            else:
                print(f"Registration failed for {email}: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            print(f"Error registering {email}: {e}")
            return None
    
    def login_user(self, email: str, password: str) -> Optional[Dict]:
        """Login user and get tokens"""
        try:
            response = self.session.post(
                f"{self.base_url}/api/v1/auth/login",
                json={"email": email, "password": password},
                headers={"Content-Type": "application/json"}
            )
            return response.json() if response.status_code == 200 else None
        except Exception as e:
            print(f"Error logging in {email}: {e}")
            return None

def create_test_users() -> List[Dict]:
    """Create comprehensive test user dataset"""
    return [
        # Regular Users
        {
            "email": "alice@example.com",
            "password": "password123",
            "name": "Alice Johnson",
            "phone_number": "+1234567890",
            "type": "regular"
        },
        {
            "email": "bob@example.com", 
            "password": "password123",
            "name": "Bob Wilson",
            "phone_number": "+1234567891",
            "type": "regular"
        },
        {
            "email": "charlie@example.com",
            "password": "password123", 
            "name": "Charlie Brown",
            "phone_number": "+1234567892",
            "type": "regular"
        },
        
        # International Users
        {
            "email": "emma@example.co.uk",
            "password": "password123",
            "name": "Emma Thompson",
            "phone_number": "+441234567890",
            "type": "international"
        },
        {
            "email": "hans@example.de",
            "password": "password123",
            "name": "Hans Mueller",
            "phone_number": "+491234567890", 
            "type": "international"
        },
        
        # Edge Cases
        {
            "email": "no-phone@example.com",
            "password": "password123",
            "name": "No Phone User",
            "phone_number": None,
            "type": "no_phone"
        },
        {
            "email": "long-name@example.com",
            "password": "password123",
            "name": "Alexander Maximilian Christopher Wellington-Smith",
            "phone_number": "+1234567893",
            "type": "long_name"
        },
        
        # Business Users
        {
            "email": "admin@company.com",
            "password": "password123",
            "name": "Corporate Admin",
            "phone_number": "+1800123456",
            "type": "corporate"
        },
        {
            "email": "events@organizer.com",
            "password": "password123",
            "name": "Event Organizer",
            "phone_number": "+1800123457",
            "type": "organizer"
        },
        
        # Test Users for Load Testing
        {
            "email": "test1@load.com",
            "password": "password123",
            "name": "Load Test User 1",
            "phone_number": "+1999000001",
            "type": "load_test"
        },
        {
            "email": "test2@load.com", 
            "password": "password123",
            "name": "Load Test User 2",
            "phone_number": "+1999000002",
            "type": "load_test"
        },
        {
            "email": "test3@load.com",
            "password": "password123", 
            "name": "Load Test User 3",
            "phone_number": "+1999000003",
            "type": "load_test"
        }
    ]

def main():
    print("🚀 User Service Seed Script")
    print("=" * 40)
    
    client = UserServiceClient()
    
    # Health check
    print("⚡ Checking User Service health...")
    if not client.health_check():
        print("❌ User Service is not healthy!")
        sys.exit(1)
    print("✅ User Service is healthy")
    
    # Get test users
    test_users = create_test_users()
    print(f"📝 Creating {len(test_users)} test users...")
    
    successful_registrations = []
    failed_registrations = []
    
    # Register users
    for user_data in test_users:
        print(f"  👤 Registering {user_data['name']} ({user_data['email']})...")
        
        result = client.register_user(
            email=user_data["email"],
            password=user_data["password"], 
            name=user_data["name"],
            phone_number=user_data["phone_number"]
        )
        
        if result:
            successful_registrations.append({
                **user_data,
                "user_id": result.get("user_id"),
                "access_token": result.get("access_token"),
                "refresh_token": result.get("refresh_token")
            })
            print(f"     ✅ Success - User ID: {result.get('user_id')}")
        else:
            failed_registrations.append(user_data)
            print(f"     ❌ Failed")
        
        time.sleep(0.1)  # Small delay to avoid overwhelming the service
    
    # Summary
    print("\n📊 Registration Summary")
    print("=" * 40)
    print(f"✅ Successful: {len(successful_registrations)}")
    print(f"❌ Failed: {len(failed_registrations)}")
    
    if failed_registrations:
        print("\n❌ Failed Registrations:")
        for user in failed_registrations:
            print(f"  - {user['email']} ({user['name']})")
    
    # Test login for a few users
    print("\n🔐 Testing Login for Sample Users...")
    login_tests = successful_registrations[:3]  # Test first 3 users
    
    for user in login_tests:
        print(f"  🔑 Testing login for {user['email']}...")
        login_result = client.login_user(user['email'], user['password'])
        
        if login_result:
            print(f"     ✅ Login successful")
        else:
            print(f"     ❌ Login failed")
    
    # Save user data for other services
    output_file = "user-service-seed-results.json"
    with open(output_file, 'w') as f:
        json.dump({
            "successful_registrations": successful_registrations,
            "failed_registrations": failed_registrations,
            "total_users": len(successful_registrations),
            "generated_at": time.strftime("%Y-%m-%d %H:%M:%S")
        }, f, indent=2, default=str)
    
    print(f"\n💾 User data saved to {output_file}")
    print("\n🎉 User Service seeding completed!")
    print("\n📚 Available Test Users by Type:")
    
    user_types = {}
    for user in successful_registrations:
        user_type = user.get('type', 'unknown')
        if user_type not in user_types:
            user_types[user_type] = []
        user_types[user_type].append(user['email'])
    
    for user_type, emails in user_types.items():
        print(f"  {user_type.title()}: {len(emails)} users")
        for email in emails:
            print(f"    - {email}")

if __name__ == "__main__":
    main()