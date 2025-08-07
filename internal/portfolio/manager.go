package portfolio

import (
	"context"
	"crypto-trading-strategies/internal/logger"
	"crypto-trading-strategies/pkg/types"
	"fmt"
	"sync"
	"time"
)

// Manager представляет менеджер портфеля
type Manager struct {
	exchange types.ExchangeClient
	logger   *logger.Logger
	mu       sync.RWMutex

	// Портфель
	portfolio *types.Portfolio
	positions map[string]*types.Position

	// Метрики
	totalInvested float64
	totalValue    float64
	lastUpdate    time.Time
}

// NewManager создает новый менеджер портфеля
func NewManager(exchange types.ExchangeClient, logger *logger.Logger) *Manager {
	return &Manager{
		exchange:  exchange,
		logger:    logger,
		portfolio: &types.Portfolio{},
		positions: make(map[string]*types.Position),
	}
}

// GetPortfolio возвращает текущий портфель
func (m *Manager) GetPortfolio() *types.Portfolio {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.portfolio
}

// GetPosition возвращает позицию по символу
func (m *Manager) GetPosition(symbol string) (*types.Position, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	position, exists := m.positions[symbol]
	return position, exists
}

// GetAllPositions возвращает все позиции
func (m *Manager) GetAllPositions() map[string]*types.Position {
	m.mu.RLock()
	defer m.mu.RUnlock()

	positions := make(map[string]*types.Position)
	for symbol, position := range m.positions {
		positions[symbol] = position
	}

	return positions
}

// UpdatePosition обновляет позицию
func (m *Manager) UpdatePosition(order types.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	symbol := order.Symbol
	position, exists := m.positions[symbol]

	if !exists {
		position = &types.Position{
			Symbol:   symbol,
			Quantity: 0,
			AvgPrice: 0,
			Timestamp: time.Now(),
		}
		m.positions[symbol] = position
	}

	// Обновляем позицию в зависимости от типа ордера
	switch order.Side {
	case types.OrderSideBuy:
		if order.Status == types.OrderStatusFilled {
			// Вычисляем новую среднюю цену
			totalCost := position.Quantity*position.AvgPrice + order.FilledAmount*order.FilledPrice
			totalQuantity := position.Quantity + order.FilledAmount
			
			if totalQuantity > 0 {
				position.AvgPrice = totalCost / totalQuantity
			}
			
			position.Quantity += order.FilledAmount
			position.Timestamp = time.Now()
			
			m.logger.Info("Позиция обновлена (покупка): %s %.8f @ %.2f (средняя: %.2f)", 
				symbol, order.FilledAmount, order.FilledPrice, position.AvgPrice)
		}

	case types.OrderSideSell:
		if order.Status == types.OrderStatusFilled {
			// Вычисляем реализованную прибыль/убыток
			if position.Quantity > 0 {
				realizedPnL := (order.FilledPrice - position.AvgPrice) * order.FilledAmount
				position.RealizedPnL += realizedPnL
				
				m.logger.Info("Реализован PnL: %s %.2f (%.2f - %.2f) * %.8f", 
					symbol, realizedPnL, order.FilledPrice, position.AvgPrice, order.FilledAmount)
			}
			
			position.Quantity -= order.FilledAmount
			position.Timestamp = time.Now()
			
			// Если позиция закрыта, удаляем её
			if position.Quantity <= 0 {
				delete(m.positions, symbol)
				m.logger.Info("Позиция закрыта: %s", symbol)
			}
		}
	}

	return nil
}

// RefreshPortfolio обновляет портфель с биржи
func (m *Manager) RefreshPortfolio(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Получаем баланс с биржи
	_, err := m.exchange.GetBalance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Обновляем позиции с текущими ценами
	for symbol, position := range m.positions {
		ticker, err := m.exchange.GetTicker(ctx, symbol)
		if err != nil {
			m.logger.Warn("Не удалось получить тикер для %s: %v", symbol, err)
			continue
		}

		position.CurrentPrice = ticker.Price
		position.UnrealizedPnL = (ticker.Price - position.AvgPrice) * position.Quantity
		position.Timestamp = time.Now()
	}

	// Обновляем общие метрики портфеля
	m.updatePortfolioMetrics()

	m.lastUpdate = time.Now()
	return nil
}

// updatePortfolioMetrics обновляет метрики портфеля
func (m *Manager) updatePortfolioMetrics() {
	var totalValue, totalProfit, totalLoss float64

	for _, position := range m.positions {
		positionValue := position.Quantity * position.CurrentPrice
		totalValue += positionValue

		if position.UnrealizedPnL > 0 {
			totalProfit += position.UnrealizedPnL
		} else {
			totalLoss += -position.UnrealizedPnL
		}
	}

	m.portfolio.TotalValue = totalValue
	m.portfolio.TotalProfit = totalProfit
	m.portfolio.TotalLoss = totalLoss
	m.portfolio.NetProfit = totalProfit - totalLoss
	m.portfolio.LastUpdate = time.Now()

	// Обновляем список позиций
	var positions []types.Position
	for _, position := range m.positions {
		positions = append(positions, *position)
	}
	m.portfolio.Positions = positions
}

// GetMetrics возвращает метрики портфеля
func (m *Manager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_value":    m.portfolio.TotalValue,
		"total_profit":   m.portfolio.TotalProfit,
		"total_loss":     m.portfolio.TotalLoss,
		"net_profit":     m.portfolio.NetProfit,
		"positions_count": len(m.positions),
		"last_update":    m.lastUpdate,
	}
}

// GetPositionSummary возвращает сводку по позициям
func (m *Manager) GetPositionSummary() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var summary []map[string]interface{}
	for symbol, position := range m.positions {
		summary = append(summary, map[string]interface{}{
			"symbol":         symbol,
			"quantity":       position.Quantity,
			"avg_price":      position.AvgPrice,
			"current_price":  position.CurrentPrice,
			"unrealized_pnl": position.UnrealizedPnL,
			"realized_pnl":   position.RealizedPnL,
			"timestamp":      position.Timestamp,
		})
	}

	return summary
}

// StartAutoRefresh запускает автоматическое обновление портфеля
func (m *Manager) StartAutoRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Автообновление портфеля остановлено")
			return
		case <-ticker.C:
			if err := m.RefreshPortfolio(ctx); err != nil {
				m.logger.Error("Ошибка при обновлении портфеля: %v", err)
			}
		}
	}
} 