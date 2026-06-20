package api

// Account represents a user account
type Account struct {
	Index            int64  `json:"index"`
	L1Address        string `json:"l1_address,omitempty"`
	Nonce            int64  `json:"nonce"`
	CollateralValue  string `json:"collateral_value"`
	PositionValue    string `json:"position_value"`
	PortfolioValue   string `json:"portfolio_value"`
	AvailableBalance string `json:"available_balance"`
	MaxWithdrawable  string `json:"max_withdrawable"`
	InitialMargin    string `json:"initial_margin"`
	MaintenanceMargin string `json:"maintenance_margin"`
	UnrealizedPnl    string `json:"unrealized_pnl"`
	MarkPrice        string `json:"mark_price,omitempty"`
	IsLiquidatable   bool   `json:"is_liquidatable"`
}

// DetailedAccount includes full account details with positions and assets
type DetailedAccount struct {
	Account
	Positions []AccountPosition `json:"positions,omitempty"`
	Assets    []AccountAsset    `json:"assets,omitempty"`
	Metadata  *AccountMetadata  `json:"metadata,omitempty"`
}

// DetailedAccounts is the response for account queries
type DetailedAccounts struct {
	BaseResponse
	Accounts []DetailedAccount `json:"accounts"`
}

// AccountPosition represents a position in a market
type AccountPosition struct {
	MarketID               int    `json:"market_id"`
	Symbol                 string `json:"symbol"`
	InitialMarginFraction  string `json:"initial_margin_fraction"`
	OpenOrderCount         int    `json:"open_order_count"`
	PendingOrderCount      int    `json:"pending_order_count"`
	PositionTiedOrderCount int    `json:"position_tied_order_count"`
	Sign                   int    `json:"sign"` // 1 = long, -1 = short
	Position               string `json:"position"` // position size
	AvgEntryPrice          string `json:"avg_entry_price"`
	PositionValue          string `json:"position_value"`
	UnrealizedPnl          string `json:"unrealized_pnl"`
	RealizedPnl            string `json:"realized_pnl"`
	LiquidationPrice       string `json:"liquidation_price"`
	TotalFundingPaidOut    string `json:"total_funding_paid_out,omitempty"`
	MarginMode             int    `json:"margin_mode"` // 0 = cross, 1 = isolated
	AllocatedMargin        string `json:"allocated_margin"`
}

// AccountAsset represents an asset balance in an account
type AccountAsset struct {
	AssetIndex       int16  `json:"asset_id"`
	AssetSymbol      string `json:"symbol,omitempty"`
	Balance          string `json:"balance"`
	AvailableBalance string `json:"available_balance"`
	LockedBalance    string `json:"locked_balance,omitempty"`
}

