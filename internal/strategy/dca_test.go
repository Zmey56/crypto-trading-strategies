package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// MockExchangeClient for testing
type MockExchangeClient struct {
	orders []types.Order
}

func (m *MockExchangeClient) PlaceOrder(ctx context.Context, order types.Order) error {
	m.orders = append(m.orders, order)
	return nil
}

func (m *MockExchangeClient) CancelOrder(ctx context.Context, orderID string) error {
	return nil
}

func (m *MockExchangeClient) GetOrder(ctx context.Context, orderID string) (*types.Order, error) {
	return nil, nil
}

func (m *MockExchangeClient) GetActiveOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	return nil, nil
}

func (m *MockExchangeClient) GetFilledOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	return nil, nil
}

func (m *MockExchangeClient) GetTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	return &types.Ticker{
		Symbol:    symbol,
		Price:     45000.0,
		Bid:       44999.9,
		Ask:       45000.1,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockExchangeClient) GetOrderBook(ctx context.Context, symbol string, limit int) (*types.OrderBook, error) {
	return nil, nil
}

func (m *MockExchangeClient) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]types.Candle, error) {
	return nil, nil
}

func (m *MockExchangeClient) GetBalance(ctx context.Context) (*types.Balance, error) {
	return &types.Balance{
		Asset:     "USDT",
		Free:      10000.0,
		Locked:    0.0,
		Total:     10000.0,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockExchangeClient) GetTradingFees(ctx context.Context, symbol string) (*types.TradingFees, error) {
	return &types.TradingFees{
		Symbol:    symbol,
		MakerFee:  0.001,
		TakerFee:  0.001,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockExchangeClient) Ping(ctx context.Context) error {
	return nil
}

func (m *MockExchangeClient) Close() error {
	return nil
}

func TestNewDCAStrategy(t *testing.T) {
	config := types.DCAConfig{
		Symbol:           "BTCUSDT",
		InvestmentAmount: 100.0,
		Interval:         24 * time.Hour,
		MaxInvestments:   100,
		Enabled:          true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)

	strategy := NewDCAStrategy(config, exchange, logger)

	if strategy == nil {
		t.Fatal("Strategy should not be nil")
	}

	if strategy.config.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", strategy.config.Symbol)
	}
}

func TestDCAStrategy_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  types.DCAConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: types.DCAConfig{
				Symbol:           "BTCUSDT",
				InvestmentAmount: 100.0,
				Interval:         24 * time.Hour,
				MaxInvestments:   100,
				Enabled:          true,
			},
			wantErr: false,
		},
		{
			name: "empty symbol",
			config: types.DCAConfig{
				Symbol:           "",
				InvestmentAmount: 100.0,
				Interval:         24 * time.Hour,
				MaxInvestments:   100,
				Enabled:          true,
			},
			wantErr: true,
		},
		{
			name: "negative investment amount",
			config: types.DCAConfig{
				Symbol:           "BTCUSDT",
				InvestmentAmount: -100.0,
				Interval:         24 * time.Hour,
				MaxInvestments:   100,
				Enabled:          true,
			},
			wantErr: true,
		},
		{
			name: "zero interval",
			config: types.DCAConfig{
				Symbol:           "BTCUSDT",
				InvestmentAmount: 100.0,
				Interval:         0,
				MaxInvestments:   100,
				Enabled:          true,
			},
			wantErr: true,
		},
		{
			name: "zero max investments",
			config: types.DCAConfig{
				Symbol:           "BTCUSDT",
				InvestmentAmount: 100.0,
				Interval:         24 * time.Hour,
				MaxInvestments:   0,
				Enabled:          true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchange := &MockExchangeClient{}
			logger := logger.New(logger.LevelInfo)
			strategy := NewDCAStrategy(tt.config, exchange, logger)

			err := strategy.ValidateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDCAStrategy_GetSignal(t *testing.T) {
	config := types.DCAConfig{
		Symbol:           "BTCUSDT",
		InvestmentAmount: 100.0,
		Interval:         24 * time.Hour,
		MaxInvestments:   100,
		Enabled:          true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewDCAStrategy(config, exchange, logger)

	marketData := types.MarketData{
		Symbol:    "BTCUSDT",
		Price:     45000.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	signal := strategy.GetSignal(marketData)

	if signal.Type != types.SignalTypeBuy {
		t.Errorf("Expected signal type BUY, got %s", signal.Type)
	}

	if signal.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", signal.Symbol)
	}

	if signal.Price != 45000.0 {
		t.Errorf("Expected price 45000.0, got %f", signal.Price)
	}
}

func TestDCAStrategy_Execute(t *testing.T) {
	config := types.DCAConfig{
		Symbol:           "BTCUSDT",
		InvestmentAmount: 100.0,
		Interval:         24 * time.Hour,
		MaxInvestments:   100,
		Enabled:          true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewDCAStrategy(config, exchange, logger)

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

	// Verify that an order was placed
	if len(exchange.orders) != 1 {
		t.Errorf("Expected 1 order, got %d", len(exchange.orders))
	}

	order := exchange.orders[0]
	if order.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", order.Symbol)
	}

	if order.Side != types.OrderSideBuy {
		t.Errorf("Expected side BUY, got %s", order.Side)
	}

	expectedQuantity := 100.0 / 45000.0
	if order.Quantity != expectedQuantity {
		t.Errorf("Expected quantity %f, got %f", expectedQuantity, order.Quantity)
	}
}

func TestDCAStrategy_CalculateQuantity(t *testing.T) {
	config := types.DCAConfig{
		Symbol:           "BTCUSDT",
		InvestmentAmount: 100.0,
		Interval:         24 * time.Hour,
		MaxInvestments:   100,
		Enabled:          true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewDCAStrategy(config, exchange, logger)

	price := 45000.0
	expectedQuantity := 100.0 / price
	actualQuantity := strategy.calculateQuantity(price)

	if actualQuantity != expectedQuantity {
		t.Errorf("Expected quantity %f, got %f", expectedQuantity, actualQuantity)
	}
}

func TestDCAStrategy_GetMetrics(t *testing.T) {
	config := types.DCAConfig{
		Symbol:           "BTCUSDT",
		InvestmentAmount: 100.0,
		Interval:         24 * time.Hour,
		MaxInvestments:   100,
		Enabled:          true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewDCAStrategy(config, exchange, logger)

	metrics := strategy.GetMetrics()

	if metrics.TotalTrades != 0 {
		t.Errorf("Expected 0 total trades, got %d", metrics.TotalTrades)
	}

	if metrics.TotalVolume != 0 {
		t.Errorf("Expected 0 total volume, got %f", metrics.TotalVolume)
	}
}

func TestDCAStrategy_GetStatus(t *testing.T) {
	config := types.DCAConfig{
		Symbol:           "BTCUSDT",
		InvestmentAmount: 100.0,
		Interval:         24 * time.Hour,
		MaxInvestments:   100,
		Enabled:          true,
	}

	exchange := &MockExchangeClient{}
	logger := logger.New(logger.LevelInfo)
	strategy := NewDCAStrategy(config, exchange, logger)

	status := strategy.GetStatus()

	if status["enabled"] != true {
		t.Errorf("Expected enabled true, got %v", status["enabled"])
	}

	if status["symbol"] != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %v", status["symbol"])
	}

	if status["buy_count"] != 0 {
		t.Errorf("Expected buy count 0, got %v", status["buy_count"])
	}

	if status["max_buys"] != 100 {
		t.Errorf("Expected max buys 100, got %v", status["max_buys"])
	}
}
