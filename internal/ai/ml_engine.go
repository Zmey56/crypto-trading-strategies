package ai

import (
	"context"
	"fmt"

	"github.com/Zmey56/crypto-arbitrage-trader/pkg/indicators"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
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

// AdaptToMarketConditions uses reinforcement learning for continuous improvement
func (ml *MLEngine) AdaptToMarketConditions(
	ctx context.Context,
	market types.MarketData,
) (*OptimizedStrategy, error) {

	// Detect current market regime
	regime := ml.regimeDetector.ClassifyMarket(market)

	// Walk-forward parameters optimization
	optimizedParams, err := ml.walkForward.OptimizeParams(
		market.Candles,
		regime,
	)
	if err != nil {
		return nil, fmt.Errorf("walk-forward optimization failed: %w", err)
	}

	// Dynamic position sizing
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
