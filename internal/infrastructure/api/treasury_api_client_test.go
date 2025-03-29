// internal/infrastructure/api/treasury_api_client_test.go
package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchExchangeRate(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping treasury API test in short mode")
	}

	// Setup a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the request has the expected path and parameters
		assert.Contains(t, r.URL.Path, "/v1/accounting/od/rates_of_exchange")

		// Return a mock response based on the query parameters
		currency := r.URL.Query().Get("filter")
		if currency == "" || !contains(currency, "currency:eq:EUR") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Send a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": [
				{
					"country": "Euro Zone",
					"currency": "Euro",
					"exchange_rate": "0.85",
					"record_date": "2023-04-10",
					"effective_date": "2023-04-10"
				}
			],
			"meta": {
				"count": 1,
				"labels": {
					"country": "country",
					"currency": "currency",
					"exchange_rate": "exchange_rate",
					"record_date": "record_date",
					"effective_date": "effective_date"
				}
			}
		}`))
	}))
	defer mockServer.Close()

	// Create client with mock server URL
	client := NewTreasuryAPIClient(nil)
	client.baseURL = mockServer.URL // Replace the real URL with our mock

	// Test successful request
	ctx := context.Background()
	date := time.Date(2023, 4, 15, 0, 0, 0, 0, time.UTC)
	rate, err := client.FetchExchangeRate(ctx, "EUR", date)

	// Assert response
	assert.NoError(t, err)
	assert.NotNil(t, rate)
	assert.Equal(t, "EUR", rate.Currency)
	assert.Equal(t, 0.85, rate.Rate)

	// Test rate date parsing
	expectedDate, _ := time.Parse("2006-01-02", "2023-04-10")
	assert.Equal(t, expectedDate, rate.Date)

	// Test error handling with invalid currency
	_, err = client.FetchExchangeRate(ctx, "", date)
	assert.Error(t, err)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
