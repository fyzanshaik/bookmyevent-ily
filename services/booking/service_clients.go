package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type UserServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
}

func NewUserServiceClient(baseURL, apiKey string) *UserServiceClient {
	return &UserServiceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		APIKey: apiKey,
	}
}

func (c *UserServiceClient) VerifyToken(ctx context.Context, token string) (*UserServiceUser, error) {
	requestBody := map[string]string{
		"token": token,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/internal/auth/verify", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var user UserServiceUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

type EventServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
}

func NewEventServiceClient(baseURL, apiKey string) *EventServiceClient {
	return &EventServiceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		APIKey: apiKey,
	}
}

func (c *EventServiceClient) GetEventForBooking(ctx context.Context, eventID uuid.UUID) (*EventServiceEvent, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/internal/events/%s", c.BaseURL, eventID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("event not found or not available for booking")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("event service returned status %d", resp.StatusCode)
	}

	var event EventServiceEvent
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &event, nil
}



func (c *EventServiceClient) UpdateAvailability(ctx context.Context, eventID uuid.UUID, quantity, version int32) (*UpdateAvailabilityResponse, error) {
	requestBody := UpdateAvailabilityRequest{
		Quantity: quantity,
		Version:  version,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", 
		fmt.Sprintf("%s/internal/events/%s/update-availability", c.BaseURL, eventID), 
		bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		var errorResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return nil, fmt.Errorf("availability update failed: %s", errorResp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("event service returned status %d", resp.StatusCode)
	}

	var updateResp UpdateAvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updateResp, nil
}

func (c *EventServiceClient) ReturnSeats(ctx context.Context, eventID uuid.UUID, quantity, version int32) (*UpdateAvailabilityResponse, error) {
	requestBody := UpdateAvailabilityRequest{
		Quantity: quantity,
		Version:  version,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", 
		fmt.Sprintf("%s/internal/events/%s/return-seats", c.BaseURL, eventID), 
		bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("event service returned status %d", resp.StatusCode)
	}

	var updateResp UpdateAvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updateResp, nil
}

func (cfg *APIConfig) InitServiceClients() {
	cfg.UserServiceClient = NewUserServiceClient(cfg.Config.UserServiceURL, cfg.Config.InternalAPIKey)
	cfg.EventServiceClient = NewEventServiceClient(cfg.Config.EventServiceURL, cfg.Config.InternalAPIKey)
}