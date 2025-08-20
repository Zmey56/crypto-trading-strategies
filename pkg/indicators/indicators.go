package indicators

import (
	"math"
)

// TechnicalIndicator represents a technical analysis indicator
type TechnicalIndicator struct {
	Name   string
	Values []float64
	Params map[string]float64
}

// SMA calculates Simple Moving Average
func SMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return []float64{}
	}

	result := make([]float64, len(prices)-period+1)
	sum := 0.0

	// Calculate initial sum
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[0] = sum / float64(period)

	// Calculate SMA for remaining periods
	for i := period; i < len(prices); i++ {
		sum = sum - prices[i-period] + prices[i]
		result[i-period+1] = sum / float64(period)
	}

	return result
}

// EMA calculates Exponential Moving Average
func EMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return []float64{}
	}

	result := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)

	// First EMA is SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate EMA for remaining periods
	for i := period; i < len(prices); i++ {
		result[i] = (prices[i] * multiplier) + (result[i-1] * (1 - multiplier))
	}

	return result
}

// RSI calculates Relative Strength Index
func RSI(prices []float64, period int) []float64 {
	if len(prices) < period+1 {
		return []float64{}
	}

	gains := make([]float64, len(prices))
	losses := make([]float64, len(prices))

	// Calculate gains and losses
	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	// Calculate average gains and losses
	avgGain := SMA(gains, period)
	avgLoss := SMA(losses, period)

	result := make([]float64, len(avgGain))
	for i := 0; i < len(avgGain); i++ {
		if avgLoss[i] == 0 {
			result[i] = 100
		} else {
			rs := avgGain[i] / avgLoss[i]
			result[i] = 100 - (100 / (1 + rs))
		}
	}

	return result
}

// MACD calculates Moving Average Convergence Divergence
func MACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) ([]float64, []float64, []float64) {
	if len(prices) < slowPeriod {
		return []float64{}, []float64{}, []float64{}
	}

	fastEMA := EMA(prices, fastPeriod)
	slowEMA := EMA(prices, slowPeriod)

	// Calculate MACD line
	macdLine := make([]float64, len(fastEMA))
	startIdx := len(slowEMA) - len(fastEMA)
	for i := 0; i < len(fastEMA); i++ {
		macdLine[i] = fastEMA[i] - slowEMA[startIdx+i]
	}

	// Calculate signal line
	signalLine := EMA(macdLine, signalPeriod)

	// Calculate histogram
	histogram := make([]float64, len(signalLine))
	for i := 0; i < len(signalLine); i++ {
		histogram[i] = macdLine[len(macdLine)-len(signalLine)+i] - signalLine[i]
	}

	return macdLine, signalLine, histogram
}

// BollingerBands calculates Bollinger Bands
func BollingerBands(prices []float64, period int, stdDev float64) ([]float64, []float64, []float64) {
	if len(prices) < period {
		return []float64{}, []float64{}, []float64{}
	}

	sma := SMA(prices, period)
	upper := make([]float64, len(sma))
	lower := make([]float64, len(sma))

	for i := 0; i < len(sma); i++ {
		// Calculate standard deviation for this period
		sum := 0.0
		for j := i; j < i+period && j < len(prices); j++ {
			sum += math.Pow(prices[j]-sma[i], 2)
		}
		deviation := math.Sqrt(sum / float64(period))

		upper[i] = sma[i] + (stdDev * deviation)
		lower[i] = sma[i] - (stdDev * deviation)
	}

	return upper, sma, lower
}

// Stochastic calculates Stochastic Oscillator
func Stochastic(highs, lows, closes []float64, kPeriod, dPeriod int) ([]float64, []float64) {
	if len(closes) < kPeriod {
		return []float64{}, []float64{}
	}

	kValues := make([]float64, len(closes)-kPeriod+1)
	for i := kPeriod - 1; i < len(closes); i++ {
		highestHigh := highs[i]
		lowestLow := lows[i]

		for j := i - kPeriod + 1; j <= i; j++ {
			if highs[j] > highestHigh {
				highestHigh = highs[j]
			}
			if lows[j] < lowestLow {
				lowestLow = lows[j]
			}
		}

		if highestHigh == lowestLow {
			kValues[i-kPeriod+1] = 50
		} else {
			kValues[i-kPeriod+1] = ((closes[i] - lowestLow) / (highestHigh - lowestLow)) * 100
		}
	}

	// Calculate %D (SMA of %K)
	dValues := SMA(kValues, dPeriod)

	return kValues, dValues
}
