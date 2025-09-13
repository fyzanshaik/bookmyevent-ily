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
    if err := json.NewEncoder(w).Encode(payload); err != nil {
        log.Printf("Error encoding JSON: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"error":"Internal server error"}`))
    }
}
