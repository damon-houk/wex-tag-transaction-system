package service

import (
	"context"
	"math"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/repository"
	"github.com/google/uuid"
)

// TransactionService handles business logic for transactions
type TransactionService struct {
	repo repository.TransactionRepository
}

// NewTransactionService creates a new transaction service
func NewTransactionService(repo repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// CreateTransaction creates and stores a new transaction
func (s *TransactionService) CreateTransaction(ctx context.Context, desc string, date time.Time, amount float64) (string, error) {
	// Round amount to nearest cent
	amount = math.Round(amount*100) / 100

	// Create transaction entity
	tx := &entity.Transaction{
		ID:          uuid.New().String(),
		Description: desc,
		Date:        date,
		Amount:      amount,
	}

	// Validate
	if err := tx.Validate(); err != nil {
		return "", err
	}

	// Store in repository
	return s.repo.Store(ctx, tx)
}

// GetTransaction retrieves a transaction by ID
func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*entity.Transaction, error) {
	return s.repo.FindByID(ctx, id)
}
