package strategy

import (
	"context"
	"crypto-trading-strategies/internal/portfolio"
	"crypto-trading-strategies/pkg/types"
	"errors"
	"time"
)

type DCAStrategy struct {
	// Основные параметры стратегии
	Symbol            string  `json:"symbol"`
	BaseOrderSize     float64 `json:"base_order_size"`
	SafetyOrderSize   float64 `json:"safety_order_size"`
	SafetyOrdersCount int     `json:"safety_orders_count"`
	PriceDeviation    float64 `json:"price_deviation"` // % для размещения safety ордеров
	TakeProfitPercent float64 `json:"take_profit_percent"`

	// Внутреннее состояние
	CurrentOrders []types.Order `json:"current_orders"`
	AveragePrice  float64       `json:"average_price"`
	TotalInvested float64       `json:"total_invested"`
	TotalQuantity float64       `json:"total_quantity"`
	IsActive      bool          `json:"is_active"`

	// Зависимости
	portfolioManager *portfolio.Manager
	riskManager      RiskManager
}

// Execute реализует основную логику DCA стратегии
func (d *DCAStrategy) Execute(ctx context.Context, market types.MarketData) error {
	// Проверка условий для входа в сделку
	if !d.IsActive {
		return d.initiateFirstOrder(ctx, market)
	}

	// Логика размещения safety ордеров при падении цены
	if d.shouldPlaceSafetyOrder(market.Price) {
		return d.placeSafetyOrder(ctx, market)
	}

	// Проверка условий для take profit
	if d.shouldTakeProfit(market.Price) {
		return d.executetakeProfit(ctx, market)
	}

	return nil
}

// initiateFirstOrder размещает первоначальный ордер
func (d *DCAStrategy) initiateFirstOrder(ctx context.Context, market types.MarketData) error {
	// Валидация размера позиции согласно риск-менеджменту
	if !d.riskManager.CanTrade(d.BaseOrderSize, d.Symbol) {
		return errors.New("risk limits exceeded for initial order")
	}

	order := types.Order{
		Symbol:    d.Symbol,
		Side:      types.BUY,
		Type:      types.MARKET,
		Quantity:  d.BaseOrderSize / market.Price,
		Price:     market.Price,
		Timestamp: time.Now(),
	}

	// Обновление внутреннего состояния
	d.updatePosition(order)
	d.IsActive = true

	return d.portfolioManager.PlaceOrder(ctx, order)
}
