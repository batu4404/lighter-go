package client

import (
	"context"
	"github.com/0xJord4n/lighter-go/types/api"
	"github.com/0xJord4n/lighter-go/types/apimodels"
)

// MinimalHTTPClient is the minimal interface for HTTP operations.
// This is maintained for backward compatibility.
type MinimalHTTPClient interface {
	GetNextNonce(ctx context.Context, accountIndex int64, apiKeyIndex uint8) (int64, error)
	GetApiKey(ctx context.Context, accountIndex int64, apiKeyIndex uint8) (string, error)
	// InvalidateApiKeys drops any cached API-key data for the account so the
	// next GetApiKey refetches from the server. Implementations that do not
	// cache may treat this as a no-op.
	InvalidateApiKeys(accountIndex int64)
}

// FullHTTPClient extends MinimalHTTPClient with complete API coverage.
// Use this interface for full access to all Lighter API endpoints.
type FullHTTPClient interface {
	MinimalHTTPClient

	// API group accessors - each returns a specialized API client
	Account() AccountAPI
	Order() OrderAPI
	Transaction() TransactionAPI
	Candlestick() CandlestickAPI
	Block() BlockAPI
	Bridge() BridgeAPI
	Info() InfoAPI
	Referral() ReferralAPI
}

// ReferralUseOpts contains the optional fields accepted by referral_use.
type ReferralUseOpts struct {
	Discord   string
	Telegram  string
	Signature string
}

// ReferralAPI provides access to referral-related endpoints
type ReferralAPI interface {
	// ReferralUse assigns the given referral code to the account identified by the
	// L1 address (POST /api/v1/referral/use). auth is a Lighter auth token for the
	// account. x is required by the API (typically a social handle); opts may be nil.
	ReferralUse(ctx context.Context, l1Address, referralCode, x, auth string, opts *ReferralUseOpts) error
}

// AccountAPI provides access to account-related endpoints
type AccountAPI interface {
	// GetAccount retrieves account details
	GetAccount(ctx context.Context, by api.QueryBy, value string) (*api.DetailedAccounts, error)

	// GetAccountsByL1Address retrieves accounts linked to an L1 address
	GetAccountsByL1Address(ctx context.Context, l1Address string) (*api.SubAccounts, error)

	// GetAccountMetadata retrieves account metadata
	GetAccountMetadata(ctx context.Context, by api.QueryBy, value string, auth string) (*api.AccountMetadatas, error)

	// GetAccountLimits retrieves account trading limits
	GetAccountLimits(ctx context.Context, accountIndex int64, auth string) (*api.AccountLimits, error)

	// GetLiquidations retrieves account liquidation history
	GetLiquidations(ctx context.Context, accountIndex int64, limit int, auth string, opts *LiquidationOpts) (*api.LiquidationInfos, error)

	// GetPositionFunding retrieves position funding history
	GetPositionFunding(ctx context.Context, accountIndex int64, limit int, auth string, opts *PositionFundingOpts) (*api.PositionFundings, error)

	// GetPnL retrieves account PnL history by account index
	GetPnL(ctx context.Context, accountIndex int64, resolution string, timestamps api.TimestampRange, countBack int, auth string, ignoreTransfers bool) (*api.AccountPnL, error)

	// GetPublicPoolsMetadata retrieves public pool metadata
	GetPublicPoolsMetadata(ctx context.Context, filter string, index int, limit int, auth string, accountIndex *int64) (*api.RespPublicPoolsMetadata, error)

	// ChangeAccountTier changes the account tier (requires auth)
	ChangeAccountTier(ctx context.Context, accountIndex int64, newTier string, auth string) (*api.RespChangeAccountTier, error)

	// GetL1Metadata retrieves L1 address metadata
	GetL1Metadata(ctx context.Context, l1Address string, auth string) (*api.L1Metadata, error)

	// GetApiKeys retrieves API keys for an account
	GetApiKeys(ctx context.Context, accountIndex int64, apiKeyIndex *uint8) (*api.AccountApiKeys, error)
}

