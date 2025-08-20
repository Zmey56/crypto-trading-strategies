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
	// Parse command line flags
	configFile := flag.String("config", "", "Path to config file")
	flag.Parse()

	// Load configuration
	var cfg *config.Config
	var err error

	if *configFile != "" {
		cfg, err = config.Load(*configFile)
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = config.LoadFromEnv()
	}

	// Create logger
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
			fmt.Printf("Failed to create logger: %v\n", err)
			os.Exit(1)
		}
	} else {
		log = logger.New(logLevel)
	}

	log.Info("🔲 Grid Bot starting...")
	log.Info("Version: %s", cfg.App.Version)
	log.Info("Exchange: %s", cfg.Exchange.Name)
	log.Info("Symbol: %s", cfg.Strategy.Grid.Symbol)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create mock exchange client (use real client in production)
	exchange := createMockExchange(cfg, log)

	// Create portfolio manager
	portfolioManager := portfolio.NewManager(exchange, log)

	// Create strategy factory
	strategyFactory := strategy.NewFactory(log)

	// Create Grid strategy
	gridStrategy, err := strategyFactory.CreateGrid(*cfg.Strategy.Grid, exchange)
	if err != nil {
		log.Error("Failed to create Grid strategy: %v", err)
		os.Exit(1)
	}

	// Validate strategy config
	if err := gridStrategy.ValidateConfig(); err != nil {
		log.Error("Strategy config validation error: %v", err)
		os.Exit(1)
	}

	// Start portfolio auto-refresh
	go portfolioManager.StartAutoRefresh(ctx, 30*time.Second)

	// Start trading loop
	go runTradingLoop(ctx, gridStrategy, exchange, log, cfg.Strategy.Grid.Symbol)

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server for monitoring (optional)
	if cfg.App.Port > 0 {
		go startHTTPServer(ctx, cfg, log, gridStrategy, portfolioManager)
	}

	log.Info("Grid Bot started and running")

	// Wait for termination signal
	<-sigChan
	log.Info("Termination signal received, stopping bot...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := gridStrategy.Shutdown(shutdownCtx); err != nil {
		log.Error("Error stopping strategy: %v", err)
	}

	log.Info("Grid Bot stopped")
}

// runTradingLoop starts the main trading loop
func runTradingLoop(ctx context.Context, strategy strategy.Strategy, exchange types.ExchangeClient, log *logger.Logger, symbol string) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds for grid
	defer ticker.Stop()

	log.Info("Trading loop started for %s", symbol)

	for {
		select {
		case <-ctx.Done():
			log.Info("Trading loop stopped")
			return
		case <-ticker.C:
			// Fetch market data
			marketData, err := getMarketData(ctx, exchange, symbol)
			if err != nil {
				log.Error("Failed to fetch market data: %v", err)
				continue
			}

			// Execute strategy
			if err := strategy.Execute(ctx, marketData); err != nil {
				log.Error("Strategy execution error: %v", err)
			}

			// Log metrics
			metrics := strategy.GetMetrics()
			log.Debug("Метрики стратегии: %+v", metrics)
		}
	}
}

// getMarketData fetches market data
func getMarketData(ctx context.Context, exchange types.ExchangeClient, symbol string) (types.MarketData, error) {
	ticker, err := exchange.GetTicker(ctx, symbol)
	if err != nil {
		return types.MarketData{}, err
	}

	return types.MarketData{
		Symbol:    symbol,
		Price:     ticker.Price,
		Volume:    ticker.Volume,
		Timestamp: ticker.Timestamp,
		Ticker:    ticker,
	}, nil
}

	// createMockExchange creates mock exchange client for demonstration
func createMockExchange(cfg *config.Config, log *logger.Logger) types.ExchangeClient {
	return &MockExchangeClient{
		config: cfg,
		logger: log,
	}
}

// MockExchangeClient represents mock exchange client
type MockExchangeClient struct {
	config *config.Config
	logger *logger.Logger
}

func (m *MockExchangeClient) PlaceOrder(ctx context.Context, order types.Order) error {
	m.logger.Info("Mock: Размещен ордер %s %.8f @ %.2f", order.Symbol, order.Quantity, order.Price)

			// Simulate successful execution
	order.Status = types.OrderStatusFilled
	order.FilledAmount = order.Quantity
	order.FilledPrice = order.Price

	return nil
}

func (m *MockExchangeClient) CancelOrder(ctx context.Context, orderID string) error {
	m.logger.Info("Mock: Отменен ордер %s", orderID)
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
	// Simulate a current BTC price with more volatility for grid testing
	basePrice := 45000.0
	volatility := float64(time.Now().Unix()%2000 - 1000) // -1000 to +1000
	price := basePrice + volatility

	return &types.Ticker{
		Symbol:    symbol,
		Price:     price,
		Bid:       price - 0.1,
		Ask:       price + 0.1,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockExchangeClient) GetOrderBook(ctx context.Context, symbol string, limit int) (*types.OrderBook, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockExchangeClient) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]types.Candle, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockExchangeClient) GetBalance(ctx context.Context) (*types.Balance, error) {
	return &types.Balance{
		Asset:     "USDT",
		Free:      10000.0,
		Locked:    0.0,
		Total:     10000.0,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockExchangeClient) GetTradingFees(ctx context.Context, symbol string) (*types.TradingFees, error) {
	return &types.TradingFees{
		Symbol:    symbol,
		MakerFee:  0.001,
		TakerFee:  0.001,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockExchangeClient) Ping(ctx context.Context) error {
	return nil
}

func (m *MockExchangeClient) Close() error {
	return nil
}

// startHTTPServer runs the HTTP server for monitoring
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

	mux.HandleFunc("GET /portfolio", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, portfolio.GetPortfolio())
	})

	mux.HandleFunc("GET /strategy/status", func(w http.ResponseWriter, r *http.Request) {
		// Try to get extended status if strategy supports it
		type statusProvider interface{ GetStatus() map[string]interface{} }
		if sp, ok := strategy.(statusProvider); ok {
			writeJSON(w, http.StatusOK, sp.GetStatus())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "no detailed status"})
	})

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"strategy":  strategy.GetMetrics(),
			"portfolio": portfolio.GetMetrics(),
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: loggingMiddleware(log, mux),
	}

	go func() {
		log.Info("HTTP сервер запущен на порту %d", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("HTTP сервер остановлен")
}

func loggingMiddleware(log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
