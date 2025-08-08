package backtest

type FeeImpactAnalysis struct {
	Strategy         string  `json:"strategy"`
	TradeFrequency   int     `json:"trades_per_month"`
	AvgFeeRate       float64 `json:"average_fee_rate"`    // %
	MonthlyFeeCost   float64 `json:"monthly_fee_cost"`    // USD
	FeeToReturnRatio float64 `json:"fee_to_return_ratio"` // %
	OptimalMinProfit float64 `json:"optimal_min_profit"`  // % per trade
}

// CalculateFeeImpact analyzes the impact of fees on profitability
func CalculateFeeImpact(strategy string, monthlyTrades int, avgFee float64, monthlyReturn float64) *FeeImpactAnalysis {
	monthlyFeeCost := float64(monthlyTrades) * avgFee
	feeToReturnRatio := (monthlyFeeCost / monthlyReturn) * 100

	// Minimum profit to cover fees and risk buffer (2.5x fee)
	optimalMinProfit := (avgFee * 2.5)

	return &FeeImpactAnalysis{
		Strategy:         strategy,
		TradeFrequency:   monthlyTrades,
		AvgFeeRate:       avgFee,
		MonthlyFeeCost:   monthlyFeeCost,
		FeeToReturnRatio: feeToReturnRatio,
		OptimalMinProfit: optimalMinProfit,
	}
}
