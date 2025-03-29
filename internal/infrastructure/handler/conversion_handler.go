// Package handler internal/infrastructure/handler/conversion_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/logger"
	"github.com/damon-houk/wex-tag-transaction-system/internal/infrastructure/middleware"
	"github.com/gorilla/mux"
)

// ConvertedTransactionResponse represents the response for the conversion endpoint
type ConvertedTransactionResponse struct {
	ID              string  `json:"id"`
	Description     string  `json:"description"`
	Date            string  `json:"date"`
	OriginalAmount  float64 `json:"original_amount"`
	Currency        string  `json:"currency"`
	ExchangeRate    float64 `json:"exchange_rate"`
	ConvertedAmount float64 `json:"converted_amount"`
	RateDate        string  `json:"rate_date"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error       string `json:"error"`
	Status      int    `json:"status"`
	Description string `json:"description,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
}

// ConversionHandler handles HTTP requests for currency conversion
type ConversionHandler struct {
	service *service.ConversionService
	logger  logger.Logger
}

// NewConversionHandler creates a new conversion handler
func NewConversionHandler(service *service.ConversionService, log logger.Logger) *ConversionHandler {
	if log == nil {
		log = logger.GetDefaultLogger()
	}

	return &ConversionHandler{
		service: service,
		logger:  log,
	}
}

// ConvertTransaction handles retrieving a transaction with currency conversion
func (h *ConversionHandler) ConvertTransaction(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	// Get ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	h.logger.Info("Handling convert transaction request", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
	})

	// Get currency from query parameter
	currency := r.URL.Query().Get("currency")
	if currency == "" {
		h.logger.Warn("Missing currency parameter", map[string]interface{}{
			"request_id": requestID,
			"id":         id,
		})
		sendErrorResponse(w, h.logger, "Missing currency parameter",
			"The 'currency' query parameter is required", http.StatusBadRequest, requestID)
		return
	}

	h.logger.Debug("Currency parameter", map[string]interface{}{
		"request_id": requestID,
		"id":         id,
		"currency":   currency,
	})

	// Currency codes should be 3 characters
	if len(currency) != 3 {
		h.logger.Warn("Invalid currency code", map[string]interface{}{
			"request_id": requestID,
			"id":         id,
			"currency":   currency,
			"length":     len(currency),
		})
		sendErrorResponse(w, h.logger, "Invalid currency code",
			"Currency code should be 3 characters (e.g., EUR, GBP, CAD)", http.StatusBadRequest, requestID)
		return
	}

	// Call service
	convertedTx, err := h.service.GetTransactionInCurrency(r.Context(), id, currency)
	if err != nil {
		// Handle different types of errors
		switch {
		case strings.Contains(err.Error(), "not found"):
			h.logger.Warn("Transaction not found", map[string]interface{}{
				"request_id": requestID,
				"id":         id,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Transaction not found",
				"The requested transaction could not be found", http.StatusNotFound, requestID)
		case strings.Contains(err.Error(), "no exchange rate available"):
			h.logger.Warn("No exchange rate available", map[string]interface{}{
				"request_id": requestID,
				"id":         id,
				"currency":   currency,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "No exchange rate available",
				"No exchange rate is available within 6 months of the transaction date for the specified currency",
				http.StatusBadRequest, requestID)
		case strings.Contains(err.Error(), "exchange rate date") && strings.Contains(err.Error(), "outside the allowed range"):
			h.logger.Warn("Exchange rate outside allowed range", map[string]interface{}{
				"request_id": requestID,
				"id":         id,
				"currency":   currency,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Exchange rate outside allowed range",
				"The available exchange rate is outside the 6-month window prior to the transaction date",
				http.StatusBadRequest, requestID)
		case strings.Contains(err.Error(), "failed to get exchange rate"):
			// Log the error for internal debugging
			h.logger.Error("Exchange rate service error", map[string]interface{}{
				"request_id": requestID,
				"id":         id,
				"currency":   currency,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Exchange rate service unavailable",
				"Unable to retrieve exchange rate data. Please try again later.",
				http.StatusServiceUnavailable, requestID)
		case strings.Contains(err.Error(), "failed to execute request"):
			// Network or API connectivity issues
			h.logger.Error("API connectivity error", map[string]interface{}{
				"request_id": requestID,
				"id":         id,
				"currency":   currency,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Service temporarily unavailable",
				"The exchange rate service is temporarily unavailable. Please try again later.",
				http.StatusServiceUnavailable, requestID)
		default:
			// Log unexpected errors for investigation
			h.logger.Error("Unexpected error in conversion handler", map[string]interface{}{
				"request_id": requestID,
				"id":         id,
				"currency":   currency,
				"error":      err.Error(),
			})
			sendErrorResponse(w, h.logger, "Internal server error",
				"An unexpected error occurred. Please try again later.",
				http.StatusInternalServerError, requestID)
		}
		return
	}

	h.logger.Info("Transaction converted successfully", map[string]interface{}{
		"request_id":       requestID,
		"id":               id,
		"currency":         currency,
		"original_amount":  convertedTx.OriginalAmount,
		"exchange_rate":    convertedTx.ExchangeRate,
		"converted_amount": convertedTx.ConvertedAmount,
	})

	// Create response
	resp := ConvertedTransactionResponse{
		ID:              convertedTx.ID,
		Description:     convertedTx.Description,
		Date:            convertedTx.Date.Format("2006-01-02"),
		OriginalAmount:  convertedTx.OriginalAmount,
		Currency:        convertedTx.Currency,
		ExchangeRate:    convertedTx.ExchangeRate,
		ConvertedAmount: convertedTx.ConvertedAmount,
		RateDate:        convertedTx.RateDate.Format("2006-01-02"),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RegisterRoutes registers the conversion handler routes
func (h *ConversionHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/transactions/{id}/convert", h.ConvertTransaction).Methods("GET")

	h.logger.Info("Conversion routes registered", map[string]interface{}{
		"routes": []string{
			"GET /transactions/{id}/convert",
		},
	})
}
