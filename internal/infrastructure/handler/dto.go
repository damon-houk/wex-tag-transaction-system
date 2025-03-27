package handler

// CreateTransactionRequest represents the request body for creating a transaction
type CreateTransactionRequest struct {
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
}

// TransactionResponse represents the response for transaction endpoints
type TransactionResponse struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
}

// CreateTransactionResponse represents the response for the create transaction endpoint
type CreateTransactionResponse struct {
	ID string `json:"id"`
}
