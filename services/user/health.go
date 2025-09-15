package user

import (
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
)

func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status":  "healthy",
		"service": "user-service",
		"message": "Service is running normally",
	}
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status":  "ready",
		"service": "user-service",
		"message": "Service is ready to accept requests",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}