package cache

import (
	"sync"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
)

// CacheEntry represents a cached exchange rate with expiration
type CacheEntry struct {
	Rate      *entity.ExchangeRate
	Timestamp time.Time
}

// ExchangeRateCache provides a thread-safe in-memory cache for exchange rates
type ExchangeRateCache struct {
	cache      map[string]CacheEntry
	expiration time.Duration
	mutex      sync.RWMutex
}

// NewExchangeRateCache creates a new exchange rate cache
func NewExchangeRateCache() *ExchangeRateCache {
	return &ExchangeRateCache{
		cache:      make(map[string]CacheEntry),
		expiration: 24 * time.Hour, // Default 24h expiration
	}
}

// generateCacheKey creates a cache key from currency and date
func generateCacheKey(currency string, date time.Time) string {
	return currency + ":" + date.Format("2006-01-02")
}

// Get retrieves an exchange rate from the cache if available and not expired
func (c *ExchangeRateCache) Get(currency string, date time.Time) *entity.ExchangeRate {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := generateCacheKey(currency, date)
	entry, exists := c.cache[key]

	// Return nil if entry doesn't exist or is expired
	if !exists || time.Since(entry.Timestamp) > c.expiration {
		return nil
	}

	return entry.Rate
}

// Put stores an exchange rate in the cache
func (c *ExchangeRateCache) Put(rate *entity.ExchangeRate, forDate time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := generateCacheKey(rate.Currency, forDate)
	c.cache[key] = CacheEntry{
		Rate:      rate,
		Timestamp: time.Now(),
	}
}

// Clear clears all entries from the cache
func (c *ExchangeRateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]CacheEntry)
}

// SetExpiration sets the cache expiration duration
func (c *ExchangeRateCache) SetExpiration(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.expiration = duration
}

// Size returns the number of items in the cache
func (c *ExchangeRateCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// CleanExpired removes expired entries from the cache
func (c *ExchangeRateCache) CleanExpired() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0
	now := time.Now()

	for key, entry := range c.cache {
		if now.Sub(entry.Timestamp) > c.expiration {
			delete(c.cache, key)
			count++
		}
	}

	return count
}
