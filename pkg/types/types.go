package types

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// MarketData represents market data snapshot
type MarketData struct {
	Symbol    string
	Price     float64
	Volume    float64
	Timestamp time.Time
	Ticker    *Ticker
	OrderBook *OrderBook
	Candles   []Candle
}

// Ticker represents current quote
type Ticker struct {
	Symbol    string
	Price     float64
	Bid       float64
	Ask       float64
	Volume    float64
	Timestamp time.Time
}

// OrderBook represents order book
type OrderBook struct {
	Symbol string
	Bids   []OrderBookEntry
	Asks   []OrderBookEntry
}

// OrderBookEntry represents an order book entry
type OrderBookEntry struct {
	Price  float64
	Amount float64
}

// Candle represents a candle
type Candle struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Timestamp time.Time
}

// Order represents a trade order
type Order struct {
	ID            string
	Symbol        string
	Side          OrderSide
	Type          OrderType
	Quantity      float64
	Price         float64
	Status        OrderStatus
	FilledAmount  float64
	FilledPrice   float64
	Timestamp     time.Time
	ExchangeOrder *ExchangeOrder
}

// OrderSide represents order side
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType represents order type
type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusRejected        OrderStatus = "REJECTED"
)

// ExchangeOrder represents an exchange-side order
type ExchangeOrder struct {
	ExchangeOrderID string
	Exchange        string
	ClientOrderID   string
}

// OrderUpdate represents an order update
type OrderUpdate struct {
	OrderID       string
	Status        OrderStatus
	FilledAmount  float64
	FilledPrice   float64
	Timestamp     time.Time
	ExchangeOrder *ExchangeOrder
}

// Balance represents account balance
type Balance struct {
	Asset     string
	Free      float64
	Locked    float64
	Total     float64
	Timestamp time.Time
}

// TradingFees represents trading fees
type TradingFees struct {
	Symbol    string
	MakerFee  float64
	TakerFee  float64
	Timestamp time.Time
}

// Signal represents a trading signal
type Signal struct {
	Type      SignalType
	Symbol    string
	Price     float64
	Quantity  float64
	Strength  float64
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// SignalType represents signal type
type SignalType string

const (
	SignalTypeBuy  SignalType = "BUY"
	SignalTypeSell SignalType = "SELL"
	SignalTypeHold SignalType = "HOLD"
)

// StrategyMetrics collects strategy performance counters
type StrategyMetrics struct {
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	TotalProfit   float64
	TotalLoss     float64
	WinRate       float64
	AverageWin    float64
	AverageLoss   float64
	ProfitFactor  float64
	MaxDrawdown   float64
	SharpeRatio   float64
	TotalVolume   float64
	LastUpdate    time.Time
}

// DCAConfig contains DCA parameters
type DCAConfig struct {
	Symbol           string        `json:"symbol"`
	InvestmentAmount float64       `json:"investment_amount"`
	Interval         time.Duration `json:"interval"`
	MaxInvestments   int           `json:"max_investments"`
	PriceThreshold   float64       `json:"price_threshold"`
	StopLoss         float64       `json:"stop_loss"`
	TakeProfit       float64       `json:"take_profit"`
	Enabled          bool          `json:"enabled"`
}

// UnmarshalJSON implements custom parsing for interval
func (d *DCAConfig) UnmarshalJSON(data []byte) error {
	type Alias DCAConfig
	aux := &struct {
		Interval string `json:"interval"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Interval != "" {
		duration, err := time.ParseDuration(aux.Interval)
		if err != nil {
			return fmt.Errorf("invalid interval format: %w", err)
		}
		d.Interval = duration
	}

	return nil
}

// GridConfig contains Grid strategy parameters
type GridConfig struct {
	Symbol             string  `json:"symbol"`
	UpperPrice         float64 `json:"upper_price"`
	LowerPrice         float64 `json:"lower_price"`
	GridLevels         int     `json:"grid_levels"`
	InvestmentPerLevel float64 `json:"investment_per_level"`
	Enabled            bool    `json:"enabled"`
}

// ComboConfig holds combined strategies configuration
type ComboConfig struct {
	Strategies []StrategyConfig `json:"strategies"`
	Enabled    bool             `json:"enabled"`
}

// StrategyConfig describes a strategy envelope
type StrategyConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// Portfolio represents a portfolio snapshot
type Portfolio struct {
	TotalValue  float64
	TotalProfit float64
	TotalLoss   float64
	NetProfit   float64
	Positions   []Position
	LastUpdate  time.Time
}

// Position represents a position
type Position struct {
	Symbol        string
	Quantity      float64
	AvgPrice      float64
	CurrentPrice  float64
	UnrealizedPnL float64
	RealizedPnL   float64
	Timestamp     time.Time
}

// ExchangeClient is the exchange interface used by strategies
type ExchangeClient interface {
	// Order management
	PlaceOrder(ctx context.Context, order Order) error
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	GetActiveOrders(ctx context.Context, symbol string) ([]Order, error)
	GetFilledOrders(ctx context.Context, symbol string) ([]Order, error)

	// Market data
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)
	GetOrderBook(ctx context.Context, symbol string, limit int) (*OrderBook, error)
	GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]Candle, error)

	// Account information
	GetBalance(ctx context.Context) (*Balance, error)
	GetTradingFees(ctx context.Context, symbol string) (*TradingFees, error)

	// Connection management
	Ping(ctx context.Context) error
	Close() error
}
