// internal/infrastructure/api/treasury_api_integration_test.go
package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTreasuryAPIIntegration(t *testing.T) {
	// This test makes actual API calls - skip in short mode and CI
	if testing.Short() {
		t.Skip("Skipping Treasury API integration test in short mode")
	}

	// Create client with actual API
	client := NewTreasuryAPIClient(nil)

	// Test with a known currency and recent date
	ctx := context.Background()

	// Use a date a few months in the past to ensure rates exist
	date := time.Now().AddDate(0, -3, 0)

	// Common currencies to test
	currencies := []string{"EUR", "CAD", "GBP", "JPY"}

	for _, currency := range currencies {
		t.Run(currency, func(t *testing.T) {
			rate, err := client.FetchExchangeRate(ctx, currency, date)

			// We don't know if we'll get a result for every currency, but the call should succeed
			if err != nil {
				// Only fail if it's not a "no exchange rate available" error
				if rate == nil && !contains(err.Error(), "no exchange rate available") {
					t.Fatalf("Failed to get exchange rate for %s: %v", currency, err)
				}
				t.Logf("No rate available for %s: %v", currency, err)
				return
			}

			// If we got a result, validate it
			assert.NotNil(t, rate)
			assert.Equal(t, currency, rate.Currency)
			assert.Greater(t, rate.Rate, 0.0)
			assert.False(t, rate.Date.IsZero())
			assert.True(t, rate.Date.Before(date) || rate.Date.Equal(date))
			assert.True(t, rate.Date.After(date.AddDate(0, -6, 0)) || rate.Date.Equal(date.AddDate(0, -6, 0)))

			t.Logf("Got exchange rate for %s: %f on %s",
				currency, rate.Rate, rate.Date.Format("2006-01-02"))
		})
	}
}
