package main

import (
	"context"
	"crypto-trading-strategies/internal/config"
	"crypto-trading-strategies/internal/logger"
	"crypto-trading-strategies/internal/portfolio"
	"crypto-trading-strategies/internal/strategy"
	"crypto-trading-strategies/pkg/types"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Парсим флаги командной строки
	configFile := flag.String("config", "", "Путь к файлу конфигурации")
	flag.Parse()

	// Загружаем конфигурацию
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

	// Создаем логгер
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

	log.Info("🤖 DCA Bot запускается...")
	log.Info("Версия: %s", cfg.App.Version)
	log.Info("Биржа: %s", cfg.Exchange.Name)
	log.Info("Символ: %s", cfg.Strategy.DCA.Symbol)

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем mock exchange client (в реальном проекте здесь будет настоящий клиент)
	exchange := createMockExchange(cfg, log)

	// Создаем менеджер портфеля
	portfolioManager := portfolio.NewManager(exchange, log)

	// Создаем фабрику стратегий
	strategyFactory := strategy.NewFactory(log)

	// Создаем DCA стратегию
	dcaStrategy, err := strategyFactory.CreateDCA(*cfg.Strategy.DCA, exchange)
	if err != nil {
		log.Error("Ошибка создания DCA стратегии: %v", err)
		os.Exit(1)
	}

	// Валидируем конфигурацию стратегии
	if err := dcaStrategy.ValidateConfig(); err != nil {
		log.Error("Ошибка валидации конфигурации: %v", err)
		os.Exit(1)
	}

	// Запускаем автообновление портфеля
	go portfolioManager.StartAutoRefresh(ctx, 30*time.Second)

	// Запускаем основной цикл торговли
	go runTradingLoop(ctx, dcaStrategy, exchange, log, cfg.Strategy.DCA.Symbol)

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем HTTP сервер для мониторинга (опционально)
	if cfg.App.Port > 0 {
		go startHTTPServer(ctx, cfg, log, dcaStrategy, portfolioManager)
	}

	log.Info("DCA Bot успешно запущен и работает")

	// Ждем сигнала завершения
	<-sigChan
	log.Info("Получен сигнал завершения, останавливаем бота...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := dcaStrategy.Shutdown(shutdownCtx); err != nil {
		log.Error("Ошибка при остановке стратегии: %v", err)
	}

	log.Info("DCA Bot остановлен")
}

// runTradingLoop запускает основной цикл торговли
func runTradingLoop(ctx context.Context, strategy strategy.Strategy, exchange types.ExchangeClient, log *logger.Logger, symbol string) {
	ticker := time.NewTicker(1 * time.Minute) // Проверяем каждую минуту
	defer ticker.Stop()

	log.Info("Запущен торговый цикл для %s", symbol)

	for {
		select {
		case <-ctx.Done():
			log.Info("Торговый цикл остановлен")
			return
		case <-ticker.C:
			// Получаем рыночные данные
			marketData, err := getMarketData(ctx, exchange, symbol)
			if err != nil {
				log.Error("Ошибка получения рыночных данных: %v", err)
				continue
			}

			// Выполняем стратегию
			if err := strategy.Execute(ctx, marketData); err != nil {
				log.Error("Ошибка выполнения стратегии: %v", err)
			}

			// Логируем метрики
			metrics := strategy.GetMetrics()
			log.Debug("Метрики стратегии: %+v", metrics)
		}
	}
}

// getMarketData получает рыночные данные
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

// createMockExchange создает mock exchange client для демонстрации
func createMockExchange(cfg *config.Config, log *logger.Logger) types.ExchangeClient {
	return &MockExchangeClient{
		config: cfg,
		logger: log,
	}
}

// MockExchangeClient представляет mock клиент биржи
type MockExchangeClient struct {
	config *config.Config
	logger *logger.Logger
}

func (m *MockExchangeClient) PlaceOrder(ctx context.Context, order types.Order) error {
	m.logger.Info("Mock: Размещен ордер %s %.8f @ %.2f", order.Symbol, order.Quantity, order.Price)

	// Имитируем успешное исполнение
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
	// Имитируем текущую цену BTC
	price := 45000.0 + float64(time.Now().Unix()%1000) // Простая имитация колебаний цены

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

// startHTTPServer запускает HTTP сервер для мониторинга
func startHTTPServer(ctx context.Context, cfg *config.Config, log *logger.Logger, strategy strategy.Strategy, portfolio *portfolio.Manager) {
	// TODO: Реализовать HTTP сервер для мониторинга
	log.Info("HTTP сервер запущен на порту %d", cfg.App.Port)

	<-ctx.Done()
	log.Info("HTTP сервер остановлен")
}
