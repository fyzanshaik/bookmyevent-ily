package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"time"

	"github.com/sqlc-dev/pqtype"
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

func ParseAmount(amountStr string) float64 {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0.0
	}
	return amount
}

func GenerateBookingReference() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[mathrand.Intn(len(charset))]
	}
	return "EVT-" + string(b)
}

func GenerateTicketURL(bookingReference string) string {
	return "https://tickets.evently.com/qr/" + bookingReference
}

func GenerateGatewayTransactionID() string {
	return "txn_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func FormatCurrency(amount float64, currency string) string {
	if currency == "" {
		currency = "USD"
	}
	return currency + " " + strconv.FormatFloat(amount, 'f', 2, 64)
}

func CalculateRefundAmount(originalAmount float64, bookingTime time.Time, eventDateTime time.Time) float64 {
	hoursUntilEvent := time.Until(eventDateTime).Hours()

	if hoursUntilEvent > 24 {
		return originalAmount
	}

	if hoursUntilEvent > 2 {
		return originalAmount * 0.5
	}

	return 0.0
}

func NullRawMessageToJSONRawMessage(nullRM pqtype.NullRawMessage) json.RawMessage {
	if !nullRM.Valid {
		return json.RawMessage("{}")
	}
	return json.RawMessage(nullRM.RawMessage)
}

func GetStringFromInterface(i any) string {
	if i == nil {
		return ""
	}
	if s, ok := i.(string); ok {
		return s
	}
	return ""
}

func GetFloatFromInterface(i any) float64 {
	if i == nil {
		return 0.0
	}
	if f, ok := i.(float64); ok {
		return f
	}
	if f, ok := i.(float32); ok {
		return float64(f)
	}
	if n, ok := i.(int); ok {
		return float64(n)
	}
	if n, ok := i.(int32); ok {
		return float64(n)
	}
	if n, ok := i.(int64); ok {
		return float64(n)
	}
	return 0.0
}
