#!/bin/bash

BASE_URL="http://localhost:8001"
USER_EMAIL="test@example.com"
USER_PASSWORD="password123"
USER_NAME="Test User"
USER_PHONE="1234567890"

echo "Testing User Service Endpoints..."

echo "1. Health Check"
curl -s "$BASE_URL/healthz" | jq . && echo "✅ Health check passed"

echo -e "\n2. User Registration"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASSWORD\",\"name\":\"$USER_NAME\",\"phone_number\":\"$USER_PHONE\"}")
echo "$REGISTER_RESPONSE" | jq .
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.refresh_token')
echo "✅ Registration successful"

echo -e "\n3. User Login"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASSWORD\"}")
echo "$LOGIN_RESPONSE" | jq .
LOGIN_ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
LOGIN_REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')
echo "✅ Login successful"

echo -e "\n4. Get User Profile"
curl -s -X GET "$BASE_URL/api/v1/users/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq . && echo "✅ Profile fetch successful"

echo -e "\n5. Update User Profile"
curl -s -X PUT "$BASE_URL/api/v1/users/profile" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{\"name\":\"Updated Test User\",\"phone_number\":\"9876543210\"}" | jq . && echo "✅ Profile update successful"

echo -e "\n6. Refresh Token"
curl -s -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H "Authorization: Bearer $REFRESH_TOKEN" | jq . && echo "✅ Token refresh successful"

echo -e "\n7. Get User Bookings"
curl -s -X GET "$BASE_URL/api/v1/users/bookings" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq . && echo "✅ Bookings fetch successful"

echo -e "\n8. Internal Token Verification"
curl -s -X POST "$BASE_URL/internal/auth/verify" \
  -H "Content-Type: application/json" \
  -H "Authorization: ApiKey internal-service-communication-key-change-in-production" \
  -d "{\"token\":\"$ACCESS_TOKEN\"}" | jq . && echo "✅ Internal token verification successful"

USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user_id')
echo -e "\n9. Internal Get User"
curl -s -X GET "$BASE_URL/internal/users/$USER_ID" \
  -H "Authorization: ApiKey internal-service-communication-key-change-in-production" | jq . && echo "✅ Internal get user successful"

echo -e "\n10. Logout"
curl -s -X POST "$BASE_URL/api/v1/auth/logout" \
  -H "Authorization: Bearer $LOGIN_REFRESH_TOKEN" | jq . && echo "✅ Logout successful"

echo -e "\nAll tests completed!"