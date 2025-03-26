// internal/config/config.go
package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	BTCPay struct {
		BaseURL       string `yaml:"base_url"`
		APIKey        string `yaml:"api_key"`
		StoreID       string `yaml:"store_id"`
		WebhookSecret string `yaml:"webhook_secret"` // Добавлено новое поле
	}
	MongoDB struct {
		URI        string `yaml:"uri"`
		Database   string `yaml:"database"`
		Collection string `yaml:"collection"`
	} `yaml:"mongodb"`
}

// LoadConfig loads configuration from a YAML file and overrides with environment variables
func LoadConfig(path string) (*Config, error) {
	// Create default config
	cfg := &Config{}

	// Read configuration file if provided
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("error opening config file: %w", err)
		}
		defer file.Close()

		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(cfg); err != nil {
			return nil, fmt.Errorf("error decoding config file: %w", err)
		}
	}

	// Override with environment variables if set
	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		cfg.Server.Port = serverPort
	}

	if serverHost := os.Getenv("SERVER_HOST"); serverHost != "" {
		cfg.Server.Host = serverHost
	}

	if btcpayBaseURL := os.Getenv("BTCPAY_URL"); btcpayBaseURL != "" {
		cfg.BTCPay.BaseURL = btcpayBaseURL
	}

	if btcpayAPIKey := os.Getenv("BTCPAY_API_KEY"); btcpayAPIKey != "" {
		cfg.BTCPay.APIKey = btcpayAPIKey
	}

	if btcpayStoreID := os.Getenv("BTCPAY_STORE_ID"); btcpayStoreID != "" {
		cfg.BTCPay.StoreID = btcpayStoreID
	}

	if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
		cfg.MongoDB.URI = mongoURI
	}

	if mongoDatabase := os.Getenv("MONGO_DATABASE"); mongoDatabase != "" {
		cfg.MongoDB.Database = mongoDatabase
	}

	if mongoCollection := os.Getenv("MONGO_COLLECTION"); mongoCollection != "" {
		cfg.MongoDB.Collection = mongoCollection
	}

	return cfg, nil
}
