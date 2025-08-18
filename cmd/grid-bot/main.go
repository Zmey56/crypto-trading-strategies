package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/config"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/portfolio"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/strategy"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

func main() {
	configFile := flag.String("config", "", "Путь к файлу конфигурации")
	flag.Parse()

	var cfg *config.Config
	var err error
	if *configFile != "" {
		cfg, err = config.Load(*configFile)
		if err != nil {
			fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = config.LoadFromEnv()
	}

	logLevel := logger.LevelInfo
	switch cfg.Logging.Level {
	case "debug":
		logLevel = logger.LevelDebug
	case "warn":
		logLevel = logger.LevelWarn
	case "error":
		logLevel = logger.LevelError
	}
	var log *logger.Logger
	if cfg.Logging.File != "" {
		log, err = logger.NewWithFile(logLevel, cfg.Logging.File)
		if err != nil {
			fmt.Printf("Ошибка создания логгера: %v\n", err)
			os.Exit(1)
		}
	} else {
		log = logger.New(logLevel)
	}

	log.Info("🕸️ Grid Bot starting...")
	if cfg.Strategy.Grid == nil {
		log.Error("Grid config not provided")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exchange := createMockExchange(cfg, log)
	portfolioManager := portfolio.NewManager(exchange, log)
	factory := strategy.NewFactory(log)

	gridStrategy, err := factory.CreateGrid(*cfg.Strategy.Grid, exchange)
	if err != nil {
		log.Error("Ошибка создания Grid стратегии: %v", err)
		os.Exit(1)
	}
	if err := gridStrategy.ValidateConfig(); err != nil {
		log.Error("Ошибка валидации Grid конфигурации: %v", err)
		os.Exit(1)
	}

	go portfolioManager.StartAutoRefresh(ctx, 30*time.Second)
	go runTradingLoop(ctx, gridStrategy, exchange, log, cfg.Strategy.Grid.Symbol)

	if cfg.App.Port > 0 {
		go startHTTPServer(ctx, cfg, log, gridStrategy, portfolioManager)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Info("Получен сигнал завершения, останавливаем бота...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	_ = gridStrategy.Shutdown(shutdownCtx)
}

// общий цикл и HTTP сервер идентичны DCA боту
func runTradingLoop(ctx context.Context, strategy strategy.Strategy, exchange types.ExchangeClient, log *logger.Logger, symbol string) {
	ticker := time.NewTicker(30 * time.Second) // Уменьшаем интервал для более частого обновления
	defer ticker.Stop()
	log.Info("🔄 Trading loop started for %s", symbol)

	// Выполняем первую итерацию сразу
	if err := executeStrategyIteration(ctx, strategy, exchange, log, symbol); err != nil {
		log.Error("Initial strategy execution error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("🛑 Trading loop stopped")
			return
		case <-ticker.C:
			if err := executeStrategyIteration(ctx, strategy, exchange, log, symbol); err != nil {
				log.Error("Strategy execution error: %v", err)
			}
		}
	}
}

func executeStrategyIteration(ctx context.Context, strategy strategy.Strategy, exchange types.ExchangeClient, log *logger.Logger, symbol string) error {
	marketData, err := getMarketData(ctx, exchange, symbol)
	if err != nil {
		return fmt.Errorf("failed to fetch market data: %w", err)
	}

	log.Debug("📊 Market data: %s @ %.2f (volume: %.2f)",
		marketData.Symbol, marketData.Price, marketData.Volume)

	if err := strategy.Execute(ctx, marketData); err != nil {
		return fmt.Errorf("strategy execution failed: %w", err)
	}

	// Логируем статус стратегии каждые 5 минут
	if time.Now().Second() < 30 {
		status := strategy.GetStatus()
		log.Info("📈 Strategy status: %d active orders, %d total trades",
			status["active_orders"], status["total_trades"])
	}

	return nil
}

func getMarketData(ctx context.Context, exchange types.ExchangeClient, symbol string) (types.MarketData, error) {
	ticker, err := exchange.GetTicker(ctx, symbol)
	if err != nil {
		return types.MarketData{}, err
	}
	return types.MarketData{Symbol: symbol, Price: ticker.Price, Volume: ticker.Volume, Timestamp: ticker.Timestamp, Ticker: ticker}, nil
}

func startHTTPServer(ctx context.Context, cfg *config.Config, log *logger.Logger, strategy strategy.Strategy, portfolio *portfolio.Manager) {
	mux := http.NewServeMux()
	writeJSON := func(w http.ResponseWriter, status int, v interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /portfolio", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, portfolio.GetPortfolio()) })
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"strategy": strategy.GetMetrics(), "portfolio": portfolio.GetMetrics()})
	})
	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.App.Port), Handler: mux}
	go func() {
		log.Info("HTTP server listening on %d", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error: %v", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("HTTP server stopped")
}

// mock exchange similar to dca-bot
type MockExchangeClient struct {
	config *config.Config
	logger *logger.Logger
}

func createMockExchange(cfg *config.Config, log *logger.Logger) types.ExchangeClient {
	return &MockExchangeClient{config: cfg, logger: log}
}
func (m *MockExchangeClient) PlaceOrder(ctx context.Context, order types.Order) error {
	m.logger.Info("Mock: Placed order %s %.8f @ %.2f", order.Symbol, order.Quantity, order.Price)
	order.Status = types.OrderStatusFilled
	order.FilledAmount = order.Quantity
	order.FilledPrice = order.Price
	return nil
}
func (m *MockExchangeClient) CancelOrder(ctx context.Context, orderID string) error {
	m.logger.Info("Mock: Canceled order %s", orderID)
	return nil
}
func (m *MockExchangeClient) GetOrder(ctx context.Context, orderID string) (*types.Order, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockExchangeClient) GetActiveOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	return nil, nil
}
func (m *MockExchangeClient) GetFilledOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	return nil, nil
}
func (m *MockExchangeClient) GetTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	price := 45000.0 + float64(time.Now().Unix()%1000)
	return &types.Ticker{Symbol: symbol, Price: price, Bid: price - 0.1, Ask: price + 0.1, Volume: 1000.0, Timestamp: time.Now()}, nil
}
func (m *MockExchangeClient) GetOrderBook(ctx context.Context, symbol string, limit int) (*types.OrderBook, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockExchangeClient) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]types.Candle, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockExchangeClient) GetBalance(ctx context.Context) (*types.Balance, error) {
	return &types.Balance{Asset: "USDT", Free: 10000.0, Locked: 0.0, Total: 10000.0, Timestamp: time.Now()}, nil
}
func (m *MockExchangeClient) GetTradingFees(ctx context.Context, symbol string) (*types.TradingFees, error) {
	return &types.TradingFees{Symbol: symbol, MakerFee: 0.001, TakerFee: 0.001, Timestamp: time.Now()}, nil
}
func (m *MockExchangeClient) Ping(ctx context.Context) error { return nil }
func (m *MockExchangeClient) Close() error                   { return nil }
