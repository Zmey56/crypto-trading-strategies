package strategy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// ComboStrategy combines multiple strategies with weighted signals
type ComboStrategy struct {
	config   types.ComboConfig
	exchange types.ExchangeClient
	logger   *logger.Logger

	strategies []Strategy
	weights    []float64

	mu      sync.RWMutex
	metrics types.StrategyMetrics
}

// NewComboStrategy creates a new combo strategy
func NewComboStrategy(config types.ComboConfig, exchange types.ExchangeClient, logger *logger.Logger) (*ComboStrategy, error) {
	if len(config.Strategies) == 0 {
		return nil, fmt.Errorf("at least one strategy is required")
	}

	cs := &ComboStrategy{
		config:   config,
		exchange: exchange,
		logger:   logger,
		weights:  make([]float64, len(config.Strategies)),
	}

	// Initialize strategies and weights
	if err := cs.initializeStrategies(); err != nil {
		return nil, fmt.Errorf("failed to initialize strategies: %w", err)
	}

	return cs, nil
}

// initializeStrategies creates individual strategies from config
func (cs *ComboStrategy) initializeStrategies() error {
	factory := NewFactory(cs.logger)
	cs.strategies = make([]Strategy, len(cs.config.Strategies))

	// Equal weights by default
	weight := 1.0 / float64(len(cs.config.Strategies))

	for i, strategyConfig := range cs.config.Strategies {
		var strategy Strategy

		switch strategyConfig.Type {
		case "dca":
			dcaConfig, err := cs.parseDCAConfig(strategyConfig.Config)
			if err != nil {
				return fmt.Errorf("invalid DCA config: %w", err)
			}
			strategy, err = factory.CreateDCA(dcaConfig, cs.exchange)
			if err != nil {
				return fmt.Errorf("failed to create DCA strategy: %w", err)
			}

		case "grid":
			gridConfig, err := cs.parseGridConfig(strategyConfig.Config)
			if err != nil {
				return fmt.Errorf("invalid Grid config: %w", err)
			}
			strategy, err = factory.CreateGrid(gridConfig, cs.exchange)
			if err != nil {
				return fmt.Errorf("failed to create Grid strategy: %w", err)
			}

		default:
			return fmt.Errorf("unsupported strategy type: %s", strategyConfig.Type)
		}

		cs.strategies[i] = strategy
		cs.weights[i] = weight
	}

	return nil
}

// parseDCAConfig converts map to DCAConfig
func (cs *ComboStrategy) parseDCAConfig(config map[string]interface{}) (types.DCAConfig, error) {
	dcaConfig := types.DCAConfig{}

	if symbol, ok := config["symbol"].(string); ok {
		dcaConfig.Symbol = symbol
	} else {
		return dcaConfig, fmt.Errorf("symbol is required for DCA strategy")
	}

	if investmentAmount, ok := config["investment_amount"].(float64); ok {
		dcaConfig.InvestmentAmount = investmentAmount
	} else {
		dcaConfig.InvestmentAmount = 100.0 // default
	}

	if intervalStr, ok := config["interval"].(string); ok {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return dcaConfig, fmt.Errorf("invalid interval: %w", err)
		}
		dcaConfig.Interval = interval
	} else {
		dcaConfig.Interval = 24 * time.Hour // default
	}

	if maxInvestments, ok := config["max_investments"].(float64); ok {
		dcaConfig.MaxInvestments = int(maxInvestments)
	} else {
		dcaConfig.MaxInvestments = 100 // default
	}

	if enabled, ok := config["enabled"].(bool); ok {
		dcaConfig.Enabled = enabled
	} else {
		dcaConfig.Enabled = true // default
	}

	return dcaConfig, nil
}

// parseGridConfig converts map to GridConfig
func (cs *ComboStrategy) parseGridConfig(config map[string]interface{}) (types.GridConfig, error) {
	gridConfig := types.GridConfig{}

	if symbol, ok := config["symbol"].(string); ok {
		gridConfig.Symbol = symbol
	} else {
		return gridConfig, fmt.Errorf("symbol is required for Grid strategy")
	}

	if upperPrice, ok := config["upper_price"].(float64); ok {
		gridConfig.UpperPrice = upperPrice
	} else {
		return gridConfig, fmt.Errorf("upper_price is required for Grid strategy")
	}

	if lowerPrice, ok := config["lower_price"].(float64); ok {
		gridConfig.LowerPrice = lowerPrice
	} else {
		return gridConfig, fmt.Errorf("lower_price is required for Grid strategy")
	}

	if gridLevels, ok := config["grid_levels"].(float64); ok {
		gridConfig.GridLevels = int(gridLevels)
	} else {
		gridConfig.GridLevels = 20 // default
	}

	if investmentPerLevel, ok := config["investment_per_level"].(float64); ok {
		gridConfig.InvestmentPerLevel = investmentPerLevel
	} else {
		gridConfig.InvestmentPerLevel = 100.0 // default
	}

	if enabled, ok := config["enabled"].(bool); ok {
		gridConfig.Enabled = enabled
	} else {
		gridConfig.Enabled = true // default
	}

	return gridConfig, nil
}

