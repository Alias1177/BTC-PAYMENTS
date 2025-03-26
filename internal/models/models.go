package models

import (
	"time"
)

// InvoiceRequest represents the request to create a new invoice
type InvoiceRequest struct {
	PriceAmount   float64 `json:"price_amount" validate:"required"`
	PriceCurrency string  `json:"price_currency" validate:"required"`
	OrderID       string  `json:"order_id" validate:"required"`
	BuyerEmail    string  `json:"buyer_email"`
	RedirectURL   string  `json:"redirect_url"`
	WebhookURL    string  `json:"webhook_url"`
}

// InvoiceResponse represents the response from creating an invoice
type InvoiceResponse struct {
	InvoiceID      string    `json:"invoice_id"`
	CheckoutURL    string    `json:"checkout_url"`
	Status         string    `json:"status"`
	ExpirationTime time.Time `json:"expiration_time"`
}

// InvoiceStatus represents the detailed status of an invoice
type InvoiceStatus struct {
	InvoiceID  string    `json:"invoice_id"`
	Status     string    `json:"status"`
	AmountPaid float64   `json:"amount_paid"`
	Currency   string    `json:"currency"`
	PaidDate   time.Time `json:"paid_date,omitempty"`
}

// Transaction represents a transaction record in the database

// WebhookEvent represents a webhook notification from BTCPay Server
type WebhookEvent struct {
	DeliveryID         string                 `json:"deliveryId"`
	WebhookID          string                 `json:"webhookId"`
	OriginalDeliveryID string                 `json:"originalDeliveryId,omitempty"`
	IsRedelivery       bool                   `json:"isRedelivery"`
	Type               string                 `json:"type"`
	Timestamp          int64                  `json:"timestamp"`
	StoreID            string                 `json:"storeId"`
	InvoiceID          string                 `json:"invoiceId"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}
