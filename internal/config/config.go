package config

import (
	"os"
	"strconv"
	"time"
)

type UserServiceConfig struct {
	Port                string
	DatabaseURL         string
	DatabaseReplicaURL  string
	JWTSecret          string
	JWTAccessDuration  time.Duration
	JWTRefreshDuration time.Duration
	InternalAPIKey     string
	LogLevel           string
	Environment        string
}

func LoadUserServiceConfig() *UserServiceConfig {
	return &UserServiceConfig{
		Port:                getEnv("USER_SERVICE_PORT", "8001"),
		DatabaseURL:         getEnvRequired("USER_SERVICE_DB_URL"),
		DatabaseReplicaURL:  getEnv("USER_SERVICE_DB_REPLICA_URL", ""),
		JWTSecret:          getEnvRequired("JWT_SECRET"),
		JWTAccessDuration:  getDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute),
		JWTRefreshDuration: getDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
		InternalAPIKey:     getEnvRequired("INTERNAL_API_KEY"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		Environment:       getEnv("ENVIRONMENT", "development"),
	}
}

// Future service configs can be added here
type EventServiceConfig struct {
	Port                string
	DatabaseURL         string
	DatabaseReplicaURL  string
	InternalAPIKey     string
	UserServiceURL     string
	LogLevel           string
}

func LoadEventServiceConfig() *EventServiceConfig {
	return &EventServiceConfig{
		Port:                getEnv("EVENT_SERVICE_PORT", "8002"),
		DatabaseURL:         getEnvRequired("EVENT_SERVICE_DB_URL"),
		DatabaseReplicaURL:  getEnv("EVENT_SERVICE_DB_REPLICA_URL", ""),
		InternalAPIKey:     getEnvRequired("INTERNAL_API_KEY"),
		UserServiceURL:     getEnvRequired("USER_SERVICE_URL"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}
}

// Helper functions
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