package indicators

import (
	"crypto-trading-strategies/pkg/types"
	"math"
)

type ATRIndicator struct {
	period    int
	values    []float64
	smoothing float64
}

// Calculate вычисляет ATR для адаптации Grid параметров
func (atr *ATRIndicator) Calculate(candles []types.Candle) float64 {
	if len(candles) < atr.period {
		return 0
	}

	trueRanges := make([]float64, len(candles)-1)

	for i := 1; i < len(candles); i++ {
		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close

		trueRange := math.Max(
			high-low,
			math.Max(
				math.Abs(high-prevClose),
				math.Abs(low-prevClose),
			),
		)
		trueRanges[i-1] = trueRange
	}

	return atr.exponentialMovingAverage(trueRanges)
}

// AdaptGridSpacing адаптирует расстояние между уровнями
func (g *GridStrategy) AdaptGridSpacing(atr float64) {
	// Установка spacing как процент от ATR (10-20%)
	optimalSpacing := atr * 0.15

	// Пересчет количества уровней для поддержания диапазона
	g.GridLevels = int((g.UpperBound - g.LowerBound) / optimalSpacing)
	g.GridSpacing = optimalSpacing
}
