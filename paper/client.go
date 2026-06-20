package paper

import (
	"fmt"
	"sync"
)

// OrderSource supplies market metadata and order book snapshots to the paper
// client. Tests inject a fake; a production adapter would wrap lighter-go's
// HTTP order API.
type OrderSource interface {
	// OrderBookDetails returns the perps order book detail(s) for a market.
	OrderBookDetails(marketID int) ([]MarketDetail, error)
	// OrderBookOrders returns a full order book snapshot for a market.
	OrderBookOrders(marketID, limit int) (BookData, error)
}

// Client is a local simulation client. It is safe for concurrent use.
type Client struct {
	mu             sync.Mutex
	accountTier    AccountTier
	source         OrderSource
	orderBookLimit int
	account        *Account
	marketConfigs  map[int]MarketConfig
	orderBooks     map[int]*InMemoryOrderBook
}

// Option configures a Client.
type Option func(*Client)

// WithOrderBookLimit sets the snapshot depth (1..250, default 100).
func WithOrderBookLimit(limit int) Option { return func(c *Client) { c.orderBookLimit = limit } }

// WithAccountTier sets the fee tier (default TierStandard).
func WithAccountTier(tier AccountTier) Option { return func(c *Client) { c.accountTier = tier } }

// NewClient creates a paper trading client backed by the given order source.
func NewClient(source OrderSource, initialCollateralUSDC float64, opts ...Option) (*Client, error) {
	if initialCollateralUSDC <= 0 {
		return nil, &ParamError{Msg: fmt.Sprintf("initial collateral must be positive, got %v", initialCollateralUSDC)}
	}
	if source == nil {
		return nil, &ParamError{Msg: "order source must not be nil"}
	}
	c := &Client{
		accountTier:    TierStandard,
		source:         source,
		orderBookLimit: 100,
		account:        newAccount(initialCollateralUSDC),
		marketConfigs:  map[int]MarketConfig{},
		orderBooks:     map[int]*InMemoryOrderBook{},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.orderBookLimit < 1 || c.orderBookLimit > 250 {
		return nil, &ParamError{Msg: "order book limit must be between 1 and 250"}
	}
	return c, nil
}

// AccountTier returns the configured fee tier.
func (c *Client) AccountTier() AccountTier {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.accountTier
}

// TrackMarketSnapshot fetches the market config and a one-shot order book
// snapshot for the market.
func (c *Client) TrackMarketSnapshot(marketID int) error {
	if err := validatePerpMarketID(marketID); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureMarketConfig(marketID); err != nil {
		return err
	}
	return c.refreshOrderBookLocked(marketID)
}

// TrackMarket loads the market config so live order book data can be fed via
// ApplyLiveSnapshot / ApplyLiveDelta. Unlike the Python client it does not open
// a WebSocket itself — the caller supplies the live feed (e.g. via lighter-go's
// ws client) and forwards snapshots/deltas to the Apply* methods.
func (c *Client) TrackMarket(marketID int) error {
	if err := validatePerpMarketID(marketID); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureMarketConfig(marketID); err != nil {
		return err
	}
	c.bookFor(marketID)
	return nil
}

// RefreshOrderBook re-fetches the snapshot for an already-tracked market.
func (c *Client) RefreshOrderBook(marketID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.marketConfigs[marketID]; !ok {
		return &ParamError{Msg: fmt.Sprintf("market %d not tracked", marketID)}
	}
	return c.refreshOrderBookLocked(marketID)
}

func (c *Client) refreshOrderBookLocked(marketID int) error {
	data, err := c.source.OrderBookOrders(marketID, c.orderBookLimit)
	if err != nil {
		return err
	}
	book := c.orderBooks[marketID]
	if book == nil {
		book = &InMemoryOrderBook{}
		c.orderBooks[marketID] = book
	}
	book.ApplySnapshot(data)
	c.checkLiquidationAndUpdateMetrics()
	return nil
}

// ApplyLiveSnapshot applies a live order book snapshot (as a WS listener would).
func (c *Client) ApplyLiveSnapshot(marketID int, data BookData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bookFor(marketID).ApplySnapshot(data)
	c.checkLiquidationAndUpdateMetrics()
}

// ApplyLiveDelta applies a live order book delta (as a WS listener would).
func (c *Client) ApplyLiveDelta(marketID int, data BookData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bookFor(marketID).ApplyDelta(data)
	c.checkLiquidationAndUpdateMetrics()
}

func (c *Client) bookFor(marketID int) *InMemoryOrderBook {
	book := c.orderBooks[marketID]
	if book == nil {
		book = &InMemoryOrderBook{}
		c.orderBooks[marketID] = book
	}
	return book
}

// CreatePaperOrder simulates a taker order against the tracked book.
func (c *Client) CreatePaperOrder(req OrderRequest) (OrderResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	config, ok := c.marketConfigs[req.MarketID]
	if !ok {
		return OrderResult{}, &ParamError{Msg: fmt.Sprintf("market %d not tracked, call TrackMarketSnapshot first", req.MarketID)}
	}
	book := c.orderBooks[req.MarketID]
	if book == nil {
		return OrderResult{}, &ParamError{Msg: fmt.Sprintf("no order book for market %d", req.MarketID)}
	}
	if err := ValidateOrder(req, config); err != nil {
		return OrderResult{}, err
	}

	fills, unfilled := SimulateMatch(req, book.Asks, book.Bids, config)

	var totalFilled, totalQuote, totalFee float64
	for _, fill := range fills {
		ApplyFill(c.account, req.MarketID, req.Side, fill.Size, fill.Price, fill.Fee, false)
		totalFilled += fill.Size
		totalQuote += fill.Size * fill.Price
		totalFee += fill.Fee
	}

	avgPrice := 0.0
	if totalFilled > 0 {
		avgPrice = totalQuote / totalFilled
	}
	c.checkLiquidationAndUpdateMetrics()

	return OrderResult{
		OrderType:   req.OrderType,
		Side:        req.Side,
		MarketID:    req.MarketID,
		Fills:       fills,
		FilledSize:  totalFilled,
		AvgPrice:    avgPrice,
		TotalFee:    totalFee,
		QuoteAmount: totalQuote,
		Unfilled:    unfilled,
	}, nil
}

// GetHealth returns the current cross-margin account health.
func (c *Client) GetHealth() AccountHealth {
	c.mu.Lock()
	defer c.mu.Unlock()
	return ComputeHealth(c.account, c.getMarkPrices(), c.marketConfigs)
}

// GetLiquidationPrice returns the liquidation price for a market's position.
func (c *Client) GetLiquidationPrice(marketID int) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return ComputeLiquidationPrice(c.account, marketID, c.getMarkPrices(), c.marketConfigs)
}

