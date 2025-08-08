package binance

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/Zmey56/crypto-arbitrage-trader/pkg/types"
    "golang.org/x/time/rate"
)

type Client struct {
	config      ExchangeConfig
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	baseURL     string

	// WebSocket connections
	wsConn      *websocket.Conn
	wsReconnect chan struct{}

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
		wsReconnect: make(chan struct{}, 1),
		logger:      logger.NewLogger("binance"),
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
	order.Status = mapBinanceOrderStatus(response.Status)
	order.Timestamp = time.Unix(response.TransactTime/1000, 0)

	c.logger.Info("Order placed successfully",
		"symbol", order.Symbol,
		"side", order.Side,
		"quantity", order.Quantity,
		"orderID", order.ID)

	return nil
}

func (c *Client) makeSignedRequest(ctx context.Context, method, endpoint string, params map[string]interface{}, result interface{}) error {
	params["timestamp"] = time.Now().Add(c.serverTimeOffset).UnixNano() / 1e6

	signature := c.generateSignature(params)
	params["signature"] = signature

	url := c.baseURL + endpoint

	req, err := c.buildHTTPRequest(method, url, params)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.config.APIKey)
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := c.handleHTTPResponse(resp, result); err != nil {
		return err
	}

	// Update rate limit tracking
	c.updateRateLimitInfo(resp.Header)

	return nil
}

func (c *Client) SubscribeToTickers(ctx context.Context, symbols []string) (<-chan types.Ticker, error) {
	tickerChan := make(chan types.Ticker, 1000)

	wsURL := c.buildWebSocketURL(symbols)
	conn, err := websocket.Dial(wsURL, "", c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.wsConn = conn

	go c.handleWebSocketMessages(ctx, tickerChan)
	go c.maintainWebSocketConnection(ctx)

	return tickerChan, nil
}

func (c *Client) handleWebSocketMessages(ctx context.Context, tickerChan chan<- types.Ticker) {
	defer close(tickerChan)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.wsReconnect:
			if err := c.reconnectWebSocket(); err != nil {
				c.logger.Error("Failed to reconnect WebSocket", "error", err)
				continue
			}
		default:
			var message BinanceTickerMessage
			if err := websocket.JSON.Receive(c.wsConn, &message); err != nil {
				c.logger.Error("WebSocket receive error", "error", err)
				c.wsReconnect <- struct{}{}
				continue
			}

			ticker := c.convertBinanceTicker(message)

			select {
			case tickerChan <- ticker:
			case <-ctx.Done():
				return
			}
		}
	}
}
