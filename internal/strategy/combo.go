package strategy

import (
	"context"
	"crypto-trading-strategies/pkg/types"
	"fmt"
	"github.com/Zmey56/crypto-trading-strategies/internal/exchange"
	"time"
)

type ComboStrategy struct {
	dcaEngine   *DCAEngine
	gridEngine  *GridEngine
	minigrids   map[float64]*Minigrid
	currentMode ComboMode

	// Параметры гибридной стратегии
	config       ComboConfig
	stateManager *ComboStateManager

	// Зависимости
	exchange         exchange.Client
	portfolioManager *portfolio.Manager
	riskManager      *risk.Manager
	logger           *logger.Logger
}

type ComboMode int

const (
	ComboModeDCA ComboMode = iota
	ComboModeGrid
	ComboModeHybrid
)

type Minigrid struct {
	ID           string           `json:"id"`
	Range        types.PriceRange `json:"range"`
	Orders       []types.Order    `json:"orders"`
	Status       MinigridStatus   `json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
	ProfitTarget float64          `json:"profit_target"`
}

func (c *ComboStrategy) Execute(ctx context.Context, market types.MarketData) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Анализ рыночных условий для определения оптимального режима
	optimalMode := c.determineOptimalMode(market)

	if c.shouldSwitchMode(optimalMode) {
		if err := c.switchMode(ctx, optimalMode); err != nil {
			return fmt.Errorf("failed to switch mode: %w", err)
		}
	}

	return c.executeCurrentMode(ctx, market)
}

func (c *ComboStrategy) determineOptimalMode(market types.MarketData) ComboMode {
	volatility := c.calculateVolatility(market.Candles)
	trend := c.analyzeTrend(market)

	// Логика выбора режима на основе рыночных условий
	switch {
	case volatility > c.config.HighVolatilityThreshold && trend == types.TrendSideways:
		return ComboModeGrid
	case trend == types.TrendBearish:
		return ComboModeDCA
	default:
		return ComboModeHybrid
	}
}

func (c *ComboStrategy) createMinigrid(dcaLevel float64, market types.MarketData) (*Minigrid, error) {
	gridRange := types.PriceRange{
		Lower: dcaLevel * (1 - c.config.MinigridSpread/100),
		Upper: dcaLevel * (1 + c.config.MinigridSpread/100),
	}

	minigrid := &Minigrid{
		ID:           utils.GenerateMinigridID(),
		Range:        gridRange,
		Status:       MinigridStatusActive,
		CreatedAt:    time.Now(),
		ProfitTarget: c.config.MinigridProfitTarget,
	}

	// Создание grid-ордеров внутри minigrid
	orders, err := c.generateMinigridOrders(minigrid, market)
	if err != nil {
		return nil, fmt.Errorf("failed to generate minigrid orders: %w", err)
	}

	minigrid.Orders = orders
	c.minigrids[dcaLevel] = minigrid

	c.logger.Info("Minigrid created",
		"id", minigrid.ID,
		"range", gridRange,
		"orders", len(orders))

	return minigrid, nil
}