// AccountMetadata represents account metadata
type AccountMetadata struct {
	AccountIndex   int64  `json:"account_index"`
	ReferralCode   string `json:"referral_code,omitempty"`
	Tier           string `json:"tier,omitempty"`
	TotalVolume    string `json:"total_volume,omitempty"`
	MakerFeeRate   string `json:"maker_fee_rate,omitempty"`
	TakerFeeRate   string `json:"taker_fee_rate,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
}

// AccountMetadatas is the response for account metadata queries
type AccountMetadatas struct {
	BaseResponse
	Metadatas []AccountMetadata `json:"metadatas"`
}

// AccountLimits represents trading limits for an account
type AccountLimits struct {
	BaseResponse
	MaxLLPPercentage    int    `json:"max_llp_percentage"`
	MaxLLPAmount        string `json:"max_llp_amount"`
	UserTier            string `json:"user_tier"`
	CanCreatePublicPool bool   `json:"can_create_public_pool"`
}

// AccountStats represents account statistics
type AccountStats struct {
	TotalVolume      string `json:"total_volume"`
	TotalTrades      int64  `json:"total_trades"`
	TotalPnl         string `json:"total_pnl"`
	WinRate          string `json:"win_rate,omitempty"`
	AverageTrade     string `json:"average_trade,omitempty"`
}

// SubAccount represents a sub-account
type SubAccount struct {
	Index        int64  `json:"index"`
	MasterIndex  int64  `json:"master_index"`
	L1Address    string `json:"l1_address,omitempty"`
}

// SubAccounts is the response for sub-account queries
type SubAccounts struct {
	BaseResponse
	MasterAccount int64        `json:"master_account"`
	SubAccounts   []SubAccount `json:"sub_accounts"`
}

// ApiKey represents an API key
type ApiKey struct {
	AccountIndex int64  `json:"account_index"`
	ApiKeyIndex  uint8  `json:"api_key_index"`
	Nonce        int64  `json:"nonce"`
	PublicKey    string `json:"public_key"`
}

// AccountApiKeys is the response for API key queries
type AccountApiKeys struct {
	BaseResponse
	ApiKeys []ApiKey `json:"api_keys"`
}

// RespChangeAccountTier is the response for changing account tier
type RespChangeAccountTier struct {
	BaseResponse
	AccountIndex int64  `json:"account_index"`
	NewTier      string `json:"new_tier"`
}

// L1Metadata represents L1 address metadata
type L1Metadata struct {
	BaseResponse
	L1Address        string `json:"l1_address"`
	LinkedAccounts   []int64 `json:"linked_accounts,omitempty"`
	TotalDeposited   string `json:"total_deposited,omitempty"`
	TotalWithdrawn   string `json:"total_withdrawn,omitempty"`
}

// LiquidationInfo represents liquidation information
type LiquidationInfo struct {
	AccountIndex     int64  `json:"account_index"`
	MarketIndex      int16  `json:"market_index"`
	Size             string `json:"size"`
	Price            string `json:"price"`
	LiquidationFee   string `json:"liquidation_fee"`
	Timestamp        int64  `json:"timestamp"`
	TxHash           string `json:"tx_hash,omitempty"`
}

// LiquidationInfos is the response for liquidation queries
type LiquidationInfos struct {
	BaseResponse
	Liquidations []LiquidationInfo `json:"liquidations"`
	Cursor       Cursor            `json:"cursor,omitempty"`
}

// PositionFunding represents funding payment for a position
type PositionFunding struct {
	MarketIndex   int16  `json:"market_index"`
	FundingRate   string `json:"funding_rate"`
	FundingAmount string `json:"funding_amount"`
	PositionSize  string `json:"position_size"`
	Side          string `json:"side"`
	Timestamp     int64  `json:"timestamp"`
}

// PositionFundings is the response for position funding queries
type PositionFundings struct {
	BaseResponse
	Fundings []PositionFunding `json:"fundings"`
	Cursor   Cursor            `json:"cursor,omitempty"`
}

// RiskInfo represents risk information for an account
type RiskInfo struct {
	AccountIndex         int64  `json:"account_index"`
	MarginRatio          string `json:"margin_ratio"`
	LiquidationRisk      string `json:"liquidation_risk"`
	MaintenanceMargin    string `json:"maintenance_margin"`
	AvailableMargin      string `json:"available_margin"`
	TotalPositionValue   string `json:"total_position_value"`
}

// TransferFeeInfo represents transfer fee information
type TransferFeeInfo struct {
	BaseResponse
	FromAccountIndex int64  `json:"from_account_index"`
	ToAccountIndex   int64  `json:"to_account_index,omitempty"`
	FeeRate          string `json:"fee_rate"`
	MinFee           string `json:"min_fee"`
	MaxFee           string `json:"max_fee"`
	TransferFeeUsdc  int64  `json:"transfer_fee_usdc"` // Fee in USDC base units (6 decimals)
}

// ValidatorInfo represents validator information
type ValidatorInfo struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	Stake     string `json:"stake"`
	IsActive  bool   `json:"is_active"`
}
