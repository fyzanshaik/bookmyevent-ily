package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

func GetCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateRequestID() string {
	id, _ := GenerateRandomString(16)
	return id
}

func StringPtrFromNullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func ParsePrice(priceStr sql.NullString) float64 {
	if !priceStr.Valid {
		return 0.0
	}

	var price float64
	if _, err := fmt.Sscanf(priceStr.String, "%f", &price); err != nil {
		return 0.0
	}
	return price
}
