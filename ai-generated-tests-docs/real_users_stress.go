package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	USER_SERVICE_URL    = "http://localhost:8001"
	EVENT_SERVICE_URL   = "http://localhost:8002"
	BOOKING_SERVICE_URL = "http://localhost:8004"
	ADMIN_EMAIL         = "fyzanadmin@mail.com"
	ADMIN_PASSWORD      = "11111111"
	VENUE_ID           = "7c07652c-75a7-4c1e-9f0f-4a0bff40ff82"
	NUM_USERS          = 300
	USER_PASSWORD      = "testpass123"
)

type User struct {
	ID          string
	Email       string
	AccessToken string
}

type UserRegistrationRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserAuthResponse struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
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

type TestResult struct {
	UserID           int
	Success          bool
	ReservationID    string
	BookingReference string
	Error            string
	ResponseTime     time.Duration
}

// Global token map for O(1) lookup
var userTokens = make(map[int]string, NUM_USERS)
var tokenMutex = sync.RWMutex{}

func main() {
	fmt.Println("🚀 Creating 300 REAL Users & Testing Concurrent Booking")
	fmt.Println("=" + fmt.Sprintf("%60s", ""))

	// Step 1: Get admin token
	adminToken, err := getAdminToken()
	if err != nil {
		log.Fatalf("Failed to get admin token: %v", err)
	}
	fmt.Println("✅ Admin authenticated")

	// Step 2: Create 300 real users in database
	fmt.Println("\n👥 Creating 300 REAL users in database...")
	startTime := time.Now()

	if err := createRealUsers(); err != nil {
		log.Fatalf("Failed to create users: %v", err)
	}

	createDuration := time.Since(startTime)
	fmt.Printf("✅ Created 300 real users in %v\n", createDuration)
	fmt.Printf("📊 Token map size: %d (O(1) lookup ready)\n\n", len(userTokens))

	// Step 3: Create test events
	event1, err := createTestEvent(adminToken, "Real Test Event 1 - 10 Seats", 10, 10)
	if err != nil {
		log.Fatalf("Failed to create event 1: %v", err)
	}

	event2, err := createTestEvent(adminToken, "Real Test Event 2 - 299 Seats", 299, 1)
	if err != nil {
		log.Fatalf("Failed to create event 2: %v", err)
	}

	fmt.Printf("✅ Event 1: %s (10 seats, max 10 per booking)\n", event1.EventID)
	fmt.Printf("✅ Event 2: %s (299 seats, max 1 per booking)\n\n", event2.EventID)

	// Step 4: Test 1 - 300 users with unique tokens → 10 seats
	fmt.Println("🎯 TEST 1: 300 REAL users (unique tokens) → 10 seats")
	fmt.Println("Expected: Only 1 user should succeed, 299 should fail/waitlist")
	test1Results := runRealUserStressTest(event1.EventID, 10, "real-test1")
	analyzeResults("REAL TEST 1", test1Results, 10, 1)

	time.Sleep(3 * time.Second) // Brief pause

	// Step 5: Test 2 - 300 users with unique tokens → 299 seats
	fmt.Println("\n🎯 TEST 2: 300 REAL users (unique tokens) → 299 seats")
	fmt.Println("Expected: 299 users succeed, 1 user should fail/waitlist")
	test2Results := runRealUserStressTest(event2.EventID, 1, "real-test2")
	analyzeResults("REAL TEST 2", test2Results, 299, 299)

	// Generate report
	generateRealUserReport(test1Results, test2Results, event1, event2)

	fmt.Println("\n🎉 Real 300-user concurrency test completed!")
}

func createRealUsers() error {
	var wg sync.WaitGroup
	userChan := make(chan User, NUM_USERS)
	errorChan := make(chan error, NUM_USERS)

	// Create users concurrently (batched for performance)
	batchSize := 50
	for i := 0; i < NUM_USERS; i += batchSize {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			end := start + batchSize
			if end > NUM_USERS {
				end = NUM_USERS
			}

			for j := start; j < end; j++ {
				user, err := createAndAuthenticateUser(j + 1)
				if err != nil {
					errorChan <- fmt.Errorf("failed to create user %d: %v", j+1, err)
					return
				}
				userChan <- user
			}
		}(i)
	}

	// Wait for all user creation to complete
	go func() {
		wg.Wait()
		close(userChan)
		close(errorChan)
	}()

	// Collect results
	userCount := 0
	for {
		select {
		case user, ok := <-userChan:
			if !ok {
				userChan = nil
			} else {
				// Store token in map for O(1) lookup
				userID := userCount + 1
				tokenMutex.Lock()
				userTokens[userID] = user.AccessToken
				tokenMutex.Unlock()
				userCount++

				if userCount%50 == 0 {
					fmt.Printf("  📝 Created %d/%d users...\n", userCount, NUM_USERS)
				}
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else {
				return err
			}
		}

		if userChan == nil && errorChan == nil {
			break
		}
	}

	if userCount != NUM_USERS {
		return fmt.Errorf("expected %d users, created %d", NUM_USERS, userCount)
	}

	return nil
}

