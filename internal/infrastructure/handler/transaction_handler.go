package handler

import (
	"encoding/json"
	"log"
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
		sendErrorResponse(w, "Invalid request body",
			"The request body could not be parsed as valid JSON", http.StatusBadRequest)
		return
	}

	// Validate description length
	if len(req.Description) > 50 {
		sendErrorResponse(w, "Description too long",
			"Description must not exceed 50 characters", http.StatusBadRequest)
		return
	}

	// Validate amount is positive
	if req.Amount <= 0 {
		sendErrorResponse(w, "Invalid amount",
			"Amount must be a positive value", http.StatusBadRequest)
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		sendErrorResponse(w, "Invalid date format",
			"Date must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

	// Don't allow future dates
	if date.After(time.Now()) {
		sendErrorResponse(w, "Future date not allowed",
			"Transaction date cannot be in the future", http.StatusBadRequest)
		return
	}

	// Call service
	id, err := h.service.CreateTransaction(r.Context(), req.Description, date, req.Amount)
	if err != nil {
		// Handle different types of errors
		switch {
		case strings.Contains(err.Error(), "description must not exceed"):
			sendErrorResponse(w, "Description too long",
				"Description must not exceed 50 characters", http.StatusBadRequest)
		case strings.Contains(err.Error(), "amount must be"):
			sendErrorResponse(w, "Invalid amount",
				"Amount must be a positive value", http.StatusBadRequest)
		default:
			log.Printf("Unexpected error in create transaction: %v", err)
			sendErrorResponse(w, "Internal server error",
				"An unexpected error occurred while creating the transaction",
				http.StatusInternalServerError)
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
			sendErrorResponse(w, "Transaction not found",
				"The requested transaction could not be found", http.StatusNotFound)
		} else {
			log.Printf("Unexpected error in get transaction: %v", err)
			sendErrorResponse(w, "Internal server error",
				"An unexpected error occurred while retrieving the transaction",
				http.StatusInternalServerError)
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
