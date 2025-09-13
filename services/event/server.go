package event

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/database"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/events"
)

func SetupRoutes(config *APIConfig) *http.ServeMux {
	mux := http.NewServeMux()

	// Health checks
	mux.HandleFunc("GET /healthz", HandleHealthz)
	mux.HandleFunc("GET /health/ready", config.HandleReadiness)

	// Admin Authentication Endpoints
	mux.HandleFunc("POST /api/v1/auth/admin/register", config.AdminRegister)
	mux.HandleFunc("POST /api/v1/auth/admin/login", config.AdminLogin)
	mux.HandleFunc("POST /api/v1/auth/admin/refresh", config.AdminRefreshToken)

	// Public Event Endpoints (shold have high Traffic)
	mux.HandleFunc("GET /api/v1/events", config.ListPublishedEvents)
	mux.HandleFunc("GET /api/v1/events/{id}", config.GetEventByID)
	mux.HandleFunc("GET /api/v1/events/{id}/availability", config.GetEventAvailability)

	// Admin Event Management (Protected)
	adminAuth := auth.RequireAdminAuth(config.Config.JWTSecret)
	mux.HandleFunc("POST /api/v1/admin/events", adminAuth(config.CreateEvent))
	mux.HandleFunc("PUT /api/v1/admin/events/{id}", adminAuth(config.UpdateEvent))
	mux.HandleFunc("DELETE /api/v1/admin/events/{id}", adminAuth(config.DeleteEvent))
	mux.HandleFunc("GET /api/v1/admin/events", adminAuth(config.ListAdminEvents))
	mux.HandleFunc("GET /api/v1/admin/events/{id}/analytics", adminAuth(config.GetEventAnalytics))

	// Admin Venue Management (Protected)
	mux.HandleFunc("POST /api/v1/admin/venues", adminAuth(config.CreateVenue))
	mux.HandleFunc("GET /api/v1/admin/venues", adminAuth(config.ListVenues))
	mux.HandleFunc("PUT /api/v1/admin/venues/{id}", adminAuth(config.UpdateVenue))
	mux.HandleFunc("DELETE /api/v1/admin/venues/{id}", adminAuth(config.DeleteVenue))

	// Internal Service Endpoints (Service-to-Service)
	internalAuth := auth.RequireInternalAuth(config.Config.InternalAPIKey)
	mux.HandleFunc("POST /internal/events/{id}/update-availability", internalAuth(config.UpdateEventAvailability))
	mux.HandleFunc("GET /internal/events/{id}", internalAuth(config.GetEventForBooking))
	mux.HandleFunc("POST /internal/events/{id}/return-seats", internalAuth(config.ReturnEventSeats))

	return mux
}

func StartServer(config *APIConfig) {
	mux := SetupRoutes(config)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + config.Config.Port,
	}

	config.Logger.Info("Starting Event Service", "port", config.Config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitEventService() (*APIConfig, *sql.DB) {
	cfg := config.LoadEventServiceConfig()
	logger := logger.New(cfg.LogLevel).WithService("event-service")

	db, err := database.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	dbQueries := events.New(db)

	apiConfig := &APIConfig{
		DB:      dbQueries,
		DB_Conn: db,
		Config:  cfg,
		Logger:  logger,
	}

	return apiConfig, db
}

func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}
