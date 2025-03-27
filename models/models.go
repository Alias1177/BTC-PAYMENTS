package models

type PaymentRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	InvoiceID string `json:"invoice_id" binding:"required"`
}

// PaymentResponse represents the data sent to frontend
type PaymentResponse struct {
	Status     bool        `json:"status"`
	Message    string      `json:"message,omitempty"`
	PaymentURL string      `json:"payment_url,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

// PaymentStatus represents payment status information
type PaymentStatus struct {
	Status   string  `json:"status"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}
