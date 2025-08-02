package strategy

import (
	"context"
	"crypto-trading-strategies/pkg/types"
)

type Strategy interface {
	Execute(ctx context.Context, market types.MarketData) error
	GetSignal(market types.MarketData) types.Signal
	ValidateConfig() error
	GetMetrics() types.StrategyMetrics
	Shutdown(ctx context.Context) error
}

type StrategyFactory interface {
	CreateDCA(config types.DCAConfig) (Strategy, error)
	CreateGrid(config types.GridConfig) (Strategy, error)
	CreateCombo(config types.ComboConfig) (Strategy, error)
}
