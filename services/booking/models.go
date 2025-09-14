package booking

import (
	"database/sql"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/bookings"
	"github.com/google/uuid"
)

type APIConfig struct {
	DB                 bookings.Querier
	DB_Conn            *sql.DB
	Config             *config.BookingServiceConfig
	Logger             *logger.Logger
	UserServiceClient  *UserServiceClient
	EventServiceClient *EventServiceClient
	RedisClient        *RedisClient
}

type CheckAvailabilityRequest struct {
	EventID  uuid.UUID `json:"event_id"`
	Quantity int32     `json:"quantity"`
}

type CheckAvailabilityResponse struct {
	Available      bool    `json:"available"`
	AvailableSeats int32   `json:"available_seats"`
	MaxPerBooking  int32   `json:"max_per_booking"`
	BasePrice      float64 `json:"base_price"`
}

type ReservationRequest struct {
	EventID        uuid.UUID `json:"event_id"`
	Quantity       int32     `json:"quantity"`
	IdempotencyKey string    `json:"idempotency_key"`
}

type ReservationResponse struct {
	ReservationID    uuid.UUID `json:"reservation_id"`
	BookingReference string    `json:"booking_reference"`
	ExpiresAt        time.Time `json:"expires_at"`
	TotalAmount      float64   `json:"total_amount"`
}

type ConfirmationRequest struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	PaymentToken  string    `json:"payment_token"`
	PaymentMethod string    `json:"payment_method"`
}

type PaymentInfo struct {
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
}

type ConfirmationResponse struct {
	BookingID        uuid.UUID   `json:"booking_id"`
	BookingReference string      `json:"booking_reference"`
	Status           string      `json:"status"`
	TicketURL        string      `json:"ticket_url"`
	Payment          PaymentInfo `json:"payment"`
}

type BookingDetailsResponse struct {
	BookingID        uuid.UUID  `json:"booking_id"`
	BookingReference string     `json:"booking_reference"`
	Event            EventInfo  `json:"event"`
	Quantity         int32      `json:"quantity"`
	TotalAmount      float64    `json:"total_amount"`
	Status           string     `json:"status"`
	PaymentStatus    string     `json:"payment_status"`
	TicketURL        string     `json:"ticket_url,omitempty"`
	BookedAt         time.Time  `json:"booked_at"`
	ConfirmedAt      *time.Time `json:"confirmed_at,omitempty"`
}

type EventInfo struct {
	Name     string    `json:"name"`
	Venue    string    `json:"venue"`
	DateTime time.Time `json:"datetime"`
}

type CancellationResponse struct {
	Message      string  `json:"message"`
	RefundStatus string  `json:"refund_status"`
	RefundAmount float64 `json:"refund_amount"`
}

type JoinWaitlistRequest struct {
	EventID  uuid.UUID `json:"event_id"`
	Quantity int32     `json:"quantity"`
}

type JoinWaitlistResponse struct {
	WaitlistID    uuid.UUID `json:"waitlist_id"`
	Position      int32     `json:"position"`
	EstimatedWait string    `json:"estimated_wait"`
	Status        string    `json:"status"`
}

type WaitlistPositionResponse struct {
	Position      int32  `json:"position"`
	TotalWaiting  int32  `json:"total_waiting"`
	Status        string `json:"status"`
	EstimatedWait string `json:"estimated_wait"`
}

type LeaveWaitlistRequest struct {
	EventID uuid.UUID `json:"event_id"`
}

type EventServiceEvent struct {
	EventID              uuid.UUID `json:"event_id"`
	AvailableSeats       int32     `json:"available_seats"`
	MaxTicketsPerBooking int32     `json:"max_tickets_per_booking"`
	BasePrice            float64   `json:"base_price"`
	Version              int32     `json:"version"`
	Status               string    `json:"status"`
	Name                 string    `json:"name"`
}

type UserServiceUser struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Valid  bool      `json:"valid"`
}

type ReservationData struct {
	UserID           uuid.UUID `json:"user_id"`
	EventID          uuid.UUID `json:"event_id"`
	Quantity         int32     `json:"quantity"`
	Amount           float64   `json:"amount"`
	BookingID        uuid.UUID `json:"booking_id"`
	BookingReference string    `json:"booking_reference"`
	ExpiresAt        time.Time `json:"expires_at"`
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
