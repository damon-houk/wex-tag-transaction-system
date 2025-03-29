// cmd/server/main.go
package main

import (
	"os"
	"path/filepath"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/api"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/db"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/handler"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/middleware"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	// Setup structured logger
	jsonLogger := logger.NewJSONLogger(os.Stdout, logger.InfoLevel)
	jsonLogger.Info("Starting WEX TAG Transaction Processing System", map[string]interface{}{
		"version":   "1.0.0",
		"timestamp": "2025-03-29T12:00:00Z",
	})

	// Setup BadgerDB
	dbPath := filepath.Join(".", "data")
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		jsonLogger.Fatal("Failed to create database directory", map[string]interface{}{
			"error": err.Error(),
			"path":  dbPath,
		})
	}

	badgerOpts := badger.DefaultOptions(dbPath)
	badgerOpts.Logger = nil // Disable Badger's default logger

	jsonLogger.Info("Opening database", map[string]interface{}{
		"path": dbPath,
	})

	badgerDB, err := badger.Open(badgerOpts)
	if err != nil {
		jsonLogger.Fatal("Failed to open database", map[string]interface{}{
			"error": err.Error(),
			"path":  dbPath,
		})
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			jsonLogger.Error("Error closing BadgerDB", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Initialize repositories and services
	txRepo := db.NewBadgerTransactionRepository(badgerDB, jsonLogger)
	treasuryClient := api.NewTreasuryAPIClient(jsonLogger)
	exchangeRateRepo := db.NewTreasuryExchangeRateRepository(treasuryClient, jsonLogger)

	// Initialize services
	txService := service.NewTransactionService(txRepo, jsonLogger)
	conversionService := service.NewConversionService(txRepo, exchangeRateRepo, jsonLogger)

	// Initialize handlers
	txHandler := handler.NewTransactionHandler(txService, jsonLogger)
	conversionHandler := handler.NewConversionHandler(conversionService, jsonLogger)

	// Setup router
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware(jsonLogger))

	// Register routes
	txHandler.RegisterRoutes(router)
	conversionHandler.RegisterRoutes(router)

	// Add health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := ":" + port
	jsonLogger.Info("Server listening", map[string]interface{}{
		"address": serverAddr,
	})

	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		jsonLogger.Fatal("Server failed", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
