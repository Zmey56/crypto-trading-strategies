package ai

import (
	"context"
	"crypto-trading-strategies/internal/indicators"
	"crypto-trading-strategies/pkg/types"
	"fmt"
)

type MLEngine struct {
	reinfLearning     *ReinforcementLearning
	walkForward       *WalkForwardOptimizer
	regimeDetector    *RegimeDetector
	adversarialTester *AdversarialTester
}

type ReinforcementLearning struct {
	rewards   map[string]float64
	penalties map[string]float64
	strategy  Strategy
}

// AdaptToMarketConditions использует reinforcement learning для
// непрерывного улучшения стратегии
func (ml *MLEngine) AdaptToMarketConditions(
	ctx context.Context,
	market types.MarketData,
) (*OptimizedStrategy, error) {

	// Определение текущего рыночного режима
	regime := ml.regimeDetector.ClassifyMarket(market)

	// Walk-forward оптимизация параметров
	optimizedParams, err := ml.walkForward.OptimizeParams(
		market.Candles,
		regime,
	)
	if err != nil {
		return nil, fmt.Errorf("walk-forward optimization failed: %w", err)
	}

	// Динамическое изменение размера позиций
	dynamicSizing := ml.calculateDynamicPositionSizing(market, regime)

	return &OptimizedStrategy{
		Parameters:   optimizedParams,
		PositionSize: dynamicSizing,
		Regime:       regime,
		Confidence:   ml.calculateConfidence(market),
	}, nil
}

type RegimeDetector struct {
	indicators []*indicators.TechnicalIndicator
	mlModel    *MachineLearningModel
}

// ClassifyMarket автоматически классифицирует рыночные условия
func (rd *RegimeDetector) ClassifyMarket(market types.MarketData) RegimeType {
	features := rd.extractFeatures(market)

	return rd.mlModel.Predict(features)
}

type RegimeType int

const (
	TrendingUp RegimeType = iota
	TrendingDown
	RangeBound
	HighVolatility
	LowVolatility
)
