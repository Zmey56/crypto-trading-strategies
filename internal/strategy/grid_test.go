package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

func TestNewGridStrategy(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         20,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewGridStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Grid strategy: %v", err)
	}

	if strategy == nil {
		t.Fatal("Strategy should not be nil")
	}
}

func TestGridStrategy_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  types.GridConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: types.GridConfig{
				Symbol:             "BTCUSDT",
				UpperPrice:         50000.0,
				LowerPrice:         40000.0,
				GridLevels:         20,
				InvestmentPerLevel: 100.0,
				Enabled:            true,
			},
			wantErr: false,
		},
		{
			name: "missing symbol",
			config: types.GridConfig{
				Symbol:             "",
				UpperPrice:         50000.0,
				LowerPrice:         40000.0,
				GridLevels:         20,
				InvestmentPerLevel: 100.0,
				Enabled:            true,
			},
			wantErr: true,
		},
		{
			name: "invalid price bounds",
			config: types.GridConfig{
				Symbol:             "BTCUSDT",
				UpperPrice:         40000.0,
				LowerPrice:         50000.0,
				GridLevels:         20,
				InvestmentPerLevel: 100.0,
				Enabled:            true,
			},
			wantErr: true,
		},
		{
			name: "invalid grid levels",
			config: types.GridConfig{
				Symbol:             "BTCUSDT",
				UpperPrice:         50000.0,
				LowerPrice:         40000.0,
				GridLevels:         1,
				InvestmentPerLevel: 100.0,
				Enabled:            true,
			},
			wantErr: true,
		},
		{
			name: "invalid investment per level",
			config: types.GridConfig{
				Symbol:             "BTCUSDT",
				UpperPrice:         50000.0,
				LowerPrice:         40000.0,
				GridLevels:         20,
				InvestmentPerLevel: 0.0,
				Enabled:            true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchange := &MockExchangeClient{}
			logger := logger.New(logger.LevelInfo)

			strategy, err := NewGridStrategy(tt.config, exchange, logger)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewGridStrategy() unexpected error = %v", err)
				}
				return
			}

			if err := strategy.ValidateConfig(); (err != nil) != tt.wantErr {
				t.Errorf("GridStrategy.ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGridStrategy_Execute(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         5,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewGridStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Grid strategy: %v", err)
	}

	// Test execution with price at lower bound
	marketData := types.MarketData{
		Symbol:    "BTCUSDT",
		Price:     40000.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	if err := strategy.Execute(ctx, marketData); err != nil {
		t.Errorf("GridStrategy.Execute() error = %v", err)
	}

	// Test execution with price at upper bound
	marketData.Price = 50000.0
	if err := strategy.Execute(ctx, marketData); err != nil {
		t.Errorf("GridStrategy.Execute() error = %v", err)
	}

	// Test execution with disabled strategy
	config.Enabled = false
	strategy, _ = NewGridStrategy(config, exchange, logger)
	if err := strategy.Execute(ctx, marketData); err != nil {
		t.Errorf("GridStrategy.Execute() should not error when disabled")
	}
}

func TestGridStrategy_GetSignal(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         5,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewGridStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Grid strategy: %v", err)
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
}

func TestGridStrategy_GetMetrics(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         5,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewGridStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Grid strategy: %v", err)
	}

	metrics := strategy.GetMetrics()
	if metrics.TotalTrades != 0 {
		t.Errorf("Expected 0 trades initially, got %d", metrics.TotalTrades)
	}
}

func TestGridStrategy_Shutdown(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         5,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy, err := NewGridStrategy(config, exchange, logger)
	if err != nil {
		t.Fatalf("Failed to create Grid strategy: %v", err)
	}

	ctx := context.Background()
	if err := strategy.Shutdown(ctx); err != nil {
		t.Errorf("GridStrategy.Shutdown() error = %v", err)
	}
}
