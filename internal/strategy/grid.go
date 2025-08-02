package strategy

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"crypto-trading-strategies/pkg/indicators"
	"crypto-trading-strategies/pkg/types"
)

type GridStrategy struct {
	config         types.GridConfig
	gridLevels     []float64
	activeOrders   map[float64]types.Order
	completedPairs map[string]GridTradePair
	totalProfit    float64
	tradeCount     int

	// Advanced features
	atrIndicator      *indicators.ATRIndicator
	volatilityAdapter *VolatilityAdapter
	profitTracker     *ProfitTracker

	// Dependencies
	exchange         exchange.Client
	portfolioManager *portfolio.Manager
	riskManager      *risk.Manager
	logger           *logger.Logger

	// Synchronization
	mutex sync.RWMutex

	// Performance optimization
	priceCache      map[float64]time.Time
	lastUpdateTime  time.Time
	updateThreshold time.Duration
}

type GridTradePair struct {
	BuyOrder  types.Order
	SellOrder types.Order
	Profit    float64
	Timestamp time.Time
}

type VolatilityAdapter struct {
	atrPeriod      int
	adaptiveFactor float64
	minSpacing     float64
	maxSpacing     float64
}

func NewGridStrategy(config types.GridConfig, deps StrategyDependencies) *GridStrategy {
	grid := &GridStrategy{
		config:           config,
		activeOrders:     make(map[float64]types.Order),
		completedPairs:   make(map[string]GridTradePair),
		exchange:         deps.Exchange,
		portfolioManager: deps.PortfolioManager,
		riskManager:      deps.RiskManager,
		logger:           deps.Logger,
		priceCache:       make(map[float64]time.Time),
		updateThreshold:  time.Second * 5,
	}

	grid.atrIndicator = indicators.NewATR(14)
	grid.volatilityAdapter = &VolatilityAdapter{
		atrPeriod:      14,
		adaptiveFactor: 0.15,
		minSpacing:     config.MinGridSpacing,
		maxSpacing:     config.MaxGridSpacing,
	}

	grid.calculateGridLevels()
	return grid
}

func (g *GridStrategy) Execute(ctx context.Context, market types.MarketData) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Throttle updates to prevent excessive API calls
	if time.Since(g.lastUpdateTime) < g.updateThreshold {
		return nil
	}
	g.lastUpdateTime = time.Now()

	if err := g.validatePriceRange(market.Price); err != nil {
		return g.handlePriceOutOfRange(ctx, market)
	}

	// Update volatility-based spacing if enabled
	if g.config.AdaptiveSpacing {
		g.updateAdaptiveSpacing(market)
	}

	return g.processGridLevels(ctx, market)
}

func (g *GridStrategy) calculateGridLevels() {
	g.gridLevels = make([]float64, g.config.GridLevels)

	switch g.config.GridType {
	case types.GridTypeArithmetic:
		spacing := (g.config.UpperBound - g.config.LowerBound) / float64(g.config.GridLevels-1)
		for i := 0; i < g.config.GridLevels; i++ {
			g.gridLevels[i] = g.config.LowerBound + float64(i)*spacing
		}
	case types.GridTypeGeometric:
		ratio := math.Pow(g.config.UpperBound/g.config.LowerBound, 1.0/float64(g.config.GridLevels-1))
		for i := 0; i < g.config.GridLevels; i++ {
			g.gridLevels[i] = g.config.LowerBound * math.Pow(ratio, float64(i))
		}
	case types.GridTypeAdaptive:
		g.calculateAdaptiveGrid()
	}

	sort.Float64s(g.gridLevels)
}

func (g *GridStrategy) processGridLevels(ctx context.Context, market types.MarketData) error {
	currentPrice := market.Price

	for _, level := range g.gridLevels {
		// Process buy opportunities
		if g.shouldPlaceBuyOrder(level, currentPrice) {
			if err := g.placeBuyOrder(ctx, level, market); err != nil {
				g.logger.Error("Failed to place buy order",
					"level", level,
					"error", err)
				continue
			}
		}

		// Process sell opportunities
		if g.shouldPlaceSellOrder(level, currentPrice) {
			if err := g.placeSellOrder(ctx, level, market); err != nil {
				g.logger.Error("Failed to place sell order",
					"level", level,
					"error", err)
				continue
			}
		}
	}

	return g.processFilledOrders(ctx)
}

func (g *GridStrategy) shouldPlaceBuyOrder(level, currentPrice float64) bool {
	// Check if price is at or below level and no active buy order exists
	if currentPrice > level+g.getToleranceRange(level) {
		return false
	}

	if _, exists := g.activeOrders[level]; exists {
		return false
	}

	return true
}

func (g *GridStrategy) placeBuyOrder(ctx context.Context, level float64, market types.MarketData) error {
	orderSize := g.calculateOrderSize(level)

	if !g.riskManager.CanTrade(orderSize, g.config.Symbol) {
		return ErrRiskLimitsExceeded
	}

	order := types.Order{
		ID:         utils.GenerateOrderID(),
		Symbol:     g.config.Symbol,
		Side:       types.BUY,
		Type:       types.LIMIT,
		Quantity:   orderSize / level,
		Price:      level,
		Status:     types.OrderStatusPending,
		Timestamp:  time.Now(),
		StrategyID: g.config.ID,
		GridLevel:  level,
	}

	if err := g.exchange.PlaceOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to place buy order at level %.2f: %w", level, err)
	}

	g.activeOrders[level] = order

	g.logger.Info("Grid buy order placed",
		"symbol", g.config.Symbol,
		"level", level,
		"quantity", order.Quantity)

	return nil
}

func (g *GridStrategy) processFilledOrders(ctx context.Context) error {
	filledOrders, err := g.exchange.GetFilledOrders(ctx, g.config.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get filled orders: %w", err)
	}

	for _, order := range filledOrders {
		if err := g.handleFilledOrder(ctx, order); err != nil {
			g.logger.Error("Failed to handle filled order",
				"orderID", order.ID,
				"error", err)
		}
	}

	return nil
}

func (g *GridStrategy) handleFilledOrder(ctx context.Context, order types.Order) error {
	level := order.GridLevel

	if order.Side == types.BUY {
		// Place corresponding sell order
		sellPrice := level * (1 + g.config.TakeProfitPercent/100)
		return g.placeSellOrderAtLevel(ctx, sellPrice, order)
	} else {
		// Complete the trade pair and calculate profit
		return g.completeGridTradePair(order)
	}
}

func (g *GridStrategy) updateAdaptiveSpacing(market types.MarketData) {
	atrValue := g.atrIndicator.Calculate(market.Candles)
	if atrValue == 0 {
		return
	}

	optimalSpacing := atrValue * g.volatilityAdapter.adaptiveFactor

	// Constrain spacing within bounds
	optimalSpacing = math.Max(optimalSpacing, g.volatilityAdapter.minSpacing)
	optimalSpacing = math.Min(optimalSpacing, g.volatilityAdapter.maxSpacing)

	// Recalculate grid levels with new spacing
	newLevels := int((g.config.UpperBound - g.config.LowerBound) / optimalSpacing)
	if newLevels != g.config.GridLevels {
		g.config.GridLevels = newLevels
		g.calculateGridLevels()

		g.logger.Info("Grid levels updated based on volatility",
			"atrValue", atrValue,
			"newSpacing", optimalSpacing,
			"newLevels", newLevels)
	}
}
