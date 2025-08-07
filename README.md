# Crypto Trading Strategies

Проект для автоматизированной торговли криптовалютами с использованием различных стратегий.

## 🚀 Возможности

- **DCA (Dollar Cost Averaging)** - стратегия усреднения стоимости
- **Grid Trading** - сеточная торговля (в разработке)
- **Combo Strategies** - комбинированные стратегии (в разработке)
- Поддержка множественных бирж (Binance, Kraken)
- Мониторинг портфеля в реальном времени
- RESTful API для управления
- Подробное логирование и метрики

## 📁 Структура проекта

```
crypto-trading-strategies/
├── cmd/                    # Исполняемые файлы
│   ├── dca-bot/           # DCA бот
│   ├── grid-bot/          # Grid бот
│   └── backtester/        # Бэктестер
├── internal/              # Внутренние пакеты
│   ├── config/            # Конфигурация
│   ├── exchange/          # Клиенты бирж
│   ├── strategy/          # Торговые стратегии
│   ├── portfolio/         # Управление портфелем
│   └── logger/            # Логирование
├── pkg/                   # Публичные пакеты
│   ├── types/             # Общие типы данных
│   └── indicators/        # Технические индикаторы
├── configs/               # Конфигурационные файлы
├── examples/              # Примеры использования
└── docs/                  # Документация
```

## 🛠️ Установка и запуск

### Требования

- Go 1.21 или выше
- API ключи от биржи (Binance, Kraken)

### Установка

```bash
# Клонирование репозитория
git clone <repository-url>
cd crypto-trading-strategies

# Установка зависимостей
go mod tidy

# Сборка
go build ./cmd/dca-bot
```

### Конфигурация

1. Скопируйте пример конфигурации:
```bash
cp configs/dca-config.json configs/my-config.json
```

2. Отредактируйте конфигурацию:
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

### Запуск DCA бота

```bash
# С конфигурационным файлом
./dca-bot -config configs/my-config.json

# С переменными окружения
export DCA_SYMBOL=BTCUSDT
export DCA_INVESTMENT_AMOUNT=100
export DCA_INTERVAL=24h
export EXCHANGE_API_KEY=your-api-key
export EXCHANGE_SECRET_KEY=your-secret-key
./dca-bot
```

## 📊 DCA Стратегия

DCA (Dollar Cost Averaging) - стратегия усреднения стоимости, которая заключается в регулярной покупке актива на фиксированную сумму независимо от цены.

### Принципы работы

1. **Регулярные инвестиции**: Покупка на фиксированную сумму через заданные интервалы
2. **Автоматическое исполнение**: Бот автоматически размещает ордера
3. **Управление рисками**: Ограничение максимального количества инвестиций
4. **Мониторинг**: Отслеживание позиций и метрик в реальном времени

### Конфигурация DCA

```json
{
  "symbol": "BTCUSDT",           // Торговая пара
  "investment_amount": 100.0,    // Сумма инвестиции в USDT
  "interval": "24h",             // Интервал между покупками
  "max_investments": 100,        // Максимальное количество покупок
  "price_threshold": 0.0,        // Порог цены (0 = без ограничений)
  "stop_loss": 0.0,              // Stop Loss (0 = отключен)
  "take_profit": 0.0,            // Take Profit (0 = отключен)
  "enabled": true                // Включить/выключить стратегию
}
```

## 🔧 API

### Endpoints

- `GET /health` - Проверка состояния
- `GET /portfolio` - Информация о портфеле
- `GET /strategy/status` - Статус стратегии
- `POST /strategy/config` - Обновление конфигурации
- `GET /metrics` - Метрики стратегии

### Пример использования API

```bash
# Получение статуса портфеля
curl http://localhost:8080/portfolio

# Обновление конфигурации
curl -X POST http://localhost:8080/strategy/config \
  -H "Content-Type: application/json" \
  -d '{"investment_amount": 150.0}'
```

## 📈 Мониторинг

### Метрики стратегии

- **Total Trades**: Общее количество сделок
- **Win Rate**: Процент прибыльных сделок
- **Total Profit/Loss**: Общая прибыль/убыток
- **Average Win/Loss**: Средняя прибыль/убыток
- **Profit Factor**: Фактор прибыли
- **Max Drawdown**: Максимальная просадка

### Логирование

Бот ведет подробные логи всех операций:

```
[INFO] 🤖 DCA Bot запускается...
[INFO] Версия: 1.0.0
[INFO] Биржа: binance
[INFO] Символ: BTCUSDT
[INFO] DCA Bot успешно запущен и работает
[INFO] Mock: Размещен ордер BTCUSDT 0.00222222 @ 45000.00
[INFO] DCA покупка выполнена: BTCUSDT 0.00222222 @ 45000.00 (покупка #1)
```

## 🛡️ Безопасность

### Рекомендации

1. **API ключи**: Используйте API ключи только для торговли, без права вывода
2. **Sandbox**: Сначала тестируйте на песочнице
3. **Лимиты**: Устанавливайте разумные лимиты на инвестиции
4. **Мониторинг**: Регулярно проверяйте логи и метрики

### Переменные окружения

```bash
# Безопасное хранение API ключей
export EXCHANGE_API_KEY=your-api-key
export EXCHANGE_SECRET_KEY=your-secret-key
export EXCHANGE_SANDBOX=true
```

## 🧪 Тестирование

### Unit тесты

```bash
go test ./internal/strategy
go test ./internal/portfolio
```

### Интеграционные тесты

```bash
go test ./test/integration
```

### Бэктестинг

```bash
go run cmd/backtester/main.go -config configs/backtest-config.json
```

## 📝 Разработка

### Добавление новой стратегии

1. Создайте новый файл в `internal/strategy/`
2. Реализуйте интерфейс `Strategy`
3. Добавьте конфигурацию в `pkg/types/types.go`
4. Обновите фабрику стратегий
5. Добавьте тесты

### Добавление новой биржи

1. Создайте клиент в `internal/exchange/`
2. Реализуйте интерфейс `ExchangeClient`
3. Добавьте в `UnifiedClient`
4. Добавьте тесты

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch
3. Внесите изменения
4. Добавьте тесты
5. Создайте Pull Request

## 📄 Лицензия

MIT License

## ⚠️ Отказ от ответственности

Этот проект предназначен только для образовательных целей. Торговля криптовалютами связана с высокими рисками. Авторы не несут ответственности за возможные финансовые потери.

## 📞 Поддержка

- Issues: [GitHub Issues](https://github.com/your-repo/issues)
- Discussions: [GitHub Discussions](https://github.com/your-repo/discussions)
- Email: support@example.com