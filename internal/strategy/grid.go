package strategy

import (
	"context"
	"crypto-trading-strategies/pkg/types"
	"crypto-trading-strategies/pkg/utils"
	"math"
	"sync"
)

type GridStrategy struct {
	// Основные параметры сетки
	Symbol     string   `json:"symbol"`
	UpperBound float64  `json:"upper_bound"` // Верхняя граница диапазона
	LowerBound float64  `json:"lower_bound"` // Нижняя граница диапазона
	GridLevels int      `json:"grid_levels"` // Количество уровней сетки
	OrderSize  float64  `json:"order_size"`  // Размер каждого ордера
	GridType   GridType `json:"grid_type"`   // ARITHMETIC или GEOMETRIC

	// Внутреннее состояние
	ActiveOrders map[float64]types.Order `json:"active_orders"`
	GridSpacing  float64                 `json:"grid_spacing"`
	TotalProfit  float64                 `json:"total_profit"`
	TradeCount   int                     `json:"trade_count"`

	// Синхронизация для concurrent access
	mutex sync.RWMutex

	// Зависимости
	portfolioManager *portfolio.Manager
	riskManager      RiskManager
}

type GridType string

const (
	ARITHMETIC GridType = "arithmetic"
	GEOMETRIC  GridType = "geometric"
)

// Execute реализует основную логику Grid стратегии
func (g *GridStrategy) Execute(ctx context.Context, market types.MarketData) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Проверка, находится ли цена в пределах сетки
	if !g.isPriceInRange(market.Price) {
		return g.handlePriceOutOfRange(ctx, market)
	}

	// Выполнение сделок при достижении уровней сетки
	return g.processGridLevels(ctx, market)
}

// calculateGridLevels вычисляет все уровни сетки
func (g *GridStrategy) calculateGridLevels() []float64 {
	levels := make([]float64, g.GridLevels)

	switch g.GridType {
	case ARITHMETIC:
		g.GridSpacing = (g.UpperBound - g.LowerBound) / float64(g.GridLevels-1)
		for i := 0; i < g.GridLevels; i++ {
			levels[i] = g.LowerBound + float64(i)*g.GridSpacing
		}

	case GEOMETRIC:
		ratio := math.Pow(g.UpperBound/g.LowerBound, 1.0/float64(g.GridLevels-1))
		for i := 0; i < g.GridLevels; i++ {
			levels[i] = g.LowerBound * math.Pow(ratio, float64(i))
		}
	}

	return levels
}

// processGridLevels обрабатывает активацию уровней сетки
func (g *GridStrategy) processGridLevels(ctx context.Context, market types.MarketData) error {
	levels := g.calculateGridLevels()
	currentPrice := market.Price

	for _, level := range levels {
		// Проверка активации уровня покупки
		if g.shouldBuyAtLevel(level, currentPrice) {
			if err := g.placeBuyOrder(ctx, level); err != nil {
				return err
			}
		}

		// Проверка активации уровня продажи
		if g.shouldSellAtLevel(level, currentPrice) {
			if err := g.placeSellOrder(ctx, level); err != nil {
				return err
			}
		}
	}

	return nil
}
