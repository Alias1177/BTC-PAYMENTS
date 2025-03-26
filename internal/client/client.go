package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BTCPayClient обеспечивает взаимодействие с API BTCPay Server
type BTCPayClient struct {
	baseURL    string
	apiKey     string
	storeID    string
	httpClient *http.Client
}

// NewBTCPayClient создает новый экземпляр клиента BTCPay Server
func NewBTCPayClient(baseURL, apiKey, storeID string) *BTCPayClient {
	return &BTCPayClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		storeID:    storeID,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// InvoiceRequest содержит данные для создания инвойса
type InvoiceRequest struct {
	PriceAmount   float64 `json:"price_amount"`
	PriceCurrency string  `json:"price_currency"`
	OrderID       string  `json:"order_id"`
	BuyerEmail    string  `json:"buyer_email,omitempty"`
	RedirectURL   string  `json:"redirect_url,omitempty"`
	WebhookURL    string  `json:"webhook_url,omitempty"`
}

// InvoiceResponse содержит ответ на создание инвойса
type InvoiceResponse struct {
	InvoiceID      string    `json:"invoice_id"`
	CheckoutURL    string    `json:"checkout_url"`
	Status         string    `json:"status"`
	ExpirationTime time.Time `json:"expiration_time"`
}

// InvoiceStatus содержит информацию о статусе инвойса
type InvoiceStatus struct {
	InvoiceID  string    `json:"invoice_id"`
	Status     string    `json:"status"`
	AmountPaid float64   `json:"amount_paid"`
	Currency   string    `json:"currency"`
	PaidDate   time.Time `json:"paid_date,omitempty"`
}

// CreateInvoice создает новый инвойс в BTCPay Server
func (c *BTCPayClient) CreateInvoice(invoiceReq InvoiceRequest) (*InvoiceResponse, error) {
	// Подготовка запроса для BTCPay Server
	btcpayReq := map[string]interface{}{
		"amount":   invoiceReq.PriceAmount,
		"currency": invoiceReq.PriceCurrency,
		"metadata": map[string]string{"orderId": invoiceReq.OrderID},
		"checkout": map[string]interface{}{"redirectURL": invoiceReq.RedirectURL},
		"receipt":  map[string]interface{}{"enabled": true},
	}

	if invoiceReq.BuyerEmail != "" {
		btcpayReq["buyer"] = map[string]string{"email": invoiceReq.BuyerEmail}
	}

	if invoiceReq.WebhookURL != "" {
		btcpayReq["notificationURL"] = invoiceReq.WebhookURL
	}

	// Конвертация в JSON
	payload, err := json.Marshal(btcpayReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	// Создание HTTP запроса
	url := fmt.Sprintf("%s/api/v1/stores/%s/invoices", c.baseURL, c.storeID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.apiKey))

	// Отправка запроса
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Чтение ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Обработка ошибок
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("BTCPay Server вернул ошибку: %d - %s", resp.StatusCode, string(body))
	}

	// Парсинг ответа
	var btcpayResp map[string]interface{}
	if err := json.Unmarshal(body, &btcpayResp); err != nil {
		return nil, fmt.Errorf("ошибка анмаршалинга ответа: %w", err)
	}

	// Извлечение информации для нашего ответа
	invoiceID, _ := btcpayResp["id"].(string)
	checkoutURL, _ := btcpayResp["checkoutLink"].(string)
	status, _ := btcpayResp["status"].(string)

	// Парсинг времени истечения
	var expirationTime time.Time
	if expiryStr, ok := btcpayResp["expirationTime"].(string); ok {
		expirationTime, err = time.Parse(time.RFC3339, expiryStr)
		if err != nil {
			return nil, fmt.Errorf("ошибка во время парсинга времени истечение")
		}
	}

	// Создание ответа
	invoiceResponse := &InvoiceResponse{
		InvoiceID:      invoiceID,
		CheckoutURL:    checkoutURL,
		Status:         status,
		ExpirationTime: expirationTime,
	}

	return invoiceResponse, nil
}

// GetInvoice получает информацию о существующем инвойсе
func (c *BTCPayClient) GetInvoice(invoiceID string) (*InvoiceStatus, error) {
	// Создание HTTP запроса
	url := fmt.Sprintf("%s/api/v1/stores/%s/invoices/%s", c.baseURL, c.storeID, invoiceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Установка заголовков
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.apiKey))

	// Отправка запроса
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Чтение ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Обработка ошибок
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("BTCPay Server вернул ошибку: %d - %s", resp.StatusCode, string(body))
	}

	// Парсинг ответа
	var btcpayResp map[string]interface{}
	if err := json.Unmarshal(body, &btcpayResp); err != nil {
		return nil, fmt.Errorf("ошибка анмаршалинга ответа: %w", err)
	}

	// Извлечение данных
	status, _ := btcpayResp["status"].(string)

	// Получение информации о сумме оплаты
	amountPaid := 0.0
	currency := ""

	if payments, ok := btcpayResp["payments"].([]interface{}); ok && len(payments) > 0 {
		payment, _ := payments[0].(map[string]interface{})
		amountPaid, _ = payment["value"].(float64)
		currency, _ = payment["currency"].(string)
	}

	// Получение даты оплаты
	var paidDate time.Time
	if paidDateStr, ok := btcpayResp["paidDate"].(string); ok {
		paidDate, err = time.Parse(time.RFC3339, paidDateStr)
		if err != nil {
			return nil, fmt.Errorf("ошибка во время получение даты оплаты")
		}
	}

	// Создание ответа
	invoiceStatus := &InvoiceStatus{
		InvoiceID:  invoiceID,
		Status:     status,
		AmountPaid: amountPaid,
		Currency:   currency,
		PaidDate:   paidDate,
	}

	return invoiceStatus, nil
}
