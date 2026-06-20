package api

// AccountPnL represents account PnL data from /api/v1/pnl
type AccountPnL struct {
	BaseResponse
	Resolution string     `json:"resolution"`
	PnL        []PnLEntry `json:"pnl"`
}

// PnLEntry represents a single PnL data point
// Fields use float64 since the API returns numbers, not strings
type PnLEntry struct {
	Timestamp       int64   `json:"timestamp"`
	TradePnL        float64 `json:"trade_pnl"`         // Perpetual trading PnL
	TradeSpotPnL    float64 `json:"trade_spot_pnl"`    // Spot trading PnL
	Inflow          float64 `json:"inflow"`            // Deposits
	Outflow         float64 `json:"outflow"`           // Withdrawals
	SpotInflow      float64 `json:"spot_inflow"`       // Spot buys
	SpotOutflow     float64 `json:"spot_outflow"`      // Spot sells
	PoolPnL         float64 `json:"pool_pnl"`          // Liquidity pool PnL
	PoolInflow      float64 `json:"pool_inflow"`       // Pool deposits
	PoolOutflow     float64 `json:"pool_outflow"`      // Pool withdrawals
	StakingPnL      float64 `json:"staking_pnl"`       // Staking rewards
	StakingInflow   float64 `json:"staking_inflow"`    // Staking deposits
	StakingOutflow  float64 `json:"staking_outflow"`   // Staking withdrawals
	PoolTotalShares float64 `json:"pool_total_shares"` // LP shares
}

// DailyPnL represents daily PnL summary
type DailyPnL struct {
	Date            string `json:"date"` // YYYY-MM-DD format
	StartValue      string `json:"start_value"`
	EndValue        string `json:"end_value"`
	HighValue       string `json:"high_value"`
	LowValue        string `json:"low_value"`
	RealizedPnl     string `json:"realized_pnl"`
	UnrealizedPnl   string `json:"unrealized_pnl"`
	TotalPnl        string `json:"total_pnl"`
	TotalReturn     string `json:"total_return"`
	TotalReturnPct  string `json:"total_return_pct"`
	Volume          string `json:"volume"`
	TradeCount      int64  `json:"trade_count"`
	Fees            string `json:"fees"`
	FundingPayments string `json:"funding_payments"`
}

// MarketPnL represents PnL for a specific market
type MarketPnL struct {
	MarketIndex    int16  `json:"market_index"`
	MarketSymbol   string `json:"market_symbol,omitempty"`
	RealizedPnl    string `json:"realized_pnl"`
	UnrealizedPnl  string `json:"unrealized_pnl"`
	TotalPnl       string `json:"total_pnl"`
	Volume         string `json:"volume"`
	TradeCount     int64  `json:"trade_count"`
	Fees           string `json:"fees"`
	FundingPayments string `json:"funding_payments"`
}

// PnLSummary represents a summary of PnL
type PnLSummary struct {
	AccountIndex   int64       `json:"account_index"`
	TotalPnl       string      `json:"total_pnl"`
	RealizedPnl    string      `json:"realized_pnl"`
	UnrealizedPnl  string      `json:"unrealized_pnl"`
	TotalVolume    string      `json:"total_volume"`
	TotalTrades    int64       `json:"total_trades"`
	TotalFees      string      `json:"total_fees"`
	WinRate        string      `json:"win_rate,omitempty"`
	ProfitFactor   string      `json:"profit_factor,omitempty"`
	ByMarket       []MarketPnL `json:"by_market,omitempty"`
}
