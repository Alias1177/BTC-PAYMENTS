package handler

import (
	"github.com/Alias1177/BTC-PAYMENTS/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

func (h *Handler) GetUserPaymentsHandler(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.PaymentResponse{
			Status:  false,
			Message: "user_id parameter is required",
		})
		return
	}

	rows, err := h.DB.Query(
		"SELECT invoice_id, status, amount, currency, created_at FROM user_payments WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, models.PaymentResponse{
			Status:  false,
			Message: "Database error",
		})
		return
	}
	defer rows.Close()

	var payments []map[string]interface{}

	for rows.Next() {
		var invoiceID, status, currency string
		var amount float64
		var createdAt time.Time

		if err := rows.Scan(&invoiceID, &status, &amount, &currency, &createdAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Используем новую сигнатуру функции с параметрами из config
		if status != "finished" && status != "failed" && status != "expired" {
			currentStatus, err := getLatestPaymentStatus(invoiceID, h.Config.APIBaseURL, h.Config.APIKEY)
			if err == nil && currentStatus != status {
				status = currentStatus
				h.DB.Exec(
					"UPDATE user_payments SET status = $1 WHERE user_id = $2 AND invoice_id = $3",
					status, userID, invoiceID,
				)
			}
		}

		payments = append(payments, map[string]interface{}{
			"invoice_id": invoiceID,
			"status":     status,
			"amount":     amount,
			"currency":   currency,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	// Добавляем возврат результата, который отсутствовал
	c.JSON(http.StatusOK, models.PaymentResponse{
		Status: true,
		Data:   payments,
	})
}
