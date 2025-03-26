package api

import (
	"chi/BTC-PAYMENTS/internal/client"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

	// TODO: Сохранение в базу данных, когда будет реализовано хранилище
	if h.storage != nil {
		// Здесь будет код сохранения транзакции
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
	// Проверка сигнатуры (для повышения безопасности)
	// signature := c.GetHeader("X-Signature")
	// TODO: Реализовать проверку сигнатуры

	var webhookEvent map[string]interface{}
	if err := c.ShouldBindJSON(&webhookEvent); err != nil {
		h.logger.Error("Ошибка парсинга webhook-события: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Извлечение и логирование информации о событии
	eventType, _ := webhookEvent["type"].(string)
	invoiceId, _ := webhookEvent["invoiceId"].(string)

	h.logger.Info("Получено webhook-событие: %s для инвойса %s", eventType, invoiceId)

	// Обработка события (будет реализовано позднее с использованием хранилища)
	if h.storage != nil {
		// TODO: Обновление статуса транзакции в базе данных
	}

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
