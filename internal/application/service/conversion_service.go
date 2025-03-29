// Package service internal/application/service/conversion_service.go
package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/repository"
)

// ConvertedTransaction represents a transaction with conversion information
type ConvertedTransaction struct {
	ID              string    `json:"id"`
	Description     string    `json:"description"`
	Date            time.Time `json:"date"`
	OriginalAmount  float64   `json:"original_amount"`
	Currency        string    `json:"currency"`
	ExchangeRate    float64   `json:"exchange_rate"`
	ConvertedAmount float64   `json:"converted_amount"`
	RateDate        time.Time `json:"rate_date"`
}

// ConversionService handles currency conversion for transactions
type ConversionService struct {
	txRepo       repository.TransactionRepository
	exchangeRepo repository.ExchangeRateRepository
}

// NewConversionService creates a new conversion service
func NewConversionService(txRepo repository.TransactionRepository, exchangeRepo repository.ExchangeRateRepository) *ConversionService {
	return &ConversionService{
		txRepo:       txRepo,
		exchangeRepo: exchangeRepo,
	}
}

// GetTransactionInCurrency retrieves a transaction converted to the specified currency
func (s *ConversionService) GetTransactionInCurrency(ctx context.Context, id, currency string) (*ConvertedTransaction, error) {
	// Get transaction
	tx, err := s.txRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	// Find applicable exchange rate
	rate, err := s.exchangeRepo.FindRate(ctx, currency, tx.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Calculate converted amount
	convertedAmount := tx.Amount * rate.Rate

	// Round to two decimal places
	convertedAmount = math.Round(convertedAmount*100) / 100

	return &ConvertedTransaction{
		ID:              tx.ID,
		Description:     tx.Description,
		Date:            tx.Date,
		OriginalAmount:  tx.Amount,
		Currency:        currency,
		ExchangeRate:    rate.Rate,
		ConvertedAmount: convertedAmount,
		RateDate:        rate.Date,
	}, nil
}
