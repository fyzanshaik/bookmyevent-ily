package middleware

import (
	"net/http"
	"time"

	"github.com/fyzanshaik/bookmyevent-ily/internal/logger"
	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.New().String()[:8]

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     200,
			}

			r.Header.Set("X-Request-ID", requestID)
			w.Header().Set("X-Request-ID", requestID)

			requestLogger := log.WithRequestID(requestID)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			
			fields := map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     rw.statusCode,
				"duration":   duration.String(),
				"ip":         getClientIP(r),
				"user_agent": r.UserAgent(),
				"size":       rw.size,
			}

			switch {
			case rw.statusCode >= 500:
				requestLogger.WithFields(fields).Error("HTTP Request")
			case rw.statusCode >= 400:
				requestLogger.WithFields(fields).Warn("HTTP Request")
			default:
				requestLogger.WithFields(fields).Info("HTTP Request")
			}
		})
	}
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	return r.RemoteAddr
}

