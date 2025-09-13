package user

import (
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/utils"
)

func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (cfg *APIConfig) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "ready",
		"service": "user-service",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}