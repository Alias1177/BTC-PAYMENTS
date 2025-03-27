package handler

import (
	"database/sql"
	"errors"
	"github.com/Alias1177/BTC-PAYMENTS/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (h *Handler) CheckPaymentHandler(c *gin.Context) {
	userID := c.Query("user_id")
	invoiceID := c.Query("invoice_id")

	if userID == "" || invoiceID == "" {
		c.JSON(http.StatusBadRequest, models.PaymentResponse{
			Status:  false,
			Message: "Both user_id and invoice_id parameters are required",
		})
		return
	}

	var amount float64
	var currency string
	var status string

	err := h.DB.QueryRow(
		"SELECT amount, currency, status FROM user_payments WHERE user_id = $1 AND invoice_id = $2",
		userID, invoiceID,
	).Scan(&amount, &currency, &status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.PaymentResponse{
				Status:  false,
				Message: "Payment not found",
			})
		} else {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, models.PaymentResponse{
				Status:  false,
				Message: "Database error",
			})
		}
		return
	}

	// Получаем обновленный статус
	currentStatus, err := getLatestPaymentStatus(invoiceID, h.Config.APIBaseURL, h.Config.APIKEY)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PaymentResponse{
			Status:  false,
			Message: "Ошибка получения статуса платежа",
		})
		return
	}

	// Если статус изменился, обновляем в БД
	if currentStatus != status {
		status = currentStatus
		_, err := h.DB.Exec(
			"UPDATE user_payments SET status = $1 WHERE user_id = $2 AND invoice_id = $3",
			status, userID, invoiceID,
		)
		if err != nil {
			log.Printf("Error updating payment status: %v", err)
		}
	}

	// Используем структуру PaymentStatus из models
	c.JSON(http.StatusOK, models.PaymentResponse{
		Status: true,
		Data: models.PaymentStatus{
			Status:   status,
			Amount:   amount,
			Currency: currency,
		},
	})
}
