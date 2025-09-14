package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/google/uuid"
)

type SearchServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	Logger     *logger.Logger
}

type SearchEventDocument struct {
	EventID        uuid.UUID `json:"event_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	VenueID        uuid.UUID `json:"venue_id"`
	VenueName      string    `json:"venue_name"`
	VenueAddress   string    `json:"venue_address,omitempty"`
	VenueCity      string    `json:"venue_city"`
	VenueState     string    `json:"venue_state,omitempty"`
	VenueCountry   string    `json:"venue_country"`
	EventType      string    `json:"event_type"`
	StartDateTime  time.Time `json:"start_datetime"`
	EndDateTime    time.Time `json:"end_datetime"`
	BasePrice      float64   `json:"base_price"`
	AvailableSeats int32     `json:"available_seats"`
	TotalCapacity  int32     `json:"total_capacity"`
	Status         string    `json:"status"`
	Version        int32     `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SearchIndexRequest struct {
	Event SearchEventDocument `json:"event"`
}

func NewSearchServiceClient(baseURL, apiKey string, logger *logger.Logger) *SearchServiceClient {
	fmt.Printf("DEBUG: NewSearchServiceClient called - baseURL: '%s', apiKey: '%s'\n", baseURL, apiKey)
	client := &SearchServiceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		APIKey: apiKey,
		Logger: logger,
	}
	fmt.Printf("DEBUG: SearchServiceClient created - BaseURL: '%s'\n", client.BaseURL)
	return client
}

func (c *SearchServiceClient) IndexEvent(ctx context.Context, event EventResponse, venue VenueResponse) error {
	fmt.Printf("DEBUG: IndexEvent called - BaseURL: '%s', APIKey: '%s'\n", c.BaseURL, c.APIKey)
	if c.BaseURL == "" {
		c.Logger.Debug("Search service URL not configured, skipping indexing")
		fmt.Printf("DEBUG: BaseURL is empty, skipping indexing\n")
		return nil
	}

	doc := SearchEventDocument{
		EventID:        event.EventID,
		Name:           event.Name,
		VenueID:        event.VenueID,
		EventType:      event.EventType,
		StartDateTime:  event.StartDatetime,
		EndDateTime:    event.EndDatetime,
		BasePrice:      event.BasePrice,
		AvailableSeats: event.AvailableSeats,
		TotalCapacity:  event.TotalCapacity,
		Status:         event.Status,
		Version:        event.Version,
		CreatedAt:      event.CreatedAt,
		UpdatedAt:      event.UpdatedAt,
		VenueName:      venue.Name,
		VenueCity:      venue.City,
		VenueCountry:   venue.Country,
	}

	if event.Description != nil {
		doc.Description = *event.Description
	}

	if venue.Address != "" {
		doc.VenueAddress = venue.Address
	}

	if venue.State != nil {
		doc.VenueState = *venue.State
	}

	request := SearchIndexRequest{
		Event: doc,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		c.Logger.Error("Failed to marshal search document", "error", err, "event_id", event.EventID)
		return fmt.Errorf("failed to marshal search document: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/internal/search/events", bytes.NewBuffer(jsonData))
	if err != nil {
		c.Logger.Error("Failed to create search request", "error", err, "event_id", event.EventID)
		return fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("Failed to index event in search service", "error", err, "event_id", event.EventID)
		fmt.Printf("Failed to index searcSeervice: %s\n", err.Error())
		return fmt.Errorf("failed to index event in search service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("Search service returned error", "status", resp.StatusCode, "body", string(body), "event_id", event.EventID)
		fmt.Printf("Search service returned error %d: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("search service returned status %d: %s", resp.StatusCode, string(body))
	}

	c.Logger.Info("Successfully indexed event in search service", "event_id", event.EventID, "event_name", event.Name)
	return nil
}

func (c *SearchServiceClient) UpdateEvent(ctx context.Context, event EventResponse, venue VenueResponse) error {
	fmt.Printf("Updating event in search service: %s ID: %s\n", event.Name, event.EventID.String())

	return c.IndexEvent(ctx, event, venue)
}

func (c *SearchServiceClient) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	if c.BaseURL == "" {
		c.Logger.Debug("Search service URL not configured, skipping deletion")
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", c.BaseURL+"/internal/search/events/"+eventID.String(), nil)
	if err != nil {
		c.Logger.Error("Failed to create delete request", "error", err, "event_id", eventID)
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("Failed to delete event from search service", "error", err, "event_id", eventID)
		return fmt.Errorf("failed to delete event from search service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("Search service delete returned error", "status", resp.StatusCode, "body", string(body), "event_id", eventID)
		return fmt.Errorf("search service returned status %d: %s", resp.StatusCode, string(body))
	}

	c.Logger.Info("Successfully deleted event from search service", "event_id", eventID)
	return nil
}
