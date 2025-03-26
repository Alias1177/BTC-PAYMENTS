package main

import (
	"chi/BTC-PAYMENTS/config"
	"chi/BTC-PAYMENTS/internal/api"
	"chi/BTC-PAYMENTS/internal/client"
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

	// Create and setup router
	router := api.SetupRouter(btcpayClient, nil, log) // We'll add storage implementation later

	// Start the server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info("Server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server: %v", err)
	}
}
