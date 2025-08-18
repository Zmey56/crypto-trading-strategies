package strategy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// GridStrategy представляет Grid торговую стратегию
type GridStrategy struct {
	config      types.GridConfig
	exchange    types.ExchangeClient
	logger      *logger.Logger
	metrics     *types.StrategyMetrics
	gridLevels  []float64
	activeOrders map[string]types.Order
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewGridStrategy создает новую Grid стратегию
func NewGridStrategy(config types.GridConfig, exchange types.ExchangeClient, logger *logger.Logger) *GridStrategy {
	ctx, cancel := context.WithCancel(context.Background())
	
	strategy := &GridStrategy{
		config:       config,
		exchange:     exchange,
		logger:       logger,
		metrics: &types.StrategyMetrics{
			LastUpdate: time.Now(),
		},
		activeOrders: make(map[string]types.Order),
		ctx:          ctx,
		cancel:       cancel,
	}
	
	strategy.calculateGridLevels()
	return strategy
}

// Execute выполняет Grid стратегию
func (g *GridStrategy) Execute(ctx context.Context, market types.MarketData) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Проверяем, включена ли стратегия
	if !g.config.Enabled {
		return nil
	}

	// Проверяем, находится ли цена в диапазоне сетки
	if market.Price < g.config.LowerPrice || market.Price > g.config.UpperPrice {
		g.logger.Warn("Цена %f вышла за пределы сетки [%f, %f]", 
			market.Price, g.config.LowerPrice, g.config.UpperPrice)
		return nil
	}

	// Находим ближайшие уровни сетки
	buyLevel, sellLevel := g.findNearestLevels(market.Price)
	
	// Проверяем возможность покупки
	if buyLevel > 0 && g.shouldPlaceBuyOrder(buyLevel, market.Price) {
		if err := g.placeBuyOrder(ctx, buyLevel, market); err != nil {
			g.logger.Error("Ошибка размещения ордера на покупку: %v", err)
			return err
		}
	}

	// Проверяем возможность продажи
	if sellLevel > 0 && g.shouldPlaceSellOrder(sellLevel, market.Price) {
		if err := g.placeSellOrder(ctx, sellLevel, market); err != nil {
			g.logger.Error("Ошибка размещения ордера на продажу: %v", err)
			return err
		}
	}

	return nil
}

// GetSignal генерирует торговый сигнал
func (g *GridStrategy) GetSignal(market types.MarketData) types.Signal {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.config.Enabled {
		return types.Signal{
			Type:      types.SignalTypeHold,
			Symbol:    market.Symbol,
			Price:     market.Price,
			Timestamp: market.Timestamp,
		}
	}

	// Проверяем диапазон цен
	if market.Price < g.config.LowerPrice || market.Price > g.config.UpperPrice {
		return types.Signal{
			Type:      types.SignalTypeHold,
			Symbol:    market.Symbol,
			Price:     market.Price,
			Timestamp: market.Timestamp,
		}
	}

	buyLevel, sellLevel := g.findNearestLevels(market.Price)
	
	if buyLevel > 0 && g.shouldPlaceBuyOrder(buyLevel, market.Price) {
		return types.Signal{
			Type:      types.SignalTypeBuy,
			Symbol:    market.Symbol,
			Price:     buyLevel,
			Quantity:  g.calculateOrderQuantity(buyLevel),
			Strength:  0.8,
			Timestamp: market.Timestamp,
			Metadata: map[string]interface{}{
				"grid_level": buyLevel,
				"strategy":   "grid",
			},
		}
	}

	if sellLevel > 0 && g.shouldPlaceSellOrder(sellLevel, market.Price) {
		return types.Signal{
			Type:      types.SignalTypeSell,
			Symbol:    market.Symbol,
			Price:     sellLevel,
			Quantity:  g.calculateOrderQuantity(sellLevel),
			Strength:  0.8,
			Timestamp: market.Timestamp,
			Metadata: map[string]interface{}{
				"grid_level": sellLevel,
				"strategy":   "grid",
			},
		}
	}

	return types.Signal{
		Type:      types.SignalTypeHold,
		Symbol:    market.Symbol,
		Price:     market.Price,
		Timestamp: market.Timestamp,
	}
}

// ValidateConfig проверяет корректность конфигурации
func (g *GridStrategy) ValidateConfig() error {
	if g.config.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if g.config.UpperPrice <= g.config.LowerPrice {
		return fmt.Errorf("upper price must be greater than lower price")
	}

	if g.config.GridLevels <= 0 {
		return fmt.Errorf("grid levels must be positive")
	}

	if g.config.InvestmentPerLevel <= 0 {
		return fmt.Errorf("investment per level must be positive")
	}

	return nil
}

// GetMetrics возвращает метрики стратегии
func (g *GridStrategy) GetMetrics() types.StrategyMetrics {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return *g.metrics
}

// Shutdown завершает работу стратегии
func (g *GridStrategy) Shutdown(ctx context.Context) error {
	g.cancel()
	
	// Отменяем все активные ордера
	for _, order := range g.activeOrders {
		if err := g.exchange.CancelOrder(ctx, order.ID); err != nil {
			g.logger.Error("Ошибка отмены ордера %s: %v", order.ID, err)
		}
	}
	
	g.logger.Info("Grid стратегия остановлена")
	return nil
}

