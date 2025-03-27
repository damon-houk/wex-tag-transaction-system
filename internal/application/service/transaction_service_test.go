package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
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

func TestCreateTransaction(t *testing.T) {
	repo := new(MockTransactionRepository)
	service := NewTransactionService(repo)
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
