package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// MockExchangeClient для тестирования Grid стратегии
type MockGridExchangeClient struct {
	orders []types.Order
}

func (m *MockGridExchangeClient) PlaceOrder(ctx context.Context, order types.Order) error {
	m.orders = append(m.orders, order)
	return nil
}

func (m *MockGridExchangeClient) CancelOrder(ctx context.Context, orderID string) error {
	return nil
}

func (m *MockGridExchangeClient) GetOrder(ctx context.Context, orderID string) (*types.Order, error) {
	return nil, nil
}

func (m *MockGridExchangeClient) GetActiveOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	return nil, nil
}

func (m *MockGridExchangeClient) GetFilledOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	return nil, nil
}

func (m *MockGridExchangeClient) GetTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	return &types.Ticker{
		Symbol:    symbol,
		Price:     45000.0,
		Bid:       44999.9,
		Ask:       45000.1,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockGridExchangeClient) GetOrderBook(ctx context.Context, symbol string, limit int) (*types.OrderBook, error) {
	return nil, nil
}

func (m *MockGridExchangeClient) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]types.Candle, error) {
	return nil, nil
}

func (m *MockGridExchangeClient) GetBalance(ctx context.Context) (*types.Balance, error) {
	return &types.Balance{
		Asset:     "USDT",
		Free:      10000.0,
		Locked:    0.0,
		Total:     10000.0,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockGridExchangeClient) GetTradingFees(ctx context.Context, symbol string) (*types.TradingFees, error) {
	return &types.TradingFees{
		Symbol:    symbol,
		MakerFee:  0.001,
		TakerFee:  0.001,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockGridExchangeClient) Ping(ctx context.Context) error {
	return nil
}

func (m *MockGridExchangeClient) Close() error {
	return nil
}

func TestNewGridStrategy(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         10,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy := NewGridStrategy(config, exchange, logger)

	if strategy == nil {
		t.Fatal("Strategy should not be nil")
	}

	if strategy.config.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", strategy.config.Symbol)
	}

	if len(strategy.gridLevels) != 10 {
		t.Errorf("Expected 10 grid levels, got %d", len(strategy.gridLevels))
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
				GridLevels:         10,
				InvestmentPerLevel: 100.0,
				Enabled:            true,
			},
			wantErr: false,
		},
		{
			name: "empty symbol",
			config: types.GridConfig{
				Symbol:             "",
				UpperPrice:         50000.0,
				LowerPrice:         40000.0,
				GridLevels:         10,
				InvestmentPerLevel: 100.0,
				Enabled:            true,
			},
			wantErr: true,
		},
		{
			name: "upper price <= lower price",
			config: types.GridConfig{
				Symbol:             "BTCUSDT",
				UpperPrice:         40000.0,
				LowerPrice:         50000.0,
				GridLevels:         10,
				InvestmentPerLevel: 100.0,
				Enabled:            true,
			},
			wantErr: true,
		},
		{
			name: "zero grid levels",
			config: types.GridConfig{
				Symbol:             "BTCUSDT",
				UpperPrice:         50000.0,
				LowerPrice:         40000.0,
				GridLevels:         0,
				InvestmentPerLevel: 100.0,
				Enabled:            true,
			},
			wantErr: true,
		},
		{
			name: "zero investment per level",
			config: types.GridConfig{
				Symbol:             "BTCUSDT",
				UpperPrice:         50000.0,
				LowerPrice:         40000.0,
				GridLevels:         10,
				InvestmentPerLevel: 0.0,
				Enabled:            true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchange := &MockGridExchangeClient{}
			logger := logger.New(logger.LevelInfo)
			strategy := NewGridStrategy(tt.config, exchange, logger)

			err := strategy.ValidateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGridStrategy_GetSignal(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         10,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewGridStrategy(config, exchange, logger)

	marketData := types.MarketData{
		Symbol:    "BTCUSDT",
		Price:     45000.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	signal := strategy.GetSignal(marketData)

	// Проверяем, что сигнал имеет правильный тип (может быть BUY, SELL или HOLD)
	if signal.Type != types.SignalTypeBuy && signal.Type != types.SignalTypeSell && signal.Type != types.SignalTypeHold {
		t.Errorf("Expected signal type BUY, SELL or HOLD, got %s", signal.Type)
	}

	if signal.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", signal.Symbol)
	}

	// Проверяем, что цена в сигнале соответствует ожидаемой
	if signal.Type == types.SignalTypeBuy || signal.Type == types.SignalTypeSell {
		expectedQuantity := 100.0 / signal.Price
		if signal.Quantity != expectedQuantity {
			t.Errorf("Expected quantity %f, got %f", expectedQuantity, signal.Quantity)
		}
	}
}

func TestGridStrategy_Execute(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         10,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewGridStrategy(config, exchange, logger)

	marketData := types.MarketData{
		Symbol:    "BTCUSDT",
		Price:     45000.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err := strategy.Execute(ctx, marketData)

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Проверяем, что ордера были размещены
	if len(exchange.orders) == 0 {
		t.Error("Expected orders to be placed")
	}
}

func TestGridStrategy_CalculateGridLevels(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         5,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewGridStrategy(config, exchange, logger)

	expectedLevels := []float64{40000.0, 42500.0, 45000.0, 47500.0, 50000.0}

	if len(strategy.gridLevels) != len(expectedLevels) {
		t.Errorf("Expected %d levels, got %d", len(expectedLevels), len(strategy.gridLevels))
	}

	for i, expected := range expectedLevels {
		if strategy.gridLevels[i] != expected {
			t.Errorf("Level %d: expected %f, got %f", i, expected, strategy.gridLevels[i])
		}
	}
}

func TestGridStrategy_FindNearestLevels(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         5,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewGridStrategy(config, exchange, logger)

	// Тест для цены в середине диапазона
	buyLevel, sellLevel := strategy.findNearestLevels(45000.0)
	if buyLevel != 45000.0 {
		t.Errorf("Expected buy level 45000.0, got %f", buyLevel)
	}
	if sellLevel != 47500.0 {
		t.Errorf("Expected sell level 47500.0, got %f", sellLevel)
	}

	// Тест для цены ниже нижней границы
	buyLevel, sellLevel = strategy.findNearestLevels(35000.0)
	if buyLevel != 0 {
		t.Errorf("Expected buy level 0, got %f", buyLevel)
	}
	if sellLevel != 0 {
		t.Errorf("Expected sell level 0, got %f", sellLevel)
	}
}

func TestGridStrategy_CalculateOrderQuantity(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         10,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewGridStrategy(config, exchange, logger)

	price := 45000.0
	expectedQuantity := 100.0 / price
	actualQuantity := strategy.calculateOrderQuantity(price)

	if actualQuantity != expectedQuantity {
		t.Errorf("Expected quantity %f, got %f", expectedQuantity, actualQuantity)
	}
}

func TestGridStrategy_GetMetrics(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         10,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewGridStrategy(config, exchange, logger)

	metrics := strategy.GetMetrics()

	if metrics.TotalTrades != 0 {
		t.Errorf("Expected 0 total trades, got %d", metrics.TotalTrades)
	}

	if metrics.TotalVolume != 0 {
		t.Errorf("Expected 0 total volume, got %f", metrics.TotalVolume)
	}
}

func TestGridStrategy_GetStatus(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         10,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewGridStrategy(config, exchange, logger)

	status := strategy.GetStatus()

	if status["enabled"] != true {
		t.Errorf("Expected enabled true, got %v", status["enabled"])
	}

	if status["symbol"] != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %v", status["symbol"])
	}

	if status["grid_levels"] != 10 {
		t.Errorf("Expected grid levels 10, got %v", status["grid_levels"])
	}

	if status["active_orders"] != 0 {
		t.Errorf("Expected active orders 0, got %v", status["active_orders"])
	}
}

func TestGridStrategy_PriceOutOfRange(t *testing.T) {
	config := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         10,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	exchange := &MockGridExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewGridStrategy(config, exchange, logger)

	// Цена выше верхней границы
	marketData := types.MarketData{
		Symbol:    "BTCUSDT",
		Price:     55000.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err := strategy.Execute(ctx, marketData)

	if err != nil {
		t.Errorf("Execute() should not return error for out-of-range price: %v", err)
	}

	// Не должно быть размещено ордеров
	if len(exchange.orders) != 0 {
		t.Errorf("Expected 0 orders for out-of-range price, got %d", len(exchange.orders))
	}
}
