// internal/application/service/transaction_service_test.go
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/damon-houk/wex-tag-transaction-system/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTransaction(t *testing.T) {
	repo := new(mocks.MockTransactionRepository)
	log := logger.NewJSONLogger(nil, logger.InfoLevel)
	service := NewTransactionService(repo, log)
	ctx := context.Background()

	t.Run("Valid transaction", func(t *testing.T) {
		// Setup
		desc := "Test transaction"
		date := time.Now()
		amount := 123.45

		// Mock expectations
		repo.On("Store", ctx, mock.MatchedBy(func(tx *entity.Transaction) bool {
			return tx.Description == desc && tx.Date == date && tx.Amount == amount
		})).Return("test-id", nil).Once()

		// Execute
		id, err := service.CreateTransaction(ctx, desc, date, amount)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "test-id", id)
		repo.AssertExpectations(t)
	})

	t.Run("Invalid description", func(t *testing.T) {
		// Setup
		desc := "This description is way too long and exceeds the 50 character limit"
		date := time.Now()
		amount := 123.45

		// Execute
		id, err := service.CreateTransaction(ctx, desc, date, amount)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "", id)
		assert.Contains(t, err.Error(), "description must not exceed 50 characters")
	})

	t.Run("Invalid amount", func(t *testing.T) {
		// Setup
		desc := "Test transaction"
		date := time.Now()
		amount := -123.45

		// Execute
		id, err := service.CreateTransaction(ctx, desc, date, amount)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "", id)
		assert.Contains(t, err.Error(), "amount must be a positive value")
	})

	t.Run("Repository error", func(t *testing.T) {
		// Setup
		desc := "Test transaction"
		date := time.Now()
		amount := 123.45

		// Mock expectations
		repo.On("Store", ctx, mock.Anything).Return("", errors.New("repository error")).Once()

		// Execute
		id, err := service.CreateTransaction(ctx, desc, date, amount)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "", id)
		assert.Equal(t, "repository error", err.Error())
		repo.AssertExpectations(t)
	})
}
