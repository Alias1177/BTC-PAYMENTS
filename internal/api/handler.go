package api

import (
	"bytes"
	"chi/BTC-PAYMENTS/internal/client"
	"chi/BTC-PAYMENTS/internal/models"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CreateInvoice обрабатывает создание нового счета
// @Summary Создание нового счета на оплату
// @Description Создает новый счет на оплату в BTCPay Server
// @Tags invoices
// @Accept json
// @Produce json
// @Param request body client.InvoiceRequest true "Параметры счета на оплату"
// @Success 201 {object} client.InvoiceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/invoices [post]
func (h *Handler) CreateInvoice(c *gin.Context) {
	var req client.InvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Ошибка валидации запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка обязательных полей
	if req.PriceAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Сумма должна быть больше нуля"})
		return
	}

	if req.PriceCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Валюта не указана"})
		return
	}

	if req.OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID заказа не указан"})
		return
	}

	h.logger.Info("Создание инвойса для заказа %s на сумму %f %s",
		req.OrderID,
		req.PriceAmount,
		req.PriceCurrency,
	)

	// Создание инвойса в BTCPay Server
	invoice, err := h.btcpayClient.CreateInvoice(req)
	if err != nil {
		h.logger.Error("Ошибка создания инвойса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать счет на оплату"})
		return
	}

	// Сохранение в базу данных ПОСЛЕ успешного создания инвойса
	if h.storage != nil {
		tx := &models.Transaction{
			InvoiceID:     invoice.InvoiceID, // Используем поле из полученного ответа
			OrderID:       req.OrderID,       // Берем из запроса на создание инвойса
			Status:        invoice.Status,    // Статус из ответа BTCPay
			PriceAmount:   req.PriceAmount,   // Сумма из запроса
			PriceCurrency: req.PriceCurrency, // Валюта из запроса
			BuyerEmail:    req.BuyerEmail,    // Email из запроса (если есть)
			CreatedAt:     time.Now(),        // Текущее время
			UpdatedAt:     time.Now(),        // Текущее время
		}

		if err := h.storage.CreateTransaction(tx); err != nil {
			h.logger.Error("Ошибка сохранения транзакции: %v", err)
			// Продолжаем выполнение, так как инвойс уже создан в BTCPay
		} else {
			h.logger.Info("Транзакция успешно сохранена в БД: %s", invoice.InvoiceID)
		}
	}

	c.JSON(http.StatusCreated, invoice)
}

