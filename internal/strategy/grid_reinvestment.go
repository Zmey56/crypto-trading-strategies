package strategy

import "time"

type ReinvestmentEngine struct {
	reinvestmentRate  float64 // Процент прибыли для реинвестирования
	profitThreshold   float64 // Минимальная прибыль для срабатывания
	compoundingPeriod time.Duration
}

// ProcessProfit обрабатывает полученную прибыль
func (re *ReinvestmentEngine) ProcessProfit(profit float64, gridStrategy *GridStrategy) error {
	if profit < re.profitThreshold {
		return nil
	}

	// Расчет суммы для реинвестирования
	reinvestAmount := profit * re.reinvestmentRate

	// Увеличение размера ордеров пропорционально прибыли
	newOrderSize := gridStrategy.OrderSize * (1 + reinvestAmount/gridStrategy.getTotalCapital())

	// Обновление параметров стратегии
	gridStrategy.OrderSize = newOrderSize

	return gridStrategy.recalculateGrid()
}
