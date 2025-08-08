package backtest

import (
    "time"

    "github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

func (e *Engine) BacktestDCA(symbol string, candles []Candle, start, end time.Time, cfg types.DCAConfig, initialBalance float64) PerformanceMetrics {
    cash := initialBalance
    qty := 0.0
    totalFees := 0.0
    trades := 0
    wins := 0

    nextBuy := start
    var equity []float64
    for _, c := range candles {
        if c.Time.Before(start) || c.Time.After(end) { continue }
        price := c.Close
        if !nextBuy.After(c.Time) && trades < cfg.MaxInvestments && cfg.InvestmentAmount > 0 && cash > 0 {
            invest := cfg.InvestmentAmount
            if invest > cash { invest = cash }
            fee := invest * e.feeRate
            totalFees += fee
            qty += (invest - fee) / price
            cash -= invest
            trades++
            nextBuy = nextBuy.Add(cfg.Interval)
        }
        equity = append(equity, cash+qty*price)
    }
    if len(equity) == 0 { return PerformanceMetrics{} }
    // wins proxy: last price above average buy -> count as win
    if qty > 0 {
        avg := (initialBalance - cash - totalFees) / qty
        if candles[len(candles)-1].Close > avg { wins = trades }
    }
    return computePerformance(equity, end.Sub(start), trades, wins, totalFees)
}


