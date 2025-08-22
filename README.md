# Crypto Trading Strategies

Project for automated cryptocurrency trading using various strategies.

## 🚀 Features

- **DCA (Dollar Cost Averaging)** - dollar cost averaging strategy ✅
- **Grid Trading** - grid trading ✅
- **Combo Strategies** - combined strategies ✅
- Multiple exchange support (Binance, Kraken) ✅
- Real-time portfolio monitoring ✅
- RESTful API for management ✅
- Detailed logging and metrics ✅
- Strategy backtesting ✅
- Docker containerization ✅
- Full test coverage ✅

## 📁 Project Structure

```
crypto-trading-strategies/
├── cmd/                    # Executable files
│   ├── dca-bot/           # DCA bot
│   ├── grid-bot/          # Grid bot
│   └── backtester/        # Backtester
├── internal/              # Internal packages
│   ├── config/            # Configuration
│   ├── exchange/          # Exchange clients
│   ├── strategy/          # Trading strategies
│   ├── portfolio/         # Portfolio management
│   └── logger/            # Logging
├── pkg/                   # Public packages
│   ├── types/             # Common data types
│   └── indicators/        # Technical indicators
├── configs/               # Configuration files
├── examples/              # Usage examples
└── docs/                  # Documentation
```

## 🛠️ Installation and Setup

### Requirements

- Go 1.21 or higher
- Exchange API keys (Binance, Kraken)

### Installation

```bash
# Clone repository
git clone <repository-url>
cd crypto-trading-strategies

# Install dependencies
go mod tidy

# Build
go build ./cmd/dca-bot
```

### Configuration

1. Copy the example configuration:
```bash
cp configs/dca-config.json configs/my-config.json
```

2. Edit the configuration:
```json
{
  "exchange": {
    "name": "binance",
    "api_key": "your-api-key",
    "secret_key": "your-secret-key",
    "sandbox": true
  },
  "strategy": {
    "dca": {
      "symbol": "BTCUSDT",
      "investment_amount": 100.0,
      "interval": "24h",
      "max_investments": 100,
      "enabled": true
    }
  }
}
```

### Running Bots

#### DCA Bot
```bash
# Build
make build

# Run with config file
make run-dca

# Or directly
./bin/dca-bot -config configs/dca-config.json

# With environment variables
export DCA_SYMBOL=BTCUSDT
export DCA_INVESTMENT_AMOUNT=100
export DCA_INTERVAL=24h
export EXCHANGE_API_KEY=your-api-key
export EXCHANGE_SECRET_KEY=your-secret-key
./bin/dca-bot
```

#### Grid Bot
```bash
# Run Grid bot
make run-grid

# Or directly
./bin/grid-bot -config configs/grid-config.json
```

#### Combo Bot
```bash
# Run Combo bot
make run-combo

# Or directly
./bin/combo-bot -config configs/combo-config.json
```

#### Docker
```bash
# Build and run all bots
docker-compose up -d

# Run only DCA bot
docker-compose up dca-bot

# Stop
docker-compose down
```

## 📊 DCA Strategy

DCA (Dollar Cost Averaging) - a dollar cost averaging strategy that involves regularly purchasing an asset for a fixed amount regardless of price.

### Working Principles

1. **Regular Investments**: Purchase for a fixed amount at specified intervals
2. **Automatic Execution**: Bot automatically places orders
3. **Risk Management**: Limiting the maximum number of investments
4. **Monitoring**: Tracking positions and metrics in real time

### DCA Configuration

```json
{
        "symbol": "BTCUSDT",           // Trading pair
      "investment_amount": 100.0,    // Investment amount in USDT
      "interval": "24h",             // Interval between purchases
      "max_investments": 100,        // Maximum number of investments
      "price_threshold": 0.0,        // Price threshold (0 = no restrictions)
      "stop_loss": 0.0,              // Stop Loss (0 = disabled)
      "take_profit": 0.0,            // Take Profit (0 = disabled)
      "enabled": true                // Enable/disable strategy
}
```

## 🔧 API

### Endpoints

- `GET /health` - Health check
- `GET /portfolio` - Portfolio information
- `GET /strategy/status` - Strategy status
- `POST /strategy/config` - Update configuration
- `GET /metrics` - Strategy metrics

### API Usage Example

```bash
# Get portfolio status
curl http://localhost:8080/portfolio

# Update configuration
curl -X POST http://localhost:8080/strategy/config \
  -H "Content-Type: application/json" \
  -d '{"investment_amount": 150.0}'
```

## 📈 Strategies

