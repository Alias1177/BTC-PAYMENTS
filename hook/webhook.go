package hook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type NowPaymentsWebhook struct {
	InvoiceID string  `json:"invoice_id"`
	Status    string  `json:"payment_status"`
	Amount    float64 `json:"actual_amount"`
	Currency  string  `json:"actual_currency"`
}

func NowPaymentsWebhookHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		signatureHeader := c.GetHeader("x-nowpayments-sig")
		if signatureHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No signature provided"})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Println("Error reading body:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		ipnSecretKey := os.Getenv("NP_WEBHOOK_SECRET")

		mac := hmac.New(sha512.New, []byte(ipnSecretKey))
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if signatureHeader != expectedSignature {
			log.Println("Invalid webhook signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var webhook NowPaymentsWebhook
		if err := c.ShouldBindJSON(&webhook); err != nil {
			log.Printf("Webhook bind error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		if webhook.Status == "finished" {
			_, err := db.Exec(
				"UPDATE user_payments SET status = $1 WHERE invoice_id = $2",
				"finished", webhook.InvoiceID,
			)

			if err != nil {
				log.Printf("Database update error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database update failed"})
				return
			}

			log.Printf("Invoice %s marked as finished", webhook.InvoiceID)
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
