package portfolio

type RiskManager struct {
	MaxDrawdown     float64 `json:"max_drawdown"`      // Maximum drawdown (%)
	MaxPositionSize float64 `json:"max_position_size"` // Maximum position size
	DailyLossLimit  float64 `json:"daily_loss_limit"`  // Daily loss limit

	// Dynamic parameters
	currentDrawdown float64
	dailyPnL        float64
	positionSizes   map[string]float64
}

// CanTrade checks whether a new order can be placed under current risk limits
func (r *RiskManager) CanTrade(orderSize float64, symbol string) bool {
	// Drawdown limits
	if r.currentDrawdown >= r.MaxDrawdown {
		return false
	}

	// Position size limit
	currentPosition := r.positionSizes[symbol]
	if currentPosition+orderSize > r.MaxPositionSize {
		return false
	}

	// Daily loss limits
	if r.dailyPnL <= -r.DailyLossLimit {
		return false
	}

	return true
}
