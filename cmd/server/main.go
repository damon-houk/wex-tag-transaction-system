package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
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
	defer badgerDB.Close()

	// Initialize repositories
	txRepo := db.NewBadgerTransactionRepository(badgerDB)

	// Initialize services
	txService := service.NewTransactionService(txRepo)

	// Initialize handlers
	txHandler := handler.NewTransactionHandler(txService)

	// Setup router
	router := mux.NewRouter()
	txHandler.RegisterRoutes(router)

	// Start server
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
