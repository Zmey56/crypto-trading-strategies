package strategy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// DCAStrategy implements a basic Dollar-Cost Averaging strategy
type DCAStrategy struct {
	config   types.DCAConfig
	exchange types.ExchangeClient
	logger   *logger.Logger
	metrics  *types.StrategyMetrics
	lastBuy  time.Time
	buyCount int
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewDCAStrategy creates a new DCA strategy instance
func NewDCAStrategy(config types.DCAConfig, exchange types.ExchangeClient, logger *logger.Logger) *DCAStrategy {
	ctx, cancel := context.WithCancel(context.Background())

	return &DCAStrategy{
		config:   config,
		exchange: exchange,
		logger:   logger,
		metrics: &types.StrategyMetrics{
			LastUpdate: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Execute runs the DCA logic
func (d *DCAStrategy) Execute(ctx context.Context, market types.MarketData) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if strategy is enabled
	if !d.config.Enabled {
		return nil
	}

	// Enforce interval between buys
	if time.Since(d.lastBuy) < d.config.Interval {
		return nil
	}

	// Respect max number of investments
	if d.buyCount >= d.config.MaxInvestments {
		d.logger.Info("Достигнуто максимальное количество инвестиций для %s", d.config.Symbol)
		return nil
	}

	// Optional price threshold
	if d.config.PriceThreshold > 0 && market.Price > d.config.PriceThreshold {
		return nil
	}

	// Execute buy
	if err := d.executeBuy(ctx, market); err != nil {
		d.logger.Error("Ошибка при выполнении покупки: %v", err)
		return err
	}

	return nil
}

// GetSignal produces a trading signal (for observability)
func (d *DCAStrategy) GetSignal(market types.MarketData) types.Signal {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Check threshold
	if d.config.PriceThreshold > 0 && market.Price > d.config.PriceThreshold {
		return types.Signal{
			Type:      types.SignalTypeHold,
			Symbol:    market.Symbol,
			Price:     market.Price,
			Timestamp: market.Timestamp,
		}
	}

	// Check interval
	if time.Since(d.lastBuy) < d.config.Interval {
		return types.Signal{
			Type:      types.SignalTypeHold,
			Symbol:    market.Symbol,
			Price:     market.Price,
			Timestamp: market.Timestamp,
		}
	}

	// Check max investments
	if d.buyCount >= d.config.MaxInvestments {
		return types.Signal{
			Type:      types.SignalTypeHold,
			Symbol:    market.Symbol,
			Price:     market.Price,
			Timestamp: market.Timestamp,
		}
	}

	return types.Signal{
		Type:      types.SignalTypeBuy,
		Symbol:    market.Symbol,
		Price:     market.Price,
		Quantity:  d.calculateQuantity(market.Price),
		Strength:  1.0,
		Timestamp: market.Timestamp,
		Metadata: map[string]interface{}{
			"buy_count": d.buyCount + 1,
			"interval":  d.config.Interval.String(),
		},
	}
}

// ValidateConfig validates configuration
func (d *DCAStrategy) ValidateConfig() error {
	if d.config.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if d.config.InvestmentAmount <= 0 {
		return fmt.Errorf("investment amount must be positive")
	}

	if d.config.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	if d.config.MaxInvestments <= 0 {
		return fmt.Errorf("max investments must be positive")
	}

	return nil
}

// GetMetrics returns strategy metrics snapshot
func (d *DCAStrategy) GetMetrics() types.StrategyMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return *d.metrics
}

// Shutdown gracefully stops the strategy
func (d *DCAStrategy) Shutdown(ctx context.Context) error {
	d.cancel()
	d.logger.Info("DCA стратегия остановлена")
	return nil
}

// executeBuy places a market buy and updates metrics
func (d *DCAStrategy) executeBuy(ctx context.Context, market types.MarketData) error {
	quantity := d.calculateQuantity(market.Price)

	order := types.Order{
		Symbol:    d.config.Symbol,
		Side:      types.OrderSideBuy,
		Type:      types.OrderTypeMarket,
		Quantity:  quantity,
		Price:     market.Price,
		Status:    types.OrderStatusNew,
		Timestamp: time.Now(),
	}

	d.logger.Info("Размещаем DCA ордер: %s %.8f @ %.2f",
		order.Symbol, order.Quantity, order.Price)

	if err := d.exchange.PlaceOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to place order: %w", err)
	}

	// Update metrics
	d.lastBuy = time.Now()
	d.buyCount++
	d.updateMetrics(order, market.Price)

	d.logger.Info("DCA покупка выполнена: %s %.8f @ %.2f (покупка #%d)",
		order.Symbol, order.Quantity, order.Price, d.buyCount)

	return nil
}

// calculateQuantity computes buy quantity by fixed investment amount
func (d *DCAStrategy) calculateQuantity(price float64) float64 {
	return d.config.InvestmentAmount / price
}

// updateMetrics updates strategy metrics counters
func (d *DCAStrategy) updateMetrics(order types.Order, price float64) {
	d.metrics.TotalTrades++
	d.metrics.TotalVolume += order.Quantity * price
	d.metrics.LastUpdate = time.Now()

	// In DCA we do not compute PnL until selling; track total volume only
}

// GetConfig returns current strategy config
func (d *DCAStrategy) GetConfig() types.DCAConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config
}

// UpdateConfig updates strategy config with validation
func (d *DCAStrategy) UpdateConfig(config types.DCAConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.validateConfig(config); err != nil {
		return err
	}

	d.config = config
	d.logger.Info("Конфигурация DCA стратегии обновлена")
	return nil
}

// validateConfig validates config struct
func (d *DCAStrategy) validateConfig(config types.DCAConfig) error {
	if config.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if config.InvestmentAmount <= 0 {
		return fmt.Errorf("investment amount must be positive")
	}

	if config.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	if config.MaxInvestments <= 0 {
		return fmt.Errorf("max investments must be positive")
	}

	return nil
}

// GetStatus returns strategy status map for API
func (d *DCAStrategy) GetStatus() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"enabled":           d.config.Enabled,
		"symbol":            d.config.Symbol,
		"buy_count":         d.buyCount,
		"max_buys":          d.config.MaxInvestments,
		"last_buy":          d.lastBuy,
		"next_buy":          d.lastBuy.Add(d.config.Interval),
		"interval":          d.config.Interval.String(),
		"investment_amount": d.config.InvestmentAmount,
	}
}
