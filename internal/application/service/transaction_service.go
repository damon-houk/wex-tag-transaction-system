package service

import (
	"context"
	"math"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/repository"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// TransactionService handles business logic for transactions
type TransactionService struct {
	repo   repository.TransactionRepository
	logger logger.Logger
}

// NewTransactionService creates a new transaction service
func NewTransactionService(repo repository.TransactionRepository, log logger.Logger) *TransactionService {
	if log == nil {
		log = logger.GetDefaultLogger()
	}

	return &TransactionService{
		repo:   repo,
		logger: log,
	}
}

// CreateTransaction creates and stores a new transaction
func (s *TransactionService) CreateTransaction(ctx context.Context, desc string, date time.Time, amount float64) (string, error) {
	requestID := middleware.GetRequestID(ctx)

	s.logger.Info("Creating new transaction", map[string]interface{}{
		"request_id":  requestID,
		"description": desc,
		"date":        date.Format("2006-01-02"),
		"amount":      amount,
	})

	// Round amount to nearest cent
	amount = math.Round(amount*100) / 100

	now := time.Now().UTC()

	// Create transaction entity
	tx := &entity.Transaction{
		ID:          uuid.New().String(),
		Description: desc,
		Date:        date,
		Amount:      amount,
		CreatedAt:   now,
	}

	// Calculate TTL for data retention
	tx.CalculateTTL()

	// Validate
	if err := tx.Validate(); err != nil {
		s.logger.Error("Transaction validation failed", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return "", err
	}

	// Store in repository
	id, err := s.repo.Store(ctx, tx)
	if err != nil {
		s.logger.Error("Failed to store transaction", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return "", err
	}

	s.logger.Info("Transaction created successfully", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
	})

	return id, nil
}

// GetTransaction retrieves a transaction by ID
func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*entity.Transaction, error) {
	requestID := middleware.GetRequestID(ctx)

	s.logger.Info("Retrieving transaction", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
	})

	tx, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve transaction", map[string]interface{}{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		})
		return nil, err
	}

	s.logger.Info("Transaction retrieved successfully", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
	})

	return tx, nil
}
