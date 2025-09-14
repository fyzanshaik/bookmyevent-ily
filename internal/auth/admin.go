package auth

import (
	"fmt"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/constants"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AdminClaims struct {
	AdminID     uuid.UUID `json:"admin_id"`
	Role        string    `json:"role"`
	Permissions string    `json:"permissions"`
	jwt.RegisteredClaims
}

func MakeAdminJWT(adminID uuid.UUID, role, permissions, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := AdminClaims{
		AdminID:     adminID,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    constants.JWTIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			Subject:   adminID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign admin token: %w", err)
	}

	return signedToken, nil
}

func ValidateAdminJWT(tokenString, tokenSecret string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse admin token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid admin token")
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok {
		return nil, fmt.Errorf("could not parse admin claims")
	}

	return claims, nil
}
