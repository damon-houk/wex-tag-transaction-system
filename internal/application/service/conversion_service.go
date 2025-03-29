// Package service internal/application/service/conversion_service.go
package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/repository"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/middleware"
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
	logger       logger.Logger
}

// NewConversionService creates a new conversion service
func NewConversionService(txRepo repository.TransactionRepository, exchangeRepo repository.ExchangeRateRepository, log logger.Logger) *ConversionService {
	if log == nil {
		log = logger.GetDefaultLogger()
	}

	return &ConversionService{
		txRepo:       txRepo,
		exchangeRepo: exchangeRepo,
		logger:       log,
	}
}

// GetTransactionInCurrency retrieves a transaction converted to the specified currency
func (s *ConversionService) GetTransactionInCurrency(ctx context.Context, id, currency string) (*ConvertedTransaction, error) {
	requestID := middleware.GetRequestID(ctx)

	s.logger.Info("Converting transaction currency", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
		"currency":   currency,
	})

	// Get transaction
	tx, err := s.txRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve transaction for conversion", map[string]interface{}{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	s.logger.Debug("Retrieved transaction for conversion", map[string]interface{}{
		"request_id":  requestID,
		"id":          id,
		"description": tx.Description,
		"date":        tx.Date.Format("2006-01-02"),
		"amount":      tx.Amount,
	})

	// Find applicable exchange rate
	s.logger.Debug("Finding exchange rate", map[string]interface{}{
		"request_id": requestID,
		"currency":   currency,
		"date":       tx.Date.Format("2006-01-02"),
	})

	rate, err := s.exchangeRepo.FindRate(ctx, currency, tx.Date)
	if err != nil {
		s.logger.Error("Failed to get exchange rate", map[string]interface{}{
			"request_id": requestID,
			"currency":   currency,
			"date":       tx.Date.Format("2006-01-02"),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	s.logger.Info("Found exchange rate", map[string]interface{}{
		"request_id": requestID,
		"currency":   currency,
		"rate_date":  rate.Date.Format("2006-01-02"),
		"rate":       rate.Rate,
	})

	// Calculate converted amount
	convertedAmount := tx.Amount * rate.Rate

	// Round to two decimal places
	convertedAmount = math.Round(convertedAmount*100) / 100

	s.logger.Info("Conversion completed", map[string]interface{}{
		"request_id":       requestID,
		"id":               id,
		"currency":         currency,
		"original_amount":  tx.Amount,
		"exchange_rate":    rate.Rate,
		"converted_amount": convertedAmount,
		"rate_date":        rate.Date.Format("2006-01-02"),
	})

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
