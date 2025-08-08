package strategy

import (
	"fmt"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

// Factory is a strategy factory
type Factory struct {
	logger *logger.Logger
}

// NewFactory creates a new strategy factory
func NewFactory(logger *logger.Logger) *Factory {
	return &Factory{
		logger: logger,
	}
}

// CreateDCA creates a DCA strategy
func (f *Factory) CreateDCA(config types.DCAConfig, exchange types.ExchangeClient) (Strategy, error) {
	if err := f.validateDCAConfig(config); err != nil {
		return nil, fmt.Errorf("invalid DCA config: %w", err)
	}

	strategy := NewDCAStrategy(config, exchange, f.logger)
	return strategy, nil
}

// CreateGrid creates a Grid strategy
func (f *Factory) CreateGrid(config types.GridConfig, exchange types.ExchangeClient) (Strategy, error) {
	if err := f.validateGridConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Grid config: %w", err)
	}
	gs, err := NewGridStrategy(config, exchange, f.logger)
	if err != nil {
		return nil, err
	}
	return gs, nil
}

// CreateCombo creates a combined strategy
func (f *Factory) CreateCombo(config types.ComboConfig, exchange types.ExchangeClient) (Strategy, error) {
	if err := f.validateComboConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Combo config: %w", err)
	}

	// TODO: Реализовать комбинированную стратегию
	return nil, fmt.Errorf("combo strategy not implemented yet")
}

// validateDCAConfig validates DCA configuration
func (f *Factory) validateDCAConfig(config types.DCAConfig) error {
	if config.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if config.InvestmentAmount <= 0 {
		return fmt.Errorf("investment amount must be positive")
	}

	if config.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	if config.MaxInvestments <= 0 {
		return fmt.Errorf("max investments must be positive")
	}

	return nil
}

// validateGridConfig validates Grid configuration
func (f *Factory) validateGridConfig(config types.GridConfig) error {
	if config.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if config.UpperPrice <= config.LowerPrice {
		return fmt.Errorf("upper price must be greater than lower price")
	}

	if config.GridLevels <= 0 {
		return fmt.Errorf("grid levels must be positive")
	}

	if config.InvestmentPerLevel <= 0 {
		return fmt.Errorf("investment per level must be positive")
	}

	return nil
}

// validateComboConfig validates combined strategy configuration
func (f *Factory) validateComboConfig(config types.ComboConfig) error {
	if len(config.Strategies) == 0 {
		return fmt.Errorf("at least one strategy is required")
	}

	for i, strategy := range config.Strategies {
		if strategy.Type == "" {
			return fmt.Errorf("strategy type is required for strategy %d", i)
		}

		if strategy.Config == nil {
			return fmt.Errorf("strategy config is required for strategy %d", i)
		}
	}

	return nil
}
