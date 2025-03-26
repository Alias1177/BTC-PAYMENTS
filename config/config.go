// internal/config/config.go
package config

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	BTCPay struct {
		BaseURL string `yaml:"base_url"`
		APIKey  string `yaml:"api_key"`
		StoreID string `yaml:"store_id"`
	} `yaml:"btcpay"`

	MongoDB struct {
		URI        string `yaml:"uri"`
		Database   string `yaml:"database"`
		Collection string `yaml:"collection"`
	} `yaml:"mongodb"`
}

//func LoadConfig(path string) (*Config, error) {
//	// Реализация загрузки конфигурации
//}
