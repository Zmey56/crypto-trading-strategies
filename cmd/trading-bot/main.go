package main

import (
	"context"
	"github.com/username/crypto-trading-strategies/internal/strategy"
	"log"
	"os"
	"os/signal"
	"syscall"

	"crypto-trading-strategies/internal/app"
	"crypto-trading-strategies/internal/config"
	"crypto-trading-strategies/internal/logger"
)

type TradingApplication struct {
	config     *config.Config
	logger     *logger.Logger
	strategies map[string]strategy.Strategy
	exchanges  map[string]exchange.Client
	portfolio  *portfolio.Manager
	metrics    *metrics.Server
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	app := NewTradingApplication(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		app.logger.Info("Shutdown signal received")
		cancel()
	}()

	if err := app.Run(ctx); err != nil {
		app.logger.Error("Application failed:", err)
		os.Exit(1)
	}
}
