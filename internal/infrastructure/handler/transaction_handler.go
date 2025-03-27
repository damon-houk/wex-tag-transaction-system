package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
	"github.com/gorilla/mux"
)

// TransactionHandler handles HTTP requests for transactions
type TransactionHandler struct {
	service *service.TransactionService
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(service *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// CreateTransaction handles the creation of a new transaction
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Don't allow future dates
	if date.After(time.Now()) {
		http.Error(w, "Transaction date cannot be in the future", http.StatusBadRequest)
		return
	}

	// Call service
	id, err := h.service.CreateTransaction(r.Context(), req.Description, date, req.Amount)
	if err != nil {
		// Handle different types of errors
		switch {
		case strings.Contains(err.Error(), "description must not exceed"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case strings.Contains(err.Error(), "amount must be"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateTransactionResponse{ID: id})
}

// GetTransaction handles retrieving a transaction by ID
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Call service
	tx, err := h.service.GetTransaction(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Create response
	resp := TransactionResponse{
		ID:          tx.ID,
		Description: tx.Description,
		Date:        tx.Date.Format("2006-01-02"),
		Amount:      tx.Amount,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RegisterRoutes registers the transaction handler routes
func (h *TransactionHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/transactions", h.CreateTransaction).Methods("POST")
	router.HandleFunc("/transactions/{id}", h.GetTransaction).Methods("GET")
}
