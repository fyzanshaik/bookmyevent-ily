package config

import (
	"os"
	"strconv"
	"time"
)

type UserServiceConfig struct {
	Port               string
	DatabaseURL        string
	DatabaseReplicaURL string
	JWTSecret          string
	JWTAccessDuration  time.Duration
	JWTRefreshDuration time.Duration
	InternalAPIKey     string
	LogLevel           string
	Environment        string
}

func LoadUserServiceConfig() *UserServiceConfig {
	return &UserServiceConfig{
		Port:               getEnv("USER_SERVICE_PORT", "8001"),
		DatabaseURL:        getEnvRequired("USER_SERVICE_DB_URL"),
		DatabaseReplicaURL: getEnv("USER_SERVICE_DB_REPLICA_URL", ""),
		JWTSecret:          getEnvRequired("JWT_SECRET"),
		JWTAccessDuration:  getDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute),
		JWTRefreshDuration: getDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
		InternalAPIKey:     getEnvRequired("INTERNAL_API_KEY"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		Environment:        getEnv("ENVIRONMENT", "development"),
	}
}

type EventServiceConfig struct {
	Port               string
	DatabaseURL        string
	DatabaseReplicaURL string
	JWTSecret          string
	JWTAccessDuration  time.Duration
	JWTRefreshDuration time.Duration
	InternalAPIKey     string
	UserServiceURL     string
	SearchServiceURL   string
	LogLevel           string
	Environment        string
}

func LoadEventServiceConfig() *EventServiceConfig {
	return &EventServiceConfig{
		Port:               getEnv("EVENT_SERVICE_PORT", "8002"),
		DatabaseURL:        getEnvRequired("EVENT_SERVICE_DB_URL"),
		DatabaseReplicaURL: getEnv("EVENT_SERVICE_DB_REPLICA_URL", ""),
		JWTSecret:          getEnvRequired("JWT_SECRET"),
		JWTAccessDuration:  getDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute),
		JWTRefreshDuration: getDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
		InternalAPIKey:     getEnvRequired("INTERNAL_API_KEY"),
		UserServiceURL:     getEnvRequired("USER_SERVICE_URL"),
		SearchServiceURL:   getEnv("SEARCH_SERVICE_URL", ""),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		Environment:        getEnv("ENVIRONMENT", "development"),
	}
}

type BookingServiceConfig struct {
	Port               string
	DatabaseURL        string
	DatabaseReplicaURL string
	RedisURL           string
	// RedisReplicaURL          string
	JWTSecret                 string
	JWTAccessDuration         time.Duration
	JWTRefreshDuration        time.Duration
	InternalAPIKey            string
	UserServiceURL            string
	EventServiceURL           string
	ReservationExpiry         time.Duration
	WaitlistOfferDuration     time.Duration
	MaxTicketsPerUser         int
	RateLimitPerMinute        int
	CleanupInterval           time.Duration
	WaitlistProcessInterval   time.Duration
	MockPaymentSuccessRate    float64
	MockPaymentProcessingTime time.Duration
	LogLevel                  string
	Environment               string
}

func LoadBookingServiceConfig() *BookingServiceConfig {
	return &BookingServiceConfig{
		Port:               getEnv("BOOKING_SERVICE_PORT", "8004"),
		DatabaseURL:        getEnvRequired("BOOKING_SERVICE_DB_URL"),
		DatabaseReplicaURL: getEnv("BOOKING_SERVICE_DB_REPLICA_URL", ""),
		RedisURL:           getEnvRequired("REDIS_URL"),
		// RedisReplicaURL:          getEnv("REDIS_REPLICA_URL", ""),  // Future: Redis read replica
		JWTSecret:                 getEnvRequired("JWT_SECRET"),
		JWTAccessDuration:         getDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute),
		JWTRefreshDuration:        getDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
		InternalAPIKey:            getEnvRequired("INTERNAL_API_KEY"),
		UserServiceURL:            getEnvRequired("USER_SERVICE_URL"),
		EventServiceURL:           getEnvRequired("EVENT_SERVICE_URL"),
		ReservationExpiry:         getDuration("BOOKING_RESERVATION_EXPIRY", 5*time.Minute),
		WaitlistOfferDuration:     getDuration("BOOKING_WAITLIST_OFFER_DURATION", 2*time.Minute),
		MaxTicketsPerUser:         getInt("BOOKING_MAX_TICKETS_PER_USER", 10),
		RateLimitPerMinute:        getInt("BOOKING_RATE_LIMIT_PER_MINUTE", 10),
		CleanupInterval:           getDuration("BOOKING_CLEANUP_INTERVAL", 60*time.Second),
		WaitlistProcessInterval:   getDuration("BOOKING_WAITLIST_PROCESSING_INTERVAL", 30*time.Second),
		MockPaymentSuccessRate:    getFloat("MOCK_PAYMENT_SUCCESS_RATE", 0.95),
		MockPaymentProcessingTime: getDuration("MOCK_PAYMENT_PROCESSING_TIME", 2*time.Second),
		LogLevel:                  getEnv("LOG_LEVEL", "info"),
		Environment:               getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Required environment variable not set: " + key)
	}
	return value
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func getBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

func getFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

type SearchServiceConfig struct {
	Port             string
	ElasticsearchURL string
	RedisURL         string
	// RedisReplicaURL     string
	EventServiceURL  string
	InternalAPIKey   string
	IndexName        string
	CacheExpiry      time.Duration
	MaxSearchResults int
	SearchTimeout    time.Duration
	LogLevel         string
	Environment      string
}

func LoadSearchServiceConfig() *SearchServiceConfig {
	return &SearchServiceConfig{
		Port:             getEnv("SEARCH_SERVICE_PORT", "8003"),
		ElasticsearchURL: getEnvRequired("ELASTICSEARCH_URL"),
		RedisURL:         getEnvRequired("REDIS_URL"),
		// RedisReplicaURL:  getEnv("REDIS_REPLICA_URL", ""),
		EventServiceURL:  getEnvRequired("EVENT_SERVICE_URL"),
		InternalAPIKey:   getEnvRequired("INTERNAL_API_KEY"),
		IndexName:        getEnv("ELASTICSEARCH_INDEX_NAME", "events"),
		CacheExpiry:      getDuration("SEARCH_CACHE_EXPIRY", 5*time.Minute),
		MaxSearchResults: getInt("SEARCH_MAX_RESULTS", 1000),
		SearchTimeout:    getDuration("SEARCH_TIMEOUT", 10*time.Second),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Environment:      getEnv("ENVIRONMENT", "development"),
	}
}