func createAndAuthenticateUser(userID int) (User, error) {
	email := fmt.Sprintf("testuser%d@example.com", userID)
	name := fmt.Sprintf("Test User %d", userID)

	// Step 1: Register user
	regReq := UserRegistrationRequest{
		Name:     name,
		Email:    email,
		Password: USER_PASSWORD,
	}

	regData, _ := json.Marshal(regReq)
	resp, err := http.Post(USER_SERVICE_URL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(regData))
	if err != nil {
		return User{}, fmt.Errorf("registration request failed: %v", err)
	}
	resp.Body.Close()

	// Step 2: Login to get token (even if registration failed, user might exist)
	loginReq := map[string]string{
		"email":    email,
		"password": USER_PASSWORD,
	}

	loginData, _ := json.Marshal(loginReq)
	loginResp, err := http.Post(USER_SERVICE_URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		return User{}, fmt.Errorf("login request failed: %v", err)
	}
	defer loginResp.Body.Close()

	var authResp UserAuthResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&authResp); err != nil {
		return User{}, fmt.Errorf("failed to decode auth response: %v", err)
	}

	if authResp.AccessToken == "" {
		return User{}, fmt.Errorf("no access token received for user %d", userID)
	}

	return User{
		ID:          authResp.UserID,
		Email:       email,
		AccessToken: authResp.AccessToken,
	}, nil
}

func runRealUserStressTest(eventID string, quantity int, testName string) []TestResult {
	fmt.Printf("⏱️  Starting REAL concurrent booking test with %d unique tokens...\n", NUM_USERS)
	startTime := time.Now()

	results := make([]TestResult, NUM_USERS)
	var wg sync.WaitGroup

	for i := 0; i < NUM_USERS; i++ {
		wg.Add(1)
		go func(userIndex int) {
			defer wg.Done()

			requestStart := time.Now()
			userID := userIndex + 1

			// O(1) token lookup from map
			tokenMutex.RLock()
			userToken, exists := userTokens[userID]
			tokenMutex.RUnlock()

			result := TestResult{
				UserID: userID,
			}

			if !exists {
				result.Error = "Token not found in map"
				result.ResponseTime = time.Since(requestStart)
				results[userIndex] = result
				return
			}

			// Try booking with unique user token
			booking := BookingRequest{
				EventID:        eventID,
				Quantity:       quantity,
				IdempotencyKey: fmt.Sprintf("%s-user-%d-%d", testName, userID, time.Now().UnixNano()),
			}

			jsonData, _ := json.Marshal(booking)
			req, err := http.NewRequest("POST", BOOKING_SERVICE_URL+"/api/v1/bookings/reserve", bytes.NewBuffer(jsonData))
			if err != nil {
				result.Error = err.Error()
				result.ResponseTime = time.Since(requestStart)
				results[userIndex] = result
				return
			}

			req.Header.Set("Authorization", "Bearer "+userToken)
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
				result.Success = true
				result.ReservationID = bookingResp.ReservationID
				result.BookingReference = bookingResp.BookingReference
			} else {
				result.Error = bookingResp.Error
				if result.Error == "" {
					result.Error = "Unknown booking error"
				}
			}

			result.ResponseTime = time.Since(requestStart)
			results[userIndex] = result
		}(i)
	}

	wg.Wait()

	totalTime := time.Since(startTime)
	fmt.Printf("✅ Real user test completed in %v\n", totalTime)

	return results
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

