# Massive Stress Test Results - 300 Concurrent Users

## Test Overview
- **Date:** 2025-09-16 15:17:25
- **Users:** 300 concurrent users
- **Test Duration:** Complete stress test of booking system under extreme load

## Test 1: 300 Users → 10 Seats Event
- **Event:** Stress Test Event 1 - 10 Seats (397b0fc6-0882-45a6-8165-b26bd2faa794)
- **Capacity:** 10 seats
- **User Request:** 10 seats each (demanding 3000 seats total)
- **Expected Result:** Only 1 user should succeed

## Test 2: 300 Users → 299 Seats Event
- **Event:** Stress Test Event 2 - 299 Seats (4c7440cd-50c3-414c-9237-a3553e689803)
- **Capacity:** 299 seats
- **User Request:** 1 seat each (demanding 300 seats total)
- **Expected Result:** 299 users succeed, 1+ users waitlisted

## Results Generated: 2025-09-16 15:17:25

For detailed analysis, see console output.
