package strategy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"crypto-trading-strategies/pkg/types"
	"crypto-trading-strategies/pkg/utils"
)

type DCAStrategy struct {
	config           types.DCAConfig
	state            DCAState
	currentOrders    []types.Order
	averagePrice     float64
	totalInvested    float64
	totalQuantity    float64
	safetyOrderCount int

	// Dependencies
	exchange         exchange.Client
	portfolioManager *portfolio.Manager
	riskManager      *risk.Manager
	logger           *logger.Logger

	// Synchronization
	mutex sync.RWMutex

	// Channels for coordination
	signalChan chan types.Signal
	stopChan   chan struct{}
}

type DCAState int

const (
	DCAStateInactive DCAState = iota
	DCAStateActive
	DCAStateProfitTaking
	DCAStateCompleted
)

func NewDCAStrategy(config types.DCAConfig, deps StrategyDependencies) *DCAStrategy {
	return &DCAStrategy{
		config:           config,
		state:            DCAStateInactive,
		exchange:         deps.Exchange,
		portfolioManager: deps.PortfolioManager,
		riskManager:      deps.RiskManager,
		logger:           deps.Logger,
		signalChan:       make(chan types.Signal, 100),
		stopChan:         make(chan struct{}),
	}
}

func (d *DCAStrategy) Execute(ctx context.Context, market types.MarketData) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.validateMarketConditions(market); err != nil {
		return fmt.Errorf("market validation failed: %w", err)
	}

	switch d.state {
	case DCAStateInactive:
		return d.initiateBaseOrder(ctx, market)
	case DCAStateActive:
		return d.processActiveState(ctx, market)
	case DCAStateProfitTaking:
		return d.processProfitTaking(ctx, market)
	default:
		return nil
	}
}

func (d *DCAStrategy) initiateBaseOrder(ctx context.Context, market types.MarketData) error {
	if !d.riskManager.CanTrade(d.config.BaseOrderSize, d.config.Symbol) {
		return ErrRiskLimitsExceeded
	}

	orderSize := d.calculateOptimalOrderSize(market)

	order := types.Order{
		ID:         utils.GenerateOrderID(),
		Symbol:     d.config.Symbol,
		Side:       types.BUY,
		Type:       types.MARKET,
		Quantity:   orderSize / market.Price,
		Price:      market.Price,
		Status:     types.OrderStatusPending,
		Timestamp:  time.Now(),
		StrategyID: d.config.ID,
	}

	if err := d.exchange.PlaceOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to place base order: %w", err)
	}

	d.updatePosition(order)
	d.state = DCAStateActive

	d.logger.Info("DCA base order placed",
		"symbol", d.config.Symbol,
		"quantity", order.Quantity,
		"price", order.Price)

	return nil
}

func (d *DCAStrategy) processActiveState(ctx context.Context, market types.MarketData) error {
	// Check for safety order conditions
	if d.shouldPlaceSafetyOrder(market.Price) {
		return d.placeSafetyOrder(ctx, market)
	}

	// Check for profit taking conditions
	if d.shouldTakeProfit(market.Price) {
		return d.initiateProfitTaking(ctx, market)
	}

	return nil
}

func (d *DCAStrategy) shouldPlaceSafetyOrder(currentPrice float64) bool {
	if d.safetyOrderCount >= d.config.SafetyOrdersCount {
		return false
	}

	priceDeviationThreshold := d.averagePrice * (1 - d.config.PriceDeviation/100)
	return currentPrice <= priceDeviationThreshold
}

func (d *DCAStrategy) placeSafetyOrder(ctx context.Context, market types.MarketData) error {
	// Progressive safety order sizing
	orderSize := d.config.SafetyOrderSize * d.calculateSafetyOrderMultiplier()

	if !d.riskManager.CanTrade(orderSize, d.config.Symbol) {
		d.logger.Warn("Safety order blocked by risk manager",
			"symbol", d.config.Symbol,
			"orderSize", orderSize)
		return nil
	}

	order := types.Order{
		ID:         utils.GenerateOrderID(),
		Symbol:     d.config.Symbol,
		Side:       types.BUY,
		Type:       types.MARKET,
		Quantity:   orderSize / market.Price,
		Price:      market.Price,
		Status:     types.OrderStatusPending,
		Timestamp:  time.Now(),
		StrategyID: d.config.ID,
		OrderTag:   fmt.Sprintf("safety_%d", d.safetyOrderCount+1),
	}

	if err := d.exchange.PlaceOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to place safety order: %w", err)
	}

	d.updatePosition(order)
	d.safetyOrderCount++

	d.logger.Info("DCA safety order placed",
		"symbol", d.config.Symbol,
		"safetyOrderNumber", d.safetyOrderCount,
		"newAveragePrice", d.averagePrice)

	return nil
}

func (d *DCAStrategy) updatePosition(order types.Order) {
	newInvestment := order.Quantity * order.Price
	newQuantity := order.Quantity

	// Calculate new weighted average price
	d.averagePrice = (d.totalInvested + newInvestment) / (d.totalQuantity + newQuantity)
	d.totalInvested += newInvestment
	d.totalQuantity += newQuantity

	d.currentOrders = append(d.currentOrders, order)
}
