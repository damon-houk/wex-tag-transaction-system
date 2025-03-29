// internal/performance_test.go
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
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/db"
	"github.com/dgraph-io/badger/v3"
)

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
	txService := service.NewTransactionService(txRepo)

	// Performance test configuration
	numTransactions := 100
	concurrency := 10

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
}
