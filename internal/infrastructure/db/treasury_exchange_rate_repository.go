// Package db internal/infrastructure/db/treasury_exchange_rate_repository.go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/repository"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
)

// ExchangeRateProvider defines an interface for providers of exchange rate data
type ExchangeRateProvider interface {
	FetchExchangeRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error)
}

// TreasuryExchangeRateRepository implements the ExchangeRateRepository interface
type TreasuryExchangeRateRepository struct {
	provider ExchangeRateProvider
	logger   logger.Logger
}

// NewTreasuryExchangeRateRepository creates a new repository for exchange rates
func NewTreasuryExchangeRateRepository(provider ExchangeRateProvider, logger logger.Logger) repository.ExchangeRateRepository {
	return &TreasuryExchangeRateRepository{
		provider: provider,
		logger:   logger,
	}
}

// FindRate finds an exchange rate for a specific currency and date
func (r *TreasuryExchangeRateRepository) FindRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error) {
	r.logger.Info("Finding exchange rate", map[string]interface{}{
		"currency": currency,
		"date":     date.Format("2006-01-02"),
	})

	// Use the provider to get the exchange rate
	rate, err := r.provider.FetchExchangeRate(ctx, currency, date)
	if err != nil {
		r.logger.Error("Failed to retrieve exchange rate", map[string]interface{}{
			"currency": currency,
			"date":     date.Format("2006-01-02"),
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to retrieve exchange rate: %w", err)
	}

	r.logger.Info("Exchange rate found", map[string]interface{}{
		"currency":     currency,
		"date":         date.Format("2006-01-02"),
		"rate":         rate.Rate,
		"rate_date":    rate.Date.Format("2006-01-02"),
		"time_to_find": time.Since(date).String(),
	})

	return rate, nil
}

// StoreRate saves an exchange rate
func (r *TreasuryExchangeRateRepository) StoreRate(ctx context.Context, rate *entity.ExchangeRate) error {
	r.logger.Info("Storing exchange rate", map[string]interface{}{
		"currency":  rate.Currency,
		"rate_date": rate.Date.Format("2006-01-02"),
		"rate":      rate.Rate,
	})

	// Currently, we don't have a persistent storage for exchange rates
	// In a real application, you might want to store this in a database
	return nil
}
