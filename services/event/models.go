package event

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/events"
	"github.com/google/uuid"
)

type APIConfig struct {
	DB           events.Querier
	DB_Conn      *sql.DB
	Config       *config.EventServiceConfig
	Logger       *logger.Logger
	SearchClient *SearchServiceClient
}

type AdminRegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Role        string `json:"role,omitempty"`
}

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AdminAuthResponse struct {
	AdminID      uuid.UUID `json:"admin_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Permissions  string    `json:"permissions,omitempty"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}

type AdminRefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CreateEventRequest struct {
	Name                 string    `json:"name"`
	Description          string    `json:"description,omitempty"`
	VenueID              uuid.UUID `json:"venue_id"`
	EventType            string    `json:"event_type"`
	StartDatetime        time.Time `json:"start_datetime"`
	EndDatetime          time.Time `json:"end_datetime"`
	TotalCapacity        int32     `json:"total_capacity"`
	BasePrice            float64   `json:"base_price"`
	MaxTicketsPerBooking int32     `json:"max_tickets_per_booking,omitempty"`
}

type UpdateEventRequest struct {
	Name                 *string    `json:"name,omitempty"`
	Description          *string    `json:"description,omitempty"`
	VenueID              *uuid.UUID `json:"venue_id,omitempty"`
	EventType            *string    `json:"event_type,omitempty"`
	StartDatetime        *time.Time `json:"start_datetime,omitempty"`
	EndDatetime          *time.Time `json:"end_datetime,omitempty"`
	TotalCapacity        *int32     `json:"total_capacity,omitempty"`
	AvailableSeats       *int32     `json:"available_seats,omitempty"`
	BasePrice            *float64   `json:"base_price,omitempty"`
	MaxTicketsPerBooking *int32     `json:"max_tickets_per_booking,omitempty"`
	Status               *string    `json:"status,omitempty"`
	Version              int32      `json:"version"`
}

type EventResponse struct {
	EventID              uuid.UUID `json:"event_id"`
	Name                 string    `json:"name"`
	Description          *string   `json:"description,omitempty"`
	VenueID              uuid.UUID `json:"venue_id"`
	VenueName            *string   `json:"venue_name,omitempty"`
	VenueAddress         *string   `json:"venue_address,omitempty"`
	VenueCity            *string   `json:"venue_city,omitempty"`
	VenueState           *string   `json:"venue_state,omitempty"`
	VenueCountry         *string   `json:"venue_country,omitempty"`
	EventType            string    `json:"event_type"`
	StartDatetime        time.Time `json:"start_datetime"`
	EndDatetime          time.Time `json:"end_datetime"`
	TotalCapacity        int32     `json:"total_capacity"`
	AvailableSeats       int32     `json:"available_seats"`
	BasePrice            float64   `json:"base_price"`
	MaxTicketsPerBooking int32     `json:"max_tickets_per_booking"`
	Status               string    `json:"status"`
	Version              int32     `json:"version"`
	CreatedBy            uuid.UUID `json:"created_by"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type EventListResponse struct {
	Events  []EventResponse `json:"events"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	HasMore bool            `json:"has_more"`
}

type EventAvailabilityResponse struct {
	AvailableSeats int32     `json:"available_seats"`
	Status         string    `json:"status"`
	LastUpdated    time.Time `json:"last_updated"`
}

type EventAnalyticsResponse struct {
	EventID             uuid.UUID `json:"event_id"`
	Name                string    `json:"name"`
	TotalCapacity       int32     `json:"total_capacity"`
	AvailableSeats      int32     `json:"available_seats"`
	TicketsSold         int32     `json:"tickets_sold"`
	CapacityUtilization float64   `json:"capacity_utilization"`
	BasePrice           float64   `json:"base_price"`
	EstimatedRevenue    float64   `json:"estimated_revenue"`
}

type CreateVenueRequest struct {
	Name         string          `json:"name"`
	Address      string          `json:"address"`
	City         string          `json:"city"`
	State        string          `json:"state,omitempty"`
	Country      string          `json:"country"`
	PostalCode   string          `json:"postal_code,omitempty"`
	Capacity     int32           `json:"capacity"`
	LayoutConfig json.RawMessage `json:"layout_config,omitempty"`
}

type UpdateVenueRequest struct {
	Name         *string          `json:"name,omitempty"`
	Address      *string          `json:"address,omitempty"`
	City         *string          `json:"city,omitempty"`
	State        *string          `json:"state,omitempty"`
	Country      *string          `json:"country,omitempty"`
	PostalCode   *string          `json:"postal_code,omitempty"`
	Capacity     *int32           `json:"capacity,omitempty"`
	LayoutConfig *json.RawMessage `json:"layout_config,omitempty"`
}

type VenueResponse struct {
	VenueID      uuid.UUID       `json:"venue_id"`
	Name         string          `json:"name"`
	Address      string          `json:"address"`
	City         string          `json:"city"`
	State        *string         `json:"state,omitempty"`
	Country      string          `json:"country"`
	PostalCode   *string         `json:"postal_code,omitempty"`
	Capacity     int32           `json:"capacity"`
	LayoutConfig json.RawMessage `json:"layout_config,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type VenueListResponse struct {
	Venues  []VenueResponse `json:"venues"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	HasMore bool            `json:"has_more"`
}

type UpdateAvailabilityRequest struct {
	Quantity int32 `json:"quantity"`
	Version  int32 `json:"version"`
}

type UpdateAvailabilityResponse struct {
	EventID        uuid.UUID `json:"event_id"`
	AvailableSeats int32     `json:"available_seats"`
	Status         string    `json:"status"`
	Version        int32     `json:"version"`
}

type ReturnSeatsRequest struct {
	Quantity int32 `json:"quantity"`
	Version  int32 `json:"version"`
}

type ReturnSeatsResponse struct {
	EventID        uuid.UUID `json:"event_id"`
	AvailableSeats int32     `json:"available_seats"`
	Status         string    `json:"status"`
	Version        int32     `json:"version"`
}

type EventForBookingResponse struct {
	EventID              uuid.UUID `json:"event_id"`
	AvailableSeats       int32     `json:"available_seats"`
	MaxTicketsPerBooking int32     `json:"max_tickets_per_booking"`
	BasePrice            float64   `json:"base_price"`
	Version              int32     `json:"version"`
	Status               string    `json:"status"`
	Name                 string    `json:"name"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error struct {
		Code      string    `json:"code"`
		Message   string    `json:"message"`
		Details   any       `json:"details,omitempty"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"error"`
}
