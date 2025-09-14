package booking

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/database"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
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
	mux.HandleFunc("GET /api/v1/bookings/user/{userId}", userAuth(config.GetUserBookings))

	mux.HandleFunc("POST /api/v1/waitlist/join", userAuth(config.JoinWaitlist))
	mux.HandleFunc("GET /api/v1/waitlist/position", userAuth(config.GetWaitlistPosition))
	mux.HandleFunc("DELETE /api/v1/waitlist/leave", userAuth(config.LeaveWaitlist))

	internalAuth := auth.RequireInternalAuth(config.Config.InternalAPIKey)
	mux.HandleFunc("GET /internal/bookings/{id}", internalAuth(config.GetBookingInternal))
	mux.HandleFunc("POST /internal/bookings/expire-reservations", internalAuth(config.ExpireReservations))

	return mux
}

func StartServer(config *APIConfig) {
	mux := SetupRoutes(config)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + config.Config.Port,
	}

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
