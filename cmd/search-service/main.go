package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fyzanshaik/bookmyevent-ily/services/search"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	validateSearchEnv()

	apiConfig, err := search.InitSearchService()
	if err != nil {
		log.Fatalf("Failed to initialize Search Service: %v", err)
	}

	search.StartServer(apiConfig)
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

func validateSearchEnv() {
	required := []string{
		"SEARCH_SERVICE_PORT",
		"ELASTICSEARCH_URL",
		"REDIS_URL",
		"EVENT_SERVICE_URL",
		"INTERNAL_API_KEY",
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
