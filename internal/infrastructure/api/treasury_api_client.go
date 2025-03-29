// Package api internal/infrastructure/api/treasury_api_client.go
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/cache"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/db"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
)

const (
	treasuryBaseURL  = "https://api.fiscaldata.treasury.gov/services/api/fiscal_service"
	exchangeRatePath = "/v1/accounting/od/rates_of_exchange"
)

// TreasuryAPIClient is a client for the Treasury API
type TreasuryAPIClient struct {
	baseURL    string
	httpClient *http.Client
	cache      *cache.ExchangeRateCache
	logger     logger.Logger
}

// Ensure TreasuryAPIClient implements the ExchangeRateProvider interface
var _ db.ExchangeRateProvider = (*TreasuryAPIClient)(nil)

// NewTreasuryAPIClient creates a new Treasury API client
func NewTreasuryAPIClient(log logger.Logger) *TreasuryAPIClient {
	// Create default HTTP client with circuit breaker configuration
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &TreasuryAPIClient{
		baseURL:    treasuryBaseURL,
		httpClient: httpClient,
		cache:      cache.NewExchangeRateCache(),
		logger:     log,
	}
}

// TreasuryResponse represents the response structure from the Treasury API
type TreasuryResponse struct {
	Data []struct {
		CountryName   string `json:"country"`
		CurrencyDesc  string `json:"currency"`
		ExchangeRate  string `json:"exchange_rate"`
		RecordDate    string `json:"record_date"`
		EffectiveDate string `json:"effective_date"`
	} `json:"data"`
	Meta struct {
		Count  int `json:"count"`
		Labels struct {
			CountryName   string `json:"country"`
			CurrencyDesc  string `json:"currency"`
			ExchangeRate  string `json:"exchange_rate"`
			RecordDate    string `json:"record_date"`
			EffectiveDate string `json:"effective_date"`
		} `json:"labels"`
	} `json:"meta"`
}

