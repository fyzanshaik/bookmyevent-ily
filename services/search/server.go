package search

import (
	"context"
	"log"
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/auth"
	"github.com/fyzanshaik/bookmyevent-ily/internal/config"
	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/fyzanshaik/bookmyevent-ily/internal/middleware"
	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(config *APIConfig) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", utils.HandleHealthz)
	mux.HandleFunc("GET /health/ready", config.HandleReadiness)

	mux.HandleFunc("GET /api/v1/search", config.SearchEvents)
	mux.HandleFunc("GET /api/v1/search/suggestions", config.GetSuggestions)
	mux.HandleFunc("GET /api/v1/search/filters", config.GetFilters)
	mux.HandleFunc("GET /api/v1/search/trending", config.GetTrendingEvents)

	internalAuth := auth.RequireInternalAuth(config.Config.InternalAPIKey)
	mux.HandleFunc("POST /internal/search/events", internalAuth(config.IndexEvent))
	mux.HandleFunc("DELETE /internal/search/events/{id}", internalAuth(config.DeleteEvent))
	mux.HandleFunc("POST /internal/search/resync", internalAuth(config.FullResync))

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

	config.Logger.Info("Starting Search Service", "port", config.Config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func InitSearchService() (*APIConfig, error) {
	cfg := config.LoadSearchServiceConfig()
	logger := logger.New(cfg.LogLevel).WithService("search-service")

	esClient, err := NewElasticsearchClient(cfg.ElasticsearchURL, cfg.IndexName, logger)
	if err != nil {
		return nil, err
	}

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(redisOptions)

	eventServiceClient := NewEventServiceClient(cfg.EventServiceURL, cfg.InternalAPIKey, logger)

	apiConfig := &APIConfig{
		Config:             cfg,
		Logger:             logger,
		ESClient:           esClient,
		RedisClient:        redisClient,
		EventServiceClient: eventServiceClient,
	}

	logger.Info("Initializing Elasticsearch index")
	if err := esClient.CreateIndex(context.Background()); err != nil {
		logger.Error("Failed to create Elasticsearch index", "error", err)
		return nil, err
	}

	return apiConfig, nil
}


