package app

import (
	"github.com/Zmey56/crypto-arbitrage-trader/internal/analytics"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/config"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/exchange"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/exchange/mock"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/portfolio"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/risk"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/strategy"
)

type Container struct {
	config           *config.Config
	logger           *logger.Logger
	exchangeClients  map[string]exchange.Client
	strategyFactory  *strategy.Factory
	portfolioManager *portfolio.Manager
	riskManager      *risk.Manager
	metricsCollector *analytics.MetricsCollector
}

func NewContainer(cfg *config.Config) (*Container, error) {
	log := logger.New(logger.LevelInfo)

	exchangeClients := make(map[string]exchange.Client)
	// Initialize default exchange client (mock for now)
	mockClient := &mock.MockClient{}
	exchangeClients["binance"] = mockClient

	riskManager := &risk.Manager{}
	portfolioManager := portfolio.NewManager(mockClient, log)

	metricsCollector := &analytics.MetricsCollector{}

	return &Container{
		config:           cfg,
		logger:           log,
		exchangeClients:  exchangeClients,
		strategyFactory:  strategy.NewFactory(log),
		portfolioManager: portfolioManager,
		riskManager:      riskManager,
		metricsCollector: metricsCollector,
	}, nil
}