// FetchExchangeRate retrieves the exchange rate from the Treasury API
func (c *TreasuryAPIClient) FetchExchangeRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error) {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	// Log request details
	c.logger.Info("Fetching exchange rate", map[string]interface{}{
		"request_id": requestID,
		"currency":   currency,
		"date":       date.Format("2006-01-02"),
	})

	// Check cache first
	if cachedRate := c.cache.Get(currency, date); cachedRate != nil {
		c.logger.Info("Cache hit for exchange rate", map[string]interface{}{
			"request_id": requestID,
			"currency":   currency,
			"date":       date.Format("2006-01-02"),
			"rate":       cachedRate.Rate,
		})
		return cachedRate, nil
	}

	// Calculate the date 6 months before the purchase date
	sixMonthsAgo := date.AddDate(0, -6, 0)

	// Build request URL with appropriate filters based on the official API documentation
	reqURL := fmt.Sprintf("%s%s?filter=currency:eq:%s,record_date:lte:%s,record_date:gte:%s&sort=-record_date&limit=1",
		c.baseURL,
		exchangeRatePath,
		url.QueryEscape(currency),
		date.Format("2006-01-02"),
		sixMonthsAgo.Format("2006-01-02"))

	c.logger.Debug("Treasury API request URL", map[string]interface{}{
		"request_id": requestID,
		"url":        reqURL,
	})

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		c.logger.Error("Failed to create request", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Request-ID", fmt.Sprintf("%v", requestID))

	// Execute request with retry logic
	var resp *http.Response
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		startTime := time.Now()
		resp, err = c.httpClient.Do(req)
		duration := time.Since(startTime)

		// Log request metrics
		c.logger.Info("API request metrics", map[string]interface{}{
			"request_id":   requestID,
			"attempt":      attempt,
			"duration_ms":  duration.Milliseconds(),
			"success":      err == nil,
			"status_code":  resp != nil && err == nil,
			"api_endpoint": "treasury_exchange_rate",
		})

		if err == nil {
			break
		}

		if attempt < maxRetries {
			// Wait with exponential backoff before retrying
			backoffTime := time.Duration(attempt*attempt) * time.Second
			c.logger.Warn("Request failed, retrying", map[string]interface{}{
				"request_id":  requestID,
				"attempt":     attempt,
				"max_retries": maxRetries,
				"error":       err.Error(),
				"backoff_ms":  backoffTime.Milliseconds(),
			})
			time.Sleep(backoffTime)

			// Create a new request for the retry
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				c.logger.Error("Failed to create request for retry", map[string]interface{}{
					"request_id": requestID,
					"error":      err.Error(),
				})
				return nil, fmt.Errorf("failed to create request for retry: %w", err)
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("X-Request-ID", fmt.Sprintf("%v", requestID))
		}
	}

	if err != nil {
		c.logger.Error("Failed to execute request after multiple attempts", map[string]interface{}{
			"request_id":  requestID,
			"max_retries": maxRetries,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request after %d attempts: %w", maxRetries, err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			c.logger.Warn("Error closing response body", map[string]interface{}{
				"request_id": requestID,
				"error":      closeErr.Error(),
			})
		}
	}()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug("Treasury API response", map[string]interface{}{
		"request_id":  requestID,
		"status_code": resp.StatusCode,
		"body_size":   len(bodyBytes),
	})

	// Check response status
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("API returned error status", map[string]interface{}{
			"request_id":  requestID,
			"status_code": resp.StatusCode,
			"body":        string(bodyBytes),
		})
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode)
	}

	// Parse response
	var treasuryResp TreasuryResponse
	if err := json.Unmarshal(bodyBytes, &treasuryResp); err != nil {
		c.logger.Error("Failed to decode response", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if any data was returned
	if len(treasuryResp.Data) == 0 {
		c.logger.Warn("No exchange rate data available", map[string]interface{}{
			"request_id": requestID,
			"currency":   currency,
			"date":       date.Format("2006-01-02"),
			"date_from":  sixMonthsAgo.Format("2006-01-02"),
		})
		return nil, fmt.Errorf("no exchange rate available within 6 months of %s for currency %s",
			date.Format("2006-01-02"),
			currency)
	}

	// Parse the exchange rate and date
	rateData := treasuryResp.Data[0]

	c.logger.Debug("Rate data retrieved", map[string]interface{}{
		"request_id":     requestID,
		"currency":       rateData.CurrencyDesc,
		"country":        rateData.CountryName,
		"date":           rateData.RecordDate,
		"rate":           rateData.ExchangeRate,
		"effective_date": rateData.EffectiveDate,
	})

	// Parse rate with better error handling
	var rate float64
	if _, err := fmt.Sscanf(rateData.ExchangeRate, "%f", &rate); err != nil {
		c.logger.Error("Failed to parse exchange rate", map[string]interface{}{
			"request_id": requestID,
			"rate_value": rateData.ExchangeRate,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to parse exchange rate '%s': %w", rateData.ExchangeRate, err)
	}

	// Validate the rate is positive
	if rate <= 0 {
		c.logger.Error("Invalid exchange rate value", map[string]interface{}{
			"request_id": requestID,
			"rate":       rate,
		})
		return nil, fmt.Errorf("invalid exchange rate value: %f", rate)
	}

	// Parse date
	rateDate, err := time.Parse("2006-01-02", rateData.RecordDate)
	if err != nil {
		c.logger.Error("Failed to parse rate date", map[string]interface{}{
			"request_id": requestID,
			"date":       rateData.RecordDate,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to parse rate date '%s': %w", rateData.RecordDate, err)
	}

	// Double-check that the rate date is within the required 6-month window
	if rateDate.Before(sixMonthsAgo) || rateDate.After(date) {
		c.logger.Error("Exchange rate date outside allowed range", map[string]interface{}{
			"request_id":           requestID,
			"rate_date":            rateDate.Format("2006-01-02"),
			"transaction_date":     date.Format("2006-01-02"),
			"six_months_prior":     sixMonthsAgo.Format("2006-01-02"),
			"days_before_tx":       date.Sub(rateDate).Hours() / 24,
			"days_after_six_month": rateDate.Sub(sixMonthsAgo).Hours() / 24,
		})
		return nil, fmt.Errorf("exchange rate date %s is outside the allowed range (must be between %s and %s inclusive)",
			rateDate.Format("2006-01-02"),
			sixMonthsAgo.Format("2006-01-02"),
			date.Format("2006-01-02"))
	}

	// Create exchange rate entity
	exchangeRate := &entity.ExchangeRate{
		Currency: currency,
		Date:     rateDate,
		Rate:     rate,
	}

	// Store in cache
	c.cache.Put(exchangeRate, date)
	c.logger.Info("Cached exchange rate", map[string]interface{}{
		"request_id":       requestID,
		"currency":         currency,
		"transaction_date": date.Format("2006-01-02"),
		"rate_date":        rateDate.Format("2006-01-02"),
		"rate":             rate,
	})

	return exchangeRate, nil
}
