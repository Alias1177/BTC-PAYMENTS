package main

import (
	"github.com/Alias1177/BTC-PAYMENTS/config"
	"github.com/Alias1177/BTC-PAYMENTS/hook" // Убедимся, что пакет hook существует в проекте
	"github.com/Alias1177/BTC-PAYMENTS/internal/handler"
	"github.com/Alias1177/BTC-PAYMENTS/repo"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConn, err := repo.SetupDB(cfg.DataBase)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	h := &handler.Handler{
		DB:     dbConn,
		Config: cfg,
	}

	api := r.Group("/api")
	{
		// Регистрируем эндпоинты внутри этой группы
		api.POST("/assign-invoice", h.AssignInvoiceHandler)
		api.GET("/check-payment", h.CheckPaymentHandler)
		api.GET("/user-payments", h.GetUserPaymentsHandler)
	}

	r.POST("/webhook", hook.NowPaymentsWebhookHandler(dbConn.Conn()))

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}
