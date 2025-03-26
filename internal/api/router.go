package api

import (
	"github.com/Alias1177/BTC-PAYMENTS/internal/client"
	"github.com/Alias1177/BTC-PAYMENTS/internal/models"
	"github.com/Alias1177/BTC-PAYMENTS/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"time"
)

// Handler содержит зависимости для обработчиков API
type Handler struct {
	btcpayClient  *client.BTCPayClient
	storage       models.Repository
	logger        *logger.Logger
	webhookSecret string
}

// NewHandler создает новый экземпляр Handler
func NewHandler(btcpayClient *client.BTCPayClient, storage models.Repository, logger *logger.Logger, webhookSecret string) *Handler {
	return &Handler{
		btcpayClient:  btcpayClient,
		storage:       storage,
		logger:        logger,
		webhookSecret: webhookSecret,
	}
}

// SetupRouter настраивает и возвращает маршрутизатор Gin
func SetupRouter(btcpayClient *client.BTCPayClient, storage models.Repository, logger *logger.Logger, webhookSecret string) *gin.Engine {
	// Установка режима работы Gin (возможно, в production понадобится режим release)
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.DebugMode)
	}

	// Создаем новый роутер Gin с логгером и recovery middleware
	router := gin.New()
	router.Use(gin.Recovery())

	// Middleware для логирования запросов
	router.Use(func(c *gin.Context) {
		// Начало обработки запроса
		t := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Обработка запроса
		c.Next()

		// Конец обработки запроса
		latency := time.Since(t)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("[HTTP] %v | %3d | %13v | %15s | %-7s %s",
			t.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	})

	// Создаем обработчик API
	handler := NewHandler(btcpayClient, storage, logger, webhookSecret)

	// Базовый маршрут для проверки работоспособности
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Настройка маршрутов API
	v1 := router.Group("/api/v1")
	{
		// Маршруты для инвойсов
		invoices := v1.Group("/invoices")
		{
			invoices.POST("", handler.CreateInvoice)
			invoices.GET(":id", handler.GetInvoice)
		}

		// Маршрут для webhook-уведомлений
		v1.POST("/webhooks/btcpay", handler.HandleWebhook)

		// Маршрут для списка транзакций
		v1.GET("/transactions", handler.ListTransactions)
	}

	// Маршруты для мониторинга
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger-документация API
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
