package config

import (
	"crypto-trading-strategies/pkg/types"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config представляет основную конфигурацию приложения
type Config struct {
	App      AppConfig      `json:"app"`
	Exchange ExchangeConfig `json:"exchange"`
	Strategy StrategyConfig `json:"strategy"`
	Logging  LoggingConfig  `json:"logging"`
}

// AppConfig представляет конфигурацию приложения
type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Port    int    `json:"port"`
	Debug   bool   `json:"debug"`
}

// ExchangeConfig представляет конфигурацию биржи
type ExchangeConfig struct {
	Name       string `json:"name"`
	APIKey     string `json:"api_key"`
	SecretKey  string `json:"secret_key"`
	Passphrase string `json:"passphrase"`
	Sandbox    bool   `json:"sandbox"`
}

// StrategyConfig представляет конфигурацию стратегий
type StrategyConfig struct {
	DCA  *types.DCAConfig  `json:"dca"`
	Grid *types.GridConfig `json:"grid"`
}

// LoggingConfig представляет конфигурацию логирования
type LoggingConfig struct {
	Level  string `json:"level"`
	File   string `json:"file"`
	Format string `json:"format"`
}

// Load загружает конфигурацию из файла
func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// LoadFromEnv загружает конфигурацию из переменных окружения
func LoadFromEnv() *Config {
	return &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "crypto-trading-bot"),
			Version: getEnv("APP_VERSION", "1.0.0"),
			Port:    getEnvAsInt("APP_PORT", 8080),
			Debug:   getEnvAsBool("APP_DEBUG", false),
		},
		Exchange: ExchangeConfig{
			Name:       getEnv("EXCHANGE_NAME", "binance"),
			APIKey:     getEnv("EXCHANGE_API_KEY", ""),
			SecretKey:  getEnv("EXCHANGE_SECRET_KEY", ""),
			Passphrase: getEnv("EXCHANGE_PASSPHRASE", ""),
			Sandbox:    getEnvAsBool("EXCHANGE_SANDBOX", true),
		},
		Strategy: StrategyConfig{
			DCA: &types.DCAConfig{
				Symbol:           getEnv("DCA_SYMBOL", "BTCUSDT"),
				InvestmentAmount: getEnvAsFloat("DCA_INVESTMENT_AMOUNT", 100.0),
				Interval:         getEnvAsDuration("DCA_INTERVAL", 24*time.Hour),
				MaxInvestments:   getEnvAsInt("DCA_MAX_INVESTMENTS", 100),
				PriceThreshold:   getEnvAsFloat("DCA_PRICE_THRESHOLD", 0.0),
				StopLoss:         getEnvAsFloat("DCA_STOP_LOSS", 0.0),
				TakeProfit:       getEnvAsFloat("DCA_TAKE_PROFIT", 0.0),
				Enabled:          getEnvAsBool("DCA_ENABLED", true),
			},
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			File:   getEnv("LOG_FILE", ""),
			Format: getEnv("LOG_FORMAT", "text"),
		},
	}
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}

	if c.Exchange.Name == "" {
		return fmt.Errorf("exchange name is required")
	}

	if c.Exchange.APIKey == "" {
		return fmt.Errorf("exchange API key is required")
	}

	if c.Exchange.SecretKey == "" {
		return fmt.Errorf("exchange secret key is required")
	}

	return nil
}

// Save сохраняет конфигурацию в файл
func (c *Config) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// Вспомогательные функции для работы с переменными окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := fmt.Sscanf(value, "%f", &defaultValue); err == nil && floatValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
} 