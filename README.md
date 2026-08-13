# Crypto Trading Strategies

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Status: Research](https://img.shields.io/badge/status-research%2Feducational-orange)]()

A Go implementation and side-by-side comparison of two classic automated trading strategies — **DCA (Dollar Cost Averaging)** and **Grid Trading** — plus a **Combo** strategy that runs both at once.

The repository ships three long-running bots, a CSV-driven backtesting engine that scores DCA against Grid on the same price series, and a small HTTP API for observing a running bot.

> **This is a research and educational project.** It is not audited, not production-hardened, and comes with no guarantee of profit. See [Project status](#-project-status) for an honest breakdown of what is implemented and what is not.

---

## Table of Contents

- [Project status](#-project-status)
- [Architecture](#-architecture)
- [Quick start](#-quick-start)
- [Project structure](#-project-structure)
- [Configuration](#-configuration)
- [Strategies](#-strategies)
- [Backtester](#-backtester)
- [HTTP API](#-http-api)
- [Docker & deployment](#-docker--deployment)
- [Testing](#-testing)
- [Extending the project](#-extending-the-project)
- [Known limitations & roadmap](#-known-limitations--roadmap)
- [Security](#-security)
- [FAQ](#-faq)
- [License & disclaimer](#-license--disclaimer)

---

## 🚦 Project status

Being upfront about maturity matters more than a long feature list, so here is what actually works today.

| Area | Status | Notes |
|---|---|---|
| DCA strategy | ✅ Implemented | Interval-based buying with a max-investment cap and optional price threshold |
| Grid strategy | ✅ Implemented | Static grid between two bounds, buy on level touch, sell at the next level up |
| Combo strategy | ✅ Implemented | Runs DCA + Grid together, equal weights, weighted signal aggregation |
| Backtesting engine | ✅ Implemented | CSV input, DCA vs Grid comparison, fee-aware |
| HTTP monitoring API | ✅ Implemented | 5 endpoints, JSON, served by each bot |
| Docker / Compose / K8s manifests | ✅ Provided | Dev and prod compose files, K8s deployment, Prometheus/Grafana configs |
| **Live exchange trading** | ⚠️ **Not wired up** | The three bots run against an **in-memory mock exchange**. See below. |
| Binance REST client | 🟡 Written, unused | `internal/exchange/binance` implements the full `ExchangeClient` interface with HMAC signing and testnet support, but no bot instantiates it |
| Kraken / other exchanges | ❌ Not implemented | `createExchangeClient` returns `unsupported exchange` for anything but Binance |
| Stop-loss / take-profit | ❌ Config only | `stop_loss` and `take_profit` are parsed and validated but never acted on |
| Test coverage | 🟡 Partial | Unit tests exist for `internal/strategy` only |
| `internal/ai`, `compliance`, `crosschain`, `risk`, `analytics`, `pkg/nlp` | 🧪 Experimental | Compile, but are not reachable from any binary — scaffolding for future work |

### About the mock exchange

Each bot in `cmd/` defines its own `MockExchangeClient` inline and uses it unconditionally:

```go
// cmd/dca-bot/main.go
exchange := createMockExchange(cfg, log)
```

The mock reports a fixed $10,000 balance, a synthetic ticker, and instantly "fills" every order at the requested price. **Running a bot therefore places no real orders and risks no real funds** — which is the intended default. Wiring in the real Binance client is a deliberate, one-line-plus-review change; see [Going live](#going-live-with-a-real-exchange).

---

## 🏗 Architecture

```
                 ┌─────────────────────────────────────────────┐
                 │            cmd/{dca,grid,combo}-bot         │
                 │                                             │
   config file ──▶  config.Load()  ──┐                         │
   or ENV vars  ──▶  config.LoadFromEnv()                      │
                 │                   │                         │
                 │                   ▼                         │
                 │        strategy.Factory                     │
                 │                   │                         │
                 │      ┌────────────┼────────────┐            │
                 │      ▼            ▼            ▼            │
                 │  DCAStrategy  GridStrategy  ComboStrategy   │
                 │      └────────────┼────────────┘            │
                 │                   │  Strategy interface     │
                 │                   ▼                         │
                 │   runTradingLoop (ticks every 1 min)        │
                 │     1. exchange.GetTicker(symbol)           │
                 │     2. strategy.Execute(ctx, marketData)    │
                 │     3. log strategy.GetMetrics()            │
                 │                   │                         │
                 │                   ▼                         │
                 │        types.ExchangeClient                 │
                 │      (MockExchangeClient by default)        │
                 │                                             │
                 │   portfolio.Manager ── auto-refresh 30s     │
                 │   net/http server  ── :8080 /health /metrics│
                 └─────────────────────────────────────────────┘

                 ┌─────────────────────────────────────────────┐
                 │              cmd/backtester                 │
                 │  CSV candles ─▶ backtest.Engine             │
                 │       ├─ BacktestDCA  ─┐                    │
                 │       └─ BacktestGrid ─┴─▶ StrategyComparison│
                 │                              (JSON stdout)  │
                 └─────────────────────────────────────────────┘
```

Every strategy satisfies one small interface, which is what makes the Combo strategy and the backtester possible:

```go
// internal/strategy/interface.go
type Strategy interface {
    Execute(ctx context.Context, market types.MarketData) error
    GetSignal(market types.MarketData) types.Signal
    ValidateConfig() error
    GetMetrics() types.StrategyMetrics
    Shutdown(ctx context.Context) error
}
```

The exchange side is abstracted behind `types.ExchangeClient` (14 methods covering orders, market data, account info and connection lifecycle), so swapping the mock for a real venue does not touch strategy code.

---

## 🚀 Quick start

### Requirements

- **Go 1.23+** (the module declares `go 1.23.0`, toolchain `go1.24.2`)
- Optional: Docker + Docker Compose for containerised runs
- No exchange account is needed — the bots run against a mock by default

### Build and run

```bash
git clone https://github.com/Zmey56/crypto-trading-strategies.git
cd crypto-trading-strategies

# Fetch dependencies (the only external one is golang.org/x/time)
make deps

# Build all four binaries into ./bin
make build

# Run the DCA bot with the sample config
make run-dca
```

You should see:

```
[INFO] 🤖 DCA Bot starting...
[INFO] Version: 1.0.0
[INFO] Exchange: binance
[INFO] Symbol: BTCUSDT
[INFO] Trading loop started for BTCUSDT
[INFO] DCA Bot started and running
[INFO] HTTP server started on port 8080
```

In another terminal:

```bash
curl -s localhost:8080/strategy/status | jq
curl -s localhost:8080/metrics | jq
```

Stop with `Ctrl+C` — the bot performs a graceful shutdown with a 30s timeout.

### Run a backtest

```bash
make run-backtest
```

### All Make targets

| Target | Description |
|---|---|
| `make build` | Build `dca-bot`, `grid-bot`, `combo-bot`, `backtester` into `bin/` |
| `make run-dca` / `run-grid` / `run-combo` | Build if needed, then run with the matching config from `configs/` |
| `make run-backtest` | Run the backtester over the bundled sample CSV |
| `make test` | `go test -v ./...` |
| `make test-coverage` | Tests + `coverage.html` |
| `make fmt` | `go fmt ./...` |
| `make lint` | `golangci-lint run` (requires golangci-lint installed) |
| `make deps` | `go mod tidy && go mod download` |
| `make setup` | Create `logs/` and `bin/` |
| `make clean` | Remove `bin/` and coverage artifacts |
| `make docker-build` | Build the `crypto-trading-bot:latest` image |
| `make help` | List all targets |

---

## 📁 Project structure

```
crypto-trading-strategies/
├── cmd/
│   ├── dca-bot/           # DCA bot: config → strategy → 1-min loop → HTTP API
│   ├── grid-bot/          # Grid bot, same shape
│   ├── combo-bot/         # Combo bot, same shape
│   └── backtester/        # CLI: CSV in, DCA-vs-Grid JSON comparison out
├── internal/
│   ├── strategy/          # ★ Core: DCA, Grid, Combo + factory + interface (+ tests)
│   ├── backtest/          # ★ Engine, CSV loader, per-strategy sims, metrics, comparison
│   ├── config/            # JSON file + environment variable loading, validation
│   ├── portfolio/         # Position tracking, P&L, periodic refresh
│   ├── exchange/
│   │   ├── binance/       # Full Binance REST client (HMAC-signed, testnet-aware) — unused by bots
│   │   ├── mock/          # Standalone mock client
│   │   └── client.go      # UnifiedClient + exchange factory
│   ├── logger/            # Levelled logger, optional file output
│   ├── analytics/         # Performance + psychological metrics (experimental)
│   ├── risk/              # Risk manager (experimental)
│   ├── ai/                # ML engine, sentiment analysis (experimental)
│   ├── compliance/        # KYC/AML, tax reporting (experimental)
│   ├── crosschain/        # Cross-chain arbitrage engine (experimental)
│   └── app/               # DI container (experimental)
├── pkg/
│   ├── types/             # ★ Shared domain types: Order, Ticker, Signal, *Config, ExchangeClient
│   ├── indicators/        # SMA, EMA, RSI, MACD, Bollinger Bands, Stochastic
│   └── nlp/               # Text processor (experimental)
├── configs/               # JSON configs consumed by the bots + YAML reference presets
├── deployments/
│   ├── docker/            # dev & prod compose, init.sql, nginx.conf
│   ├── kubernetes/        # namespace, configmap, secret template, deployment, ingress
│   ├── monitoring/        # Prometheus config, Grafana dashboards, alerting rules
│   └── scripts/           # deploy.sh, backup-database.sh
├── test/data/             # Sample OHLCV CSV
├── Dockerfile             # Multi-stage build, non-root runtime user
├── docker-compose.yml     # All three bots + backtester
└── Makefile
```

★ = the packages that carry the actual trading logic.

---

## ⚙️ Configuration

A bot resolves its configuration in one of two ways:

1. **`-config path/to/file.json`** — loads and validates a JSON file (`config.Load`).
2. **No flag** — builds the config entirely from environment variables (`config.LoadFromEnv`).

These are mutually exclusive: passing `-config` ignores the environment completely.

### JSON schema

```json
{
  "app": {
    "name": "crypto-dca-bot",
    "version": "1.0.0",
    "port": 8080,
    "debug": false
  },
  "exchange": {
    "name": "binance",
    "api_key": "your-api-key-here",
    "secret_key": "your-secret-key-here",
    "passphrase": "",
    "sandbox": true
  },
  "strategy": {
    "dca":   { "...": "see DCA section" },
    "grid":  { "...": "see Grid section" },
    "combo": { "...": "see Combo section" }
  },
  "logging": {
    "level": "info",
    "file": "logs/dca-bot.log",
    "format": "text"
  }
}
```

`Validate()` requires a non-empty `app.name`, `exchange.name`, `exchange.api_key` and `exchange.secret_key`. The placeholder values in the shipped configs satisfy this — the mock exchange never checks them.

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `APP_NAME` | `crypto-trading-bot` | Application name |
| `APP_VERSION` | `1.0.0` | Reported in logs and `/health` |
| `APP_PORT` | `8080` | HTTP API port; set to `0` to disable the server |
| `APP_DEBUG` | `false` | Debug flag |
| `EXCHANGE_NAME` | `binance` | Exchange identifier |
| `EXCHANGE_API_KEY` | *(empty)* | API key |
| `EXCHANGE_SECRET_KEY` | *(empty)* | API secret |
| `EXCHANGE_PASSPHRASE` | *(empty)* | Passphrase (venues that need one) |
| `EXCHANGE_SANDBOX` | `true` | `true` → `testnet.binance.vision`, `false` → `api.binance.com` |
| `DCA_SYMBOL` | `BTCUSDT` | Trading pair |
| `DCA_INVESTMENT_AMOUNT` | `100.0` | Quote-currency amount per buy |
| `DCA_INTERVAL` | `24h` | Go duration string (`30m`, `4h`, `24h`, …) |
| `DCA_MAX_INVESTMENTS` | `100` | Hard cap on the number of buys |
| `DCA_PRICE_THRESHOLD` | `0.0` | Only buy at or below this price; `0` disables |
| `DCA_STOP_LOSS` | `0.0` | ⚠️ Parsed but not implemented |
| `DCA_TAKE_PROFIT` | `0.0` | ⚠️ Parsed but not implemented |
| `DCA_ENABLED` | `true` | Enable the strategy |
| `LOG_LEVEL` | `info` | `debug` \| `info` \| `warn` \| `error` |
| `LOG_FILE` | *(empty)* | Log file path; empty → stdout only |
| `LOG_FORMAT` | `text` | Log format |

Booleans accept `true/1/yes` and `false/0/no`. Note that `LoadFromEnv` only populates the **DCA** strategy block — Grid and Combo bots need a JSON config file.

### A note on the YAML files

`configs/beginner-dca-config.yaml`, `configs/dca-strategies.yaml` and `configs/grid-strategies.yaml` are **reference presets, not runtime configuration**. No YAML parser is linked into the project (`golang.org/x/time` is the sole dependency), and `config.Load` only decodes JSON. Treat them as documented parameter sets to copy into JSON or into your own tooling.

---

## 📊 Strategies

### DCA (Dollar Cost Averaging)

Buy a fixed quote amount at a fixed interval, regardless of price. The intent is to smooth the average entry price over time rather than to time the market.

**Execution logic** (`internal/strategy/dca.go`), evaluated on every 1-minute tick:

1. Skip if `enabled` is false.
2. Skip if less than `interval` has elapsed since the last buy.
3. Skip if `buyCount >= max_investments`.
4. Skip if `price_threshold > 0` and the current price is **above** it.
5. Otherwise place a market buy for `investment_amount / price` units and record the fill.

```json
{
  "symbol": "BTCUSDT",
  "investment_amount": 100.0,
  "interval": "24h",
  "max_investments": 100,
  "price_threshold": 0.0,
  "stop_loss": 0.0,
  "take_profit": 0.0,
  "enabled": true
}
```

| Field | Type | Meaning |
|---|---|---|
| `symbol` | string | Trading pair, e.g. `BTCUSDT` |
| `investment_amount` | float | Quote currency spent per buy |
| `interval` | duration string | Minimum time between buys |
| `max_investments` | int | Stop after this many buys |
| `price_threshold` | float | Buy only at or below this price; `0` = always |
| `stop_loss` / `take_profit` | float | ⚠️ Validated, currently no effect |
| `enabled` | bool | Master switch |

DCA is the only strategy exposing a live-tunable config over HTTP — see [`POST /strategy/config`](#post-strategyconfig).

---

### Grid Trading

Place a ladder of price levels between a lower and an upper bound, buy when price drops onto an empty level, and sell that position once price reaches the next level up. The strategy monetises oscillation rather than direction.

**Level construction** (`internal/strategy/grid.go`):

```
step = (upper_price - lower_price) / (grid_levels - 1)
level[i] = lower_price + i * step
```

**Execution logic**, on every tick, iterating levels in ascending order:

- **Buy** if `price <= level` and that level holds no position → market buy `investment_per_level / price` units, record `{quantity, avgPrice}` against the level.
- **Sell** if the level holds a position and `price >= level[i+1]` → market sell the whole position, realise `(price - avgPrice) * quantity`, clear the level.

Win rate, profit factor and total volume are recomputed after each pass.

```json
{
  "symbol": "BTCUSDT",
  "upper_price": 50000.0,
  "lower_price": 40000.0,
  "grid_levels": 20,
  "investment_per_level": 100.0,
  "enabled": true
}
```

| Field | Type | Meaning |
|---|---|---|
| `symbol` | string | Trading pair |
| `upper_price` | float | Top of the grid |
| `lower_price` | float | Bottom of the grid; must be `> 0` and `< upper_price` |
| `grid_levels` | int | Number of levels; must be `> 1` |
| `investment_per_level` | float | Quote currency committed per level |
| `enabled` | bool | Master switch |

Total capital at risk is roughly `grid_levels × investment_per_level`. The grid is **static** — if price leaves the `[lower, upper]` band the strategy stops trading and holds whatever it accumulated.

`GetSignal` returns `HOLD` for the Grid strategy by design: all decisions happen inside `Execute`, so there is no meaningful single signal to report.

---

### Combo

Runs several strategies over the same market data and aggregates their signals.

**Behaviour** (`internal/strategy/combo.go`):

- Instantiates each child strategy via the factory; only `dca` and `grid` types are supported.
- Assigns **equal weights**: `weight = 1 / len(strategies)`. Weights are not configurable today.
- `Execute` calls every child in order; a failing child is logged and skipped rather than aborting the pass.
- `GetSignal` accumulates `strength × weight` per child, normalises by the total weight, and downgrades the result to `HOLD` if the normalised strength falls below **0.3**.
- `GetMetrics` sums trades, volume and P&L across children and recomputes win rate and profit factor on the aggregate.

```json
{
  "combo": {
    "strategies": [
      { "type": "dca",  "config": { "symbol": "BTCUSDT", "investment_amount": 50.0, "interval": "24h", "max_investments": 50, "enabled": true } },
      { "type": "grid", "config": { "symbol": "BTCUSDT", "upper_price": 50000.0, "lower_price": 40000.0, "grid_levels": 10, "investment_per_level": 50.0, "enabled": true } }
    ],
    "enabled": true
  }
}
```

Note that each child strategy places its own orders independently — the Combo strategy aggregates *reporting*, not *execution*. Budget accordingly.

---

## 🔬 Backtester

`cmd/backtester` loads an OHLCV CSV, runs the DCA and Grid simulations over the same candles, and prints a JSON comparison to stdout.

### CSV format

Header row required, timestamps in RFC3339:

```csv
timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,42000,42100,41900,42050,100
2024-01-01T01:00:00Z,42050,42200,42000,42150,110
```

Rows with fewer than 6 fields are skipped. Simulations use the **close** price of each candle.

### Flags

| Flag | Default | Description |
|---|---|---|
| `-data` | *(required)* | Path to the CSV file |
| `-start` | *(required)* | Window start, RFC3339 |
| `-end` | *(required)* | Window end, RFC3339 |
| `-symbol` | `BTCUSDT` | Symbol label |
| `-initial` | `10000` | Starting balance |
| `-fee` | `0.001` | Taker fee rate applied to both sides |
| `-dca-interval` | `24h` | DCA buy interval |
| `-dca-amount` | `100` | DCA amount per buy |
| `-dca-max` | `100` | DCA max buys |
| `-grid-lower` | `30000` | Grid lower bound |
| `-grid-upper` | `60000` | Grid upper bound |
| `-grid-levels` | `20` | Grid level count |
| `-grid-invest` | `100` | Grid investment per level |

### Example

```bash
./bin/backtester \
  -data test/data/BTCUSDT-1h.csv \
  -start 2024-01-01T00:00:00Z \
  -end 2024-01-31T23:59:59Z \
  -fee 0.001
```

Actual output against the bundled sample:

```json
{
  "dca_results": {
    "total_return": 0.005939417301425998,
    "annualized_return": 0.06995425487437146,
    "max_drawdown": 0.011878129111200144,
    "sharpe_ratio": 0.1913275728352993,
    "trade_count": 1,
    "win_rate": 100,
    "total_fees": 0.1,
    "volatility_impact": 0.0025870365937091704
  },
  "grid_results": {
    "total_return": 0.07128084858887274,
    "annualized_return": 0.842505107195346,
    "max_drawdown": 0.14246015037365323,
    "sharpe_ratio": 0.19141434747847563,
    "trade_count": 12,
    "win_rate": 0,
    "total_fees": 1.2,
    "volatility_impact": 0.031047571440384992
  },
  "backtest_period": 2678399000000000,
  "market_condition": "sideways"
}
```

`market_condition` is a coarse heuristic over first-vs-last close in the window: `> +10%` → `bull`, `< -10%` → `bear`, otherwise `sideways`.

> ⚠️ **The bundled `test/data/BTCUSDT-1h.csv` contains only 14 hourly candles.** It exists to make the pipeline runnable, not to produce meaningful statistics. The numbers above are a smoke test, not a result. Supply your own multi-month dataset before drawing any conclusion.

Model simplifications worth knowing about: fills are assumed instant at the close price, there is no slippage or spread model, no partial fills, no order rejection, and no funding costs. Fees are the only friction modelled.

---

## 🔧 HTTP API

Every bot starts a `net/http` server on `app.port` (disable by setting the port to `0`). All responses are JSON.

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe → `{"status":"ok"}` |
| `GET` | `/portfolio` | Current portfolio snapshot: total value, positions, P&L |
| `GET` | `/strategy/status` | Strategy-specific state, if the strategy implements `GetStatus()` |
| `GET` | `/metrics` | `{"strategy": StrategyMetrics, "portfolio": {...}}` |
| `POST` | `/strategy/config` | Partial live reconfiguration (DCA only) |

Default ports: DCA `8080`, Grid `8081`, Combo `8082`.

### `POST /strategy/config`

Accepts a partial JSON object; only the fields present are applied, and the merged config is re-validated before it takes effect.

Supported keys: `investment_amount`, `max_investments`, `price_threshold`, `stop_loss`, `take_profit`, `enabled`, `interval` (duration string).

```bash
curl -X POST localhost:8080/strategy/config \
  -H 'Content-Type: application/json' \
  -d '{"investment_amount": 150.0, "interval": "12h"}'
# {"status":"updated"}
```

Strategies that do not implement the updater interface return `400` with `{"error":"strategy does not support config updates"}` — this is the case for Grid and Combo.

> 🔒 **The API has no authentication whatsoever.** Anyone who can reach the port can change trading parameters. Bind it to localhost or put it behind the provided nginx config with proper auth before exposing it anywhere.

### `StrategyMetrics` shape

`total_trades`, `winning_trades`, `losing_trades`, `total_profit`, `total_loss`, `win_rate`, `average_win`, `average_loss`, `profit_factor`, `max_drawdown`, `sharpe_ratio`, `total_volume`, `last_update`.

---

## 🐳 Docker & deployment

### Local Compose

The root `docker-compose.yml` builds one image and starts all three bots plus a one-shot backtester:

```bash
docker compose up -d          # all services
docker compose up dca-bot     # just the DCA bot
docker compose logs -f dca-bot
docker compose down
```

All three services pin `EXCHANGE_SANDBOX=true` and mount `./configs` and `./logs` from the host.

The image is a multi-stage build: `golang:1.23-alpine` compiles static `CGO_ENABLED=0` binaries, and the runtime stage is a bare `alpine` running as the non-root user `appuser` (uid 1001).

### Development stack

`deployments/docker/docker-compose.dev.yml` adds live-reload volume mounts, a [Delve](https://github.com/go-delve/delve) debugger on port `2345`, Postgres, and pgAdmin.

It **requires** two secrets to be present — there are no default passwords, by design:

```bash
cd deployments/docker
cat > .env <<'EOF'
POSTGRES_PASSWORD=<choose-a-strong-password>
PGADMIN_PASSWORD=<choose-a-strong-password>
EOF
docker compose -f docker-compose.dev.yml up -d
```

Omitting them fails fast:

```
error while interpolating services.postgres-dev.environment.[]:
required variable POSTGRES_PASSWORD is missing a value
```

### Production stack & Kubernetes

- `deployments/docker/docker-compose.prod.yml` — production topology with nginx in front.
- `deployments/kubernetes/` — `namespace`, `configmap`, `secret` (template with `<base64-encoded-…>` placeholders — fill it from your own secret store, do not commit real values), `dca-bot-deployment`, `ingress`.
- `deployments/monitoring/` — Prometheus scrape config, Grafana dashboards, alerting rules.
- `deployments/scripts/deploy.sh` — refuses to run unless `POSTGRES_PASSWORD` is exported.

See [`deployments/README.md`](deployments/README.md) for the full walkthrough.

> Before deploying: the K8s ConfigMap ships `exchange-sandbox: "false"` and `"sandbox": false`. Review that setting deliberately — combined with a real exchange client it means live trading.

---

## 🧪 Testing

```bash
make test                      # go test -v ./...
make test-coverage             # writes coverage.html
go test ./internal/strategy    # the package that actually has tests
```

Current state: `internal/strategy` has unit tests covering DCA, Grid and Combo (`dca_test.go`, `grid_test.go`, `combo_test.go`). Every other package reports `[no test files]`. Contributions here are the single most useful thing you could send.

Static checks:

```bash
make fmt      # go fmt ./...
make lint     # golangci-lint run
```

---

## 🔌 Extending the project

### Adding a new strategy

1. Create `internal/strategy/your_strategy.go` and implement the five-method `Strategy` interface.
2. Add a config struct to `pkg/types/types.go` (with a custom `UnmarshalJSON` if it carries a duration field — see `DCAConfig`).
3. Add a `CreateYourStrategy` method plus a `validateYourStrategyConfig` to `internal/strategy/factory.go`.
4. Register the new `type` string in `ComboStrategy.initializeStrategies` if it should be combinable.
5. Add a field to `config.StrategyConfig`.
6. Write tests in `internal/strategy/your_strategy_test.go`.

### Adding a new exchange

1. Create `internal/exchange/yourvenue/client.go` implementing all of `types.ExchangeClient`.
2. Add a `case "yourvenue":` branch to `createExchangeClient` in `internal/exchange/client.go`.
3. Handle rate limiting (`golang.org/x/time/rate` is already a dependency) and request signing.
4. Test against the venue's sandbox before anything else.

### Going live with a real exchange

The bots deliberately hardcode the mock. To trade for real you would:

1. Replace `createMockExchange(cfg, log)` in the relevant `cmd/*/main.go` with a real client — `binance.NewClient(...)` is already written and testnet-aware.
2. Verify `sandbox: true` resolves to `https://testnet.binance.vision` and run there for a meaningful period first.
3. Read the whole order path yourself. The mock never rejects, never partially fills and never rate-limits; real venues do all three.
4. Use API keys scoped to trading only, with withdrawals disabled and IP allow-listing on.
5. Add position limits and a kill switch. Neither exists in this codebase.

Do not skip steps 2–5.

---

## ⚠️ Known limitations & roadmap

**Limitations**

- Bots trade against a mock; nothing is wired to a live venue out of the box.
- `stop_loss` / `take_profit` are inert configuration fields.
- Combo weights are fixed at equal and cannot be configured.
- The grid is static — no re-centring when price exits the band.
- The trading loop polls on a fixed 1-minute ticker; there is no WebSocket/streaming path.
- The HTTP API is unauthenticated.
- `LoadFromEnv` covers DCA only.
- Backtests ignore slippage, spread, partial fills and funding.
- The module path (`github.com/Zmey56/crypto-arbitrage-trader`) does not match the repository name — a leftover from an earlier project.
- Test coverage is limited to `internal/strategy`.

**Possible directions**

- Wire the Binance client behind a config switch instead of a code edit
- Implement stop-loss / take-profit
- Configurable Combo weights and a signal-driven (rather than parallel) execution mode
- Dynamic / re-centring grid
- WebSocket market data
- Auth on the HTTP API
- Broader test coverage and a larger sample dataset
- Adapt the Grid logic to Uniswap v3 ranges

---

## 🛡 Security

**If you ever point this at a real account:**

1. **Key scope** — create API keys with trading permission only. Never enable withdrawals. Enable IP allow-listing.
2. **Never commit secrets** — `.env` is gitignored. Configs in `configs/` contain placeholders only; keep it that way. The K8s `secret.yaml` is a template, not a store.
3. **Sandbox first** — `EXCHANGE_SANDBOX=true` routes to `testnet.binance.vision`. Prove the behaviour there.
4. **Cap the downside** — `max_investments` for DCA and `grid_levels × investment_per_level` for Grid are your only built-in exposure limits. Set them consciously.
5. **Don't expose the API** — it is unauthenticated by design-omission.
6. **Watch the logs** — every order the strategies place is logged at `INFO`.

Found a security issue? Please open an issue, or reach out privately via the contact links below rather than posting exploit details publicly.

---

## ❓ FAQ

**Is this a "get-rich-quick bot"?**
No. It is an educational and research project for studying and comparing DCA and Grid strategies. Results depend entirely on market conditions and your own risk management.

**Do I need real API keys?**
No. The shipped configs contain placeholders and the bots use a mock exchange. Everything runs locally.

**Which exchanges are supported?**
A complete Binance REST client is implemented (mainnet + testnet) but is not currently wired into the bots. No other venue is implemented. Adding one means implementing `types.ExchangeClient`.

**Does this execute real trades?**
Not as shipped. The bots place orders against an in-memory mock. Live trading requires a deliberate code change — see [Going live](#going-live-with-a-real-exchange).

**How safe is this?**
No data leaves your machine, and keys (if you add them) stay local. But the project is unaudited, has no authentication on its API, and has no kill switch. Do not use it with meaningful funds.

**What metrics do the backtests report?**
Total return, annualized return, max drawdown, Sharpe ratio, trade count, win rate, total fees and volatility impact — for DCA and Grid side by side, plus a coarse market-condition label. Extend them in `internal/backtest`.

**Can I add new strategies?**
Yes — implement the five-method `Strategy` interface. See [Extending the project](#-extending-the-project).

**How reproducible are the results?**
The backtester is fully deterministic: same CSV and same flags produce identical output. There is no randomness in the simulation path. Reproduce with `make run-backtest`.

**Can this be ported to DeFi / AMMs?**
Not today — the code targets CEX REST APIs. Adapting Grid logic to Uniswap v3 ranges is on the roadmap.

**Why is the commit history small?**
It is a side project maintained for research. Issues and pull requests are welcome.

---

## 📚 Related reading

### Comparing Strategies: DCA vs Grid Trading

A detailed write-up of the comparison this repository implements — methodology, performance across market regimes, risk and drawdown analysis, and practical implementation notes:

**[Comparing Strategies: DCA vs Grid Trading](https://medium.com/@alsgladkikh/comparing-strategies-dca-vs-grid-trading-2724fa809576)**

---

## 🚀 From theory to practice

For those who want to move from theory to practice, I publish daily actionable DCA signals:

- **Twitter**: [@Algo_Adviser](https://twitter.com/Algo_Adviser)
- **Telegram**: [AlgoAdviser](https://t.me/AlgoAdviser)

### Get started with Bitsgap

You can get a **7-day PRO trial** on Bitsgap through my referral link: **[bitsgap.com/?ref=algo-adviser](https://bitsgap.com/?ref=algo-adviser)** *(referral link — I may receive a commission)*.

Traders who register via this link will receive priority access to my private channels in the future, featuring a wider range of DCA signals, exclusive GRID setups, portfolio management techniques and market analysis.

---

## 🤝 Contributing

1. Fork the repository and create a feature branch.
2. Keep the change focused; run `make fmt`, `make lint` and `make test`.
3. Add tests — especially for packages that currently have none.
4. Open a pull request describing what changed and why.

Bug reports and questions are welcome via [GitHub Issues](https://github.com/Zmey56/crypto-trading-strategies/issues).

---

## 📄 License & disclaimer

Released under the **MIT License**.

**Disclaimer.** This software is provided for educational and research purposes only. It is not financial advice. Cryptocurrency trading carries substantial risk, including the total loss of your capital. The code is unaudited and provided "as is", without warranty of any kind. The authors accept no liability for any financial loss arising from its use. You are solely responsible for any funds you put at risk.
