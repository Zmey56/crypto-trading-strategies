.PHONY: build test clean run-dca run-grid run-backtest

build:
	go build -o bin/dca-bot ./cmd/dca-bot
	go build -o bin/grid-bot ./cmd/grid-bot
	go build -o bin/backtester ./cmd/backtester

test:
	go test -v ./...

clean:
	rm -rf bin/

run-backtest:
	bin/backtester -data test/data/BTCUSDT-1h.csv -symbol BTCUSDT -start 2024-01-01T00:00:00Z -end 2024-02-01T00:00:00Z
