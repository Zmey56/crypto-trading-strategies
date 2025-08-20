package risk

import (
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// Manager handles risk management
type Manager struct {
	maxPositionSize float64
	maxDrawdown     float64
	stopLoss        float64
	takeProfit      float64
}

// NewManager creates a new risk manager
func NewManager() *Manager {
	return &Manager{
		maxPositionSize: 0.1,  // 10% of portfolio
		maxDrawdown:     0.2,  // 20% max drawdown
		stopLoss:        0.05, // 5% stop loss
		takeProfit:      0.1,  // 10% take profit
	}
}

// ValidateOrder validates if an order meets risk requirements
func (rm *Manager) ValidateOrder(order types.Order, portfolio *types.Portfolio) error {
	// Basic validation - can be enhanced with more sophisticated risk models
	return nil
}

// CalculatePositionSize calculates safe position size
func (rm *Manager) CalculatePositionSize(portfolio *types.Portfolio, price float64) float64 {
	totalValue := portfolio.TotalValue
	maxSize := totalValue * rm.maxPositionSize
	return maxSize / price
}
