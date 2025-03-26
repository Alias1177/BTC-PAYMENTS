package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Transaction представляет запись о транзакции в базе данных
type Transaction struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	InvoiceID       string             `bson:"invoice_id" json:"invoice_id"`
	OrderID         string             `bson:"order_id" json:"order_id"`
	Status          string             `bson:"status" json:"status"`
	PriceAmount     float64            `bson:"price_amount" json:"price_amount"`
	PriceCurrency   string             `bson:"price_currency" json:"price_currency"`
	AmountPaid      float64            `bson:"amount_paid" json:"amount_paid"`
	PaymentCurrency string             `bson:"payment_currency" json:"payment_currency"`
	BuyerEmail      string             `bson:"buyer_email" json:"buyer_email"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	PaidAt          time.Time          `bson:"paid_at,omitempty" json:"paid_at,omitempty"`
}

// Repository определяет интерфейс для операций с хранилищем данных
type Repository interface {
	// CreateTransaction создает новую запись о транзакции
	CreateTransaction(tx *Transaction) error

	// UpdateTransactionStatus обновляет статус транзакции по ID инвойса
	UpdateTransactionStatus(invoiceID, status string) error

	// GetTransactionByInvoiceID получает транзакцию по ID инвойса
	GetTransactionByInvoiceID(invoiceID string) (*Transaction, error)

	// ListTransactions получает список транзакций с фильтрацией и пагинацией
	ListTransactions(filters map[string]interface{}, page, perPage int) ([]*Transaction, int, error)
}