// GetPosition returns a copy of the position for a market, or nil if flat.
func (c *Client) GetPosition(marketID int) *Position {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p := c.account.Positions[marketID]; p != nil {
		return p.clone()
	}
	return nil
}

// GetAccount returns a deep copy of the account.
func (c *Client) GetAccount() *Account {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.account.clone()
}

// GetCollateral returns the current collateral.
func (c *Client) GetCollateral() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.account.Collateral
}

// GetTrades returns a copy of the trade log.
func (c *Client) GetTrades() []Trade {
	c.mu.Lock()
	defer c.mu.Unlock()
	trades := make([]Trade, len(c.account.Trades))
	copy(trades, c.account.Trades)
	return trades
}

// GetPortfolioValue returns total account value; errors if positions exist but
// no mark prices are available.
func (c *Client) GetPortfolioValue() (float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	markPrices := c.getMarkPrices()
	if len(c.account.Positions) > 0 && len(markPrices) == 0 {
		return 0, &ParamError{Msg: "no mark prices available"}
	}
	return ComputeTotalAccountValue(c.account, markPrices), nil
}

func (c *Client) checkLiquidationAndUpdateMetrics() []int {
	liquidated := CheckAndLiquidate(c.account, c.getMarkPrices(), c.marketConfigs)
	UpdatePositionMetrics(c.account, c.getMarkPrices(), c.marketConfigs)
	return liquidated
}

func (c *Client) getMarkPrices() map[int]float64 {
	prices := map[int]float64{}
	for marketID, book := range c.orderBooks {
		mark, ok := book.MidPrice()
		if !ok {
			if config, hasCfg := c.marketConfigs[marketID]; hasCfg && config.LastTradePrice > 0 {
				mark, ok = config.LastTradePrice, true
			}
		}
		if ok && mark > 0 {
			prices[marketID] = mark
		}
	}
	return prices
}

func (c *Client) ensureMarketConfig(marketID int) error {
	if _, ok := c.marketConfigs[marketID]; ok {
		return nil
	}
	details, err := c.source.OrderBookDetails(marketID)
	if err != nil {
		return err
	}
	for _, detail := range details {
		if detail.MarketID != marketID {
			continue
		}
		config, err := marketConfigFromDetail(detail, c.accountTier)
		if err != nil {
			return err
		}
		c.marketConfigs[marketID] = config
		return nil
	}
	return &ParamError{Msg: fmt.Sprintf("perps order book detail not found for market %d", marketID)}
}

func validatePerpMarketID(marketID int) error {
	if marketID >= 2048 {
		return &ParamError{Msg: fmt.Sprintf("paper trading only supports perp markets (market_id < 2048), got %d", marketID)}
	}
	return nil
}
