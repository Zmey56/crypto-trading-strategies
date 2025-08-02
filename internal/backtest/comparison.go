package backtest

import (
	"crypto-trading-strategies/pkg/types"
	"time"
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

// CompareStrategies выполняет сравнительный анализ стратегий
func (engine *BacktestEngine) CompareStrategies(
	symbol string,
	startDate, endDate time.Time,
	initialBalance float64,
) (*StrategyComparison, error) {

	// Определение рыночных условий
	marketCondition := engine.analyzeMarketCondition(symbol, startDate, endDate)

	// Бэктест DCA стратегии
	dcaConfig := &DCAConfig{
		BaseOrderSize:     initialBalance * 0.1,
		SafetyOrderSize:   initialBalance * 0.05,
		SafetyOrdersCount: 5,
		PriceDeviation:    3.0,
		TakeProfitPercent: 2.0,
	}

	dcaResults, err := engine.BacktestDCA(symbol, startDate, endDate, dcaConfig)
	if err != nil {
		return nil, err
	}

	// Бэктест Grid стратегии
	gridConfig := &GridConfig{
		GridLevels:    20,
		OrderSize:     initialBalance * 0.05,
		GridType:      GEOMETRIC,
		TakeProfitPct: 1.0,
	}

	gridResults, err := engine.BacktestGrid(symbol, startDate, endDate, gridConfig)
	if err != nil {
		return nil, err
	}

	return &StrategyComparison{
		DCAResults:  dcaResults,
		GridResults: gridResults,
		Period:      endDate.Sub(startDate),
		MarketType:  marketCondition,
	}, nil
}
