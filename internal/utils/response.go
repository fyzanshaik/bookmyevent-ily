package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fyzanshaik/bookmyevent-ily/internal/constants"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(code)
	
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(payload); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}
}

func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}
