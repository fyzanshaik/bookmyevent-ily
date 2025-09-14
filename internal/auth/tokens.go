package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

func MakeRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	refreshToken := hex.EncodeToString(randomBytes)
	return refreshToken, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authorizationHeader := headers.Get("Authorization")
	if authorizationHeader == "" {
		return "", fmt.Errorf("authorization header doesn't exist")
	}

	splitHeader := strings.Split(authorizationHeader, " ")
	if len(splitHeader) < 2 || splitHeader[0] != "Bearer" {
		return "", fmt.Errorf("bearer token doesn't exist")
	}

	return splitHeader[1], nil
}

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("X-API-Key")
	if apiKey != "" {
		return apiKey, nil
	}

	authorizationHeader := headers.Get("Authorization")
	if authorizationHeader == "" {
		return "", fmt.Errorf("API key not found in X-API-Key header or Authorization header")
	}

	splitHeader := strings.Split(authorizationHeader, " ")
	if len(splitHeader) < 2 || splitHeader[0] != "ApiKey" {
		return "", fmt.Errorf("api key doesn't exist in Authorization header")
	}

	return splitHeader[1], nil
}