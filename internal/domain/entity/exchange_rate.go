package entity

import (
	"time"
)

// ExchangeRate represents a currency exchange rate at a specific date
type ExchangeRate struct {
	Currency string    `json:"currency"`
	Date     time.Time `json:"date"`
	Rate     float64   `json:"rate"`
}
