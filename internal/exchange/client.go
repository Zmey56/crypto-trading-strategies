package exchange

import (
	"context"
	"crypto-trading-strategies/pkg/types"
	"fmt"
	"strings"
)

type Client interface {
	// Order management
	PlaceOrder(ctx context.Context, order types.Order) error
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*types.Order, error)
	GetActiveOrders(ctx context.Context, symbol string) ([]types.Order, error)
	GetFilledOrders(ctx context.Context, symbol string) ([]types.Order, error)

	// Market data
	GetTicker(ctx context.Context, symbol string) (*types.Ticker, error)
	GetOrderBook(ctx context.Context, symbol string, limit int) (*types.OrderBook, error)
	GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]types.Candle, error)

	// Account information
	GetBalance(ctx context.Context) (*types.Balance, error)
	GetTradingFees(ctx context.Context, symbol string) (*types.TradingFees, error)

	// WebSocket streams
	SubscribeToTickers(ctx context.Context, symbols []string) (<-chan types.Ticker, error)
	SubscribeToOrderUpdates(ctx context.Context) (<-chan types.OrderUpdate, error)

	// Connection management
	Ping(ctx context.Context) error
	Close() error
}

type ExchangeConfig struct {
	Name       string
	APIKey     string
	SecretKey  string
	Passphrase string
	Sandbox    bool
	RateLimit  RateLimitConfig
	Retry      RetryConfig
}

type UnifiedClient struct {
	clients map[string]Client
	router  *RequestRouter
	monitor *HealthMonitor
	logger  *logger.Logger
}

func NewUnifiedClient(configs []ExchangeConfig) (*UnifiedClient, error) {
	clients := make(map[string]Client)

	for _, config := range configs {
		client, err := createExchangeClient(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s client: %w", config.Name, err)
		}
		clients[config.Name] = client
	}

	return &UnifiedClient{
		clients: clients,
		router:  NewRequestRouter(),
		monitor: NewHealthMonitor(),
	}, nil
}

func createExchangeClient(config ExchangeConfig) (Client, error) {
	switch strings.ToLower(config.Name) {
	case "binance":
		return binance.NewClient(config)
	case "kraken":
		return kraken.NewClient(config)
	case "coinbase":
		return coinbase.NewClient(config)
	default:
		return nil, fmt.Errorf("unsupported exchange: %s", config.Name)
	}
}
