package booking

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/database"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/fyzanshaik/bookmyevent-ily/internal/middleware"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/bookings"
)

func SetupRoutes(config *APIConfig) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", HandleHealthz)
	mux.HandleFunc("GET /health/ready", config.HandleReadiness)

	userAuth := auth.RequireAuth(config.Config.JWTSecret)
	mux.HandleFunc("GET /api/v1/bookings/check-availability", config.CheckAvailability)
	mux.HandleFunc("POST /api/v1/bookings/reserve", userAuth(config.ReserveSeats))

	mux.HandleFunc("POST /api/v1/bookings/confirm", userAuth(config.ConfirmBooking))

	mux.HandleFunc("GET /api/v1/bookings/{id}", userAuth(config.GetBookingDetails))
	mux.HandleFunc("DELETE /api/v1/bookings/{id}", userAuth(config.CancelBooking))
	mux.HandleFunc("POST /api/v1/bookings/{id}/expire", userAuth(config.ManualExpireReservation))
	mux.HandleFunc("GET /api/v1/bookings/user/{userId}", userAuth(config.GetUserBookings))

	mux.HandleFunc("POST /api/v1/waitlist/join", userAuth(config.JoinWaitlist))
	mux.HandleFunc("GET /api/v1/waitlist/position", userAuth(config.GetWaitlistPosition))
	mux.HandleFunc("DELETE /api/v1/waitlist/leave", userAuth(config.LeaveWaitlist))

	internalAuth := auth.RequireInternalAuth(config.Config.InternalAPIKey)
	mux.HandleFunc("GET /internal/bookings/{id}", internalAuth(config.GetBookingInternal))
	mux.HandleFunc("POST /internal/bookings/expire-reservations", internalAuth(config.ExpireReservations))
	mux.HandleFunc("POST /internal/bookings/force-expire-all", internalAuth(config.ForceExpireAll))

	return mux
}

func StartServer(config *APIConfig) {
	mux := SetupRoutes(config)

	handler := middleware.CORS(mux)

	server := &http.Server{
		Handler: handler,
		Addr:    ":" + config.Config.Port,
	}

	go config.startReservationExpiryWorker()

	config.Logger.Info("Starting Booking Service", "port", config.Config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitBookingService() (*APIConfig, *sql.DB) {
	cfg := config.LoadBookingServiceConfig()
	logger := logger.New(cfg.LogLevel).WithService("booking-service")

	db, err := database.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	dbQueries := bookings.New()

	apiConfig := &APIConfig{
		DB:      dbQueries,
		DB_Conn: db,
		Config:  cfg,
		Logger:  logger,
	}

	apiConfig.InitServiceClients()

	redisClient, err := NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	apiConfig.RedisClient = redisClient

	return apiConfig, db
}

func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

func (cfg *APIConfig) startReservationExpiryWorker() {
	ticker := time.NewTicker(30 * time.Second) 
	defer ticker.Stop()

	cfg.Logger.Info("Started reservation expiry worker", "interval", "30s")

	for range ticker.C {
		cfg.processExpiredReservations()
	}
}

func (cfg *APIConfig) processExpiredReservations() {
	ctx := context.Background()
	
	expiredBookings, err := cfg.DB.GetExpiredBookings(ctx, cfg.DB_Conn, 100)
	if err != nil {
		cfg.Logger.Error("Failed to get expired bookings in background worker", "error", err)
		return
	}

	if len(expiredBookings) == 0 {
		return 
	}

	processed := 0
	for _, booking := range expiredBookings {
		_, err := cfg.DB.UpdateBookingStatus(ctx, cfg.DB_Conn, bookings.UpdateBookingStatusParams{
			BookingID: booking.BookingID,
			Status:    "expired",
		})
		if err != nil {
			cfg.Logger.Error("Failed to expire booking in background worker", "error", err, "booking_id", booking.BookingID)
			continue
		}

		event, err := cfg.EventServiceClient.GetEventForBooking(ctx, booking.EventID)
		if err != nil {
			cfg.Logger.Error("Failed to get event for seat return in background worker", "error", err, "event_id", booking.EventID)
		} else {
			_, err = cfg.EventServiceClient.ReturnSeats(ctx, booking.EventID, booking.Quantity, event.Version)
			if err != nil {
				cfg.Logger.Error("Failed to return seats in background worker", "error", err, "booking_id", booking.BookingID, "event_id", booking.EventID)
			}
		}

		cfg.RedisClient.DeleteReservation(ctx, booking.BookingID)
		cfg.RedisClient.InvalidateEventAvailabilityCache(ctx, booking.EventID)
		processed++

		cfg.Logger.Info("Expired booking processed by background worker",
			"booking_id", booking.BookingID,
			"event_id", booking.EventID,
			"quantity", booking.Quantity)

		cfg.ProcessWaitlist(ctx, booking.EventID, booking.Quantity)
	}

	if processed > 0 {
		cfg.Logger.Info("Background reservation expiry completed", "processed", processed, "total", len(expiredBookings))
	}
}
