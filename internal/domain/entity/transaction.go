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
}

// Validate ensures the transaction meets all requirements
func (t *Transaction) Validate() error {
	if len(t.Description) > 50 {
		return errors.New("description must not exceed 50 characters")
	}

	if t.Amount <= 0 {
		return errors.New("amount must be a positive value")
	}

	return nil
}
