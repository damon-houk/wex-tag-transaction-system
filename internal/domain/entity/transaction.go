package entity

import (
	"errors"
	"time"
)

// Transaction represents a purchase transaction
type Transaction struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	TTL         int64     `json:"ttl,omitempty"` // Time-to-live for DynamoDB
}

// Validate ensures the transaction meets all requirements
func (t *Transaction) Validate() error {
	if len(t.Description) > 50 {
		return errors.New("description must not exceed 50 characters")
	}

	if t.Amount <= 0 {
		return errors.New("amount must be a positive value")
	}

	if t.Date.After(time.Now()) {
		return errors.New("transaction date cannot be in the future")
	}

	return nil
}

// CalculateTTL calculates the TTL for data retention (1 year)
func (t *Transaction) CalculateTTL() {
	// Set TTL to one year from creation date
	oneYearFromNow := t.CreatedAt.AddDate(1, 0, 0)
	t.TTL = oneYearFromNow.Unix()
}
