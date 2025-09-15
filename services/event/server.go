package event

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/database"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/fyzanshaik/bookmyevent-ily/internal/middleware"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/events"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
)

func SetupRoutes(config *APIConfig) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", utils.HandleHealthz)
	mux.HandleFunc("GET /health/ready", config.HandleReadiness)

	mux.HandleFunc("POST /api/v1/auth/admin/register", config.AdminRegister)
	mux.HandleFunc("POST /api/v1/auth/admin/login", config.AdminLogin)
	mux.HandleFunc("POST /api/v1/auth/admin/refresh", config.AdminRefreshToken)
	mux.HandleFunc("POST /api/v1/auth/admin/logout", config.AdminLogout)

	//Never to be used by the client, added only for testing purposed and a fallback from using elastisearch
	mux.HandleFunc("GET /api/v1/events", config.ListPublishedEvents)
	mux.HandleFunc("GET /api/v1/events/{id}", config.GetEventByID)
	mux.HandleFunc("GET /api/v1/events/{id}/availability", config.GetEventAvailability)

	adminAuth := auth.RequireAdminAuth(config.Config.JWTSecret)
	mux.HandleFunc("POST /api/v1/admin/events", adminAuth(config.CreateEvent))
	mux.HandleFunc("PUT /api/v1/admin/events/{id}", adminAuth(config.UpdateEvent))
	mux.HandleFunc("DELETE /api/v1/admin/events/{id}", adminAuth(config.DeleteEvent))
	mux.HandleFunc("GET /api/v1/admin/events", adminAuth(config.ListAdminEvents))
	//Analytics KEY DO NOT FORGET TO ADD INC CLIENT
	mux.HandleFunc("GET /api/v1/admin/events/{id}/analytics", adminAuth(config.GetEventAnalytics))

	mux.HandleFunc("POST /api/v1/admin/venues", adminAuth(config.CreateVenue))
	mux.HandleFunc("GET /api/v1/admin/venues", adminAuth(config.ListVenues))
	mux.HandleFunc("PUT /api/v1/admin/venues/{id}", adminAuth(config.UpdateVenue))
	mux.HandleFunc("DELETE /api/v1/admin/venues/{id}", adminAuth(config.DeleteVenue))

	internalAuth := auth.RequireInternalAuth(config.Config.InternalAPIKey)
	mux.HandleFunc("POST /internal/events/{id}/update-availability", internalAuth(config.UpdateEventAvailability))
	mux.HandleFunc("GET /internal/events/{id}", internalAuth(config.GetEventForBooking))
	mux.HandleFunc("POST /internal/events/{id}/return-seats", internalAuth(config.ReturnEventSeats))

	return mux
}

func StartServer(config *APIConfig) {
	mux := SetupRoutes(config)

	handler := middleware.CORS(mux)
	handler = middleware.LoggingMiddleware(config.Logger)(handler)

	server := &http.Server{
		Handler: handler,
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

	var searchClient *SearchServiceClient
	fmt.Printf("DEBUG: SearchServiceURL from config: '%s'\n", cfg.SearchServiceURL)
	if cfg.SearchServiceURL != "" {
		searchClient = NewSearchServiceClient(cfg.SearchServiceURL, cfg.InternalAPIKey, logger)
		logger.Info("Search service client initialized", "search_service_url", cfg.SearchServiceURL)
		fmt.Printf("DEBUG: SearchServiceClient created successfully\n")
	} else {
		logger.Info("Search service URL not configured, search indexing disabled")
		fmt.Printf("DEBUG: SearchServiceURL is empty, no SearchClient created\n")
	}

	apiConfig := &APIConfig{
		DB:           dbQueries,
		DB_Conn:      db,
		Config:       cfg,
		Logger:       logger,
		SearchClient: searchClient,
	}

	return apiConfig, db
}
