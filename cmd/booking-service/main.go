package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fyzanshaik/bookmyevent-ily/services/booking"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	validateBookingEnv()

	appConfig, db := booking.InitBookingService()
	defer db.Close()

	booking.StartServer(appConfig)
}

func loadEnv() {
	wd, _ := os.Getwd()
	for {
		envPath := filepath.Join(wd, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				log.Fatalf("Error loading .env: %v", err)
			}
			return
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			log.Fatal(".env file not found")
		}
		wd = parent
	}
}

func validateBookingEnv() {
	required := []string{
		"BOOKING_SERVICE_PORT",
		"BOOKING_SERVICE_DB_URL",
		"REDIS_URL",
		"JWT_SECRET",
		"INTERNAL_API_KEY",
		"USER_SERVICE_URL",
		"EVENT_SERVICE_URL",
	}

	var missing []string
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		log.Fatalf("Missing env vars: %v", missing)
	}

	fmt.Println("Environment loaded")
}