### DCA (Dollar Cost Averaging)
Dollar cost averaging strategy that involves regularly purchasing an asset for a fixed amount regardless of price.

**Working Principles:**
- Regular investments at specified intervals
- Automatic order execution
- Risk management through limits
- Real-time monitoring

### Grid Trading
Grid trading with order placement at different price levels.

**Working Principles:**
- Creating a grid of orders between upper and lower prices
- Automatic buying when price falls
- Automatic selling when price rises
- Profiting from volatility

### Combo Strategy
Combined strategy that combines multiple strategies with weighted coefficients.

**Working Principles:**
- Combining signals from different strategies
- Weighted decision making
- Risk diversification
- Adaptability to market conditions

## 📊 Monitoring

### Strategy Metrics

- **Total Trades**: Total number of trades
- **Win Rate**: Percentage of profitable trades
- **Total Profit/Loss**: Total profit/loss
- **Average Win/Loss**: Average profit/loss
- **Profit Factor**: Profit factor
- **Max Drawdown**: Maximum drawdown
- **Sharpe Ratio**: Sharpe ratio
- **Total Volume**: Total trading volume

### Logging

The bot maintains detailed logs of all operations:

```
[INFO] 🤖 DCA Bot starting...
[INFO] Version: 1.0.0
[INFO] Exchange: binance
[INFO] Symbol: BTCUSDT
[INFO] DCA Bot successfully started and running
[INFO] Mock: Placed order BTCUSDT 0.00222222 @ 45000.00
[INFO] DCA buy executed: BTCUSDT 0.00222222 @ 45000.00 (buy #1)
```

## 🛡️ Security

### Recommendations

1. **API Keys**: Use API keys only for trading, without withdrawal rights
2. **Sandbox**: Test on sandbox first
3. **Limits**: Set reasonable investment limits
4. **Monitoring**: Regularly check logs and metrics

### Environment Variables

```bash
# Secure API key storage
export EXCHANGE_API_KEY=your-api-key
export EXCHANGE_SECRET_KEY=your-secret-key
export EXCHANGE_SANDBOX=true
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test ./internal/strategy
go test ./internal/portfolio
```

### Backtesting

```bash
# Run backtester
make run-backtest

# Or directly
./bin/backtester -data test/data/BTCUSDT-1h.csv -start 2024-01-01T00:00:00Z -end 2024-01-31T23:59:59Z
```

### Code Quality Check

```bash
# Format code
make fmt

# Linting
make lint
```

## 📝 Development

### Adding New Strategy

1. Create a new file in `internal/strategy/`
2. Implement the `Strategy` interface
3. Add configuration to `pkg/types/types.go`
4. Update strategy factory
5. Add tests

### Adding New Exchange

1. Create client in `internal/exchange/`
2. Implement `ExchangeClient` interface
3. Add to `UnifiedClient`
4. Add tests

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes
4. Add tests
5. Create Pull Request

## 📄 License

MIT License

## ⚠️ Disclaimer

This project is for educational purposes only. Cryptocurrency trading involves high risks. The authors are not responsible for possible financial losses.

## 📞 Support

- Issues: [GitHub Issues](https://github.com/your-repo/issues)
- Discussions: [GitHub Discussions](https://github.com/your-repo/discussions)
- Email: support@example.com

## 📚 Related Articles

### Comparing Strategies: DCA vs Grid Trading

For a detailed comparison of DCA and Grid trading strategies, including performance analysis, risk assessment, and practical implementation insights, check out our comprehensive article:

**[Comparing Strategies: DCA vs Grid Trading](https://medium.com/@alsgladkikh/comparing-strategies-dca-vs-grid-trading-2724fa809576)**

This article provides:
- Deep dive into both DCA and Grid trading methodologies
- Performance comparison across different market conditions
- Risk analysis and drawdown scenarios
- Practical implementation guidelines
- Real-world examples and case studies

## 🚀 From Theory to Practice

For those who want to move from theory to practice, I provide daily actionable DCA signals:

- **Twitter**: [@Algo_Adviser](https://twitter.com/Algo_Adviser)
- **Telegram**: [AlgoAdviser](https://t.me/AlgoAdviser)

### 🎯 Get Started with Bitsgap

You can also get a **7-day PRO trial** on Bitsgap through my referral link: **[bitsgap.com/?ref=algo-adviser](https://bitsgap.com/?ref=algo-adviser)**

Traders who register via this link will receive priority access to my private channels in the future, featuring:
- A wider range of DCA signals across more coins
- Exclusive GRID strategy setups
- Advanced portfolio management techniques
- Real-time market analysis and insights