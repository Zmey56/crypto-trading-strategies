package strategy

type AdvancedGridConfig struct {
	Symbol     string  `json:"symbol"`
	UpperBound float64 `json:"upper_bound"`
	LowerBound float64 `json:"lower_bound"`
	GridLevels int     `json:"grid_levels"`
	OrderSize  float64 `json:"order_size"`

	// Продвинутые параметры для опытных трейдеров
	VolatilityFactor  float64 `json:"volatility_factor"`
	AdaptiveSpacing   bool    `json:"adaptive_spacing"`
	RebalanceInterval string  `json:"rebalance_interval"`

	RiskParameters struct {
		MaxDrawdown     float64 `json:"max_drawdown"`
		StopLossPercent float64 `json:"stop_loss_percent"`
		TakeProfitRatio float64 `json:"take_profit_ratio"`
	} `json:"risk_parameters"`
}

func NewAdvancedGridStrategy(config AdvancedGridConfig) *GridStrategy {
	return &GridStrategy{
		config:           config,
		volatilityEngine: NewVolatilityEngine(config.VolatilityFactor),
		riskManager:      NewAdvancedRiskManager(config.RiskParameters),
		adaptiveLogic:    config.AdaptiveSpacing,
	}
}
