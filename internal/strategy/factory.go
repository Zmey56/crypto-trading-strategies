package strategy

import (
	"crypto-trading-strategies/internal/logger"
	"crypto-trading-strategies/pkg/types"
	"fmt"
)

// Factory представляет фабрику стратегий
type Factory struct {
	logger *logger.Logger
}

// NewFactory создает новую фабрику стратегий
func NewFactory(logger *logger.Logger) *Factory {
	return &Factory{
		logger: logger,
	}
}

// CreateDCA создает DCA стратегию
func (f *Factory) CreateDCA(config types.DCAConfig, exchange types.ExchangeClient) (Strategy, error) {
	if err := f.validateDCAConfig(config); err != nil {
		return nil, fmt.Errorf("invalid DCA config: %w", err)
	}

	strategy := NewDCAStrategy(config, exchange, f.logger)
	return strategy, nil
}

// CreateGrid создает Grid стратегию
func (f *Factory) CreateGrid(config types.GridConfig, exchange types.ExchangeClient) (Strategy, error) {
	if err := f.validateGridConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Grid config: %w", err)
	}

	// TODO: Реализовать Grid стратегию
	return nil, fmt.Errorf("grid strategy not implemented yet")
}

// CreateCombo создает комбинированную стратегию
func (f *Factory) CreateCombo(config types.ComboConfig, exchange types.ExchangeClient) (Strategy, error) {
	if err := f.validateComboConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Combo config: %w", err)
	}

	// TODO: Реализовать комбинированную стратегию
	return nil, fmt.Errorf("combo strategy not implemented yet")
}

// validateDCAConfig проверяет конфигурацию DCA стратегии
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

// validateGridConfig проверяет конфигурацию Grid стратегии
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

// validateComboConfig проверяет конфигурацию комбинированной стратегии
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