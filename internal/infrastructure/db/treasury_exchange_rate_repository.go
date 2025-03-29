// Package db internal/infrastructure/db/treasury_exchange_rate_repository.go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/repository"
)

// ExchangeRateProvider defines an interface for providers of exchange rate data
type ExchangeRateProvider interface {
	FetchExchangeRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error)
}

// TreasuryExchangeRateRepository implements the ExchangeRateRepository interface
type TreasuryExchangeRateRepository struct {
	provider ExchangeRateProvider
}

// NewTreasuryExchangeRateRepository creates a new repository for exchange rates
func NewTreasuryExchangeRateRepository(provider ExchangeRateProvider) repository.ExchangeRateRepository {
	return &TreasuryExchangeRateRepository{
		provider: provider,
	}
}

// FindRate finds an exchange rate for a specific currency and date
func (r *TreasuryExchangeRateRepository) FindRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error) {
	// Use the provider to get the exchange rate
	rate, err := r.provider.FetchExchangeRate(ctx, currency, date)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve exchange rate: %w", err)
	}

	return rate, nil
}

// StoreRate saves an exchange rate
func (r *TreasuryExchangeRateRepository) StoreRate(ctx context.Context, rate *entity.ExchangeRate) error {
	// Currently, we don't have a persistent storage for exchange rates
	// In a real application, you might want to store this in a database
	return nil
}
