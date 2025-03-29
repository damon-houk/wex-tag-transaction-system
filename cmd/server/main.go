// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/api"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/db"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/handler"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting WEX TAG Transaction Processing System")

	// Setup BadgerDB
	dbPath := filepath.Join(".", "data")
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	badgerOpts := badger.DefaultOptions(dbPath)
	badgerOpts.Logger = nil // Disable Badger's default logger

	badgerDB, err := badger.Open(badgerOpts)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			log.Printf("Error closing BadgerDB: %v", err)
		}
	}()

	// Create loggers
	treasuryLogger := log.New(os.Stdout, "[TREASURY API] ", log.LstdFlags)

	// Initialize repositories and services
	txRepo := db.NewBadgerTransactionRepository(badgerDB)
	treasuryClient := api.NewTreasuryAPIClient(treasuryLogger)
	exchangeRateRepo := db.NewTreasuryExchangeRateRepository(treasuryClient)

	// Initialize services
	txService := service.NewTransactionService(txRepo)
	conversionService := service.NewConversionService(txRepo, exchangeRateRepo)

	// Initialize handlers
	txHandler := handler.NewTransactionHandler(txService)
	conversionHandler := handler.NewConversionHandler(conversionService)

	// Setup router
	router := mux.NewRouter()
	txHandler.RegisterRoutes(router)
	conversionHandler.RegisterRoutes(router)

	// Add middleware for logging
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Start server
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
