package handler

import (
	"fmt"
	"github.com/Alias1177/BTC-PAYMENTS/config"
	"log"

	"github.com/Alias1177/BTC-PAYMENTS/models"
	"github.com/Alias1177/BTC-PAYMENTS/repo"
	"github.com/gin-gonic/gin"
	"net/http"
)

var invoiceAmounts = map[string]struct {
	amount   float64
	currency string
}{
	"6428428501": {13.9, "USD"},
	"5611069835": {64.9, "USD"},
	"6068449530": {1999.99, "USD"},
}

type Handler struct {
	DB     repo.Postgres
	Config *config.Config
}

func (h *Handler) AssignInvoiceHandler(c *gin.Context) {
	var request models.PaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.PaymentResponse{
			Status:  false,
			Message: "Invalid request format",
		})
		return
	}

	invoice, ok := invoiceAmounts[request.InvoiceID]
	if !ok {
		c.JSON(http.StatusBadRequest, models.PaymentResponse{
			Status:  false,
			Message: "Unknown invoice ID",
		})
		return
	}

	_, err := h.DB.Exec(
		"INSERT INTO user_payments (user_id, invoice_id, amount, currency) VALUES ($1, $2, $3, $4)",
		request.UserID, request.InvoiceID, invoice.amount, invoice.currency,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, models.PaymentResponse{
			Status:  false,
			Message: "Failed to save payment data",
		})
		return
	}

	c.JSON(http.StatusOK, models.PaymentResponse{
		Status:     true,
		PaymentURL: fmt.Sprintf("https://nowpayments.io/payment/?iid=%s", request.InvoiceID),
	})
}
