package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/repository"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/middleware"
	"github.com/dgraph-io/badger/v3"
)

// BadgerTransactionRepository implements the transaction repository interface using BadgerDB
type BadgerTransactionRepository struct {
	db     *badger.DB
	logger logger.Logger
}

// NewBadgerTransactionRepository creates a new BadgerDB transaction repository
func NewBadgerTransactionRepository(db *badger.DB, log logger.Logger) repository.TransactionRepository {
	if log == nil {
		log = logger.GetDefaultLogger()
	}

	return &BadgerTransactionRepository{
		db:     db,
		logger: log,
	}
}

// Store saves a transaction and returns its ID
func (r *BadgerTransactionRepository) Store(ctx context.Context, tx *entity.Transaction) (string, error) {
	requestID := middleware.GetRequestID(ctx)

	// Set CreatedAt if not already set
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now().UTC()
		tx.CalculateTTL() // Calculate TTL for data retention
	}

	r.logger.Debug("Storing transaction", map[string]interface{}{
		"request_id":  requestID,
		"id":          tx.ID,
		"description": tx.Description,
		"date":        tx.Date.Format("2006-01-02"),
		"amount":      tx.Amount,
		"created_at":  tx.CreatedAt.Format(time.RFC3339),
		"ttl":         tx.TTL,
	})

	// Serialize transaction to JSON
	data, err := json.Marshal(tx)
	if err != nil {
		r.logger.Error("Failed to marshal transaction", map[string]interface{}{
			"request_id": requestID,
			"id":         tx.ID,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	// Store in BadgerDB
	err = r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("tx:"+tx.ID), data)
	})

	if err != nil {
		r.logger.Error("Failed to store transaction in database", map[string]interface{}{
			"request_id": requestID,
			"id":         tx.ID,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("failed to store transaction: %w", err)
	}

	r.logger.Info("Transaction stored successfully", map[string]interface{}{
		"request_id": requestID,
		"id":         tx.ID,
	})

	return tx.ID, nil
}

// FindByID retrieves a transaction by its unique identifier
func (r *BadgerTransactionRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	requestID := middleware.GetRequestID(ctx)

	r.logger.Debug("Finding transaction by ID", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
	})

	var tx entity.Transaction

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("tx:" + id))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &tx)
		})
	})

	if err == badger.ErrKeyNotFound {
		r.logger.Warn("Transaction not found", map[string]interface{}{
			"request_id": requestID,
			"id":         id,
		})
		return nil, fmt.Errorf("transaction not found: %s", id)
	}

	if err != nil {
		r.logger.Error("Failed to retrieve transaction", map[string]interface{}{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	r.logger.Debug("Transaction found", map[string]interface{}{
		"request_id":  requestID,
		"id":          tx.ID,
		"description": tx.Description,
		"date":        tx.Date.Format("2006-01-02"),
		"amount":      tx.Amount,
		"created_at":  tx.CreatedAt.Format(time.RFC3339),
		"ttl":         tx.TTL,
	})

	// Check if transaction should be expired based on TTL
	if tx.TTL > 0 && time.Now().Unix() > tx.TTL {
		r.logger.Warn("Transaction has expired but was not deleted", map[string]interface{}{
			"request_id": requestID,
			"id":         id,
			"ttl":        tx.TTL,
			"now":        time.Now().Unix(),
		})
		// In production using DynamoDB, this would be handled automatically
		// For BadgerDB, we could implement a background cleanup process
		// For now, we'll still return the transaction
	}

	return &tx, nil
}