// LiquidationOpts contains options for liquidation queries
type LiquidationOpts struct {
	MarketID *int16
	Cursor   string
}

// PositionFundingOpts contains options for position funding queries
type PositionFundingOpts struct {
	MarketID *int16
	Cursor   string
	Side     string
}

// OrderAPI provides access to order-related endpoints
type OrderAPI interface {
	// GetActiveOrders retrieves active orders for an account
	GetActiveOrders(ctx context.Context, accountIndex int64, marketID *int16, auth string) (*api.Orders, error)

	// GetInactiveOrders retrieves order history
	GetInactiveOrders(ctx context.Context, accountIndex int64, marketID *int16, opts *InactiveOrdersOpts) (*api.Orders, error)

	// GetOrderBooks retrieves order book snapshots
	GetOrderBooks(ctx context.Context, marketID *int16, filter api.MarketFilter) (*api.OrderBooks, error)

	// GetOrderBookDetails retrieves detailed order book data
	// Pass nil for marketID to get all markets
	GetOrderBookDetails(ctx context.Context, marketID *int16, filter api.MarketFilter) (*api.OrderBookDetails, error)

	// GetOrderBookOrders retrieves individual orders in the order book
	GetOrderBookOrders(ctx context.Context, marketID int16, limit int) (*apimodels.OrderBookOrders, error)

	// GetRecentTrades retrieves recent trades for a market
	GetRecentTrades(ctx context.Context, marketID int16, limit int) (*api.Trades, error)

	// GetTrades retrieves trade history
	GetTrades(ctx context.Context, marketID int16, accountIndex *int64, opts *TradesOpts) (*api.Trades, error)

	// GetAssetDetails retrieves asset information
	GetAssetDetails(ctx context.Context, assetID *int16) (*api.AssetDetails, error)

	// GetExchangeStats retrieves exchange-wide statistics
	GetExchangeStats(ctx context.Context) (*api.ExchangeStats, error)
}

// InactiveOrdersOpts contains options for inactive order queries.
// The /accountInactiveOrders endpoint supports only limit and cursor; it has
// no status/sort parameters (matching lighter-python order_api.py).
type InactiveOrdersOpts struct {
	Limit  int
	Cursor string
}

// TradesOpts contains options for trade queries
type TradesOpts struct {
	Limit     int
	Cursor    string
	SortBy    string
	SortOrder string
}

// TransactionAPI provides access to transaction-related endpoints
type TransactionAPI interface {
	// SendTx submits a signed transaction
	SendTx(ctx context.Context, txType uint8, txInfo string, priceProtection *api.PriceProtection) (*api.RespSendTx, error)

	// SendTxWithIndices submits a transaction with explicit account and API key indices
	SendTxWithIndices(ctx context.Context, txType uint8, txInfo string, priceProtection *api.PriceProtection, accountIndex *int64, apiKeyIndex *uint8, auth string) (*api.RespSendTx, error)

	// SendTxBatch submits multiple transactions
	SendTxBatch(ctx context.Context, txTypes []uint8, txInfos []string) (*api.RespSendTxBatch, error)

	// SendTxBatchWithAuth submits multiple transactions with auth header
	SendTxBatchWithAuth(ctx context.Context, txTypes []uint8, txInfos []string, auth string) (*api.RespSendTxBatch, error)

	// GetTx retrieves a transaction by hash or sequence index
	GetTx(ctx context.Context, by api.QueryBy, value string) (*api.EnrichedTx, error)

	// GetTxs retrieves transactions with pagination
	GetTxs(ctx context.Context, index *int64, limit int) (*api.Txs, error)

	// GetAccountTxs retrieves transactions for an account
	GetAccountTxs(ctx context.Context, by api.QueryBy, value string, limit int, types []api.TxType) (*api.Txs, error)

	// GetTxFromL1TxHash retrieves a transaction by its L1 hash
	GetTxFromL1TxHash(ctx context.Context, hash string) (*api.EnrichedTx, error)

	// GetDepositHistory retrieves deposit history
	GetDepositHistory(ctx context.Context, accountIndex int64, l1Address string, filter string, cursor string) (*api.DepositHistory, error)

	// GetWithdrawHistory retrieves withdrawal history
	GetWithdrawHistory(ctx context.Context, accountIndex int64, filter string, cursor string) (*api.WithdrawHistory, error)

	// GetTransferHistory retrieves transfer history
	GetTransferHistory(ctx context.Context, accountIndex int64, cursor string) (*api.TransferHistory, error)

	// GetTransferFeeInfo retrieves transfer fee information
	GetTransferFeeInfo(ctx context.Context, accountIndex int64, toAccountIndex *int64, auth string) (*api.TransferFeeInfo, error)

	// GetWithdrawalDelay retrieves current withdrawal delay
	GetWithdrawalDelay(ctx context.Context) (*api.RespWithdrawalDelay, error)
}

