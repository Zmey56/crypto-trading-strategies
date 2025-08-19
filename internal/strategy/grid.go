package strategy

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// GridStrategy is a simple grid trading implementation with evenly spaced levels
type GridStrategy struct {
	config   types.GridConfig
	exchange types.ExchangeClient
	logger   *logger.Logger

	mu        sync.RWMutex
	levels    []float64                // sorted levels (low -> high)
	positions map[float64]gridPosition // position size per level

	metrics types.StrategyMetrics
}

type gridPosition struct {
	quantity float64
	avgPrice float64
}

func NewGridStrategy(config types.GridConfig, exchange types.ExchangeClient, logger *logger.Logger) (*GridStrategy, error) {
	if config.GridLevels < 2 {
		return nil, fmt.Errorf("grid levels must be >= 2")
	}
	gs := &GridStrategy{
		config:    config,
		exchange:  exchange,
		logger:    logger,
		positions: make(map[float64]gridPosition),
	}
	gs.buildLevels()
	return gs, nil
}

func (g *GridStrategy) buildLevels() {
	step := (g.config.UpperPrice - g.config.LowerPrice) / float64(g.config.GridLevels-1)
	levels := make([]float64, g.config.GridLevels)
	for i := 0; i < g.config.GridLevels; i++ {
		levels[i] = g.config.LowerPrice + float64(i)*step
	}
	sort.Float64s(levels)
	g.levels = levels
}

func (g *GridStrategy) ValidateConfig() error {
	if g.config.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if g.config.LowerPrice <= 0 || g.config.UpperPrice <= g.config.LowerPrice {
		return fmt.Errorf("invalid grid bounds")
	}
	if g.config.GridLevels <= 1 {
		return fmt.Errorf("grid levels must be > 1")
	}
	if g.config.InvestmentPerLevel <= 0 {
		return fmt.Errorf("investment per level must be positive")
	}
	return nil
}

func (g *GridStrategy) Execute(ctx context.Context, market types.MarketData) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.config.Enabled {
		return nil
	}

	price := market.Price
	// BUY when price crosses down to or below a level with empty position
	for i, level := range g.levels {
		pos := g.positions[level]
		if price <= level && pos.quantity == 0 {
			qty := g.config.InvestmentPerLevel / price
			order := types.Order{Symbol: g.config.Symbol, Side: types.OrderSideBuy, Type: types.OrderTypeMarket, Quantity: qty, Price: price, Status: types.OrderStatusNew, Timestamp: time.Now()}
			if err := g.exchange.PlaceOrder(ctx, order); err != nil {
				return fmt.Errorf("grid buy failed: %w", err)
			}
			g.positions[level] = gridPosition{quantity: qty, avgPrice: price}
			g.metrics.TotalTrades++
			g.metrics.TotalVolume += qty * price
			g.logger.Info("Grid BUY @ level %.2f qty=%.8f price=%.2f", level, qty, price)
		}

		// SELL when price reaches next higher level and we have a position at current level
		if pos.quantity > 0 && i+1 < len(g.levels) {
			nextLevel := g.levels[i+1]
			if price >= nextLevel {
				qty := pos.quantity
				order := types.Order{Symbol: g.config.Symbol, Side: types.OrderSideSell, Type: types.OrderTypeMarket, Quantity: qty, Price: price, Status: types.OrderStatusNew, Timestamp: time.Now()}
				if err := g.exchange.PlaceOrder(ctx, order); err != nil {
					return fmt.Errorf("grid sell failed: %w", err)
				}
				realized := (price - pos.avgPrice) * qty
				g.metrics.TotalTrades++
				g.metrics.TotalVolume += qty * price
				if realized >= 0 {
					g.metrics.WinningTrades++
					g.metrics.TotalProfit += realized
				} else {
					g.metrics.LosingTrades++
					g.metrics.TotalLoss += -realized
				}
				g.positions[level] = gridPosition{}
				g.logger.Info("Grid SELL from level %.2f qty=%.8f price=%.2f pnl=%.2f", level, qty, price, realized)
			}
		}
	}

	g.metrics.LastUpdate = time.Now()
	if g.metrics.TotalTrades > 0 {
		totalWins := float64(g.metrics.WinningTrades)
		totalTrades := float64(g.metrics.TotalTrades)
		g.metrics.WinRate = (totalWins / totalTrades) * 100.0
		if g.metrics.TotalLoss > 0 {
			g.metrics.ProfitFactor = g.metrics.TotalProfit / g.metrics.TotalLoss
		}
	}
	return nil
}

func (g *GridStrategy) GetSignal(market types.MarketData) types.Signal {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// For simplicity, return HOLD; the strategy performs execution inside Execute
	return types.Signal{Type: types.SignalTypeHold, Symbol: market.Symbol, Price: market.Price, Timestamp: market.Timestamp}
}

func (g *GridStrategy) GetMetrics() types.StrategyMetrics {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.metrics
}

func (g *GridStrategy) Shutdown(ctx context.Context) error {
	g.logger.Info("Grid strategy stopped")
	return nil
}
