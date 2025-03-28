package cache

import (
	"testing"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestExchangeRateCache(t *testing.T) {
	cache := NewExchangeRateCache()

	// Test initial state
	assert.Equal(t, 0, cache.Size())

	// Test storing and retrieving
	date := time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)
	rate := &entity.ExchangeRate{
		Currency: "EUR",
		Date:     date,
		Rate:     0.85,
	}

	cache.Put(rate, date)
	assert.Equal(t, 1, cache.Size())

	// Test retrieval
	retrieved := cache.Get("EUR", date)
	assert.NotNil(t, retrieved)
	assert.Equal(t, rate.Currency, retrieved.Currency)
	assert.Equal(t, rate.Rate, retrieved.Rate)

	// Test non-existent retrieval
	nonexistent := cache.Get("GBP", date)
	assert.Nil(t, nonexistent)

	// Test expiration
	cache.SetExpiration(10 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	expired := cache.Get("EUR", date)
	assert.Nil(t, expired)

	// Test cleaning expired entries
	cache.Put(rate, date)
	time.Sleep(20 * time.Millisecond)
	count := cache.CleanExpired()
	assert.Equal(t, 1, count)
	assert.Equal(t, 0, cache.Size())

	// Test clearing
	cache.SetExpiration(1 * time.Hour)
	cache.Put(rate, date)
	assert.Equal(t, 1, cache.Size())
	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}
