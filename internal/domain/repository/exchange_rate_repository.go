package repository

import (
	"context"
	"time"

	"github.com/yourusername/wex-tag-transaction-system/internal/domain/entity"
)

// ExchangeRateRepository defines the interface for exchange rate access
type ExchangeRateRepository interface {
	// FindRate finds an exchange rate for a specific currency and date
	FindRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error)

	// StoreRate saves an exchange rate
	StoreRate(ctx context.Context, rate *entity.ExchangeRate) error
}
