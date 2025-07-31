package portfolio

type RiskManager struct {
	MaxDrawdown     float64 `json:"max_drawdown"`      // Максимальная просадка (%)
	MaxPositionSize float64 `json:"max_position_size"` // Максимальный размер позиции
	DailyLossLimit  float64 `json:"daily_loss_limit"`  // Дневной лимит убытков

	// Динамические параметры
	currentDrawdown float64
	dailyPnL        float64
	positionSizes   map[string]float64
}

// CanTrade проверяет возможность размещения ордера
func (r *RiskManager) CanTrade(orderSize float64, symbol string) bool {
	// Проверка лимитов просадки
	if r.currentDrawdown >= r.MaxDrawdown {
		return false
	}

	// Проверка размера позиции
	currentPosition := r.positionSizes[symbol]
	if currentPosition+orderSize > r.MaxPositionSize {
		return false
	}

	// Проверка дневных лимитов
	if r.dailyPnL <= -r.DailyLossLimit {
		return false
	}

	return true
}
