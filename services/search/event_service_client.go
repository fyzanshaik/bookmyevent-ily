package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/google/uuid"
)

type EventServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	Logger     *logger.Logger
}

type EventServiceEvent struct {
	EventID       uuid.UUID `json:"event_id"`
	Name          string    `json:"name"`
	Description   *string   `json:"description,omitempty"`
	VenueID       uuid.UUID `json:"venue_id"`
	VenueName     *string   `json:"venue_name,omitempty"`
	VenueAddress  *string   `json:"venue_address,omitempty"`
	VenueCity     *string   `json:"venue_city,omitempty"`
	VenueState    *string   `json:"venue_state,omitempty"`
	VenueCountry  *string   `json:"venue_country,omitempty"`
	EventType     string    `json:"event_type"`
	StartDatetime time.Time `json:"start_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
	TotalCapacity int32     `json:"total_capacity"`
	AvailableSeats int32    `json:"available_seats"`
	BasePrice     float64   `json:"base_price"`
	Status        string    `json:"status"`
	Version       int32     `json:"version"`
	CreatedBy     uuid.UUID `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type EventServiceListResponse struct {
	Events  []EventServiceEvent `json:"events"`
	Total   int64               `json:"total"`
	Page    int                 `json:"page"`
	Limit   int                 `json:"limit"`
	HasMore bool                `json:"has_more"`
}

func NewEventServiceClient(baseURL, apiKey string, logger *logger.Logger) *EventServiceClient {
	return &EventServiceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		APIKey: apiKey,
		Logger: logger,
	}
}

func (c *EventServiceClient) GetAllPublishedEvents(ctx context.Context) ([]EventServiceEvent, error) {
	var allEvents []EventServiceEvent
	page := 1
	limit := 100

	for {
		url := fmt.Sprintf("%s/api/v1/events?page=%d&limit=%d", c.BaseURL, page, limit)
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("event service returned status %d: %s", resp.StatusCode, string(body))
		}

		var response EventServiceListResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allEvents = append(allEvents, response.Events...)

		c.Logger.Info("Fetched events batch", 
			"page", page, 
			"events_in_batch", len(response.Events),
			"total_so_far", len(allEvents),
			"has_more", response.HasMore)

		if !response.HasMore {
			break
		}

		page++
		
		if page > 100 {
			c.Logger.Warn("Reached maximum page limit, stopping fetch")
			break
		}
	}

	return allEvents, nil
}

func (c *EventServiceClient) GetEvent(ctx context.Context, eventID uuid.UUID) (*EventServiceEvent, error) {
	url := fmt.Sprintf("%s/api/v1/events/%s", c.BaseURL, eventID.String())
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("event not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("event service returned status %d: %s", resp.StatusCode, string(body))
	}

	var event EventServiceEvent
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &event, nil
}
