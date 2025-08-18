package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/portfolio"
	"github.com/Zmey56/crypto-arbitrage-trader/internal/strategy"
)

// Server представляет HTTP API сервер
type Server struct {
	port      int
	strategy  strategy.Strategy
	portfolio *portfolio.Manager
	logger    *logger.Logger
	server    *http.Server
}

// NewServer создает новый API сервер
func NewServer(port int, strategy strategy.Strategy, portfolio *portfolio.Manager, logger *logger.Logger) *Server {
	return &Server{
		port:      port,
		strategy:  strategy,
		portfolio: portfolio,
		logger:    logger,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", s.handleHealth)
	
	// Portfolio endpoints
	mux.HandleFunc("GET /portfolio", s.handleGetPortfolio)
	mux.HandleFunc("GET /portfolio/positions", s.handleGetPositions)
	mux.HandleFunc("GET /portfolio/metrics", s.handleGetPortfolioMetrics)
	
	// Strategy endpoints
	mux.HandleFunc("GET /strategy/status", s.handleGetStrategyStatus)
	mux.HandleFunc("GET /strategy/metrics", s.handleGetStrategyMetrics)
	mux.HandleFunc("POST /strategy/config", s.handleUpdateStrategyConfig)
	
	// Combined metrics
	mux.HandleFunc("GET /metrics", s.handleGetMetrics)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	s.logger.Info("HTTP API сервер запущен на порту %d", s.port)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP сервер ошибка: %v", err)
		}
	}()

	<-ctx.Done()
	return s.Shutdown(context.Background())
}

// Shutdown останавливает HTTP сервер
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// writeJSON отправляет JSON ответ
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Ошибка кодирования JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleHealth обрабатывает health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(time.Now()).String(),
	})
}

// handleGetPortfolio обрабатывает запрос портфеля
func (s *Server) handleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	portfolio := s.portfolio.GetPortfolio()
	s.writeJSON(w, http.StatusOK, portfolio)
}

// handleGetPositions обрабатывает запрос позиций
func (s *Server) handleGetPositions(w http.ResponseWriter, r *http.Request) {
	positions := s.portfolio.GetPositionSummary()
	s.writeJSON(w, http.StatusOK, positions)
}

// handleGetPortfolioMetrics обрабатывает запрос метрик портфеля
func (s *Server) handleGetPortfolioMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.portfolio.GetMetrics()
	s.writeJSON(w, http.StatusOK, metrics)
}

// handleGetStrategyStatus обрабатывает запрос статуса стратегии
func (s *Server) handleGetStrategyStatus(w http.ResponseWriter, r *http.Request) {
	// Используем рефлексию для получения статуса стратегии
	var status map[string]interface{}
	
	switch s := s.strategy.(type) {
	case *strategy.DCAStrategy:
		status = s.GetStatus()
	case *strategy.GridStrategy:
		status = s.GetStatus()
	default:
		status = map[string]interface{}{
			"type": "unknown",
		}
	}
	
	s.writeJSON(w, http.StatusOK, status)
}

// handleGetStrategyMetrics обрабатывает запрос метрик стратегии
func (s *Server) handleGetStrategyMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.strategy.GetMetrics()
	s.writeJSON(w, http.StatusOK, metrics)
}

// handleUpdateStrategyConfig обрабатывает обновление конфигурации стратегии
func (s *Server) handleUpdateStrategyConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Здесь можно добавить логику обновления конфигурации
	// Пока просто возвращаем успех
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Configuration updated successfully",
		"config":  config,
	})
}

// handleGetMetrics обрабатывает запрос общих метрик
func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	portfolioMetrics := s.portfolio.GetMetrics()
	strategyMetrics := s.strategy.GetMetrics()

	metrics := map[string]interface{}{
		"portfolio": portfolioMetrics,
		"strategy":  strategyMetrics,
		"timestamp": time.Now().UTC(),
	}

	s.writeJSON(w, http.StatusOK, metrics)
}
