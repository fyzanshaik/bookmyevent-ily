package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type VenueRequest struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Capacity    int32  `json:"capacity"`
}

type VenueResponse struct {
	VenueID string `json:"venue_id"`
	Name    string `json:"name"`
}

type EventRequest struct {
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	VenueID              string    `json:"venue_id"`
	EventType            string    `json:"event_type"`
	StartDatetime        time.Time `json:"start_datetime"`
	EndDatetime          time.Time `json:"end_datetime"`
	TotalCapacity        int32     `json:"total_capacity"`
	BasePrice            float64   `json:"base_price"`
	MaxTicketsPerBooking int32     `json:"max_tickets_per_booking"`
}

type EventResponse struct {
	EventID string `json:"event_id"`
	Name    string `json:"name"`
}

func main() {
	log.Println("🚀 Starting BookMyEvent initialization...")

	userServiceURL := getEnvOrDefault("USER_SERVICE_URL", "http://user-service:8001")
	eventServiceURL := getEnvOrDefault("EVENT_SERVICE_URL", "http://event-service:8002")

	log.Printf("User Service URL: %s", userServiceURL)
	log.Printf("Event Service URL: %s", eventServiceURL)

	waitForServices(userServiceURL, eventServiceURL)

	// Create users
	log.Println("👥 Creating users...")
	createUser(userServiceURL, "atlanuser1@mail.com", "11111111", "Atlan User 1")
	createUser(userServiceURL, "atlanuser2@mail.com", "11111111", "Atlan User 2")

	// Create admin
	log.Println("👑 Creating admin...")
	createAdmin(eventServiceURL, "atlanadmin@mail.com", "11111111", "Atlan Admin")

	// Get admin token
	log.Println("🔑 Getting admin token...")
	adminToken := loginAdmin(eventServiceURL, "atlanadmin@mail.com", "11111111")
	if adminToken == "" {
		log.Println("❌ Failed to get admin token")
		return
	}

	// Create test venue
	log.Println("🏢 Creating test venue...")
	venueID := createTestVenue(eventServiceURL, adminToken)
	if venueID == "" {
		log.Println("❌ Failed to create venue")
		return
	}

	// Create 10 events
	log.Println("🎭 Creating 10 events...")
	eventIDs := createTenEvents(eventServiceURL, adminToken, venueID)

	// Publish all events
	log.Println("📢 Publishing events...")
	publishEvents(eventServiceURL, adminToken, eventIDs)

	log.Println("✅ BookMyEvent initialization completed!")
	log.Printf("📊 Created: 2 users, 1 admin, 1 venue, %d events", len(eventIDs))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func waitForServices(userURL, eventURL string) {
	services := map[string]string{
		"User Service":  userURL + "/healthz",
		"Event Service": eventURL + "/healthz",
	}

	for name, url := range services {
		log.Printf("⏳ Waiting for %s...", name)
		for i := 0; i < 60; i++ {
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				log.Printf("✅ %s ready", name)
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func createUser(baseURL, email, password, name string) {
	url := baseURL + "/api/v1/auth/register"
	reqBody := RegisterRequest{Email: email, Password: password, Name: name}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("❌ Failed to create user %s: %v", email, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		log.Printf("✅ Created user: %s", email)
	} else if resp.StatusCode == 409 || resp.StatusCode == 400 {
		log.Printf("ℹ️  User %s already exists", email)
	} else {
		log.Printf("⚠️  Failed to create user %s (status: %d)", email, resp.StatusCode)
	}
}

func createAdmin(baseURL, email, password, name string) {
	url := baseURL + "/api/v1/auth/admin/register"
	reqBody := RegisterRequest{Email: email, Password: password, Name: name}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("❌ Failed to create admin %s: %v", email, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		log.Printf("✅ Created admin: %s", email)
	} else if resp.StatusCode == 409 || resp.StatusCode == 400 {
		log.Printf("ℹ️  Admin %s already exists", email)
	} else {
		log.Printf("⚠️  Failed to create admin %s (status: %d)", email, resp.StatusCode)
	}
}

func loginAdmin(baseURL, email, password string) string {
	url := baseURL + "/api/v1/auth/admin/login"
	reqBody := LoginRequest{Email: email, Password: password}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("❌ Failed to login admin: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("❌ Failed to login admin (status: %d)", resp.StatusCode)
		return ""
	}

	var loginResp LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)
	log.Printf("✅ Admin logged in")
	return loginResp.AccessToken
}

func createTestVenue(baseURL, token string) string {
	url := baseURL + "/api/v1/admin/venues"
	venue := VenueRequest{
		Name:     "Test Venue",
		Address:  "Test Address",
		City:     "Test City",
		Country:  "India",
		Capacity: 1000,
	}

	jsonBody, _ := json.Marshal(venue)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to create venue: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		var venueResp VenueResponse
		json.NewDecoder(resp.Body).Decode(&venueResp)
		log.Printf("✅ Created venue: %s (ID: %s)", venue.Name, venueResp.VenueID)
		return venueResp.VenueID
	}

	log.Printf("⚠️  Failed to create venue (status: %d)", resp.StatusCode)
	return ""
}

func createTenEvents(baseURL, token, venueID string) []string {
	events := []EventRequest{
		{Name: "Diwali Cultural Festival", Description: "Festival of lights celebration", VenueID: venueID, EventType: "cultural", StartDatetime: time.Now().Add(30 * 24 * time.Hour), EndDatetime: time.Now().Add(30*24*time.Hour + 4*time.Hour), TotalCapacity: 300, BasePrice: 500.0, MaxTicketsPerBooking: 5},
		{Name: "Tech Summit 2025", Description: "Technology conference", VenueID: venueID, EventType: "conference", StartDatetime: time.Now().Add(45 * 24 * time.Hour), EndDatetime: time.Now().Add(45*24*time.Hour + 8*time.Hour), TotalCapacity: 500, BasePrice: 2500.0, MaxTicketsPerBooking: 3},
		{Name: "Food Festival", Description: "Street food celebration", VenueID: venueID, EventType: "festival", StartDatetime: time.Now().Add(20 * 24 * time.Hour), EndDatetime: time.Now().Add(20*24*time.Hour + 6*time.Hour), TotalCapacity: 800, BasePrice: 300.0, MaxTicketsPerBooking: 8},
		{Name: "Music Concert", Description: "Classical music evening", VenueID: venueID, EventType: "concert", StartDatetime: time.Now().Add(25 * 24 * time.Hour), EndDatetime: time.Now().Add(25*24*time.Hour + 3*time.Hour), TotalCapacity: 200, BasePrice: 1200.0, MaxTicketsPerBooking: 4},
		{Name: "Holi Celebration", Description: "Festival of colors", VenueID: venueID, EventType: "cultural", StartDatetime: time.Now().Add(60 * 24 * time.Hour), EndDatetime: time.Now().Add(60*24*time.Hour + 5*time.Hour), TotalCapacity: 600, BasePrice: 400.0, MaxTicketsPerBooking: 6},
		{Name: "Fashion Show", Description: "Indian fashion showcase", VenueID: venueID, EventType: "entertainment", StartDatetime: time.Now().Add(50 * 24 * time.Hour), EndDatetime: time.Now().Add(50*24*time.Hour + 2*time.Hour), TotalCapacity: 400, BasePrice: 1800.0, MaxTicketsPerBooking: 2},
		{Name: "Dance Workshop", Description: "Bollywood dance training", VenueID: venueID, EventType: "workshop", StartDatetime: time.Now().Add(15 * 24 * time.Hour), EndDatetime: time.Now().Add(15*24*time.Hour + 3*time.Hour), TotalCapacity: 150, BasePrice: 600.0, MaxTicketsPerBooking: 2},
		{Name: "Startup Pitch", Description: "Entrepreneur competition", VenueID: venueID, EventType: "business", StartDatetime: time.Now().Add(40 * 24 * time.Hour), EndDatetime: time.Now().Add(40*24*time.Hour + 4*time.Hour), TotalCapacity: 300, BasePrice: 1000.0, MaxTicketsPerBooking: 3},
		{Name: "Art Exhibition", Description: "Contemporary art showcase", VenueID: venueID, EventType: "exhibition", StartDatetime: time.Now().Add(35 * 24 * time.Hour), EndDatetime: time.Now().Add(35*24*time.Hour + 6*time.Hour), TotalCapacity: 250, BasePrice: 800.0, MaxTicketsPerBooking: 4},
		{Name: "Sports Meet", Description: "Athletic competition", VenueID: venueID, EventType: "sports", StartDatetime: time.Now().Add(55 * 24 * time.Hour), EndDatetime: time.Now().Add(55*24*time.Hour + 8*time.Hour), TotalCapacity: 700, BasePrice: 350.0, MaxTicketsPerBooking: 6},
	}

	var eventIDs []string
	for _, event := range events {
		url := baseURL + "/api/v1/admin/events"
		jsonBody, _ := json.Marshal(event)

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ Failed to create event %s: %v", event.Name, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 201 || resp.StatusCode == 200 {
			var eventResp EventResponse
			json.NewDecoder(resp.Body).Decode(&eventResp)
			eventIDs = append(eventIDs, eventResp.EventID)
			log.Printf("✅ Created event: %s", event.Name)
		} else {
			log.Printf("⚠️  Failed to create event %s (status: %d)", event.Name, resp.StatusCode)
		}
		time.Sleep(100 * time.Millisecond)
	}

	return eventIDs
}

func publishEvents(baseURL, token string, eventIDs []string) {
	for _, eventID := range eventIDs {
		url := fmt.Sprintf("%s/api/v1/admin/events/%s", baseURL, eventID)

		publishReq := map[string]interface{}{
			"status":  "published",
			"version": 1,
		}

		jsonBody, _ := json.Marshal(publishReq)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ Failed to publish event %s: %v", eventID, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			log.Printf("✅ Published event: %s", eventID)
		} else {
			log.Printf("⚠️  Failed to publish event %s (status: %d)", eventID, resp.StatusCode)
		}
		time.Sleep(100 * time.Millisecond)
	}
}