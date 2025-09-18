#!/bin/bash

# High Concurrency Booking Test Script
# Tests 10 users booking 5 tickets simultaneously on an event with only 5 seats

EVENT_ID="ea16b403-8b5d-44bf-962e-b1155c97d147"
BOOKING_SERVICE_URL="http://localhost:8004"

echo "🚀 Starting High Concurrency Booking Test"
echo "Event ID: $EVENT_ID"
echo "Test: 10 users booking 5 tickets each on event with 5 seats"
echo "Expected: Only 1 user should get reservation"
echo ""

# Using our existing 2 users + creating 8 more with unique tokens
declare -a USER_TOKENS=(
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJldmVudGx5Iiwic3ViIjoiNzJiNTI3ZWItZGNlMC00Mjc3LTg3MmYtMjA3OWE1N2FiNjc3IiwiZXhwIjoxNzU4MDE1NDYxLCJpYXQiOjE3NTgwMTQ1NjF9.Yaddm06iEeubQEmmEfqfwZPDPKZshAsEIql-QEGoTPI"  # User 1
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJldmVudGx5Iiwic3ViIjoiMWVkZDU1ZTYtMzFkNi00MWE3LWEwZTgtNzAyYjEzZWVjNGMyIiwiZXhwIjoxNzU4MDE1NDkxLCJpYXQiOjE3NTgwMTQ1OTF9.1k53SwVr0vodfUxnFYkjNpiihMwFuToxly1TvU3DZ_c"  # User 2
)

# For this test, I'll simulate 10 users by using the same 2 users with different idempotency keys
# In a real system, you'd have 10 different user tokens

echo "📊 Checking initial availability..."
curl -s -X GET "$BOOKING_SERVICE_URL/api/v1/bookings/check-availability?event_id=$EVENT_ID&quantity=1" | jq .
echo ""

echo "🏁 Starting concurrent booking test..."
echo "Time: $(date)"

# Create temporary files to capture responses
mkdir -p /tmp/booking_test_results
rm -f /tmp/booking_test_results/*

# Launch 10 concurrent booking requests
for i in {1..10}; do
    user_token_index=$((($i - 1) % 2))  # Alternate between user 1 and 2 tokens
    user_token="${USER_TOKENS[$user_token_index]}"
    idempotency_key="concurrent-test-user-$i-$(date +%s)"

    # Launch each booking in background
    (
        echo "User $i attempting to book..."
        response=$(curl -s -X POST "$BOOKING_SERVICE_URL/api/v1/bookings/reserve" \
            -H "Authorization: Bearer $user_token" \
            -H "Content-Type: application/json" \
            -d "{\"event_id\": \"$EVENT_ID\", \"quantity\": 5, \"idempotency_key\": \"$idempotency_key\"}")

        echo "$response" > "/tmp/booking_test_results/user_$i.json"
        echo "User $i: $(echo "$response" | jq -r 'if .reservation_id then "SUCCESS - Got reservation: " + .reservation_id else "FAILED - " + .error end')"
    ) &
done

# Wait for all background processes to complete
wait

echo ""
echo "📈 Test Results Analysis:"
echo "========================"

successful_bookings=0
failed_bookings=0
success_file=""

for i in {1..10}; do
    result_file="/tmp/booking_test_results/user_$i.json"
    if [ -f "$result_file" ]; then
        if grep -q "reservation_id" "$result_file"; then
            successful_bookings=$((successful_bookings + 1))
            success_file="$result_file"
            echo "✅ User $i: SUCCESS - $(cat "$result_file" | jq -r '.booking_reference // "N/A"')"
        else
            failed_bookings=$((failed_bookings + 1))
            error_msg=$(cat "$result_file" | jq -r '.error // "Unknown error"')
            echo "❌ User $i: FAILED - $error_msg"
        fi
    else
        echo "⚠️  User $i: No result file found"
        failed_bookings=$((failed_bookings + 1))
    fi
done

echo ""
echo "📊 Final Summary:"
echo "================"
echo "Successful bookings: $successful_bookings"
echo "Failed bookings: $failed_bookings"

if [ $successful_bookings -eq 1 ]; then
    echo "🎉 TEST PASSED! Exactly 1 user got reservation as expected"
    if [ -n "$success_file" ]; then
        echo "Successful reservation details:"
        cat "$success_file" | jq .
    fi
else
    echo "⚠️  TEST RESULT: $successful_bookings users got reservations (expected: 1)"
fi

echo ""
echo "📋 Checking final availability..."
curl -s -X GET "$BOOKING_SERVICE_URL/api/v1/bookings/check-availability?event_id=$EVENT_ID&quantity=1" | jq .

echo ""
echo "🧹 Cleaning up test files..."
rm -rf /tmp/booking_test_results

echo ""
echo "✅ High Concurrency Test Completed!"