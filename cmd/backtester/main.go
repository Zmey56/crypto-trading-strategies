package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/backtest"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

func main() {
	data := flag.String("data", "", "Path to CSV (timestamp,open,high,low,close,volume)")
	symbol := flag.String("symbol", "BTCUSDT", "Symbol")
	start := flag.String("start", "", "Start (RFC3339)")
	end := flag.String("end", "", "End (RFC3339)")
	initBal := flag.Float64("initial", 10000, "Initial balance")
	dcaInterval := flag.String("dca-interval", "24h", "DCA interval")
	dcaAmt := flag.Float64("dca-amount", 100, "DCA investment amount")
	dcaMax := flag.Int("dca-max", 100, "DCA max investments")
	gridLower := flag.Float64("grid-lower", 30000, "Grid lower bound")
	gridUpper := flag.Float64("grid-upper", 60000, "Grid upper bound")
	gridLevels := flag.Int("grid-levels", 20, "Grid levels")
	gridInv := flag.Float64("grid-invest", 100, "Grid investment per level")
	fee := flag.Float64("fee", 0.001, "Taker fee rate")
	flag.Parse()

	if *data == "" || *start == "" || *end == "" {
		fmt.Fprintln(os.Stderr, "usage: backtester -data file.csv -start RFC3339 -end RFC3339 [opts]")
		os.Exit(2)
	}

	startT, err := time.Parse(time.RFC3339, *start)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	endT, err := time.Parse(time.RFC3339, *end)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	d, err := time.ParseDuration(*dcaInterval)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	eng := backtest.NewEngine(*fee)
	candles, err := eng.LoadCSV(*data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dcaCfg := types.DCAConfig{Symbol: *symbol, InvestmentAmount: *dcaAmt, Interval: d, MaxInvestments: *dcaMax, Enabled: true}
	gridCfg := types.GridConfig{Symbol: *symbol, UpperPrice: *gridUpper, LowerPrice: *gridLower, GridLevels: *gridLevels, InvestmentPerLevel: *gridInv, Enabled: true}
	cmp, err := eng.CompareStrategies(*symbol, candles, startT, endT, *initBal, dcaCfg, gridCfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(cmp)
}
