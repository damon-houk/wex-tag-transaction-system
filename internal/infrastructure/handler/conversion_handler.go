// Package handler Package handler internal/infrastructure/handler/conversion_handler.go
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/damon-houk/wex-tag-transaction-system/internal/application/service"
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
}

// ConversionHandler handles HTTP requests for currency conversion
type ConversionHandler struct {
	service *service.ConversionService
}

// NewConversionHandler creates a new conversion handler
func NewConversionHandler(service *service.ConversionService) *ConversionHandler {
	return &ConversionHandler{service: service}
}

// ConvertTransaction handles retrieving a transaction with currency conversion
func (h *ConversionHandler) ConvertTransaction(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get currency from query parameter
	currency := r.URL.Query().Get("currency")
	if currency == "" {
		sendErrorResponse(w, "Missing currency parameter",
			"The 'currency' query parameter is required", http.StatusBadRequest)
		return
	}

	// Currency codes should be 3 characters
	if len(currency) != 3 {
		sendErrorResponse(w, "Invalid currency code",
			"Currency code should be 3 characters (e.g., EUR, GBP, CAD)", http.StatusBadRequest)
		return
	}

	// Call service
	convertedTx, err := h.service.GetTransactionInCurrency(r.Context(), id, currency)
	if err != nil {
		// Handle different types of errors
		switch {
		case strings.Contains(err.Error(), "not found"):
			sendErrorResponse(w, "Transaction not found",
				"The requested transaction could not be found", http.StatusNotFound)
		case strings.Contains(err.Error(), "no exchange rate available"):
			sendErrorResponse(w, "No exchange rate available",
				"No exchange rate is available within 6 months of the transaction date for the specified currency",
				http.StatusBadRequest)
		case strings.Contains(err.Error(), "exchange rate date") && strings.Contains(err.Error(), "outside the allowed range"):
			sendErrorResponse(w, "Exchange rate outside allowed range",
				"The available exchange rate is outside the 6-month window prior to the transaction date",
				http.StatusBadRequest)
		case strings.Contains(err.Error(), "failed to get exchange rate"):
			// Log the error for internal debugging
			log.Printf("Exchange rate error: %v", err)
			sendErrorResponse(w, "Exchange rate service unavailable",
				"Unable to retrieve exchange rate data. Please try again later.",
				http.StatusServiceUnavailable)
		case strings.Contains(err.Error(), "failed to execute request"):
			// Network or API connectivity issues
			log.Printf("API connectivity error: %v", err)
			sendErrorResponse(w, "Service temporarily unavailable",
				"The exchange rate service is temporarily unavailable. Please try again later.",
				http.StatusServiceUnavailable)
		default:
			// Log unexpected errors for investigation
			log.Printf("Unexpected error in conversion handler: %v", err)
			sendErrorResponse(w, "Internal server error",
				"An unexpected error occurred. Please try again later.",
				http.StatusInternalServerError)
		}
		return
	}

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
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(w http.ResponseWriter, message, description string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Error:       message,
		Status:      statusCode,
		Description: description,
	}

	json.NewEncoder(w).Encode(resp)
}