// GetInvoice обрабатывает получение информации о счете
// @Summary Получение информации о счете
// @Description Получает текущий статус счета на оплату из BTCPay Server
// @Tags invoices
// @Produce json
// @Param id path string true "ID счета"
// @Success 200 {object} client.InvoiceStatus
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/invoices/{id} [get]
func (h *Handler) GetInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID счета не указан"})
		return
	}

	h.logger.Info("Получение информации о счете %s", invoiceID)

	// Получение инвойса из BTCPay Server
	invoice, err := h.btcpayClient.GetInvoice(invoiceID)
	if err != nil {
		h.logger.Error("Ошибка получения инвойса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить информацию о счете"})
		return
	}

	// Если инвойс не найден (обработка конкретных случаев в зависимости от ответа BTCPay Server)
	if invoice.Status == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Счет не найден"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func verifyWebhookSignature(signature, payload, secret string) bool {
	// Создание HMAC с использованием SHA-256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Сравнение вычисленной подписи с полученной
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// HandleWebhook обрабатывает webhook-уведомления от BTCPay Server
// @Summary Обработка уведомлений от BTCPay Server
// @Description Принимает и обрабатывает webhook-уведомления о изменении статуса счетов
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/webhooks/btcpay [post]
func (h *Handler) HandleWebhook(c *gin.Context) {
	// Получение сигнатуры для верификации
	signature := c.GetHeader("BTCPay-Sig")

	// Чтение тела запроса (выполняем только один раз)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Ошибка чтения тела webhook-запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка чтения запроса"})
		return
	}

	// Восстановление тела для последующего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Проверка сигнатуры (если применимо)
	if signature != "" && h.webhookSecret != "" {
		isValid := verifyWebhookSignature(signature, string(body), h.webhookSecret)
		if !isValid {
			h.logger.Warn("Недействительная подпись webhook: %s", signature)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительная подпись"})
			return
		}
	}

	// Парсинг тела запроса
	var event models.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		h.logger.Error("Ошибка парсинга webhook-события: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Остальной код обработки события
	h.logger.Info("Получено webhook-событие типа '%s' для инвойса %s",
		event.Type, event.InvoiceID)

	// Обработка события в зависимости от типа
	switch event.Type {
	case "InvoiceCreated":
		// Обычно не требует действий, так как инвойс уже создан через API
		break

	case "InvoiceReceivedPayment":
		// Обновление информации о платеже
		if h.storage != nil {
			// Получаем актуальные данные из BTCPay Server
			invoice, err := h.btcpayClient.GetInvoice(event.InvoiceID)
			if err != nil {
				h.logger.Error("Ошибка получения информации об инвойсе: %v", err)
				break
			}

			// Обновляем информацию о платеже в БД
			err = h.storage.UpdateTransactionPaymentInfo(
				event.InvoiceID,
				invoice.AmountPaid,
				invoice.Currency,
			)
			if err != nil {
				h.logger.Error("Ошибка обновления информации о платеже: %v", err)
			}
		}
		break

	case "InvoiceProcessing":
		// Инвойс в процессе обработки (платеж получен, но еще не подтвержден)
		if h.storage != nil {
			err := h.storage.UpdateTransactionStatus(event.InvoiceID, "processing")
			if err != nil {
				h.logger.Error("Ошибка обновления статуса транзакции: %v", err)
			}
		}
		break

	case "InvoiceSettled", "InvoiceCompleted":
		// Инвойс успешно оплачен и подтвержден
		if h.storage != nil {
			err := h.storage.UpdateTransactionStatus(event.InvoiceID, "completed")
			if err != nil {
				h.logger.Error("Ошибка обновления статуса транзакции: %v", err)
			}
		}
		break

	case "InvoiceExpired", "InvoiceInvalid":
		// Инвойс просрочен или недействителен
		if h.storage != nil {
			status := strings.ToLower(event.Type[7:]) // "expired" или "invalid"
			err := h.storage.UpdateTransactionStatus(event.InvoiceID, status)
			if err != nil {
				h.logger.Error("Ошибка обновления статуса транзакции: %v", err)
			}
		}
		break

	default:
		h.logger.Warn("Получено неизвестное webhook-событие типа: %s", event.Type)
	}

	// Отправляем успешный ответ для подтверждения получения события
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// ListTransactions обрабатывает запрос на получение списка транзакций
// @Summary Получение списка транзакций
// @Description Возвращает список транзакций с возможностью фильтрации и пагинации
// @Tags transactions
// @Produce json
// @Param status query string false "Статус транзакции"
// @Param date_from query string false "Дата начала (формат: YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания (формат: YYYY-MM-DD)"
// @Param page query int false "Номер страницы (по умолчанию: 1)"
// @Param per_page query int false "Количество записей на странице (по умолчанию: 20)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/transactions [get]
func (h *Handler) ListTransactions(c *gin.Context) {
	// Получение параметров запроса
	status := c.Query("status")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if err != nil || perPage < 1 {
		perPage = 20
	}

	h.logger.Info("Запрос списка транзакций: статус=%s, с %s по %s, страница %d, элементов %d",
		status, dateFrom, dateTo, page, perPage)

	// Фильтры для запроса к хранилищу
	filters := make(map[string]interface{})
	if status != "" {
		filters["status"] = status
	}

	// Преобразование дат
	if dateFrom != "" {
		fromDate, err := time.Parse("2006-01-02", dateFrom)
		if err == nil {
			filters["date_from"] = fromDate
		}
	}

	if dateTo != "" {
		toDate, err := time.Parse("2006-01-02", dateTo)
		if err == nil {
			// Устанавливаем конец дня
			toDate = toDate.Add(24*time.Hour - 1*time.Second)
			filters["date_to"] = toDate
		}
	}

	// Пока хранилище не реализовано, возвращаем пустой список
	if h.storage == nil {
		c.JSON(http.StatusOK, gin.H{
			"transactions": []interface{}{},
			"total":        0,
			"page":         page,
			"per_page":     perPage,
		})
		return
	}

	// TODO: Получение транзакций из хранилища
	// transactions, total, err := h.storage.ListTransactions(filters, page, perPage)

	// Заглушка до реализации хранилища
	c.JSON(http.StatusOK, gin.H{
		"transactions": []interface{}{},
		"total":        0,
		"page":         page,
		"per_page":     perPage,
		"filters":      filters,
	})
}
