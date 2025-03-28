package internal

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/db"
	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/mock"
)

// MockTreasuryAPI implements the TreasuryAPI interface for testing
type MockTreasuryAPI struct {
	mock.Mock
}

func (m *MockTreasuryAPI) GetExchangeRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error) {
	// Return predefined rates for common currencies
	rates := map[string]float64{
		"EUR": 0.85,
		"GBP": 0.75,
		"CAD": 1.25,
		"JPY": 110.0,
	}

	rate, ok := rates[currency]
	if !ok {
		return nil, fmt.Errorf("no exchange rate available for %s", currency)
	}

	return &entity.ExchangeRate{
		Currency: currency,
		Date:     date.AddDate(0, 0, -5), // 5 days before transaction date
		Rate:     rate,
	}, nil
}

func TestPerformance(t *testing.T) {
	// Skip in short mode or CI
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup test database
	dbPath, err := os.MkdirTemp("", "badger-perf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dbPath)

	badgerOpts := badger.DefaultOptions(dbPath).WithLogger(nil)
	badgerDB, err := badger.Open(badgerOpts)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer badgerDB.Close()

	// Initialize repositories and services
	txRepo := db.NewBadgerTransactionRepository(badgerDB)
	treasuryAPI := &MockTreasuryAPI{}
	txService := service.NewTransactionService(txRepo)
	conversionService := service.NewConversionService(txRepo, treasuryAPI)

	// Performance test configuration
	numTransactions := 100
	concurrency := 10

	// Preload test data
	t.Log("Preloading test data...")
	txIDs := preloadTestData(t, txService, numTransactions)

	// Test transaction creation performance
	t.Run("Transaction Creation", func(t *testing.T) {
		startTime := time.Now()

		wg := sync.WaitGroup{}
		wg.Add(concurrency)

		txPerWorker := numTransactions / concurrency

		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				defer wg.Done()

				ctx := context.Background()
				for j := 0; j < txPerWorker; j++ {
					desc := fmt.Sprintf("Test transaction %d-%d", workerID, j)
					amount := 100.0 + float64(rand.Intn(10000))/100.0
					date := time.Now().AddDate(0, 0, -rand.Intn(30))

					_, err := txService.CreateTransaction(ctx, desc, date, amount)
					if err != nil {
						t.Logf("Error creating transaction: %v", err)
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// Calculate throughput
		throughput := float64(numTransactions) / duration.Seconds()
		t.Logf("Transaction creation: %d transactions in %v (%.2f tx/sec)",
			numTransactions, duration, throughput)
	})

	// Test transaction retrieval performance
	t.Run("Transaction Retrieval", func(t *testing.T) {
		startTime := time.Now()

		wg := sync.WaitGroup{}
		wg.Add(concurrency)

		txPerWorker := numTransactions / concurrency

		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				defer wg.Done()

				ctx := context.Background()
				for j := 0; j < txPerWorker; j++ {
					idx := (workerID*txPerWorker + j) % len(txIDs)
					_, err := txService.GetTransaction(ctx, txIDs[idx])
					if err != nil {
						t.Logf("Error retrieving transaction: %v", err)
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// Calculate throughput
		throughput := float64(numTransactions) / duration.Seconds()
		t.Logf("Transaction retrieval: %d transactions in %v (%.2f tx/sec)",
			numTransactions, duration, throughput)
	})

	// Test currency conversion performance (using mock API to avoid rate limits)
	t.Run("Currency Conversion", func(t *testing.T) {
		startTime := time.Now()

		wg := sync.WaitGroup{}
		wg.Add(concurrency)

		txPerWorker := numTransactions / concurrency

		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				defer wg.Done()

				ctx := context.Background()
				currencies := []string{"EUR", "GBP", "CAD", "JPY"}

				for j := 0; j < txPerWorker; j++ {
					idx := (workerID*txPerWorker + j) % len(txIDs)
					currency := currencies[j%len(currencies)]

					_, err := conversionService.GetTransactionInCurrency(ctx, txIDs[idx], currency)
					if err != nil {
						t.Logf("Error converting transaction: %v", err)
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// Calculate throughput
		throughput := float64(numTransactions) / duration.Seconds()
		t.Logf("Currency conversion: %d conversions in %v (%.2f tx/sec)",
			numTransactions, duration, throughput)
	})
}

// preloadTestData creates test transactions and returns their IDs
func preloadTestData(t *testing.T, txService *service.TransactionService, count int) []string {
	ids := make([]string, count)
	ctx := context.Background()

	for i := 0; i < count; i++ {
		desc := fmt.Sprintf("Preloaded transaction %d", i)
		amount := 100.0 + float64(i)
		date := time.Now().AddDate(0, 0, -i%30)

		id, err := txService.CreateTransaction(ctx, desc, date, amount)
		if err != nil {
			t.Fatalf("Failed to preload test data: %v", err)
		}

		ids[i] = id
	}

	return ids
}
