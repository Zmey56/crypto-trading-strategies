package backtest

import (
    "sort"
    "time"

    "github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

func (e *Engine) BacktestGrid(symbol string, candles []Candle, start, end time.Time, cfg types.GridConfig, initialBalance float64) PerformanceMetrics {
    if cfg.GridLevels < 2 { return PerformanceMetrics{} }
    step := (cfg.UpperPrice - cfg.LowerPrice) / float64(cfg.GridLevels-1)
    levels := make([]float64, cfg.GridLevels)
    for i := 0; i < cfg.GridLevels; i++ { levels[i] = cfg.LowerPrice + float64(i)*step }
    sort.Float64s(levels)

    type pos struct{ qty, avg float64 }
    positions := make(map[int]pos)

    cash := initialBalance
    totalFees := 0.0
    trades := 0
    wins := 0
    var equity []float64

    for _, c := range candles {
        if c.Time.Before(start) || c.Time.After(end) { continue }
        p := c.Close
        // buy
        for i, level := range levels {
            if p <= level {
                if positions[i].qty == 0 && cash >= cfg.InvestmentPerLevel {
                    fee := cfg.InvestmentPerLevel * e.feeRate
                    qty := (cfg.InvestmentPerLevel - fee) / p
                    positions[i] = pos{ qty: qty, avg: p }
                    cash -= cfg.InvestmentPerLevel
                    totalFees += fee
                    trades++
                }
            }
        }
        // sell
        for i := 0; i < len(levels)-1; i++ {
            next := levels[i+1]
            if positions[i].qty > 0 && p >= next {
                qty := positions[i].qty
                proceeds := qty * p
                fee := proceeds * e.feeRate
                cash += proceeds - fee
                if p >= positions[i].avg { wins++ }
                totalFees += fee
                positions[i] = pos{}
                trades++
            }
        }
        // equity
        invQty := 0.0
        for _, ps := range positions { invQty += ps.qty }
        equity = append(equity, cash+invQty*p)
    }

    return computePerformance(equity, end.Sub(start), trades, wins, totalFees)
}


