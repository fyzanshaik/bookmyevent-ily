package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	USER_SERVICE_URL    = "http://localhost:8001"
	EVENT_SERVICE_URL   = "http://localhost:8002"
	BOOKING_SERVICE_URL = "http://localhost:8004"
	ADMIN_EMAIL         = "fyzanadmin@mail.com"
	ADMIN_PASSWORD      = "11111111"
	VENUE_ID            = "7c07652c-75a7-4c1e-9f0f-4a0bff40ff82"
)

type User struct {
	ID          string
	Email       string
	AccessToken string
}

type AdminAuth struct {
	AdminID     string `json:"admin_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	AccessToken string `json:"access_token"`
}

type Event struct {
	EventID        string  `json:"event_id"`
	Name           string  `json:"name"`
	TotalCapacity  int     `json:"total_capacity"`
	AvailableSeats int     `json:"available_seats"`
	BasePrice      float64 `json:"base_price"`
	Status         string  `json:"status"`
	Version        int     `json:"version"`
}

type BookingRequest struct {
	EventID        string `json:"event_id"`
	Quantity       int    `json:"quantity"`
	IdempotencyKey string `json:"idempotency_key"`
}

type BookingResponse struct {
	ReservationID    string  `json:"reservation_id"`
	BookingReference string  `json:"booking_reference"`
	ExpiresAt        string  `json:"expires_at"`
	TotalAmount      float64 `json:"total_amount"`
	Error            string  `json:"error"`
}

type WaitlistRequest struct {
	EventID  string `json:"event_id"`
	Quantity int    `json:"quantity"`
}

type WaitlistResponse struct {
	WaitlistID    string `json:"waitlist_id"`
	Position      int    `json:"position"`
	EstimatedWait string `json:"estimated_wait"`
	Status        string `json:"status"`
	Error         string `json:"error"`
}

type TestResult struct {
	UserID           int
	Success          bool
	ReservationID    string
	BookingReference string
	Error            string
	ResponseTime     time.Duration
	WaitlistPosition int
	IsWaitlisted     bool
}

func main() {
	fmt.Println("🚀 Starting Massive Stress Test: 300 Concurrent Users")
	fmt.Println(strings.Repeat("=", 60))

	// Get admin token
	adminToken, err := getAdminToken()
	if err != nil {
		log.Fatalf("Failed to get admin token: %v", err)
	}
	fmt.Printf("✅ Admin authenticated\n\n")

	// Create test users
	fmt.Println("👥 Creating 300 test users...")
	users, err := createTestUsers(300)
	if err != nil {
		log.Fatalf("Failed to create test users: %v", err)
	}
	fmt.Printf("✅ Created %d test users\n\n", len(users))

	// Create test events
	event1, err := createTestEvent(adminToken, "Stress Test Event 1 - 10 Seats", 10, 10)
	if err != nil {
		log.Fatalf("Failed to create event 1: %v", err)
	}

	event2, err := createTestEvent(adminToken, "Stress Test Event 2 - 299 Seats", 299, 1)
	if err != nil {
		log.Fatalf("Failed to create event 2: %v", err)
	}

	fmt.Printf("✅ Event 1 Created: %s (10 seats, max 10 per booking)\n", event1.EventID)
	fmt.Printf("✅ Event 2 Created: %s (299 seats, max 1 per booking)\n\n", event2.EventID)

	// Test 1: 300 users try to book 10 seats each on 10-seat event
	fmt.Println("🎯 TEST 1: 300 users booking 10 seats each on 10-seat event")
	fmt.Println("Expected: Only 1 user should succeed")
	test1Results := runStressTest(users, event1.EventID, 10, "test1")
	analyzeResults("TEST 1", test1Results, 10, 1) // 10 total seats, expect 1 winner

	time.Sleep(2 * time.Second) // Brief pause between tests

	// Test 2: 300 users try to book 1 seat each on 299-seat event
	fmt.Println("\n🎯 TEST 2: 300 users booking 1 seat each on 299-seat event")
	fmt.Println("Expected: 299 users succeed, 1 user gets waitlisted")
	test2Results := runStressTest(users, event2.EventID, 1, "test2")
	analyzeResults("TEST 2", test2Results, 299, 299) // 299 total seats, expect 299 winners

	// Generate final report
	generateReport(test1Results, test2Results, event1, event2)

	fmt.Println("\n🎉 Massive stress test completed!")
}

func getAdminToken() (string, error) {
	payload := map[string]string{
		"email":    ADMIN_EMAIL,
		"password": ADMIN_PASSWORD,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(EVENT_SERVICE_URL+"/api/v1/auth/admin/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var adminAuth AdminAuth
	if err := json.NewDecoder(resp.Body).Decode(&adminAuth); err != nil {
		return "", err
	}

	return adminAuth.AccessToken, nil
}

func createTestUsers(count int) ([]User, error) {
	users := make([]User, count)

	// Get fresh tokens for our existing users
	baseUsers, err := getBaseUserTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to get base user tokens: %v", err)
	}

	// Distribute users (alternating between the 2 real users)
	for i := 0; i < count; i++ {
		users[i] = User{
			ID:          fmt.Sprintf("stress-user-%d", i+1),
			Email:       fmt.Sprintf("stresstest%d@example.com", i+1),
			AccessToken: baseUsers[i%2].AccessToken, // Alternate between the 2 real tokens
		}
	}

	return users, nil
}

func getBaseUserTokens() ([]User, error) {
	baseUsers := []User{}

	// User 1
	user1Token, err := getUserToken("fyzanshaik@mail.com", "11111111")
	if err != nil {
		return nil, err
	}
	baseUsers = append(baseUsers, User{
		ID:          "72b527eb-dce0-4277-872f-2079a57ab677",
		Email:       "fyzanshaik@mail.com",
		AccessToken: user1Token,
	})

	// User 2
	user2Token, err := getUserToken("fyzanshaik2@mail.com", "11111111")
	if err != nil {
		return nil, err
	}
	baseUsers = append(baseUsers, User{
		ID:          "1edd55e6-31d6-41a7-a0e8-702b13eec4c2",
		Email:       "fyzanshaik2@mail.com",
		AccessToken: user2Token,
	})

	return baseUsers, nil
}

type UserAuth struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
}

func getUserToken(email, password string) (string, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(USER_SERVICE_URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var userAuth UserAuth
	if err := json.NewDecoder(resp.Body).Decode(&userAuth); err != nil {
		return "", err
	}

	return userAuth.AccessToken, nil
}

func createTestEvent(adminToken, name string, capacity, maxPerBooking int) (*Event, error) {
	payload := map[string]interface{}{
		"name":                    name,
		"description":             fmt.Sprintf("Stress test event with %d seats", capacity),
		"venue_id":                VENUE_ID,
		"event_type":              "workshop",
		"start_datetime":          "2025-12-01T14:00:00Z",
		"end_datetime":            "2025-12-01T18:00:00Z",
		"total_capacity":          capacity,
		"base_price":              50.0,
		"max_tickets_per_booking": maxPerBooking,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", EVENT_SERVICE_URL+"/api/v1/admin/events", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var event Event
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, err
	}

	// Publish the event
	publishPayload := map[string]interface{}{
		"status":  "published",
		"version": event.Version,
	}

	jsonData, _ = json.Marshal(publishPayload)
	req, err = http.NewRequest("PUT", EVENT_SERVICE_URL+"/api/v1/admin/events/"+event.EventID, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the updated event
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, err
	}

	return &event, nil
}

func runStressTest(users []User, eventID string, quantity int, testName string) []TestResult {
	fmt.Printf("⏱️  Starting concurrent booking test...\n")
	startTime := time.Now()

	results := make([]TestResult, len(users))
	var wg sync.WaitGroup

	for i, user := range users {
		wg.Add(1)
		go func(userIndex int, u User) {
			defer wg.Done()

			requestStart := time.Now()
			idempotencyKey := fmt.Sprintf("%s-user-%d-%d", testName, userIndex, time.Now().UnixNano())

			result := TestResult{
				UserID: userIndex + 1,
			}

			// Try to make booking
			booking := BookingRequest{
				EventID:        eventID,
				Quantity:       quantity,
				IdempotencyKey: idempotencyKey,
			}

			jsonData, _ := json.Marshal(booking)
			req, err := http.NewRequest("POST", BOOKING_SERVICE_URL+"/api/v1/bookings/reserve", bytes.NewBuffer(jsonData))
			if err != nil {
				result.Error = err.Error()
				result.ResponseTime = time.Since(requestStart)
				results[userIndex] = result
				return
			}

			req.Header.Set("Authorization", "Bearer "+u.AccessToken)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				result.Error = err.Error()
				result.ResponseTime = time.Since(requestStart)
				results[userIndex] = result
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				result.Error = err.Error()
				result.ResponseTime = time.Since(requestStart)
				results[userIndex] = result
				return
			}

			var bookingResp BookingResponse
			if err := json.Unmarshal(body, &bookingResp); err == nil && bookingResp.ReservationID != "" {
				// Success!
				result.Success = true
				result.ReservationID = bookingResp.ReservationID
				result.BookingReference = bookingResp.BookingReference
			} else {
				// Booking failed, try to join waitlist
				result.Error = bookingResp.Error
				if result.Error == "" {
					result.Error = "Unknown booking error"
				}

				// Try joining waitlist
				waitlistReq := WaitlistRequest{
					EventID:  eventID,
					Quantity: quantity,
				}

				waitlistData, _ := json.Marshal(waitlistReq)
				waitlistHttpReq, err := http.NewRequest("POST", BOOKING_SERVICE_URL+"/api/v1/waitlist/join", bytes.NewBuffer(waitlistData))
				if err == nil {
					waitlistHttpReq.Header.Set("Authorization", "Bearer "+u.AccessToken)
					waitlistHttpReq.Header.Set("Content-Type", "application/json")

					waitlistResp, err := client.Do(waitlistHttpReq)
					if err == nil {
						defer waitlistResp.Body.Close()

						var waitlist WaitlistResponse
						if err := json.NewDecoder(waitlistResp.Body).Decode(&waitlist); err == nil && waitlist.Position > 0 {
							result.IsWaitlisted = true
							result.WaitlistPosition = waitlist.Position
						}
					}
				}
			}

			result.ResponseTime = time.Since(requestStart)
			results[userIndex] = result
		}(i, user)
	}

	wg.Wait()

	totalTime := time.Since(startTime)
	fmt.Printf("✅ Test completed in %v\n", totalTime)

	return results
}

func analyzeResults(testName string, results []TestResult, totalSeats, expectedWinners int) {
	fmt.Printf("\n📊 %s ANALYSIS:\n", testName)
	fmt.Println(strings.Repeat("=", 40))

	successCount := 0
	errorCount := 0
	waitlistCount := 0
	var totalResponseTime time.Duration
	errorTypes := make(map[string]int)

	for _, result := range results {
		totalResponseTime += result.ResponseTime

		if result.Success {
			successCount++
			fmt.Printf("✅ User %d: SUCCESS - %s\n", result.UserID, result.BookingReference)
		} else {
			errorCount++
			if result.IsWaitlisted {
				waitlistCount++
				fmt.Printf("📋 User %d: WAITLISTED (Position %d)\n", result.UserID, result.WaitlistPosition)
			} else {
				errorTypes[result.Error]++
				// Only show first few errors to avoid spam
				if errorCount <= 5 {
					fmt.Printf("❌ User %d: FAILED - %s\n", result.UserID, result.Error)
				}
			}
		}
	}

	if errorCount > 5 {
		fmt.Printf("... and %d more failures\n", errorCount-5)
	}

	avgResponseTime := totalResponseTime / time.Duration(len(results))

	fmt.Printf("\n📈 SUMMARY:\n")
	fmt.Printf("Total Users: %d\n", len(results))
	fmt.Printf("Successful Bookings: %d (Expected: %d)\n", successCount, expectedWinners)
	fmt.Printf("Failed Bookings: %d\n", errorCount-waitlistCount)
	fmt.Printf("Waitlisted Users: %d\n", waitlistCount)
	fmt.Printf("Average Response Time: %v\n", avgResponseTime)

	fmt.Printf("\nError Breakdown:\n")
	for errorType, count := range errorTypes {
		fmt.Printf("  - %s: %d users\n", errorType, count)
	}

	// Test validation
	if testName == "TEST 1" {
		if successCount == 1 {
			fmt.Printf("🎉 TEST 1 PASSED: Exactly 1 user got reservation as expected!\n")
		} else {
			fmt.Printf("⚠️  TEST 1 RESULT: %d users got reservations (expected: 1)\n", successCount)
		}
	} else if testName == "TEST 2" {
		if successCount == expectedWinners && waitlistCount >= 1 {
			fmt.Printf("🎉 TEST 2 PASSED: %d users got seats, %d waitlisted!\n", successCount, waitlistCount)
		} else {
			fmt.Printf("⚠️  TEST 2 RESULT: %d successful, %d waitlisted (expected: %d successful, 1+ waitlisted)\n", successCount, waitlistCount, expectedWinners)
		}
	}
}

func generateReport(test1Results, test2Results []TestResult, event1, event2 *Event) {
	reportContent := fmt.Sprintf(`# Massive Stress Test Results - 300 Concurrent Users

## Test Overview
- **Date:** %s
- **Users:** 300 concurrent users
- **Test Duration:** Complete stress test of booking system under extreme load

## Test 1: 300 Users → 10 Seats Event
- **Event:** %s (%s)
- **Capacity:** %d seats
- **User Request:** 10 seats each (demanding 3000 seats total)
- **Expected Result:** Only 1 user should succeed

## Test 2: 300 Users → 299 Seats Event
- **Event:** %s (%s)
- **Capacity:** %d seats
- **User Request:** 1 seat each (demanding 300 seats total)
- **Expected Result:** 299 users succeed, 1+ users waitlisted

## Results Generated: %s

For detailed analysis, see console output.
`,
		time.Now().Format("2006-01-02 15:04:05"),
		event1.Name, event1.EventID, event1.TotalCapacity,
		event2.Name, event2.EventID, event2.TotalCapacity,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// Write to file
	err := os.WriteFile("massive_stress_test_report.md", []byte(reportContent), 0644)
	if err != nil {
		fmt.Printf("Failed to write report: %v\n", err)
	} else {
		fmt.Printf("📄 Report saved to: massive_stress_test_report.md\n")
	}
}
