package portfolio

import "fmt"

type StrategyAllocation struct {
	Strategy    string             `json:"strategy"`
	Allocation  float64            `json:"allocation"`   // Процент от общего капитала
	MaxDrawdown float64            `json:"max_drawdown"` // Максимальная просадка
	Performance PerformanceMetrics `json:"performance"`
}

type DiversificationManager struct {
	allocations map[string]*StrategyAllocation
	rebalancer  *RebalanceEngine
	monitor     *PerformanceMonitor

	// Параметры диверсификации
	config struct {
		MaxStrategyAllocation float64 `json:"max_strategy_allocation"` // 40%
		MinCorrelation        float64 `json:"min_correlation"`         // 0.3
		RebalanceThreshold    float64 `json:"rebalance_threshold"`     // 5%
	}
}

func (dm *DiversificationManager) OptimizeAllocation(
	strategies []Strategy,
	targetReturn float64,
	maxRisk float64,
) (*AllocationPlan, error) {

	// Анализ корреляции между стратегиями
	correlationMatrix := dm.calculateCorrelations(strategies)

	// Оптимизация по Марковицу для crypto portfolio
	optimizer := NewModernPortfolioOptimizer()
	allocation, err := optimizer.Optimize(OptimizationParams{
		Strategies:   strategies,
		TargetReturn: targetReturn,
		MaxRisk:      maxRisk,
		Correlations: correlationMatrix,
		Constraints:  dm.config,
	})

	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	return allocation, nil
}
