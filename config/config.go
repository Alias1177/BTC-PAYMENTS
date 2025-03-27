package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
)

type Config struct {
	APIKEY     string `env:"APIKEY"` // Исправлено с ApiKey
	DataBase   string `env:"DATABASE_URL"`
	Port       string `env:"SERVER_PORT"`
	APIBaseURL string `env:"API_BASE_URL"` // Исправлено с ApiUrl
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadConfig(path, cfg)
	if err != nil {
		log.Printf("Warning: не удалось прочитать конфигурационный файл: %v", err)
	}

	// Читаем переменные окружения
	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