// Execute runs all strategies and combines their signals
func (cs *ComboStrategy) Execute(ctx context.Context, market types.MarketData) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.config.Enabled {
		return nil
	}

	// Execute all strategies
	for i, strategy := range cs.strategies {
		if err := strategy.Execute(ctx, market); err != nil {
			cs.logger.Error("Strategy %d execution failed: %v", i, err)
			continue
		}
	}

	// Update combined metrics
	cs.updateMetrics()

	return nil
}

// GetSignal combines signals from all strategies with weights
func (cs *ComboStrategy) GetSignal(market types.MarketData) types.Signal {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if !cs.config.Enabled {
		return types.Signal{
			Type:      types.SignalTypeHold,
			Symbol:    market.Symbol,
			Price:     market.Price,
			Timestamp: market.Timestamp,
		}
	}

	var totalStrength float64
	var weightedSignal types.Signal

	// Collect signals from all strategies
	for i, strategy := range cs.strategies {
		signal := strategy.GetSignal(market)
		weight := cs.weights[i]

		// Weight the signal
		switch signal.Type {
		case types.SignalTypeBuy:
			weightedSignal.Type = types.SignalTypeBuy
			weightedSignal.Strength += signal.Strength * weight
			totalStrength += weight
		case types.SignalTypeSell:
			weightedSignal.Type = types.SignalTypeSell
			weightedSignal.Strength += signal.Strength * weight
			totalStrength += weight
		}
	}

	// Normalize strength
	if totalStrength > 0 {
		weightedSignal.Strength /= totalStrength
	}

	// Set common fields
	weightedSignal.Symbol = market.Symbol
	weightedSignal.Price = market.Price
	weightedSignal.Timestamp = market.Timestamp

	// If no clear signal, hold
	if weightedSignal.Strength < 0.3 {
		weightedSignal.Type = types.SignalTypeHold
		weightedSignal.Strength = 0.0
	}

	return weightedSignal
}

// ValidateConfig validates combo configuration
func (cs *ComboStrategy) ValidateConfig() error {
	if len(cs.config.Strategies) == 0 {
		return fmt.Errorf("at least one strategy is required")
	}

	for i, strategy := range cs.config.Strategies {
		if strategy.Type == "" {
			return fmt.Errorf("strategy type is required for strategy %d", i)
		}

		if strategy.Config == nil {
			return fmt.Errorf("strategy config is required for strategy %d", i)
		}
	}

	return nil
}

// GetMetrics returns combined metrics from all strategies
func (cs *ComboStrategy) GetMetrics() types.StrategyMetrics {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return cs.metrics
}

// Shutdown gracefully stops all strategies
func (cs *ComboStrategy) Shutdown(ctx context.Context) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i, strategy := range cs.strategies {
		if err := strategy.Shutdown(ctx); err != nil {
			cs.logger.Error("Failed to shutdown strategy %d: %v", i, err)
		}
	}

	cs.logger.Info("Combo strategy stopped")
	return nil
}

// updateMetrics aggregates metrics from all strategies
func (cs *ComboStrategy) updateMetrics() {
	var totalTrades, winningTrades, losingTrades int
	var totalProfit, totalLoss, totalVolume float64

	for _, strategy := range cs.strategies {
		metrics := strategy.GetMetrics()
		totalTrades += metrics.TotalTrades
		winningTrades += metrics.WinningTrades
		losingTrades += metrics.LosingTrades
		totalProfit += metrics.TotalProfit
		totalLoss += metrics.TotalLoss
		totalVolume += metrics.TotalVolume
	}

	cs.metrics.TotalTrades = totalTrades
	cs.metrics.WinningTrades = winningTrades
	cs.metrics.LosingTrades = losingTrades
	cs.metrics.TotalProfit = totalProfit
	cs.metrics.TotalLoss = totalLoss
	cs.metrics.TotalVolume = totalVolume
	cs.metrics.LastUpdate = time.Now()

	// Calculate derived metrics
	if totalTrades > 0 {
		cs.metrics.WinRate = float64(winningTrades) / float64(totalTrades) * 100.0
	}

	if winningTrades > 0 {
		cs.metrics.AverageWin = totalProfit / float64(winningTrades)
	}

	if losingTrades > 0 {
		cs.metrics.AverageLoss = totalLoss / float64(losingTrades)
	}

	if totalLoss > 0 {
		cs.metrics.ProfitFactor = totalProfit / totalLoss
	}
}

// GetStatus returns combo strategy status
func (cs *ComboStrategy) GetStatus() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":      cs.config.Enabled,
		"strategies":   len(cs.strategies),
		"total_trades": cs.metrics.TotalTrades,
		"win_rate":     cs.metrics.WinRate,
		"last_update":  cs.metrics.LastUpdate,
	}

	// Add individual strategy statuses
	strategyStatuses := make([]map[string]interface{}, len(cs.strategies))
	for i, strategy := range cs.strategies {
		if statusProvider, ok := strategy.(interface{ GetStatus() map[string]interface{} }); ok {
			strategyStatuses[i] = statusProvider.GetStatus()
		} else {
			strategyStatuses[i] = map[string]interface{}{
				"type": fmt.Sprintf("strategy_%d", i),
			}
		}
	}
	status["strategy_details"] = strategyStatuses

	return status
}
