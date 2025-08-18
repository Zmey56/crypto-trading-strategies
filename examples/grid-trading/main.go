package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/portfolio"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/strategy"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
)

func main() {
	fmt.Println("🕸️ Grid Trading Strategy Example")
	fmt.Println("==================================")

	// Создаем конфигурацию Grid стратегии
	gridConfig := types.GridConfig{
		Symbol:             "BTCUSDT",
		UpperPrice:         50000.0,
		LowerPrice:         40000.0,
		GridLevels:         10,
		InvestmentPerLevel: 100.0,
		Enabled:            true,
	}

	// Создаем логгер
	log := logger.New(logger.LevelInfo)

	// Создаем mock exchange
	exchange := createMockExchange(log)

	// Создаем portfolio manager
	portfolioManager := portfolio.NewManager(exchange, log)

	// Создаем factory и strategy
	factory := strategy.NewFactory(log)
	gridStrategy, err := factory.CreateGrid(gridConfig, exchange)
	if err != nil {
		log.Error("Ошибка создания Grid стратегии: %v", err)
		return
	}

	// Валидируем конфигурацию
	if err := gridStrategy.ValidateConfig(); err != nil {
		log.Error("Ошибка валидации Grid конфигурации: %v", err)
		return
	}

	fmt.Printf("✅ Grid стратегия создана для %s\n", gridConfig.Symbol)
	fmt.Printf("📊 Диапазон цен: $%.2f - $%.2f\n", gridConfig.LowerPrice, gridConfig.UpperPrice)
	fmt.Printf("🔢 Количество уровней: %d\n", gridConfig.GridLevels)
	fmt.Printf("💰 Инвестиции на уровень: $%.2f\n", gridConfig.InvestmentPerLevel)

	// Получаем статус стратегии
	status := gridStrategy.GetStatus()
	fmt.Printf("📈 Статус: %d активных ордеров, %d всего сделок\n",
		status["active_orders"], status["total_trades"])

	// Симулируем торговлю
	ctx := context.Background()
	simulateTrading(ctx, gridStrategy, exchange, log, gridConfig.Symbol)

	// Финальная статистика
	fmt.Println("\n📊 Финальная статистика:")
	metrics := gridStrategy.GetMetrics()
	fmt.Printf("Всего сделок: %d\n", metrics.TotalTrades)
	fmt.Printf("Общий объем: $%.2f\n", metrics.TotalVolume)
	fmt.Printf("Общая прибыль: $%.2f\n", metrics.TotalProfit)
	fmt.Printf("Винрейт: %.2f%%\n", metrics.WinRate*100)

	portfolio := portfolioManager.GetPortfolio()
	fmt.Printf("Общая стоимость портфолио: $%.2f\n", portfolio.TotalValue)
	fmt.Printf("Чистая прибыль: $%.2f\n", portfolio.NetProfit)
}

func simulateTrading(ctx context.Context, strategy strategy.Strategy, exchange types.ExchangeClient, log *logger.Logger, symbol string) {
	fmt.Println("\n🔄 Симуляция торговли...")

	// Симулируем изменения цены
	prices := []float64{45000, 42000, 48000, 41000, 49000, 43000, 47000, 44000, 46000, 45000}

	for i, price := range prices {
		fmt.Printf("\n📊 Итерация %d: Цена BTC = $%.2f\n", i+1, price)

		// Создаем данные рынка
		marketData := types.MarketData{
			Symbol:    symbol,
			Price:     price,
			Volume:    1000.0,
			Timestamp: time.Now(),
		}

		// Выполняем стратегию
		if err := strategy.Execute(ctx, marketData); err != nil {
			log.Error("Ошибка выполнения стратегии: %v", err)
			continue
		}

		// Получаем сигнал
		signal := strategy.GetSignal(marketData)
		if signal.Type != types.SignalTypeHold {
			fmt.Printf("📡 Сигнал: %s @ $%.2f (количество: %.8f)\n",
				signal.Type, signal.Price, signal.Quantity)
		}

		// Показываем статус
		status := strategy.GetStatus()
		fmt.Printf("📈 Активных ордеров: %d, Всего сделок: %d\n",
			status["active_orders"], status["total_trades"])

		// Небольшая пауза между итерациями
		time.Sleep(500 * time.Millisecond)
	}
}

type MockExchangeClient struct {
	logger *logger.Logger
}

func createMockExchange(log *logger.Logger) types.ExchangeClient {
	return &MockExchangeClient{logger: log}
}

func (m *MockExchangeClient) PlaceOrder(ctx context.Context, order types.Order) error {
	m.logger.Info("Mock: Размещен ордер %s %.8f @ %.2f", order.Symbol, order.Quantity, order.Price)
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
	return &types.Ticker{
		Symbol:    symbol,
		Price:     45000.0,
		Bid:       44999.9,
		Ask:       45000.1,
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
