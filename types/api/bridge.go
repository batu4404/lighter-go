package api

// Bridge represents a bridge transaction
type Bridge struct {
	BridgeIndex  int64  `json:"bridge_index"`
	L1Address    string `json:"l1_address"`
	Direction    string `json:"direction"` // "deposit" or "withdraw"
	AssetIndex   int16  `json:"asset_index"`
	AssetSymbol  string `json:"asset_symbol,omitempty"`
	Amount       string `json:"amount"`
	Fee          string `json:"fee,omitempty"`
	L1TxHash     string `json:"l1_tx_hash,omitempty"`
	L2TxHash     string `json:"l2_tx_hash,omitempty"`
	Status       string `json:"status"` // "pending", "confirmed", "failed"
	IsFastBridge bool   `json:"is_fast_bridge"`
	CreatedAt    int64  `json:"created_at"`
	ConfirmedAt  int64  `json:"confirmed_at,omitempty"`
}

// RespGetFastBridgeInfo is the response for fast bridge info
type RespGetFastBridgeInfo struct {
	BaseResponse
	Enabled            bool   `json:"enabled"`
	MinAmount          string `json:"min_amount"`
	MaxAmount          string `json:"max_amount"`
	Fee                string `json:"fee"`
	FeeRate            string `json:"fee_rate"`
	AvailableLiquidity string `json:"available_liquidity"`
	EstimatedTime      int64  `json:"estimated_time"` // In seconds
}

// BridgeSupportedNetwork represents a supported bridge network
type BridgeSupportedNetwork struct {
	ChainId          int64  `json:"chain_id"`
	Name             string `json:"name"`
	RpcUrl           string `json:"rpc_url,omitempty"`
	ExplorerUrl      string `json:"explorer_url,omitempty"`
	ContractAddress  string `json:"contract_address"`
	IsMainnet        bool   `json:"is_mainnet"`
	MinConfirmations int    `json:"min_confirmations"`
}

// BridgeNetworks is the response for supported networks query
type BridgeNetworks struct {
	BaseResponse
	Networks []BridgeSupportedNetwork `json:"networks"`
}

// RespCreateIntentAddress is the response for creating an intent address
type RespCreateIntentAddress struct {
	BaseResponse
	IntentAddress string `json:"intent_address"`
}

// FastWithdrawPoolInfo is the response for fast withdraw pool info
type FastWithdrawPoolInfo struct {
	BaseResponse
	ToAccountIndex int64  `json:"to_account_index"`
	WithdrawLimit  string `json:"withdraw_limit"`
}

// RespFastWithdraw is the response for fast withdraw submission
type RespFastWithdraw struct {
	BaseResponse
	TxHash string `json:"tx_hash"`
}
