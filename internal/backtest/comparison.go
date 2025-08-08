package backtest

import (
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

type StrategyComparison struct {
	DCAResults  PerformanceMetrics `json:"dca_results"`
	GridResults PerformanceMetrics `json:"grid_results"`
	Period      time.Duration      `json:"backtest_period"`
	MarketType  MarketCondition    `json:"market_condition"`
}

type PerformanceMetrics struct {
	TotalReturn      float64 `json:"total_return"`      // %
	AnnualizedReturn float64 `json:"annualized_return"` // %
	MaxDrawdown      float64 `json:"max_drawdown"`      // %
	SharpeRatio      float64 `json:"sharpe_ratio"`
	TradeCount       int     `json:"trade_count"`
	WinRate          float64 `json:"win_rate"`          // %
	TotalFees        float64 `json:"total_fees"`        // USD
	VolatilityImpact float64 `json:"volatility_impact"` // %
}

type MarketCondition string

const (
	BEAR_MARKET     MarketCondition = "bear"
	BULL_MARKET     MarketCondition = "bull"
	SIDEWAYS_MARKET MarketCondition = "sideways"
	HIGH_VOLATILITY MarketCondition = "high_vol"
)

// CompareStrategies performs comparative analysis of strategies
func (e *Engine) CompareStrategies(symbol string, candles []Candle, startDate, endDate time.Time, initialBalance float64, dcaCfg types.DCAConfig, gridCfg types.GridConfig) (*StrategyComparison, error) {
	marketCondition := analyzeMarketCondition(candles, startDate, endDate)
	dca := e.BacktestDCA(symbol, candles, startDate, endDate, dcaCfg, initialBalance)
	grid := e.BacktestGrid(symbol, candles, startDate, endDate, gridCfg, initialBalance)
	return &StrategyComparison{DCAResults: dca, GridResults: grid, Period: endDate.Sub(startDate), MarketType: marketCondition}, nil
}

func analyzeMarketCondition(candles []Candle, start, end time.Time) MarketCondition {
	// Simple heuristic: last vs first close price
	var first, last float64
	for _, c := range candles {
		if c.Time.Before(start) || c.Time.After(end) {
			continue
		}
		if first == 0 {
			first = c.Close
		}
		last = c.Close
	}
	if first == 0 {
		return SIDEWAYS_MARKET
	}
	chg := (last/first - 1)
	switch {
	case chg > 0.1:
		return BULL_MARKET
	case chg < -0.1:
		return BEAR_MARKET
	default:
		return SIDEWAYS_MARKET
	}
}
