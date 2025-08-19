package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Zmey56/crypto-arbitrage-trader/internal/logger"
	"github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
	"golang.org/x/time/rate"
)

// ExchangeConfig holds Binance exchange configuration
type ExchangeConfig struct {
	APIKey    string
	SecretKey string
	Sandbox   bool
	RateLimit RateLimitConfig
	Retry     RetryConfig
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
}

// BinanceOrderResponse represents Binance order response
type BinanceOrderResponse struct {
	OrderID       string `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	Status        string `json:"status"`
	TransactTime  int64  `json:"transactTime"`
}

// BinanceTickerMessage represents WebSocket ticker message
type BinanceTickerMessage struct {
	Symbol    string `json:"s"`
	Price     string `json:"c"`
	Volume    string `json:"v"`
	Timestamp int64  `json:"E"`
}

type Client struct {
	config      ExchangeConfig
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	baseURL     string

	// Internal state
	serverTimeOffset time.Duration
	lastWeightUpdate time.Time
	currentWeight    int

	logger *logger.Logger
}

func NewClient(config ExchangeConfig) (*Client, error) {
	client := &Client{
		config:      config,
		httpClient:  createHTTPClient(),
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit.RequestsPerSecond), config.RateLimit.Burst),
		baseURL:     getBinanceURL(config.Sandbox),
		logger:      logger.New(logger.LevelInfo),
	}

	if err := client.syncServerTime(); err != nil {
		return nil, fmt.Errorf("failed to sync server time: %w", err)
	}

	return client, nil
}

func (c *Client) PlaceOrder(ctx context.Context, order types.Order) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	params := c.buildOrderParams(order)

	var response BinanceOrderResponse
	if err := c.makeSignedRequest(ctx, "POST", "/api/v3/order", params, &response); err != nil {
		return c.handleOrderError(err, order)
	}

	// Update order with exchange response
	order.ID = response.OrderID
	order.Status = c.mapBinanceOrderStatus(response.Status)
	order.Timestamp = time.Unix(response.TransactTime/1000, 0)

	c.logger.Info("Order placed successfully: %s %.8f @ %.2f", order.Symbol, order.Quantity, order.Price)

	return nil
}

func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	params := map[string]interface{}{
		"orderId": orderID,
	}

	return c.makeSignedRequest(ctx, "DELETE", "/api/v3/order", params, nil)
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*types.Order, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	params := map[string]interface{}{
		"orderId": orderID,
	}

	var response map[string]interface{}
	if err := c.makeSignedRequest(ctx, "GET", "/api/v3/order", params, &response); err != nil {
		return nil, err
	}

	return c.parseOrderResponse(response), nil
}

func (c *Client) GetActiveOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	params := map[string]interface{}{
		"symbol": symbol,
	}

	var response []map[string]interface{}
	if err := c.makeSignedRequest(ctx, "GET", "/api/v3/openOrders", params, &response); err != nil {
		return nil, err
	}

	orders := make([]types.Order, len(response))
	for i, orderData := range response {
		orders[i] = *c.parseOrderResponse(orderData)
	}

	return orders, nil
}

func (c *Client) GetFilledOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	params := map[string]interface{}{
		"symbol": symbol,
		"limit":  1000,
	}

	var response []map[string]interface{}
	if err := c.makeSignedRequest(ctx, "GET", "/api/v3/allOrders", params, &response); err != nil {
		return nil, err
	}

	orders := make([]types.Order, 0, len(response))
	for _, orderData := range response {
		if status, ok := orderData["status"].(string); ok && status == "FILLED" {
			order := c.parseOrderResponse(orderData)
			orders = append(orders, *order)
		}
	}

	return orders, nil
}

func (c *Client) GetTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	params := map[string]interface{}{
		"symbol": symbol,
	}

	var response map[string]interface{}
	if err := c.makeRequest(ctx, "GET", "/api/v3/ticker/24hr", params, &response); err != nil {
		return nil, err
	}

	return c.parseTickerResponse(response), nil
}

func (c *Client) GetOrderBook(ctx context.Context, symbol string, limit int) (*types.OrderBook, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	params := map[string]interface{}{
		"symbol": symbol,
		"limit":  limit,
	}

	var response map[string]interface{}
	if err := c.makeRequest(ctx, "GET", "/api/v3/depth", params, &response); err != nil {
		return nil, err
	}

	return c.parseOrderBookResponse(response), nil
}

func (c *Client) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]types.Candle, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	params := map[string]interface{}{
		"symbol":   symbol,
		"interval": interval,
		"limit":    limit,
	}

	var response [][]interface{}
	if err := c.makeRequest(ctx, "GET", "/api/v3/klines", params, &response); err != nil {
		return nil, err
	}

	return c.parseCandlesResponse(response), nil
}

func (c *Client) GetBalance(ctx context.Context) (*types.Balance, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	var response map[string]interface{}
	if err := c.makeSignedRequest(ctx, "GET", "/api/v3/account", nil, &response); err != nil {
		return nil, err
	}

	// Parse balances from account info
	balances, ok := response["balances"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid balance response")
	}

	// For simplicity, return USDT balance
	for _, balance := range balances {
		if balanceMap, ok := balance.(map[string]interface{}); ok {
			if asset, ok := balanceMap["asset"].(string); ok && asset == "USDT" {
				free, _ := strconv.ParseFloat(balanceMap["free"].(string), 64)
				locked, _ := strconv.ParseFloat(balanceMap["locked"].(string), 64)
				total := free + locked

				return &types.Balance{
					Asset:     asset,
					Free:      free,
					Locked:    locked,
					Total:     total,
					Timestamp: time.Now(),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("USDT balance not found")
}

func (c *Client) GetTradingFees(ctx context.Context, symbol string) (*types.TradingFees, error) {
	// Binance has fixed fees for most users
	return &types.TradingFees{
		Symbol:    symbol,
		MakerFee:  0.001, // 0.1%
		TakerFee:  0.001, // 0.1%
		Timestamp: time.Now(),
	}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	return c.makeRequest(ctx, "GET", "/api/v3/ping", nil, nil)
}

func (c *Client) Close() error {
	return nil
}

// Helper methods

func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

func getBinanceURL(sandbox bool) string {
	if sandbox {
		return "https://testnet.binance.vision"
	}
	return "https://api.binance.com"
}

func (c *Client) syncServerTime() error {
	var response map[string]interface{}
	if err := c.makeRequest(context.Background(), "GET", "/api/v3/time", nil, &response); err != nil {
		return err
	}

	serverTime, ok := response["serverTime"].(float64)
	if !ok {
		return fmt.Errorf("invalid server time response")
	}

	c.serverTimeOffset = time.Duration(serverTime)*time.Millisecond - time.Duration(time.Now().UnixNano())*time.Nanosecond
	return nil
}

func (c *Client) buildOrderParams(order types.Order) map[string]interface{} {
	params := map[string]interface{}{
		"symbol":   order.Symbol,
		"side":     string(order.Side),
		"type":     string(order.Type),
		"quantity": fmt.Sprintf("%.8f", order.Quantity),
	}

	if order.Type == types.OrderTypeLimit {
		params["price"] = fmt.Sprintf("%.8f", order.Price)
		params["timeInForce"] = "GTC"
	}

	return params
}

func (c *Client) makeSignedRequest(ctx context.Context, method, endpoint string, params map[string]interface{}, result interface{}) error {
	timestamp := time.Now().Add(c.serverTimeOffset).UnixNano() / 1e6
	params["timestamp"] = timestamp

	signature := c.generateSignature(params)
	params["signature"] = signature

	return c.makeRequest(ctx, method, endpoint, params, result)
}

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, params map[string]interface{}, result interface{}) error {
	url := c.baseURL + endpoint

	var req *http.Request
	var err error

	if method == "GET" {
		req, err = c.buildGETRequest(ctx, url, params)
	} else {
		req, err = c.buildPOSTRequest(ctx, url, params)
	}

	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	if method != "GET" {
		req.Header.Set("X-MBX-APIKEY", c.config.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := c.handleHTTPResponse(resp, result); err != nil {
		return err
	}

	return nil
}

func (c *Client) buildGETRequest(ctx context.Context, requestURL string, params map[string]interface{}) (*http.Request, error) {
	if len(params) > 0 {
		values := make(map[string][]string)
		for key, value := range params {
			values[key] = []string{fmt.Sprintf("%v", value)}
		}
		query := url.Values(values)
		requestURL += "?" + query.Encode()
	}

	return http.NewRequestWithContext(ctx, "GET", requestURL, nil)
}

func (c *Client) buildPOSTRequest(ctx context.Context, requestURL string, params map[string]interface{}) (*http.Request, error) {
	values := make(map[string][]string)
	for key, value := range params {
		values[key] = []string{fmt.Sprintf("%v", value)}
	}
	query := url.Values(values)

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, strings.NewReader(query.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (c *Client) generateSignature(params map[string]interface{}) string {
	query := make(url.Values)
	for key, value := range params {
		query.Set(key, fmt.Sprintf("%v", value))
	}

	h := hmac.New(sha256.New, []byte(c.config.SecretKey))
	h.Write([]byte(query.Encode()))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) handleHTTPResponse(resp *http.Response, result interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

func (c *Client) handleOrderError(err error, order types.Order) error {
	c.logger.Error("Order placement failed: %v", err)
	return fmt.Errorf("order placement failed: %w", err)
}

func (c *Client) mapBinanceOrderStatus(status string) types.OrderStatus {
	switch status {
	case "NEW":
		return types.OrderStatusNew
	case "PARTIALLY_FILLED":
		return types.OrderStatusPartiallyFilled
	case "FILLED":
		return types.OrderStatusFilled
	case "CANCELED":
		return types.OrderStatusCanceled
	case "REJECTED":
		return types.OrderStatusRejected
	default:
		return types.OrderStatusNew
	}
}

func (c *Client) parseOrderResponse(data map[string]interface{}) *types.Order {
	orderID, _ := data["orderId"].(string)
	symbol, _ := data["symbol"].(string)
	side, _ := data["side"].(string)
	orderType, _ := data["type"].(string)
	status, _ := data["status"].(string)

	quantity, _ := strconv.ParseFloat(data["origQty"].(string), 64)
	price, _ := strconv.ParseFloat(data["price"].(string), 64)
	filledQty, _ := strconv.ParseFloat(data["executedQty"].(string), 64)

	transactTime, _ := data["transactTime"].(float64)

	return &types.Order{
		ID:           orderID,
		Symbol:       symbol,
		Side:         types.OrderSide(side),
		Type:         types.OrderType(orderType),
		Quantity:     quantity,
		Price:        price,
		Status:       c.mapBinanceOrderStatus(status),
		FilledAmount: filledQty,
		FilledPrice:  price,
		Timestamp:    time.Unix(int64(transactTime)/1000, 0),
	}
}

func (c *Client) parseTickerResponse(data map[string]interface{}) *types.Ticker {
	symbol, _ := data["symbol"].(string)
	price, _ := strconv.ParseFloat(data["lastPrice"].(string), 64)
	volume, _ := strconv.ParseFloat(data["volume"].(string), 64)
	timestamp, _ := data["closeTime"].(float64)

	return &types.Ticker{
		Symbol:    symbol,
		Price:     price,
		Bid:       price - 0.1, // Approximate
		Ask:       price + 0.1, // Approximate
		Volume:    volume,
		Timestamp: time.Unix(int64(timestamp)/1000, 0),
	}
}

func (c *Client) parseOrderBookResponse(data map[string]interface{}) *types.OrderBook {
	symbol, _ := data["symbol"].(string)

	bidsData, _ := data["bids"].([][]interface{})
	asksData, _ := data["asks"].([][]interface{})

	bids := make([]types.OrderBookEntry, len(bidsData))
	for i, bid := range bidsData {
		price, _ := strconv.ParseFloat(bid[0].(string), 64)
		amount, _ := strconv.ParseFloat(bid[1].(string), 64)
		bids[i] = types.OrderBookEntry{Price: price, Amount: amount}
	}

	asks := make([]types.OrderBookEntry, len(asksData))
	for i, ask := range asksData {
		price, _ := strconv.ParseFloat(ask[0].(string), 64)
		amount, _ := strconv.ParseFloat(ask[1].(string), 64)
		asks[i] = types.OrderBookEntry{Price: price, Amount: amount}
	}

	return &types.OrderBook{
		Symbol: symbol,
		Bids:   bids,
		Asks:   asks,
	}
}

func (c *Client) parseCandlesResponse(data [][]interface{}) []types.Candle {
	candles := make([]types.Candle, len(data))

	for i, candle := range data {
		timestamp, _ := candle[0].(float64)
		open, _ := strconv.ParseFloat(candle[1].(string), 64)
		high, _ := strconv.ParseFloat(candle[2].(string), 64)
		low, _ := strconv.ParseFloat(candle[3].(string), 64)
		close, _ := strconv.ParseFloat(candle[4].(string), 64)
		volume, _ := strconv.ParseFloat(candle[5].(string), 64)

		candles[i] = types.Candle{
			Symbol:    "", // Will be set by caller
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			Timestamp: time.Unix(int64(timestamp)/1000, 0),
		}
	}

	return candles
}
