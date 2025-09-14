package user

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/database"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/fyzanshaik/bookmyevent-ily/internal/middleware"
	"github.com/fyzanshaik/bookmyevent-ily/internal/repository/users"
)

func SetupRoutes(config *APIConfig) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", HandleHealthz)
	mux.HandleFunc("GET /health/ready", config.HandleReadiness)

	mux.HandleFunc("POST /api/v1/auth/register", config.AddUser)
	mux.HandleFunc("POST /api/v1/auth/login", config.LoginUser)
	mux.HandleFunc("POST /api/v1/auth/refresh", config.RefreshToken)
	mux.HandleFunc("POST /api/v1/auth/logout", config.RevokeToken)

	authMiddleware := auth.RequireAuth(config.Config.JWTSecret)
	mux.HandleFunc("GET /api/v1/users/profile", authMiddleware(config.GetProfile))
	mux.HandleFunc("PUT /api/v1/users/profile", authMiddleware(config.UpdateProfile))
	mux.HandleFunc("GET /api/v1/users/bookings", authMiddleware(config.GetUserBookings))

	internalAuth := auth.RequireInternalAuth(config.Config.InternalAPIKey)
	mux.HandleFunc("POST /internal/auth/verify", internalAuth(config.HandleInternalVerify))
	mux.HandleFunc("GET /internal/users/{userId}", internalAuth(config.HandleInternalGetUser))

	return mux
}

func StartServer(config *APIConfig) {
	mux := SetupRoutes(config)

	handler := middleware.CORS(mux)

	server := &http.Server{
		Handler: handler,
		Addr:    ":" + config.Config.Port,
	}

	config.Logger.Info("Starting User Service", "port", config.Config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitUserService() (*APIConfig, *sql.DB) {
	cfg := config.LoadUserServiceConfig()
	logger := logger.New(cfg.LogLevel).WithService("user-service")

	db, err := database.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	dbQueries := users.New()

	apiConfig := &APIConfig{
		DB:      dbQueries,
		DB_Conn: db,
		Config:  cfg,
		Logger:  logger,
	}

	return apiConfig, db
}
