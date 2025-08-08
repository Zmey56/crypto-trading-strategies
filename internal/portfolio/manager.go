package portfolio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// Manager handles portfolio state and calculations
type Manager struct {
	exchange types.ExchangeClient
	logger   *logger.Logger
	mu       sync.RWMutex

	// Portfolio data
	portfolio *types.Portfolio
	positions map[string]*types.Position

	// Aggregated metrics
	totalInvested float64
	totalValue    float64
	lastUpdate    time.Time
}

// NewManager creates a new portfolio manager
func NewManager(exchange types.ExchangeClient, logger *logger.Logger) *Manager {
	return &Manager{
		exchange:  exchange,
		logger:    logger,
		portfolio: &types.Portfolio{},
		positions: make(map[string]*types.Position),
	}
}

// GetPortfolio returns the current portfolio snapshot
func (m *Manager) GetPortfolio() *types.Portfolio {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.portfolio
}

// GetPosition returns a position by symbol
func (m *Manager) GetPosition(symbol string) (*types.Position, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	position, exists := m.positions[symbol]
	return position, exists
}

// GetAllPositions returns all positions map
func (m *Manager) GetAllPositions() map[string]*types.Position {
	m.mu.RLock()
	defer m.mu.RUnlock()

	positions := make(map[string]*types.Position)
	for symbol, position := range m.positions {
		positions[symbol] = position
	}

	return positions
}

// UpdatePosition updates position by applying an executed order
func (m *Manager) UpdatePosition(order types.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	symbol := order.Symbol
	position, exists := m.positions[symbol]

	if !exists {
		position = &types.Position{
			Symbol:    symbol,
			Quantity:  0,
			AvgPrice:  0,
			Timestamp: time.Now(),
		}
		m.positions[symbol] = position
	}

	// Update position depending on order side
	switch order.Side {
	case types.OrderSideBuy:
		if order.Status == types.OrderStatusFilled {
			// Recalculate average price
			totalCost := position.Quantity*position.AvgPrice + order.FilledAmount*order.FilledPrice
			totalQuantity := position.Quantity + order.FilledAmount

			if totalQuantity > 0 {
				position.AvgPrice = totalCost / totalQuantity
			}

			position.Quantity += order.FilledAmount
			position.Timestamp = time.Now()

			m.logger.Info("Position updated (buy): %s %.8f @ %.2f (avg: %.2f)",
				symbol, order.FilledAmount, order.FilledPrice, position.AvgPrice)
		}

	case types.OrderSideSell:
		if order.Status == types.OrderStatusFilled {
			// Compute realized PnL
			if position.Quantity > 0 {
				realizedPnL := (order.FilledPrice - position.AvgPrice) * order.FilledAmount
				position.RealizedPnL += realizedPnL

				m.logger.Info("Realized PnL: %s %.2f (%.2f - %.2f) * %.8f",
					symbol, realizedPnL, order.FilledPrice, position.AvgPrice, order.FilledAmount)
			}

			position.Quantity -= order.FilledAmount
			position.Timestamp = time.Now()

			// Remove position if fully closed
			if position.Quantity <= 0 {
				delete(m.positions, symbol)
				m.logger.Info("Position closed: %s", symbol)
			}
		}
	}

	return nil
}

// RefreshPortfolio syncs portfolio with exchange market data
func (m *Manager) RefreshPortfolio(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Fetch balance from exchange (unused in mock)
	_, err := m.exchange.GetBalance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Update positions with current prices
	for symbol, position := range m.positions {
		ticker, err := m.exchange.GetTicker(ctx, symbol)
		if err != nil {
			m.logger.Warn("Failed to fetch ticker for %s: %v", symbol, err)
			continue
		}

		position.CurrentPrice = ticker.Price
		position.UnrealizedPnL = (ticker.Price - position.AvgPrice) * position.Quantity
		position.Timestamp = time.Now()
	}

	// Recompute aggregated portfolio metrics
	m.updatePortfolioMetrics()

	m.lastUpdate = time.Now()
	return nil
}

// updatePortfolioMetrics recomputes totals
func (m *Manager) updatePortfolioMetrics() {
	var totalValue, totalProfit, totalLoss float64

	for _, position := range m.positions {
		positionValue := position.Quantity * position.CurrentPrice
		totalValue += positionValue

		if position.UnrealizedPnL > 0 {
			totalProfit += position.UnrealizedPnL
		} else {
			totalLoss += -position.UnrealizedPnL
		}
	}

	m.portfolio.TotalValue = totalValue
	m.portfolio.TotalProfit = totalProfit
	m.portfolio.TotalLoss = totalLoss
	m.portfolio.NetProfit = totalProfit - totalLoss
	m.portfolio.LastUpdate = time.Now()

	// Refresh positions slice
	var positions []types.Position
	for _, position := range m.positions {
		positions = append(positions, *position)
	}
	m.portfolio.Positions = positions
}

// GetMetrics returns portfolio metrics summary
func (m *Manager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_value":     m.portfolio.TotalValue,
		"total_profit":    m.portfolio.TotalProfit,
		"total_loss":      m.portfolio.TotalLoss,
		"net_profit":      m.portfolio.NetProfit,
		"positions_count": len(m.positions),
		"last_update":     m.lastUpdate,
	}
}

// GetPositionSummary returns human-friendly positions summary
func (m *Manager) GetPositionSummary() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var summary []map[string]interface{}
	for symbol, position := range m.positions {
		summary = append(summary, map[string]interface{}{
			"symbol":         symbol,
			"quantity":       position.Quantity,
			"avg_price":      position.AvgPrice,
			"current_price":  position.CurrentPrice,
			"unrealized_pnl": position.UnrealizedPnL,
			"realized_pnl":   position.RealizedPnL,
			"timestamp":      position.Timestamp,
		})
	}

	return summary
}

// StartAutoRefresh periodically refreshes portfolio metrics
func (m *Manager) StartAutoRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Portfolio auto-refresh stopped")
			return
		case <-ticker.C:
			if err := m.RefreshPortfolio(ctx); err != nil {
				m.logger.Error("Portfolio refresh error: %v", err)
			}
		}
	}
}
