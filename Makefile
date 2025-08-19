# Crypto Trading Strategies Makefile

# Variables
BINARY_DIR = bin
DCA_BOT = $(BINARY_DIR)/dca-bot
GRID_BOT = $(BINARY_DIR)/grid-bot
COMBO_BOT = $(BINARY_DIR)/combo-bot
BACKTESTER = $(BINARY_DIR)/backtester

# Go build flags
LDFLAGS = -ldflags "-X main.version=1.0.0"

# Default target
.PHONY: all
all: build

# Create binary directory
$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

# Build all binaries
.PHONY: build
build: $(BINARY_DIR) $(DCA_BOT) $(GRID_BOT) $(COMBO_BOT) $(BACKTESTER)

# Build DCA bot
$(DCA_BOT): cmd/dca-bot/main.go
	go build $(LDFLAGS) -o $(DCA_BOT) ./cmd/dca-bot

# Build Grid bot
$(GRID_BOT): cmd/grid-bot/main.go
	go build $(LDFLAGS) -o $(GRID_BOT) ./cmd/grid-bot

# Build Combo bot
$(COMBO_BOT): cmd/combo-bot/main.go
	go build $(LDFLAGS) -o $(COMBO_BOT) ./cmd/combo-bot

# Build Backtester
$(BACKTESTER): cmd/backtester/main.go
	go build $(LDFLAGS) -o $(BACKTESTER) ./cmd/backtester

# Run tests
.PHONY: test
test:
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

# Run DCA bot
.PHONY: run-dca
run-dca: $(DCA_BOT)
	$(DCA_BOT) -config configs/dca-config.json

# Run Grid bot
.PHONY: run-grid
run-grid: $(GRID_BOT)
	$(GRID_BOT) -config configs/grid-config.json

# Run Combo bot
.PHONY: run-combo
run-combo: $(COMBO_BOT)
	$(COMBO_BOT) -config configs/combo-config.json

# Run backtester
.PHONY: run-backtest
run-backtest: $(BACKTESTER)
	$(BACKTESTER) -data test/data/BTCUSDT-1h.csv -start 2024-01-01T00:00:00Z -end 2024-01-31T23:59:59Z

# Install dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	golangci-lint run

# Create logs directory
.PHONY: setup
setup:
	mkdir -p logs
	mkdir -p $(BINARY_DIR)

# Docker build
.PHONY: docker-build
docker-build:
	docker build -t crypto-trading-bot:latest .

# Docker run DCA bot
.PHONY: docker-run-dca
docker-run-dca:
	docker run -p 8080:8080 crypto-trading-bot:latest $(DCA_BOT) -config configs/dca-config.json

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build all binaries"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  run-dca        - Run DCA bot"
	@echo "  run-grid       - Run Grid bot"
	@echo "  run-combo      - Run Combo bot"
	@echo "  run-backtest   - Run backtester"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  setup          - Create necessary directories"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run-dca - Run DCA bot in Docker"
	@echo "  help           - Show this help"
