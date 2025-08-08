package backtest

import (
    "encoding/csv"
    "fmt"
    "io"
    "os"
    "strconv"
    "time"
)

type Candle struct {
    Time   time.Time
    Open   float64
    High   float64
    Low    float64
    Close  float64
    Volume float64
}

type Engine struct {
    feeRate float64 // taker fee rate e.g. 0.001
}

func NewEngine(feeRate float64) *Engine { return &Engine{ feeRate: feeRate } }

func (e *Engine) LoadCSV(path string) ([]Candle, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()
    r := csv.NewReader(f)
    r.FieldsPerRecord = -1
    var out []Candle
    // expect header: timestamp,open,high,low,close,volume
    _, _ = r.Read()
    for {
        rec, err := r.Read()
        if err == io.EOF { break }
        if err != nil { return nil, err }
        if len(rec) < 6 { continue }
        ts, _ := time.Parse(time.RFC3339, rec[0])
        open, _ := strconv.ParseFloat(rec[1], 64)
        high, _ := strconv.ParseFloat(rec[2], 64)
        low, _ := strconv.ParseFloat(rec[3], 64)
        closeP, _ := strconv.ParseFloat(rec[4], 64)
        vol, _ := strconv.ParseFloat(rec[5], 64)
        out = append(out, Candle{ Time: ts, Open: open, High: high, Low: low, Close: closeP, Volume: vol })
    }
    if len(out) == 0 { return nil, fmt.Errorf("no candles loaded") }
    return out, nil
}


