package service

import (
	"context"
	"time"

	"github.com/yourusername/wex-tag-transaction-system/internal/domain/entity"
)

// TreasuryAPI defines the interface for interacting with the Treasury API
type TreasuryAPI interface {
	// GetExchangeRate retrieves an exchange rate for a currency and date
	GetExchangeRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error)
}