func createTestEvent(adminToken, name string, capacity, maxPerBooking int) (*Event, error) {
	payload := map[string]interface{}{
		"name":                     name,
		"description":              fmt.Sprintf("Real user test event with %d seats", capacity),
		"venue_id":                 VENUE_ID,
		"event_type":               "workshop",
		"start_datetime":           "2025-12-15T14:00:00Z",
		"end_datetime":             "2025-12-15T18:00:00Z",
		"total_capacity":           capacity,
		"base_price":               75.0,
		"max_tickets_per_booking":  maxPerBooking,
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

	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, err
	}

	return &event, nil
}

func analyzeResults(testName string, results []TestResult, totalSeats, expectedWinners int) {
	fmt.Printf("\n📊 %s ANALYSIS:\n", testName)
	fmt.Println("========================================")

	successCount := 0
	errorCount := 0
	var totalResponseTime time.Duration
	errorTypes := make(map[string]int)

	for _, result := range results {
		totalResponseTime += result.ResponseTime

		if result.Success {
			successCount++
			if successCount <= 5 { // Show first 5 successes
				fmt.Printf("✅ User %d: SUCCESS - %s\n", result.UserID, result.BookingReference)
			}
		} else {
			errorCount++
			errorTypes[result.Error]++
			if errorCount <= 5 { // Show first 5 errors
				fmt.Printf("❌ User %d: FAILED - %s\n", result.UserID, result.Error)
			}
		}
	}

	if successCount > 5 {
		fmt.Printf("✅ ... and %d more successful bookings\n", successCount-5)
	}
	if errorCount > 5 {
		fmt.Printf("❌ ... and %d more failures\n", errorCount-5)
	}

	avgResponseTime := totalResponseTime / time.Duration(len(results))

	fmt.Printf("\n📈 SUMMARY:\n")
	fmt.Printf("Total Real Users: %d\n", len(results))
	fmt.Printf("Successful Bookings: %d (Expected: %d)\n", successCount, expectedWinners)
	fmt.Printf("Failed Bookings: %d\n", errorCount)
	fmt.Printf("Average Response Time: %v\n", avgResponseTime)

	fmt.Printf("\nError Breakdown:\n")
	for errorType, count := range errorTypes {
		fmt.Printf("  - %s: %d users\n", errorType, count)
	}

	// Validation
	if testName == "REAL TEST 1" {
		if successCount == 1 {
			fmt.Printf("🎉 REAL TEST 1 PASSED: Exactly 1 user with unique token succeeded!\n")
		} else {
			fmt.Printf("⚠️  REAL TEST 1 RESULT: %d users succeeded (expected: 1)\n", successCount)
		}
	} else if testName == "REAL TEST 2" {
		if successCount == expectedWinners {
			fmt.Printf("🎉 REAL TEST 2 PASSED: %d users with unique tokens succeeded!\n", successCount)
		} else {
			fmt.Printf("⚠️  REAL TEST 2 RESULT: %d users succeeded (expected: %d)\n", successCount, expectedWinners)
		}
	}
}

func generateRealUserReport(test1Results, test2Results []TestResult, event1, event2 *Event) {
	reportContent := fmt.Sprintf(`# Real 300-User Stress Test Results

## Test Overview
- **Date:** %s
- **Real Users:** 300 actual users created in database
- **Unique Tokens:** 300 individual JWT tokens (O(1) map lookup)
- **Test Type:** True concurrency test with unique authentication

## Test 1: 300 Real Users → 10 Seats Event
- **Event:** %s (%s)
- **Capacity:** %d seats
- **User Request:** 10 seats each
- **Token Type:** Unique per user (no alternating)

## Test 2: 300 Real Users → 299 Seats Event
- **Event:** %s (%s)
- **Capacity:** %d seats
- **User Request:** 1 seat each
- **Token Type:** Unique per user (no alternating)

## Key Improvements Over Previous Test
- ✅ Real database users (not simulated)
- ✅ Unique JWT tokens per user
- ✅ O(1) token lookup performance
- ✅ No rate limiting due to token reuse
- ✅ True concurrency testing

## Results Generated: %s
`,
		time.Now().Format("2006-01-02 15:04:05"),
		event1.Name, event1.EventID, event1.TotalCapacity,
		event2.Name, event2.EventID, event2.TotalCapacity,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	err := os.WriteFile("real_300_users_test_report.md", []byte(reportContent), 0644)
	if err != nil {
		fmt.Printf("Failed to write report: %v\n", err)
	} else {
		fmt.Printf("📄 Real user test report saved to: real_300_users_test_report.md\n")
	}
}