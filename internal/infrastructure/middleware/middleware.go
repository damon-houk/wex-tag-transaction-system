// internal/infrastructure/middleware/middleware.go
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/google/uuid"
)

// Keys for context values
type contextKey string

const (
	requestIDKey contextKey = "request_id"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add ID to context
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs requests and responses
func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Create a response wrapper to capture status code
			wrapper := newResponseWrapper(w)

			// Get request ID from context
			requestID, ok := r.Context().Value(requestIDKey).(string)
			if !ok || requestID == "" {
				requestID = "unknown"
			}

			log.Info("Request received", map[string]interface{}{
				"request_id":     requestID,
				"method":         r.Method,
				"path":           r.URL.Path,
				"query":          r.URL.RawQuery,
				"remote_addr":    r.RemoteAddr,
				"user_agent":     r.UserAgent(),
				"content_type":   r.Header.Get("Content-Type"),
				"content_length": r.ContentLength,
			})

			// Call next handler
			next.ServeHTTP(wrapper, r)

			// Log response
			duration := time.Since(startTime)
			log.Info("Response sent", map[string]interface{}{
				"request_id":     requestID,
				"method":         r.Method,
				"path":           r.URL.Path,
				"status":         wrapper.statusCode,
				"duration_ms":    duration.Milliseconds(),
				"content_type":   wrapper.Header().Get("Content-Type"),
				"content_length": wrapper.contentLength,
			})
		})
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok || requestID == "" {
		return "unknown"
	}
	return requestID
}

// responseWrapper wraps http.ResponseWriter to capture the status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode    int
	contentLength int64
}

// newResponseWrapper creates a new response wrapper
func newResponseWrapper(w http.ResponseWriter) *responseWrapper {
	return &responseWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status
	}
}

// WriteHeader captures the status code
func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the content length
func (rw *responseWrapper) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.contentLength += int64(n)
	return n, err
}
