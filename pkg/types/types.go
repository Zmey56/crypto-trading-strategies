package types

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// MarketData представляет рыночные данные
type MarketData struct {
	Symbol    string
	Price     float64
	Volume    float64
	Timestamp time.Time
	Ticker    *Ticker
	OrderBook *OrderBook
	Candles   []Candle
}

// Ticker представляет тикер с текущими ценами
type Ticker struct {
	Symbol    string
	Price     float64
	Bid       float64
	Ask       float64
	Volume    float64
	Timestamp time.Time
}

// OrderBook представляет книгу ордеров
type OrderBook struct {
	Symbol string
	Bids   []OrderBookEntry
	Asks   []OrderBookEntry
}

// OrderBookEntry представляет запись в книге ордеров
type OrderBookEntry struct {
	Price  float64
	Amount float64
}

// Candle представляет свечу
type Candle struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Timestamp time.Time
}

// Order представляет торговый ордер
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

// OrderSide представляет сторону ордера
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType представляет тип ордера
type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

// OrderStatus представляет статус ордера
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusRejected        OrderStatus = "REJECTED"
)

// ExchangeOrder представляет ордер на бирже
type ExchangeOrder struct {
	ExchangeOrderID string
	Exchange        string
	ClientOrderID   string
}

// OrderUpdate представляет обновление ордера
type OrderUpdate struct {
	OrderID       string
	Status        OrderStatus
	FilledAmount  float64
	FilledPrice   float64
	Timestamp     time.Time
	ExchangeOrder *ExchangeOrder
}

// Balance представляет баланс аккаунта
type Balance struct {
	Asset     string
	Free      float64
	Locked    float64
	Total     float64
	Timestamp time.Time
}

// TradingFees представляет торговые комиссии
type TradingFees struct {
	Symbol    string
	MakerFee  float64
	TakerFee  float64
	Timestamp time.Time
}

// Signal представляет торговый сигнал
type Signal struct {
	Type      SignalType
	Symbol    string
	Price     float64
	Quantity  float64
	Strength  float64
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// SignalType представляет тип сигнала
type SignalType string

const (
	SignalTypeBuy  SignalType = "BUY"
	SignalTypeSell SignalType = "SELL"
	SignalTypeHold SignalType = "HOLD"
)

// StrategyMetrics представляет метрики стратегии
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

// DCAConfig представляет конфигурацию DCA стратегии
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

// UnmarshalJSON реализует кастомный парсинг JSON для DCAConfig
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

// GridConfig представляет конфигурацию Grid стратегии
type GridConfig struct {
	Symbol             string  `json:"symbol"`
	UpperPrice         float64 `json:"upper_price"`
	LowerPrice         float64 `json:"lower_price"`
	GridLevels         int     `json:"grid_levels"`
	InvestmentPerLevel float64 `json:"investment_per_level"`
	Enabled            bool    `json:"enabled"`
}

// ComboConfig представляет конфигурацию комбинированной стратегии
type ComboConfig struct {
	Strategies []StrategyConfig `json:"strategies"`
	Enabled    bool             `json:"enabled"`
}

// StrategyConfig представляет базовую конфигурацию стратегии
type StrategyConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// Portfolio представляет портфель
type Portfolio struct {
	TotalValue  float64
	TotalProfit float64
	TotalLoss   float64
	NetProfit   float64
	Positions   []Position
	LastUpdate  time.Time
}

// Position представляет позицию
type Position struct {
	Symbol        string
	Quantity      float64
	AvgPrice      float64
	CurrentPrice  float64
	UnrealizedPnL float64
	RealizedPnL   float64
	Timestamp     time.Time
}

// ExchangeClient представляет интерфейс для работы с биржей
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
