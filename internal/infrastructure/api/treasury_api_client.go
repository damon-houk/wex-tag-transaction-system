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
)

const (
	treasuryBaseURL  = "https://api.fiscaldata.treasury.gov/services/api/fiscal_service"
	exchangeRatePath = "/v1/accounting/od/rates_of_exchange"
)

// TreasuryAPIClient implements the Treasury API interface
type TreasuryAPIClient struct {
	baseURL    string
	httpClient *http.Client
	cache      *cache.ExchangeRateCache
}

// NewTreasuryAPIClient creates a new Treasury API client
func NewTreasuryAPIClient(httpClient *http.Client) *TreasuryAPIClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &TreasuryAPIClient{
		baseURL:    treasuryBaseURL,
		httpClient: httpClient,
		cache:      cache.NewExchangeRateCache(),
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

// GetExchangeRate retrieves the exchange rate for a currency on or before a specific date
func (c *TreasuryAPIClient) GetExchangeRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error) {
	// Check cache first
	if cachedRate := c.cache.Get(currency, date); cachedRate != nil {
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

	fmt.Printf("Treasury API request URL: %s\n", reqURL)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Accept header to ensure JSON response
	req.Header.Add("Accept", "application/json")

	// Execute request with retry logic
	var resp *http.Response
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}

		if attempt < maxRetries {
			// Wait with exponential backoff before retrying
			backoffTime := time.Duration(attempt*attempt) * time.Second
			fmt.Printf("Request failed (attempt %d/%d): %v. Retrying in %v...\n",
				attempt, maxRetries, err, backoffTime)
			time.Sleep(backoffTime)

			// Create a new request for the retry
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create request for retry: %w", err)
			}
			req.Header.Add("Accept", "application/json")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute request after %d attempts: %w", maxRetries, err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			// Log the error but don't override the original return error if there was one
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
	}()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	bodyString := string(bodyBytes)
	fmt.Printf("Treasury API response status: %d\n", resp.StatusCode)
	fmt.Printf("Treasury API response body: %s\n", bodyString)

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status: %d, body: %s", resp.StatusCode, bodyString)
	}

	// Parse response
	var treasuryResp TreasuryResponse
	if err := json.Unmarshal(bodyBytes, &treasuryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if any data was returned
	if len(treasuryResp.Data) == 0 {
		return nil, fmt.Errorf("no exchange rate available within 6 months of %s for currency %s",
			date.Format("2006-01-02"),
			currency)
	}

	// Parse the exchange rate and date
	rateData := treasuryResp.Data[0]

	// Debug the rate data
	fmt.Printf("Rate data: %+v\n", rateData)

	// Parse rate with better error handling
	var rate float64
	if _, err := fmt.Sscanf(rateData.ExchangeRate, "%f", &rate); err != nil {
		return nil, fmt.Errorf("failed to parse exchange rate '%s': %w", rateData.ExchangeRate, err)
	}

	// Validate the rate is positive
	if rate <= 0 {
		return nil, fmt.Errorf("invalid exchange rate value: %f", rate)
	}

	// Parse date
	rateDate, err := time.Parse("2006-01-02", rateData.RecordDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rate date '%s': %w", rateData.RecordDate, err)
	}

	// Create exchange rate entity
	exchangeRate := &entity.ExchangeRate{
		Currency: currency,
		Date:     rateDate,
		Rate:     rate,
	}

	// Store in cache
	c.cache.Put(exchangeRate, date)

	return exchangeRate, nil
}
