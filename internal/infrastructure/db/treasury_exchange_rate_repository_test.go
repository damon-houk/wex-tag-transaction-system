// internal/infrastructure/db/treasury_exchange_rate_repository_test.go
package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/damon-houk/wex-tag-transaction-system/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestTreasuryExchangeRateRepository(t *testing.T) {
	mockProvider := new(mocks.MockExchangeRateProvider)
	log := logger.NewJSONLogger(nil, logger.InfoLevel)
	repo := NewTreasuryExchangeRateRepository(mockProvider, log)

	ctx := context.Background()
	testDate := time.Date(2023, 4, 15, 0, 0, 0, 0, time.UTC)
	expectedRate := &entity.ExchangeRate{
		Currency: "EUR",
		Date:     testDate.AddDate(0, 0, -5),
		Rate:     0.85,
	}

	t.Run("Successful rate retrieval", func(t *testing.T) {
		// Set up mock expectations
		mockProvider.On("FetchExchangeRate", ctx, "EUR", testDate).Return(expectedRate, nil).Once()

		// Test FindRate
		rate, err := repo.FindRate(ctx, "EUR", testDate)
		assert.NoError(t, err)
		assert.Equal(t, expectedRate, rate)

		// Verify mock was called
		mockProvider.AssertExpectations(t)
	})

	t.Run("API client error", func(t *testing.T) {
		// Set up mock expectations for error case
		mockProvider.On("FetchExchangeRate", ctx, "XYZ", testDate).
			Return(nil, errors.New("currency not supported")).Once()

		// Test FindRate with error
		rate, err := repo.FindRate(ctx, "XYZ", testDate)
		assert.Error(t, err)
		assert.Nil(t, rate)
		assert.Contains(t, err.Error(), "failed to retrieve exchange rate")

		// Verify mock was called
		mockProvider.AssertExpectations(t)
	})

	t.Run("StoreRate", func(t *testing.T) {
		// Test StoreRate (currently a no-op)
		err := repo.StoreRate(ctx, expectedRate)
		assert.NoError(t, err)
	})
}
