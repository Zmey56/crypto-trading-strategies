package mock

import (
	"context"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// MockClient implements ExchangeClient interface for testing
type MockClient struct {
	balances map[string]*types.Balance
	orders   map[string]*types.Order
}

// NewMockClient creates a new mock exchange client
func NewMockClient() *MockClient {
	return &MockClient{
		balances: map[string]*types.Balance{
			"USDT": {
				Asset:     "USDT",
				Free:      10000.0,
				Locked:    0.0,
				Total:     10000.0,
				Timestamp: time.Now(),
			},
			"BTC": {
				Asset:     "BTC",
				Free:      0.0,
				Locked:    0.0,
				Total:     0.0,
				Timestamp: time.Now(),
			},
		},
		orders: make(map[string]*types.Order),
	}
}

// PlaceOrder places a mock order
func (mc *MockClient) PlaceOrder(ctx context.Context, order types.Order) error {
	order.ID = generateOrderID()
	order.Status = types.OrderStatusFilled
	order.Timestamp = time.Now()

	// Simulate order execution
	if order.Side == types.OrderSideBuy {
		mc.balances["USDT"].Free -= order.Quantity * order.Price
		mc.balances["USDT"].Total = mc.balances["USDT"].Free + mc.balances["USDT"].Locked
		mc.balances["BTC"].Free += order.Quantity
		mc.balances["BTC"].Total = mc.balances["BTC"].Free + mc.balances["BTC"].Locked
	} else {
		mc.balances["USDT"].Free += order.Quantity * order.Price
		mc.balances["USDT"].Total = mc.balances["USDT"].Free + mc.balances["USDT"].Locked
		mc.balances["BTC"].Free -= order.Quantity
		mc.balances["BTC"].Total = mc.balances["BTC"].Free + mc.balances["BTC"].Locked
	}

	mc.orders[order.ID] = &order
	return nil
}

// CancelOrder cancels a mock order
func (mc *MockClient) CancelOrder(ctx context.Context, orderID string) error {
	if order, exists := mc.orders[orderID]; exists {
		order.Status = types.OrderStatusCanceled
	}
	return nil
}

// GetOrder gets a mock order
func (mc *MockClient) GetOrder(ctx context.Context, orderID string) (*types.Order, error) {
	if order, exists := mc.orders[orderID]; exists {
		return order, nil
	}
	return nil, nil
}

// GetActiveOrders gets active mock orders
func (mc *MockClient) GetActiveOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	var activeOrders []types.Order
	for _, order := range mc.orders {
		if order.Symbol == symbol && order.Status == types.OrderStatusNew {
			activeOrders = append(activeOrders, *order)
		}
	}
	return activeOrders, nil
}

// GetFilledOrders gets filled mock orders
func (mc *MockClient) GetFilledOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	var filledOrders []types.Order
	for _, order := range mc.orders {
		if order.Symbol == symbol && order.Status == types.OrderStatusFilled {
			filledOrders = append(filledOrders, *order)
		}
	}
	return filledOrders, nil
}

// GetTicker gets mock ticker data
func (mc *MockClient) GetTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	return &types.Ticker{
		Symbol:    symbol,
		Price:     45000.0, // Mock BTC price
		Bid:       44999.0,
		Ask:       45001.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}, nil
}

// GetOrderBook gets mock order book
func (mc *MockClient) GetOrderBook(ctx context.Context, symbol string, limit int) (*types.OrderBook, error) {
	return &types.OrderBook{
		Symbol: symbol,
		Bids: []types.OrderBookEntry{
			{Price: 44999.0, Amount: 1.0},
			{Price: 44998.0, Amount: 2.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 45001.0, Amount: 1.0},
			{Price: 45002.0, Amount: 2.0},
		},
	}, nil
}

// GetCandles gets mock candle data
func (mc *MockClient) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]types.Candle, error) {
	now := time.Now()
	var candles []types.Candle

	for i := 0; i < limit; i++ {
		candle := types.Candle{
			Symbol:    symbol,
			Open:      45000.0 + float64(i),
			High:      45010.0 + float64(i),
			Low:       44990.0 + float64(i),
			Close:     45005.0 + float64(i),
			Volume:    1000.0,
			Timestamp: now.Add(-time.Duration(i) * time.Hour),
		}
		candles = append(candles, candle)
	}

	return candles, nil
}

// GetBalance gets mock balance
func (mc *MockClient) GetBalance(ctx context.Context) (*types.Balance, error) {
	// Return USDT balance as default
	return mc.balances["USDT"], nil
}

// GetTradingFees gets mock trading fees
func (mc *MockClient) GetTradingFees(ctx context.Context, symbol string) (*types.TradingFees, error) {
	return &types.TradingFees{
		Symbol:    symbol,
		MakerFee:  0.001,
		TakerFee:  0.001,
		Timestamp: time.Now(),
	}, nil
}

// Ping pings the mock exchange
func (mc *MockClient) Ping(ctx context.Context) error {
	return nil
}

// Close closes the mock client
func (mc *MockClient) Close() error {
	return nil
}

// generateOrderID generates a mock order ID
func generateOrderID() string {
	return "mock_order_" + time.Now().Format("20060102150405")
}
