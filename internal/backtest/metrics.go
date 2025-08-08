package backtest

import (
    "math"
    "time"
)

func computePerformance(equity []float64, period time.Duration, trades int, wins int, totalFees float64) PerformanceMetrics {
    if len(equity) == 0 { return PerformanceMetrics{} }
    start := equity[0]
    end := equity[len(equity)-1]
    totalReturn := (end/start - 1.0) * 100.0

    years := period.Hours() / (24*365)
    annualized := 0.0
    if years > 0 { annualized = (math.Pow(end/start, 1/years) - 1) * 100.0 }

    maxDD := computeMaxDrawdown(equity) * 100.0
    sharpe := computeSharpe(equity)
    winRate := 0.0
    if trades > 0 { winRate = float64(wins)/float64(trades) * 100.0 }
    volImpact := computeVolImpact(equity) * 100.0

    return PerformanceMetrics{
        TotalReturn:      totalReturn,
        AnnualizedReturn: annualized,
        MaxDrawdown:      maxDD,
        SharpeRatio:      sharpe,
        TradeCount:       trades,
        WinRate:          winRate,
        TotalFees:        totalFees,
        VolatilityImpact: volImpact,
    }
}

func computeMaxDrawdown(e []float64) float64 {
    peak := e[0]
    maxDD := 0.0
    for _, v := range e {
        if v > peak { peak = v }
        dd := (peak - v) / peak
        if dd > maxDD { maxDD = dd }
    }
    return maxDD
}

func computeSharpe(e []float64) float64 {
    if len(e) < 2 { return 0 }
    // simple daily returns approximation per step
    rets := make([]float64, 0, len(e)-1)
    for i := 1; i < len(e); i++ {
        if e[i-1] == 0 { continue }
        rets = append(rets, (e[i]/e[i-1])-1)
    }
    if len(rets) == 0 { return 0 }
    mean := 0.0
    for _, r := range rets { mean += r }
    mean /= float64(len(rets))
    var v float64
    for _, r := range rets { d := r - mean; v += d*d }
    v /= float64(len(rets))
    sd := math.Sqrt(v)
    if sd == 0 { return 0 }
    // Using risk-free ~0 and step Sharpe; for article demo this is sufficient
    return mean / sd
}

func computeVolImpact(e []float64) float64 {
    if len(e) < 2 { return 0 }
    // proxy: std of returns
    rets := make([]float64, 0, len(e)-1)
    for i := 1; i < len(e); i++ {
        if e[i-1] == 0 { continue }
        rets = append(rets, (e[i]/e[i-1])-1)
    }
    if len(rets) == 0 { return 0 }
    mean := 0.0
    for _, r := range rets { mean += r }
    mean /= float64(len(rets))
    var v float64
    for _, r := range rets { d := r - mean; v += d*d }
    v /= float64(len(rets))
    return math.Sqrt(v)
}


