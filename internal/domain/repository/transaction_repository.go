package repository

import (
	"context"

	"github.com/yourusername/wex-tag-transaction-system/internal/domain/entity"
)

// TransactionRepository defines the interface for transaction storage
type TransactionRepository interface {
	// Store saves a transaction and returns its ID
	Store(ctx context.Context, transaction *entity.Transaction) (string, error)

	// FindByID retrieves a transaction by its unique identifier
	FindByID(ctx context.Context, id string) (*entity.Transaction, error)
}
