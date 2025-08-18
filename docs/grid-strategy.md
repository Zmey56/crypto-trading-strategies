# Grid Trading Strategy

## Обзор

Grid Trading Strategy (Стратегия сеточной торговли) - это автоматизированная торговая стратегия, которая размещает ордера на покупку и продажу на заранее определенных ценовых уровнях (сетке). Стратегия работает в диапазоне цен и автоматически покупает при падении цены и продает при росте.

## Принцип работы

1. **Определение диапазона цен**: Устанавливается верхняя и нижняя граница цен
2. **Создание сетки**: В диапазоне создается равномерная сетка из N уровней
3. **Размещение ордеров**: На каждом уровне размещаются ордера на покупку и продажу
4. **Автоматическое исполнение**: При достижении цены уровня ордер исполняется
5. **Повторное размещение**: После исполнения ордера на том же уровне размещается противоположный ордер

## Конфигурация

```json
{
  "grid": {
    "symbol": "BTCUSDT",
    "upper_price": 50000.0,
    "lower_price": 40000.0,
    "grid_levels": 20,
    "investment_per_level": 100.0,
    "enabled": true
  }
}
```

### Параметры

- `symbol` - Торговая пара (например, BTCUSDT)
- `upper_price` - Верхняя граница цен сетки
- `lower_price` - Нижняя граница цен сетки
- `grid_levels` - Количество уровней в сетке
- `investment_per_level` - Сумма инвестиций на каждый уровень
- `enabled` - Включение/выключение стратегии

## Алгоритм

### 1. Инициализация сетки

```go
func (g *GridStrategy) calculateGridLevels() {
    priceRange := g.config.UpperPrice - g.config.LowerPrice
    step := priceRange / float64(g.config.GridLevels-1)
    
    g.gridLevels = make([]float64, g.config.GridLevels)
    for i := 0; i < g.config.GridLevels; i++ {
        g.gridLevels[i] = g.config.LowerPrice + float64(i)*step
    }
}
```

### 2. Поиск ближайших уровней

```go
func (g *GridStrategy) findNearestLevels(price float64) (buyLevel, sellLevel float64) {
    // Если цена вне диапазона, не размещаем ордера
    if price < g.config.LowerPrice || price > g.config.UpperPrice {
        return 0, 0
    }
    
    for i, level := range g.gridLevels {
        if price <= level {
            buyLevel = level
            if i < len(g.gridLevels)-1 {
                sellLevel = g.gridLevels[i+1]
            }
            break
        }
    }
    return
}
```

### 3. Размещение ордеров

```go
func (g *GridStrategy) shouldPlaceBuyOrder(level, currentPrice float64) bool {
    // Проверяем, нет ли уже активного ордера на этом уровне
    for _, order := range g.activeOrders {
        if order.Price == level && order.Side == types.OrderSideBuy {
            return false
        }
    }
    
    // Размещаем ордер на покупку, если цена на уровне или ниже
    return currentPrice <= level
}
```

## Преимущества

1. **Автоматизация**: Полностью автоматизированная торговля
2. **Диверсификация**: Распределение рисков по множеству уровней
3. **Прибыль от волатильности**: Заработок на колебаниях цены
4. **Контроль рисков**: Ограниченный диапазон цен
5. **Предсказуемость**: Известные уровни входа и выхода

## Риски

1. **Трендовые рынки**: Может быть неэффективна в сильных трендах
2. **Проскальзывание**: Разница между ожидаемой и фактической ценой исполнения
3. **Комиссии**: Множественные сделки увеличивают комиссионные расходы
4. **Ликвидность**: Необходимость достаточной ликвидности на всех уровнях

## Пример использования

```go
// Создание конфигурации
gridConfig := types.GridConfig{
    Symbol:             "BTCUSDT",
    UpperPrice:         50000.0,
    LowerPrice:         40000.0,
    GridLevels:         10,
    InvestmentPerLevel: 100.0,
    Enabled:            true,
}

// Создание стратегии
factory := strategy.NewFactory(logger)
gridStrategy, err := factory.CreateGrid(gridConfig, exchange)

// Выполнение стратегии
marketData := types.MarketData{
    Symbol:    "BTCUSDT",
    Price:     45000.0,
    Volume:    1000.0,
    Timestamp: time.Now(),
}

err = gridStrategy.Execute(ctx, marketData)
```

## Мониторинг

### Метрики стратегии

- `TotalTrades` - Общее количество сделок
- `TotalVolume` - Общий объем торгов
- `TotalProfit` - Общая прибыль
- `WinRate` - Процент прибыльных сделок
- `ActiveOrders` - Количество активных ордеров

### API эндпоинты

- `GET /strategy/status` - Статус стратегии
- `GET /strategy/metrics` - Метрики стратегии
- `GET /portfolio` - Состояние портфолио

## Оптимизация

### Параметры для оптимизации

1. **Количество уровней**: Больше уровней = больше сделок, но выше комиссии
2. **Размер инвестиций**: Больше инвестиций = больше прибыль, но выше риски
3. **Диапазон цен**: Узкий диапазон = больше сделок, широкий = меньше сделок
4. **Интервал обновления**: Частые обновления = быстрая реакция, но больше нагрузка

### Рекомендации

1. Начинайте с небольшого количества уровней (5-10)
2. Используйте исторические данные для тестирования
3. Мониторьте комиссионные расходы
4. Адаптируйте параметры под волатильность актива

## Тестирование

Для тестирования стратегии используется backtesting модуль:

```bash
# Запуск backtesting
go run cmd/backtester/main.go -config configs/backtest-config.json

# Тестирование на исторических данных
go run examples/grid-trading/main.go
```

## Заключение

Grid Trading Strategy - это эффективный инструмент для автоматизированной торговли в боковых рынках. При правильной настройке параметров и мониторинге может приносить стабильную прибыль от волатильности рынка.
