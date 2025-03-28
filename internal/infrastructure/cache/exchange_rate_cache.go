package cache

import (
	"sync"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
)

// ExchangeRateCache provides a thread-safe in-memory cache for exchange rates
type ExchangeRateCache struct {
	cache map[string]*entity.ExchangeRate
	mutex sync.RWMutex
}

// NewExchangeRateCache creates a new exchange rate cache
func NewExchangeRateCache() *ExchangeRateCache {
	return &ExchangeRateCache{
		cache: make(map[string]*entity.ExchangeRate),
	}
}

// generateCacheKey creates a cache key from currency and date
func generateCacheKey(currency string, date time.Time) string {
	return currency + ":" + date.Format("2006-01-02")
}

// Get retrieves an exchange rate from the cache if available
func (c *ExchangeRateCache) Get(currency string, date time.Time) *entity.ExchangeRate {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := generateCacheKey(currency, date)
	rate, exists := c.cache[key]
	if !exists {
		return nil
	}
	return rate
}

// Put stores an exchange rate in the cache
func (c *ExchangeRateCache) Put(rate *entity.ExchangeRate, forDate time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := generateCacheKey(rate.Currency, forDate)
	c.cache[key] = rate
}

// Clear clears all entries from the cache
func (c *ExchangeRateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*entity.ExchangeRate)
}