// CandlestickAPI provides access to market data endpoints
type CandlestickAPI interface {
	// GetCandlesticks retrieves OHLCV data
	GetCandlesticks(ctx context.Context, marketID int16, resolution api.CandlestickResolution, timestamps api.TimestampRange, countBack int) (*api.Candlesticks, error)

	// GetFundings retrieves funding data
	GetFundings(ctx context.Context, marketID int16, resolution api.FundingResolution, timestamps api.TimestampRange, countBack int) (*api.Fundings, error)

	// GetFundingRates retrieves current funding rates
	GetFundingRates(ctx context.Context) (*api.FundingRates, error)
}

// BlockAPI provides access to block-related endpoints
type BlockAPI interface {
	// GetBlock retrieves a block by height or commitment
	GetBlock(ctx context.Context, by api.QueryBy, value string) (*api.Blocks, error)

	// GetBlocks retrieves blocks with pagination
	GetBlocks(ctx context.Context, index *int64, limit int, sort string) (*api.Blocks, error)

	// GetBlockTxs retrieves transactions for a block
	GetBlockTxs(ctx context.Context, by api.QueryBy, value string) (*api.Txs, error)

	// GetCurrentHeight retrieves the current block height
	GetCurrentHeight(ctx context.Context) (*api.CurrentHeight, error)
}

// BridgeAPI provides access to bridge-related endpoints
type BridgeAPI interface {
	// GetFastBridgeInfo retrieves fast bridge information
	GetFastBridgeInfo(ctx context.Context) (*api.RespGetFastBridgeInfo, error)

	// CreateIntentAddress creates an intent address for external deposits
	CreateIntentAddress(ctx context.Context, chainID int64, fromAddr string, amount string, isExternalDeposit bool) (*api.RespCreateIntentAddress, error)

	// GetFastWithdrawInfo retrieves fast withdraw pool information
	GetFastWithdrawInfo(ctx context.Context, accountIndex int64, auth string) (*api.FastWithdrawPoolInfo, error)

	// SubmitFastWithdraw submits a fast withdraw transaction
	SubmitFastWithdraw(ctx context.Context, txInfo string, toAddress string, auth string) (*api.RespFastWithdraw, error)
}

// InfoAPI provides access to general information endpoints
type InfoAPI interface {
	// GetStatus retrieves service status
	GetStatus(ctx context.Context) (*api.Status, error)

	// GetInfo retrieves exchange information
	GetInfo(ctx context.Context) (*api.ZkLighterInfo, error)

	// GetAnnouncements retrieves exchange announcements
	GetAnnouncements(ctx context.Context) (*api.Announcements, error)

	// Export exports account data
	Export(ctx context.Context, accountIndex int64, marketID int16, exportType api.ExportType) (*api.ExportData, error)
}
