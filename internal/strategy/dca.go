package strategy

import (
	"context"
	"crypto-trading-strategies/internal/logger"
	"crypto-trading-strategies/pkg/types"
	"fmt"
	"sync"
	"time"
)

// DCAStrategy представляет DCA стратегию
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

// NewDCAStrategy создает новую DCA стратегию
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

// Execute выполняет DCA стратегию
func (d *DCAStrategy) Execute(ctx context.Context, market types.MarketData) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Проверяем, включена ли стратегия
	if !d.config.Enabled {
		return nil
	}

	// Проверяем интервал между покупками
	if time.Since(d.lastBuy) < d.config.Interval {
		return nil
	}

	// Проверяем максимальное количество инвестиций
	if d.buyCount >= d.config.MaxInvestments {
		d.logger.Info("Достигнуто максимальное количество инвестиций для %s", d.config.Symbol)
		return nil
	}

	// Проверяем условия для покупки
	if d.config.PriceThreshold > 0 && market.Price > d.config.PriceThreshold {
		return nil
	}

	// Выполняем покупку
	if err := d.executeBuy(ctx, market); err != nil {
		d.logger.Error("Ошибка при выполнении покупки: %v", err)
		return err
	}

	return nil
}

// GetSignal генерирует торговый сигнал
func (d *DCAStrategy) GetSignal(market types.MarketData) types.Signal {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Проверяем порог цены
	if d.config.PriceThreshold > 0 && market.Price > d.config.PriceThreshold {
		return types.Signal{
			Type:      types.SignalTypeHold,
			Symbol:    market.Symbol,
			Price:     market.Price,
			Timestamp: market.Timestamp,
		}
	}

	// Проверяем интервал
	if time.Since(d.lastBuy) < d.config.Interval {
		return types.Signal{
			Type:      types.SignalTypeHold,
			Symbol:    market.Symbol,
			Price:     market.Price,
			Timestamp: market.Timestamp,
		}
	}

	// Проверяем максимальное количество инвестиций
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

// ValidateConfig проверяет корректность конфигурации
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

// GetMetrics возвращает метрики стратегии
func (d *DCAStrategy) GetMetrics() types.StrategyMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return *d.metrics
}

// Shutdown завершает работу стратегии
func (d *DCAStrategy) Shutdown(ctx context.Context) error {
	d.cancel()
	d.logger.Info("DCA стратегия остановлена")
	return nil
}

// executeBuy выполняет покупку
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

	// Обновляем метрики
	d.lastBuy = time.Now()
	d.buyCount++
	d.updateMetrics(order, market.Price)

	d.logger.Info("DCA покупка выполнена: %s %.8f @ %.2f (покупка #%d)",
		order.Symbol, order.Quantity, order.Price, d.buyCount)

	return nil
}

// calculateQuantity вычисляет количество для покупки
func (d *DCAStrategy) calculateQuantity(price float64) float64 {
	return d.config.InvestmentAmount / price
}

// updateMetrics обновляет метрики стратегии
func (d *DCAStrategy) updateMetrics(order types.Order, price float64) {
	d.metrics.TotalTrades++
	d.metrics.TotalVolume += order.Quantity * price
	d.metrics.LastUpdate = time.Now()

	// В DCA стратегии мы не рассчитываем прибыль/убыток до продажи
	// но можем отслеживать общий объем торгов
}

// GetConfig возвращает конфигурацию стратегии
func (d *DCAStrategy) GetConfig() types.DCAConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config
}

// UpdateConfig обновляет конфигурацию стратегии
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

// validateConfig проверяет конфигурацию
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

// GetStatus возвращает статус стратегии
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
