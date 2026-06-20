// Package paper is a local simulation engine for testing trading strategies
// against real Lighter order book data, ported from lighter-python's
// paper_client module to keep the Go and Python SDKs at feature parity.
//
// It maintains a virtual account with simulated collateral, executes taker-only
// fills against snapshot or live order book state, and tracks positions, PnL,
// and account health locally.
//
// Limitations (same as the Python implementation):
//   - Perp markets only — spot markets are not supported.
//   - Taker-only — supports MARKET and IOC order types, no resting limit orders.
//   - Cross-margin only — isolated margin is not simulated.
//   - No funding simulation.
//   - Read-only simulation — no transactions are submitted to the exchange.
package paper

// feeTick is the denominator for the documented fee schedule fractions.
const feeTick = 1_000_000

// AccountTier mirrors the publicly documented account tiers and their
// taker/maker fee fractions. See https://docs.lighter.xyz/trading/trading-fees
type AccountTier struct {
	Name     string
	TakerFee float64
	MakerFee float64
}

var (
	TierStandard = AccountTier{"STANDARD", 0.0, 0.0}
	TierPremium  = AccountTier{"PREMIUM", 280.0 / feeTick, 40.0 / feeTick}
	TierPremium1 = AccountTier{"PREMIUM_1", 273.0 / feeTick, 39.0 / feeTick}
	TierPremium2 = AccountTier{"PREMIUM_2", 266.0 / feeTick, 38.0 / feeTick}
	TierPremium3 = AccountTier{"PREMIUM_3", 252.0 / feeTick, 36.0 / feeTick}
	TierPremium4 = AccountTier{"PREMIUM_4", 238.0 / feeTick, 34.0 / feeTick}
	TierPremium5 = AccountTier{"PREMIUM_5", 224.0 / feeTick, 32.0 / feeTick}
	TierPremium6 = AccountTier{"PREMIUM_6", 210.0 / feeTick, 30.0 / feeTick}
	TierPremium7 = AccountTier{"PREMIUM_7", 196.0 / feeTick, 28.0 / feeTick}
)

// OrderType is the supported taker order type.
type OrderType int

const (
	OrderMarket OrderType = 0
	OrderIOC    OrderType = 1
)

// OrderSide is the trade direction.
type OrderSide int

const (
	SideBuy  OrderSide = 0
	SideSell OrderSide = 1
)

// HealthStatus mirrors PaperHealthStatus. Values 2 and 3 are reserved: the
// paper sim collapses partial/full liquidation since liquidation runs
// atomically with every mark update.
type HealthStatus int

const (
	StatusHealthy        HealthStatus = 0
	StatusPreLiquidation HealthStatus = 1
	StatusBankruptcy     HealthStatus = 4
)

// OrderRequest is a taker order to simulate.
type OrderRequest struct {
	MarketID   int
	Side       OrderSide
	BaseAmount float64
	Price      float64
	OrderType  OrderType
}

// Fill is one matched slice of an order.
type Fill struct {
	Price   float64
	Size    float64
	Fee     float64
	IsMaker bool
}

// OrderResult summarizes a simulated order.
type OrderResult struct {
	OrderType   OrderType
	Side        OrderSide
	MarketID    int
	Fills       []Fill
	FilledSize  float64
	AvgPrice    float64
	TotalFee    float64
	QuoteAmount float64
	Unfilled    float64
}

// Position is an open position in one market.
type Position struct {
	MarketID         int
	Size             float64
	EntryQuote       float64
	AvgEntryPrice    float64
	MarkPrice        float64
	UnrealizedPnl    float64
	RealizedPnl      float64
	LiquidationPrice float64
}

func (p Position) clone() *Position {
	cp := p
	return &cp
}

// Trade is a recorded fill against the account.
type Trade struct {
	MarketID      int
	Side          OrderSide
	Size          float64
	Price         float64
	Fee           float64
	RealizedPnl   float64
	IsLiquidation bool
}

// AccountHealth is a snapshot of cross-margin account health.
type AccountHealth struct {
	Status                       HealthStatus
	TotalAccountValue            float64
	InitialMarginRequirement     float64
	MaintenanceMarginRequirement float64
	MarginUsage                  float64
	Leverage                     float64
	// HasBeenLiquidated is sticky for the lifetime of the Client: set true the
	// first time a paper liquidation fires and never cleared. Reads as "this
	// session has experienced at least one liquidation".
	HasBeenLiquidated bool
}

// Account is the virtual margin account.
type Account struct {
	InitialCollateral float64
	Collateral        float64
	Positions         map[int]*Position
	Trades            []Trade
	HasBeenLiquidated bool
}

func newAccount(collateral float64) *Account {
	return &Account{
		InitialCollateral: collateral,
		Collateral:        collateral,
		Positions:         map[int]*Position{},
	}
}

func (a *Account) clone() *Account {
	positions := make(map[int]*Position, len(a.Positions))
	for id, p := range a.Positions {
		positions[id] = p.clone()
	}
	trades := make([]Trade, len(a.Trades))
	copy(trades, a.Trades)
	return &Account{
		InitialCollateral: a.InitialCollateral,
		Collateral:        a.Collateral,
		Positions:         positions,
		Trades:            trades,
		HasBeenLiquidated: a.HasBeenLiquidated,
	}
}

// MarketConfig holds the per-market parameters used by the simulation.
type MarketConfig struct {
	MarketID                     int
	Symbol                       string
	SizeDecimals                 int
	PriceDecimals                int
	DefaultInitialMarginFraction int
	MinInitialMarginFraction     int
	MaintenanceMarginFraction    int
	CloseoutMarginFraction       int
	TakerFee                     float64
	MakerFee                     float64
	MinBaseAmount                float64
	MinQuoteAmount               float64
	LastTradePrice               float64
}

// MarketDetail is the subset of an exchange "order book detail" the engine
// needs to build a MarketConfig. An OrderSource returns these.
type MarketDetail struct {
	MarketID                     int
	Symbol                       string
	MarketType                   string
	SizeDecimals                 int
	PriceDecimals                int
	DefaultInitialMarginFraction int
	MinInitialMarginFraction     int
	MaintenanceMarginFraction    int
	CloseoutMarginFraction       int
	MinBaseAmount                float64
	MinQuoteAmount               float64
	LastTradePrice               float64
}

func marketConfigFromDetail(d MarketDetail, tier AccountTier) (MarketConfig, error) {
	if d.MarketType != "perp" {
		return MarketConfig{}, &ParamError{Msg: "paper trading only supports perp markets, got " + d.MarketType}
	}
	return MarketConfig{
		MarketID:                     d.MarketID,
		Symbol:                       d.Symbol,
		SizeDecimals:                 d.SizeDecimals,
		PriceDecimals:                d.PriceDecimals,
		DefaultInitialMarginFraction: d.DefaultInitialMarginFraction,
		MinInitialMarginFraction:     d.MinInitialMarginFraction,
		MaintenanceMarginFraction:    d.MaintenanceMarginFraction,
		CloseoutMarginFraction:       d.CloseoutMarginFraction,
		TakerFee:                     tier.TakerFee,
		MakerFee:                     tier.MakerFee,
		MinBaseAmount:                d.MinBaseAmount,
		MinQuoteAmount:               d.MinQuoteAmount,
		LastTradePrice:               d.LastTradePrice,
	}, nil
}

// ParamError is returned for invalid requests / configuration.
type ParamError struct{ Msg string }

func (e *ParamError) Error() string { return e.Msg }
