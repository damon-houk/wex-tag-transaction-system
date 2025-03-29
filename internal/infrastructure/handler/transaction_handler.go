package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/middleware"
	"github.com/gorilla/mux"
)

// TransactionHandler handles HTTP requests for transactions
type TransactionHandler struct {
	service *service.TransactionService
	logger  logger.Logger
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(service *service.TransactionService, log logger.Logger) *TransactionHandler {
	if log == nil {
		log = logger.GetDefaultLogger()
	}

	return &TransactionHandler{
		service: service,
		logger:  log,
	}
}

// CreateTransaction handles the creation of a new transaction
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	h.logger.Info("Handling create transaction request", map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
	})

	// Parse request body
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		sendErrorResponse(w, h.logger, "Invalid request body",
			"The request body could not be parsed as valid JSON", http.StatusBadRequest, requestID)
		return
	}

	h.logger.Debug("Request parsed", map[string]interface{}{
		"request_id":  requestID,
		"description": req.Description,
		"date":        req.Date,
		"amount":      req.Amount,
	})

	// Validate description length
	if len(req.Description) > 50 {
		h.logger.Warn("Description too long", map[string]interface{}{
			"request_id":  requestID,
			"description": req.Description,
			"length":      len(req.Description),
			"max_allowed": 50,
		})
		sendErrorResponse(w, h.logger, "Description too long",
			"Description must not exceed 50 characters", http.StatusBadRequest, requestID)
		return
	}

	// Validate amount is positive
	if req.Amount <= 0 {
		h.logger.Warn("Invalid amount", map[string]interface{}{
			"request_id": requestID,
			"amount":     req.Amount,
		})
		sendErrorResponse(w, h.logger, "Invalid amount",
			"Amount must be a positive value", http.StatusBadRequest, requestID)
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		h.logger.Warn("Invalid date format", map[string]interface{}{
			"request_id": requestID,
			"date":       req.Date,
			"error":      err.Error(),
		})
		sendErrorResponse(w, h.logger, "Invalid date format",
			"Date must be in YYYY-MM-DD format", http.StatusBadRequest, requestID)
		return
	}

	// Don't allow future dates
	if date.After(time.Now()) {
		h.logger.Warn("Future date not allowed", map[string]interface{}{
			"request_id": requestID,
			"date":       req.Date,
		})
		sendErrorResponse(w, h.logger, "Future date not allowed",
			"Transaction date cannot be in the future", http.StatusBadRequest, requestID)
		return
	}

	// Call service
	id, err := h.service.CreateTransaction(r.Context(), req.Description, date, req.Amount)
	if err != nil {
		// Handle different types of errors
		switch {
		case strings.Contains(err.Error(), "description must not exceed"):
			h.logger.Warn("Description validation failed", map[string]interface{}{
				"request_id": requestID,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Description too long",
				"Description must not exceed 50 characters", http.StatusBadRequest, requestID)
		case strings.Contains(err.Error(), "amount must be"):
			h.logger.Warn("Amount validation failed", map[string]interface{}{
				"request_id": requestID,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Invalid amount",
				"Amount must be a positive value", http.StatusBadRequest, requestID)
		default:
			h.logger.Error("Unexpected error in create transaction", map[string]interface{}{
				"request_id": requestID,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Internal server error",
				"An unexpected error occurred while creating the transaction",
				http.StatusInternalServerError, requestID)
		}
		return
	}

	h.logger.Info("Transaction created successfully", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
	})

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateTransactionResponse{ID: id})
}

// GetTransaction handles retrieving a transaction by ID
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	// Get ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	h.logger.Info("Handling get transaction request", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
	})

	// Call service
	tx, err := h.service.GetTransaction(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.logger.Warn("Transaction not found", map[string]interface{}{
				"request_id": requestID,
				"id":         id,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Transaction not found",
				"The requested transaction could not be found", http.StatusNotFound, requestID)
		} else {
			h.logger.Error("Unexpected error in get transaction", map[string]interface{}{
				"request_id": requestID,
				"id":         id,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Internal server error",
				"An unexpected error occurred while retrieving the transaction",
				http.StatusInternalServerError, requestID)
		}
		return
	}

	h.logger.Info("Transaction retrieved successfully", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
	})

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

	h.logger.Info("Transaction routes registered", map[string]interface{}{
		"routes": []string{
			"POST /transactions",
			"GET /transactions/{id}",
		},
	})
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(w http.ResponseWriter, log logger.Logger, message, description string, statusCode int, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Error:       message,
		Status:      statusCode,
		Description: description,
		RequestID:   requestID,
	}

	log.Debug("Sending error response", map[string]interface{}{
		"request_id":  requestID,
		"status_code": statusCode,
		"message":     message,
	})

	json.NewEncoder(w).Encode(resp)
}
