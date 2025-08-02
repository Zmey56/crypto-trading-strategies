package app

import "fmt"

type Container struct {
	config           *config.Config
	logger           *logger.Logger
	exchangeClients  map[string]exchange.Client
	strategyFactory  strategy.Factory
	portfolioManager *portfolio.Manager
	riskManager      *risk.Manager
	metricsCollector *metrics.Collector
}

func NewContainer(cfg *config.Config) (*Container, error) {
	logger := logger.NewStructuredLogger(cfg.Logger)

	exchangeClients, err := initializeExchanges(cfg.Exchanges)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize exchanges: %w", err)
	}

	riskManager := risk.NewManager(cfg.Risk)
	portfolioManager := portfolio.NewManager(exchangeClients, riskManager)

	return &Container{
		config:           cfg,
		logger:           logger,
		exchangeClients:  exchangeClients,
		strategyFactory:  strategy.NewFactory(),
		portfolioManager: portfolioManager,
		riskManager:      riskManager,
		metricsCollector: metrics.NewCollector(),
	}, nil
}
