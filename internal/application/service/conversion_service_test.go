// internal/application/service/conversion_service_test.go
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository is a mock implementation of the transaction repository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Store(ctx context.Context, tx *entity.Transaction) (string, error) {
	args := m.Called(ctx, tx)
	return args.String(0), args.Error(1)
}

func (m *MockTransactionRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

// MockExchangeRateRepository is a mock implementation of the exchange rate repository
type MockExchangeRateRepository struct {
	mock.Mock
}

func (m *MockExchangeRateRepository) FindRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error) {
	args := m.Called(ctx, currency, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ExchangeRate), args.Error(1)
}

func (m *MockExchangeRateRepository) StoreRate(ctx context.Context, rate *entity.ExchangeRate) error {
	args := m.Called(ctx, rate)
	return args.Error(0)
}

func TestGetTransactionInCurrency(t *testing.T) {
	repo := new(MockTransactionRepository)
	exchangeRepo := new(MockExchangeRateRepository)
	log := logger.NewJSONLogger(nil, logger.InfoLevel)
	service := NewConversionService(repo, exchangeRepo, log)
	ctx := context.Background()

	t.Run("Successful conversion", func(t *testing.T) {
		// Setup
		txID := "test-id"
		currency := "EUR"

		tx := &entity.Transaction{
			ID:          txID,
			Description: "Test transaction",
			Date:        time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			Amount:      100.00,
		}

		rate := &entity.ExchangeRate{
			Currency: currency,
			Date:     time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC),
			Rate:     0.85,
		}

		// Mock expectations
		repo.On("FindByID", ctx, txID).Return(tx, nil).Once()
		exchangeRepo.On("FindRate", ctx, currency, tx.Date).Return(rate, nil).Once()

		// Execute
		result, err := service.GetTransactionInCurrency(ctx, txID, currency)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, txID, result.ID)
		assert.Equal(t, tx.Description, result.Description)
		assert.Equal(t, tx.Date, result.Date)
		assert.Equal(t, tx.Amount, result.OriginalAmount)
		assert.Equal(t, currency, result.Currency)
		assert.Equal(t, rate.Rate, result.ExchangeRate)
		assert.Equal(t, 85.00, result.ConvertedAmount) // 100.00 * 0.85 = 85.00
		assert.Equal(t, rate.Date, result.RateDate)

		repo.AssertExpectations(t)
		exchangeRepo.AssertExpectations(t)
	})

	t.Run("Transaction not found", func(t *testing.T) {
		// Setup
		txID := "non-existent-id"
		currency := "EUR"

		// Mock expectations
		repo.On("FindByID", ctx, txID).Return(nil, errors.New("transaction not found")).Once()

		// Execute
		result, err := service.GetTransactionInCurrency(ctx, txID, currency)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to retrieve transaction")

		repo.AssertExpectations(t)
	})

	t.Run("Exchange rate not available", func(t *testing.T) {
		// Setup
		txID := "test-id"
		currency := "XYZ" // Non-existent currency

		tx := &entity.Transaction{
			ID:          txID,
			Description: "Test transaction",
			Date:        time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			Amount:      100.00,
		}

		// Mock expectations
		repo.On("FindByID", ctx, txID).Return(tx, nil).Once()
		exchangeRepo.On("FindRate", ctx, currency, tx.Date).
			Return(nil, errors.New("no exchange rate available")).Once()

		// Execute
		result, err := service.GetTransactionInCurrency(ctx, txID, currency)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get exchange rate")

		repo.AssertExpectations(t)
		exchangeRepo.AssertExpectations(t)
	})

	t.Run("Rounding of converted amount", func(t *testing.T) {
		// Setup
		txID := "test-id"
		currency := "EUR"

		tx := &entity.Transaction{
			ID:          txID,
			Description: "Test transaction",
			Date:        time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			Amount:      100.00,
		}

		rate := &entity.ExchangeRate{
			Currency: currency,
			Date:     time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC),
			Rate:     0.8333, // This will result in a repeating decimal
		}

		// Mock expectations
		repo.On("FindByID", ctx, txID).Return(tx, nil).Once()
		exchangeRepo.On("FindRate", ctx, currency, tx.Date).Return(rate, nil).Once()

		// Execute
		result, err := service.GetTransactionInCurrency(ctx, txID, currency)

		// Assert
		assert.NoError(t, err)
		// 100.00 * 0.8333 = 83.33 (should be rounded to 2 decimal places)
		assert.Equal(t, 83.33, result.ConvertedAmount)

		repo.AssertExpectations(t)
		exchangeRepo.AssertExpectations(t)
	})
}
