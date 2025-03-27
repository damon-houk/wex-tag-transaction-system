package handler

import (
	"encoding/json"
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
		http.Error(w, "Currency parameter is required", http.StatusBadRequest)
		return
	}

	// Call service
	convertedTx, err := h.service.GetTransactionInCurrency(r.Context(), id, currency)
	if err != nil {
		// Handle different types of errors
		switch {
		case strings.Contains(err.Error(), "not found"):
			http.Error(w, err.Error(), http.StatusNotFound)
		case strings.Contains(err.Error(), "no exchange rate available"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