// calculateGridLevels вычисляет уровни сетки
func (g *GridStrategy) calculateGridLevels() {
	g.gridLevels = make([]float64, g.config.GridLevels)
	
	// Арифметическая прогрессия
	step := (g.config.UpperPrice - g.config.LowerPrice) / float64(g.config.GridLevels-1)
	
	for i := 0; i < g.config.GridLevels; i++ {
		g.gridLevels[i] = g.config.LowerPrice + float64(i)*step
	}
	
	g.logger.Info("Создана сетка с %d уровнями от %.2f до %.2f", 
		g.config.GridLevels, g.config.LowerPrice, g.config.UpperPrice)
}

// findNearestLevels находит ближайшие уровни сетки
func (g *GridStrategy) findNearestLevels(price float64) (buyLevel, sellLevel float64) {
	// Если цена ниже нижней границы, не размещаем ордера
	if price < g.config.LowerPrice {
		return 0, 0
	}
	
	// Если цена выше верхней границы, не размещаем ордера
	if price > g.config.UpperPrice {
		return 0, 0
	}
	
	for i, level := range g.gridLevels {
		if price <= level {
			// Нашли уровень для покупки
			buyLevel = level
			if i < len(g.gridLevels)-1 {
				sellLevel = g.gridLevels[i+1]
			}
			break
		}
	}
	return
}

// shouldPlaceBuyOrder проверяет, нужно ли размещать ордер на покупку
func (g *GridStrategy) shouldPlaceBuyOrder(level, currentPrice float64) bool {
	// Проверяем, есть ли уже активный ордер на этом уровне
	for _, order := range g.activeOrders {
		if order.Price == level && order.Side == types.OrderSideBuy {
			return false
		}
	}
	
	// Размещаем ордер на покупку, если цена находится на уровне или ниже
	return currentPrice <= level
}

// shouldPlaceSellOrder проверяет, нужно ли размещать ордер на продажу
func (g *GridStrategy) shouldPlaceSellOrder(level, currentPrice float64) bool {
	// Проверяем, есть ли уже активный ордер на этом уровне
	for _, order := range g.activeOrders {
		if order.Price == level && order.Side == types.OrderSideSell {
			return false
		}
	}
	
	// Размещаем ордер на продажу, если цена находится на уровне или выше
	return currentPrice >= level
}

// placeBuyOrder размещает ордер на покупку
func (g *GridStrategy) placeBuyOrder(ctx context.Context, level float64, market types.MarketData) error {
	quantity := g.calculateOrderQuantity(level)
	
	order := types.Order{
		ID:        generateOrderID(),
		Symbol:    g.config.Symbol,
		Side:      types.OrderSideBuy,
		Type:      types.OrderTypeLimit,
		Quantity:  quantity,
		Price:     level,
		Status:    types.OrderStatusNew,
		Timestamp: time.Now(),
	}

	g.logger.Info("Размещаем Grid ордер на покупку: %s %.8f @ %.2f", 
		order.Symbol, order.Quantity, order.Price)

	if err := g.exchange.PlaceOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to place buy order: %w", err)
	}

	g.activeOrders[order.ID] = order
	g.updateMetrics(order, market.Price)

	return nil
}

// placeSellOrder размещает ордер на продажу
func (g *GridStrategy) placeSellOrder(ctx context.Context, level float64, market types.MarketData) error {
	quantity := g.calculateOrderQuantity(level)
	
	order := types.Order{
		ID:        generateOrderID(),
		Symbol:    g.config.Symbol,
		Side:      types.OrderSideSell,
		Type:      types.OrderTypeLimit,
		Quantity:  quantity,
		Price:     level,
		Status:    types.OrderStatusNew,
		Timestamp: time.Now(),
	}

	g.logger.Info("Размещаем Grid ордер на продажу: %s %.8f @ %.2f", 
		order.Symbol, order.Quantity, order.Price)

	if err := g.exchange.PlaceOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to place sell order: %w", err)
	}

	g.activeOrders[order.ID] = order
	g.updateMetrics(order, market.Price)

	return nil
}

// calculateOrderQuantity вычисляет количество для ордера
func (g *GridStrategy) calculateOrderQuantity(price float64) float64 {
	return g.config.InvestmentPerLevel / price
}

// updateMetrics обновляет метрики стратегии
func (g *GridStrategy) updateMetrics(order types.Order, price float64) {
	g.metrics.TotalTrades++
	g.metrics.TotalVolume += order.Quantity * price
	g.metrics.LastUpdate = time.Now()
}

// GetStatus возвращает статус стратегии
func (g *GridStrategy) GetStatus() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return map[string]interface{}{
		"enabled":              g.config.Enabled,
		"symbol":               g.config.Symbol,
		"grid_levels":          len(g.gridLevels),
		"active_orders":        len(g.activeOrders),
		"lower_price":          g.config.LowerPrice,
		"upper_price":          g.config.UpperPrice,
		"investment_per_level": g.config.InvestmentPerLevel,
		"total_trades":         g.metrics.TotalTrades,
		"total_volume":         g.metrics.TotalVolume,
	}
}

// generateOrderID генерирует уникальный ID для ордера
func generateOrderID() string {
	return fmt.Sprintf("grid_%d", time.Now().UnixNano())
}
