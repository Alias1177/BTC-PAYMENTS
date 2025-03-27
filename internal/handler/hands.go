package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func getLatestPaymentStatus(invoiceID, APIBaseURL, APIKey string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/payment/by-invoice-id/%s", APIBaseURL, invoiceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "pending", nil
	}

	var result struct {
		Data []struct {
			PaymentStatus string `json:"payment_status"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "pending", nil
	}

	return result.Data[len(result.Data)-1].PaymentStatus, nil
}
