// internal/infrastructure/handler/integration_test.go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/db"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/handler"
	"github.com/damon-houk/wex-tag-transaction-system/internal/mocks"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupTestServer creates a test server with mocked dependencies
func setupTestServer(exchangeRateRepo *mocks.MockExchangeRateRepository) (*httptest.Server, *badger.DB, func(), error) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "badger-test")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Open BadgerDB with options for testing
	badgerOpts := badger.DefaultOptions(tempDir)
	badgerOpts.Logger = nil       // Disable logging
	badgerOpts.SyncWrites = false // Improve performance for tests

	badgerDB, err := badger.Open(badgerOpts)
	if err != nil {
		os.RemoveAll(tempDir) // Clean up the directory if DB fails to open
		return nil, nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create repository and services
	txRepo := db.NewBadgerTransactionRepository(badgerDB)
	txService := service.NewTransactionService(txRepo)
	conversionService := service.NewConversionService(txRepo, exchangeRateRepo)

	// Create handlers
	txHandler := handler.NewTransactionHandler(txService)
	conversionHandler := handler.NewConversionHandler(conversionService)

	// Setup router
	router := mux.NewRouter()
	txHandler.RegisterRoutes(router)
	conversionHandler.RegisterRoutes(router)

	// Create test server
	server := httptest.NewServer(router)

	// Return cleanup function
	cleanup := func() {
		server.Close()
		badgerDB.Close()
		os.RemoveAll(tempDir)
	}

	return server, badgerDB, cleanup, nil
}

func TestTransactionCreationAndRetrieval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup mock exchange rate repository
	mockExchangeRateRepo := new(mocks.MockExchangeRateRepository)

	// Setup test server
	server, _, cleanup, err := setupTestServer(mockExchangeRateRepo)
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}
	defer cleanup()

	// Define test transaction
	transactionJSON := `{
		"description": "Test transaction",
		"date": "2023-04-15",
		"amount": 123.45
	}`

	// Step 1: Create a transaction
	resp, err := http.Post(
		server.URL+"/transactions",
		"application/json",
		bytes.NewBufferString(transactionJSON),
	)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Parse response to get the transaction ID
	var createResp handler.CreateTransactionResponse
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	assert.NotEmpty(t, createResp.ID, "Expected a transaction ID")

	// Step 2: Retrieve the transaction
	resp2, err := http.Get(server.URL + "/transactions/" + createResp.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve transaction: %v", err)
	}
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// Parse response
	var txResp handler.TransactionResponse
	err = json.NewDecoder(resp2.Body).Decode(&txResp)
	if err != nil {
		t.Fatalf("Failed to decode transaction response: %v", err)
	}

	// Verify transaction data
	assert.Equal(t, createResp.ID, txResp.ID)
	assert.Equal(t, "Test transaction", txResp.Description)
	assert.Equal(t, "2023-04-15", txResp.Date)
	assert.Equal(t, 123.45, txResp.Amount)
}

