package main

import (
	"log"

	"github.com/fyzanshaik/bookmyevent-ily/services/search"
)

func main() {
	apiConfig, err := search.InitSearchService()
	if err != nil {
		log.Fatalf("Failed to initialize Search Service: %v", err)
	}

	search.StartServer(apiConfig)
}
