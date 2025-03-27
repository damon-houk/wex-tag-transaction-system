package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/dgraph-io/badger/v3"
)

// BadgerTransactionRepository implements the transaction repository interface using BadgerDB
type BadgerTransactionRepository struct {
	db *badger.DB
}

// NewBadgerTransactionRepository creates a new BadgerDB transaction repository
func NewBadgerTransactionRepository(db *badger.DB) *BadgerTransactionRepository {
	return &BadgerTransactionRepository{db: db}
}

// Store saves a transaction and returns its ID
func (r *BadgerTransactionRepository) Store(ctx context.Context, tx *entity.Transaction) (string, error) {
	// Serialize transaction to JSON
	data, err := json.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	// Store in BadgerDB
	err = r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("tx:"+tx.ID), data)
	})

	if err != nil {
		return "", fmt.Errorf("failed to store transaction: %w", err)
	}

	return tx.ID, nil
}

// FindByID retrieves a transaction by its unique identifier
func (r *BadgerTransactionRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
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
		return nil, fmt.Errorf("transaction not found: %s", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	return &tx, nil
}