func TestCurrencyConversion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup mock exchange rate repository
	mockExchangeRateRepo := new(mocks.MockExchangeRateRepository)

	// Setup test server
	server, badgerDB, cleanup, err := setupTestServer(mockExchangeRateRepo)
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}
	defer cleanup()

	// Insert a test transaction directly into the database
	txRepo := db.NewBadgerTransactionRepository(badgerDB)
	testDate, err := time.Parse("2006-01-02", "2023-04-15")
	if err != nil {
		t.Fatalf("Failed to parse test date: %v", err)
	}

	testTx := &entity.Transaction{
		ID:          "test-transaction-id",
		Description: "Test transaction",
		Date:        testDate,
		Amount:      123.45,
	}
	_, err = txRepo.Store(context.Background(), testTx)
	assert.NoError(t, err, "Failed to store test transaction")

	// Setup mock for exchange rate
	mockRate := &entity.ExchangeRate{
		Currency: "EUR",
		Date:     testDate.AddDate(0, 0, -5), // 5 days before the transaction
		Rate:     0.85,
	}
	mockExchangeRateRepo.On("FindRate", mock.Anything, "EUR", testDate).Return(mockRate, nil)

	// Request currency conversion
	resp, err := http.Get(server.URL + "/transactions/test-transaction-id/convert?currency=EUR")
	if err != nil {
		t.Fatalf("Failed to get transaction with conversion: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var convResp handler.ConvertedTransactionResponse
	err = json.NewDecoder(resp.Body).Decode(&convResp)
	if err != nil {
		t.Fatalf("Failed to decode conversion response: %v", err)
	}

	// Verify conversion data
	assert.Equal(t, "test-transaction-id", convResp.ID)
	assert.Equal(t, "Test transaction", convResp.Description)
	assert.Equal(t, "2023-04-15", convResp.Date)
	assert.Equal(t, 123.45, convResp.OriginalAmount)
	assert.Equal(t, "EUR", convResp.Currency)
	assert.Equal(t, 0.85, convResp.ExchangeRate)
	assert.Equal(t, 104.93, convResp.ConvertedAmount) // 123.45 * 0.85 = 104.9325, rounded to 104.93

	// Verify mock was called
	mockExchangeRateRepo.AssertExpectations(t)
}

func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup mock exchange rate repository
	mockExchangeRateRepo := new(mocks.MockExchangeRateRepository)

	// Setup test server
	server, badgerDB, cleanup, err := setupTestServer(mockExchangeRateRepo)
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}
	defer cleanup()

	t.Run("Invalid transaction date", func(t *testing.T) {
		invalidJSON := `{
			"description": "Test transaction",
			"date": "invalid-date",
			"amount": 123.45
		}`

		resp, err := http.Post(
			server.URL+"/transactions",
			"application/json",
			bytes.NewBufferString(invalidJSON),
		)
		if err != nil {
			t.Fatalf("Failed to send invalid date request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Description too long", func(t *testing.T) {
		invalidJSON := `{
			"description": "This description is way too long and exceeds the 50 character limit set by the requirements",
			"date": "2023-04-15",
			"amount": 123.45
		}`

		resp, err := http.Post(
			server.URL+"/transactions",
			"application/json",
			bytes.NewBufferString(invalidJSON),
		)
		if err != nil {
			t.Fatalf("Failed to send long description request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Negative amount", func(t *testing.T) {
		invalidJSON := `{
			"description": "Test transaction",
			"date": "2023-04-15",
			"amount": -123.45
		}`

		resp, err := http.Post(
			server.URL+"/transactions",
			"application/json",
			bytes.NewBufferString(invalidJSON),
		)
		if err != nil {
			t.Fatalf("Failed to send negative amount request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Future date", func(t *testing.T) {
		futureDate := time.Now().AddDate(1, 0, 0) // 1 year in the future

		invalidJSON := fmt.Sprintf(`{
			"description": "Test transaction",
			"date": "%s",
			"amount": 123.45
		}`, futureDate.Format("2006-01-02"))

		resp, err := http.Post(
			server.URL+"/transactions",
			"application/json",
			bytes.NewBufferString(invalidJSON),
		)
		if err != nil {
			t.Fatalf("Failed to send future date request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Transaction not found", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/transactions/non-existent-id")
		if err != nil {
			t.Fatalf("Failed to send non-existent ID request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Missing currency parameter", func(t *testing.T) {
		// First, create a test transaction in the database
		testDate := time.Now().AddDate(0, 0, -30) // 30 days ago
		testTx := &entity.Transaction{
			ID:          "missing-currency-test-id",
			Description: "Test transaction",
			Date:        testDate,
			Amount:      123.45,
		}

		txRepo := db.NewBadgerTransactionRepository(badgerDB)
		_, err := txRepo.Store(context.Background(), testTx)
		assert.NoError(t, err, "Failed to store test transaction")

		// Test missing currency parameter
		resp, err := http.Get(server.URL + "/transactions/missing-currency-test-id/convert")
		if err != nil {
			t.Fatalf("Failed to send missing currency request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Verify error response contains appropriate message
		var errorResp handler.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		assert.NoError(t, err, "Failed to decode error response")
		assert.Contains(t, errorResp.Error, "Missing currency parameter")
	})

	t.Run("Invalid currency code", func(t *testing.T) {
		// Test with invalid currency code (too short)
		resp, err := http.Get(server.URL + "/transactions/any-id/convert?currency=E")
		if err != nil {
			t.Fatalf("Failed to send invalid currency request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Test with invalid currency code (too long)
		resp, err = http.Get(server.URL + "/transactions/any-id/convert?currency=EURO")
		if err != nil {
			t.Fatalf("Failed to send invalid currency request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("No exchange rate available", func(t *testing.T) {
		// Setup a transaction in the database
		testDate := time.Now().AddDate(0, 0, -30) // 30 days ago
		testTx := &entity.Transaction{
			ID:          "no-rate-test-id",
			Description: "Test transaction",
			Date:        testDate,
			Amount:      123.45,
		}

		txRepo := db.NewBadgerTransactionRepository(badgerDB)
		_, err := txRepo.Store(context.Background(), testTx)
		assert.NoError(t, err, "Failed to store test transaction")

		// Use mock.Anything for the context and time parameters to avoid timezone issues
		mockExchangeRateRepo.On("FindRate", mock.Anything, "XYZ", mock.Anything).
			Return(nil, fmt.Errorf("no exchange rate available within 6 months of %s for currency XYZ",
				testDate.Format("2006-01-02"))).Once()

		// Test conversion with a currency that has no rate
		resp, err := http.Get(server.URL + "/transactions/no-rate-test-id/convert?currency=XYZ")
		if err != nil {
			t.Fatalf("Failed to send no exchange rate request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Verify error response contains appropriate message
		var errorResp handler.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		assert.NoError(t, err, "Failed to decode error response")
		assert.Contains(t, errorResp.Error, "No exchange rate available")
	})
}
