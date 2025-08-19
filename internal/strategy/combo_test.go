package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

func TestNewComboStrategy(t *testing.T) {
	config := types.ComboConfig{
		Strategies: []types.StrategyConfig{
			{
				Type: "dca",
				Config: map[string]interface{}{
					"symbol":            "BTCUSDT",
					"investment_amount": 100.0,
					"interval":          "24h",
					"max_investments":   100.0,
					"enabled":           true,
				},
			},
			{
				Type: "grid",
				Config: map[string]interface{}{
					"symbol":               "BTCUSDT",
					"upper_price":          50000.0,
					"lower_price":          40000.0,
					"grid_levels":          20.0,
					"investment_per_level": 100.0,
					"enabled":              true,
				},
			},
		},
		Enabled: true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewComboStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Combo strategy: %v", err)
	}

	if strategy == nil {
		t.Fatal("Strategy should not be nil")
	}
}

func TestNewComboStrategy_EmptyStrategies(t *testing.T) {
	config := types.ComboConfig{
		Strategies: []types.StrategyConfig{},
		Enabled:    true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	_, err := NewComboStrategy(config, exchange, logger)
	if err == nil {
		t.Fatal("Expected error for empty strategies")
	}
}

func TestComboStrategy_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  types.ComboConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: types.ComboConfig{
				Strategies: []types.StrategyConfig{
					{
						Type: "dca",
						Config: map[string]interface{}{
							"symbol":            "BTCUSDT",
							"investment_amount": 100.0,
							"interval":          "24h",
							"max_investments":   100.0,
							"enabled":           true,
						},
					},
				},
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "empty strategies",
			config: types.ComboConfig{
				Strategies: []types.StrategyConfig{},
				Enabled:    true,
			},
			wantErr: true,
		},
		{
			name: "missing strategy type",
			config: types.ComboConfig{
				Strategies: []types.StrategyConfig{
					{
						Type:   "",
						Config: map[string]interface{}{},
					},
				},
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "missing strategy config",
			config: types.ComboConfig{
				Strategies: []types.StrategyConfig{
					{
						Type:   "dca",
						Config: nil,
					},
				},
				Enabled: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchange := &MockExchangeClient{}
			logger := logger.New(logger.LevelInfo)

			strategy, err := NewComboStrategy(tt.config, exchange, logger)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewComboStrategy() unexpected error = %v", err)
				}
				return
			}

			if err := strategy.ValidateConfig(); (err != nil) != tt.wantErr {
				t.Errorf("ComboStrategy.ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComboStrategy_Execute(t *testing.T) {
	config := types.ComboConfig{
		Strategies: []types.StrategyConfig{
			{
				Type: "dca",
				Config: map[string]interface{}{
					"symbol":            "BTCUSDT",
					"investment_amount": 100.0,
					"interval":          "24h",
					"max_investments":   100.0,
					"enabled":           true,
				},
			},
		},
		Enabled: true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewComboStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Combo strategy: %v", err)
	}

	marketData := types.MarketData{
		Symbol:    "BTCUSDT",
		Price:     45000.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	if err := strategy.Execute(ctx, marketData); err != nil {
		t.Errorf("ComboStrategy.Execute() error = %v", err)
	}

	// Test execution with disabled strategy
	config.Enabled = false
	strategy, _ = NewComboStrategy(config, exchange, logger)
	if err := strategy.Execute(ctx, marketData); err != nil {
		t.Errorf("ComboStrategy.Execute() should not error when disabled")
	}
}

func TestComboStrategy_GetSignal(t *testing.T) {
	config := types.ComboConfig{
		Strategies: []types.StrategyConfig{
			{
				Type: "dca",
				Config: map[string]interface{}{
					"symbol":            "BTCUSDT",
					"investment_amount": 100.0,
					"interval":          "24h",
					"max_investments":   100.0,
					"enabled":           true,
				},
			},
		},
		Enabled: true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewComboStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Combo strategy: %v", err)
	}

	marketData := types.MarketData{
		Symbol:    "BTCUSDT",
		Price:     45000.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	signal := strategy.GetSignal(marketData)
	if signal.Symbol != marketData.Symbol {
		t.Errorf("Expected symbol %s, got %s", marketData.Symbol, signal.Symbol)
	}

	if signal.Price != marketData.Price {
		t.Errorf("Expected price %f, got %f", marketData.Price, signal.Price)
	}

	// Test with disabled strategy
	config.Enabled = false
	strategy, _ = NewComboStrategy(config, exchange, logger)
	signal = strategy.GetSignal(marketData)
	if signal.Type != types.SignalTypeHold {
		t.Errorf("Expected HOLD signal when disabled, got %s", signal.Type)
	}
}

func TestComboStrategy_GetMetrics(t *testing.T) {
	config := types.ComboConfig{
		Strategies: []types.StrategyConfig{
			{
				Type: "dca",
				Config: map[string]interface{}{
					"symbol":            "BTCUSDT",
					"investment_amount": 100.0,
					"interval":          "24h",
					"max_investments":   100.0,
					"enabled":           true,
				},
			},
		},
		Enabled: true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewComboStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Combo strategy: %v", err)
	}

	metrics := strategy.GetMetrics()
	if metrics.TotalTrades != 0 {
		t.Errorf("Expected 0 trades initially, got %d", metrics.TotalTrades)
	}
}

func TestComboStrategy_Shutdown(t *testing.T) {
	config := types.ComboConfig{
		Strategies: []types.StrategyConfig{
			{
				Type: "dca",
				Config: map[string]interface{}{
					"symbol":            "BTCUSDT",
					"investment_amount": 100.0,
					"interval":          "24h",
					"max_investments":   100.0,
					"enabled":           true,
				},
			},
		},
		Enabled: true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewComboStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Combo strategy: %v", err)
	}

	ctx := context.Background()
	if err := strategy.Shutdown(ctx); err != nil {
		t.Errorf("ComboStrategy.Shutdown() error = %v", err)
	}
}

func TestComboStrategy_GetStatus(t *testing.T) {
	config := types.ComboConfig{
		Strategies: []types.StrategyConfig{
			{
				Type: "dca",
				Config: map[string]interface{}{
					"symbol":            "BTCUSDT",
					"investment_amount": 100.0,
					"interval":          "24h",
					"max_investments":   100.0,
					"enabled":           true,
				},
			},
		},
		Enabled: true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewComboStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Combo strategy: %v", err)
	}

	status := strategy.GetStatus()
	if status["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", status["enabled"])
	}

	if status["strategies"] != 1 {
		t.Errorf("Expected 1 strategy, got %v", status["strategies"])
	}
}
