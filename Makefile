.PHONY: build test clean run-dca run-grid

build:
	go build -o bin/dca-bot ./cmd/dca-bot
	go build -o bin/grid-bot ./cmd/grid-bot

test:
	go test -v ./...

clean:
	rm -rf bin/
