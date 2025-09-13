package utils

import (
	"crypto/rand"
	"encoding/hex"
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