// internal/infrastructure/middleware/middleware_test.go
package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware(t *testing.T) {
	// Setup
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from context
		requestID := r.Context().Value(requestIDKey)
		assert.NotNil(t, requestID)

		// Write it to the response for testing
		w.Write([]byte(requestID.(string)))
	})

	middleware := RequestIDMiddleware(nextHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Test with no existing request ID
	middleware.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
	assert.Equal(t, requestID, w.Body.String())

	// Test with existing request ID
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-id-123")
	w = httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Verify existing ID was preserved
	assert.Equal(t, "test-id-123", w.Header().Get("X-Request-ID"))
	assert.Equal(t, "test-id-123", w.Body.String())
}

func TestGetRequestID(t *testing.T) {
	// Test with valid request ID
	ctx := context.WithValue(context.Background(), requestIDKey, "test-id-123")
	assert.Equal(t, "test-id-123", GetRequestID(ctx))

	// Test with no request ID
	assert.Equal(t, "unknown", GetRequestID(context.Background()))
}

func TestMiddlewareChain(t *testing.T) {
	// This test verifies that the middleware chain correctly preserves the request ID
	var buf bytes.Buffer
	log := logger.NewJSONLogger(&buf, logger.InfoLevel)

	// Create a chain of middleware with a handler that returns the request ID
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		w.Write([]byte(requestID))
	})

	// Apply RequestIDMiddleware then LoggingMiddleware
	chain := RequestIDMiddleware(LoggingMiddleware(log)(finalHandler))

	// Create a request with a known ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-id-123")
	w := httptest.NewRecorder()

	// Process the request
	chain.ServeHTTP(w, req)

	// Check that the final handler received the request ID
	assert.Equal(t, "test-id-123", w.Body.String())

	// Check that the request ID appears in logs
	logs := buf.String()
	assert.Contains(t, logs, "test-id-123", "Request ID should be in logs")
}
