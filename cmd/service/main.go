package main

import (
	"chi/BTC-PAYMENTS/config"
	"chi/BTC-PAYMENTS/internal/api"
	"chi/BTC-PAYMENTS/internal/client"
	"chi/BTC-PAYMENTS/internal/storage"
	"chi/BTC-PAYMENTS/pkg/logger"
	"flag"
	"fmt"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	logLevel := flag.Int("log-level", int(logger.INFO), "Log level (0-4)")
	flag.Parse()

	// Create logger
	log := logger.New(logger.LogLevel(*logLevel))
	log.Info("Starting BTC-PAYMENTS service")

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	// Create BTCPay client
	btcpayClient := client.NewBTCPayClient(
		cfg.BTCPay.BaseURL,
		cfg.BTCPay.APIKey,
		cfg.BTCPay.StoreID,
	)

	// Инициализация MongoDB репозитория
	repository, err := storage.NewMongoRepository(
		cfg.MongoDB.URI,
		cfg.MongoDB.Database,
		cfg.MongoDB.Collection,
		log,
	)
	if err != nil {
		log.Fatal("Ошибка инициализации MongoDB: %v", err)
	}
	defer repository.Close()

	// Создание и настройка маршрутизатора с хранилищем
	// Получение webhookSecret из конфигурации
	webhookSecret := cfg.BTCPay.WebhookSecret

	// Создание и настройка маршрутизатора
	router := api.SetupRouter(btcpayClient, repository, log, webhookSecret)
	// Start the server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info("Server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server: %v", err)
	}
}
