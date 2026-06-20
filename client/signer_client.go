// Package client provides the core trading client for the Lighter API.
//
// SignerClient is the main entry point for trading operations. It provides
// high-level methods for creating and submitting orders, managing positions,
// and interacting with the Lighter exchange.
//
// Example:
//
//	// Create a client (recommended - network handles chain ID automatically)
//	httpClient := http.NewFullClientForNetwork(client.Mainnet)
//	client, err := client.NewSignerClientForNetwork(httpClient, client.Mainnet, privateKey, 0, accountIndex, nil)
//
//	// Or create with explicit URL and chain ID
//	httpClient := http.NewFullClient("https://mainnet.zklighter.elliot.ai")
//	client, err := client.NewSignerClient(httpClient, privateKey, 304, 0, accountIndex, nil)
//
//	// Create a market buy order
//	txInfo, err := client.CreateMarketOrder(0, 100000, true, nil) // Buy 0.01 ETH
//
//	// Submit the order
//	resp, err := client.SendAndSubmit(txInfo)
//
//	// Create a limit order
//	txInfo, err := client.CreateLimitOrder(0, 100000, 350000, true, expiry, nil)
//
//	// Create stop loss / take profit orders
//	txInfo, err := client.CreateStopLossOrder(0, 100000, 340000, false, expiry, nil)
//	txInfo, err := client.CreateTakeProfitOrder(0, 100000, 360000, false, expiry, nil)
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/0xJord4n/lighter-go/nonce"
	"github.com/0xJord4n/lighter-go/signer"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/0xJord4n/lighter-go/types/api"
	"github.com/0xJord4n/lighter-go/types/txtypes"
)

// SignerClient extends TxClient with convenience methods for common trading patterns.
// It provides higher-level APIs similar to the Python SDK's SignerClient.
type SignerClient struct {
	*TxClient
	fullHTTP     FullHTTPClient
	nonceManager nonce.Manager
}

// NewSignerClient creates a SignerClient with full HTTP capabilities.
// If nonceManager is nil, a new OptimisticNonceManager will be created.
func NewSignerClient(httpClient FullHTTPClient, privateKey string, chainId uint32, apiKeyIndex uint8, accountIndex int64, nonceManager nonce.Manager) (*SignerClient, error) {
	txClient, err := createTxClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex)
	if err != nil {
		return nil, err
	}

	if nonceManager == nil {
		nonceManager = nonce.NewOptimisticManager(httpClient)
	}

	return &SignerClient{
		TxClient:     txClient,
		fullHTTP:     httpClient,
		nonceManager: nonceManager,
	}, nil
}

// NewSignerClientForNetwork creates a SignerClient using the chain ID from the specified network.
// The httpClient should be created using http.NewFullClientForNetwork(network) or http.NewFullClient(network.APIURL()).
//
// Example:
//
//	httpClient := http.NewFullClientForNetwork(client.Mainnet)
//	client, err := client.NewSignerClientForNetwork(httpClient, client.Mainnet, privateKey, 0, accountIndex, nil)
func NewSignerClientForNetwork(httpClient FullHTTPClient, network Network, privateKey string, apiKeyIndex uint8, accountIndex int64, nonceManager nonce.Manager) (*SignerClient, error) {
	return NewSignerClient(httpClient, privateKey, network.ChainID(), apiKeyIndex, accountIndex, nonceManager)
}

// createTxClient is a helper that creates a TxClient without registering it globally
func createTxClient(httpClient MinimalHTTPClient, privateKey string, chainId uint32, apiKeyIndex uint8, accountIndex int64) (*TxClient, error) {
	// Use the existing CreateClient function but we need to get the client back
	return CreateClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex)
}

// FullHTTP returns the full HTTP client for direct API access
func (c *SignerClient) FullHTTP() FullHTTPClient {
	return c.fullHTTP
}

// NonceManager returns the nonce manager
func (c *SignerClient) NonceManager() nonce.Manager {
	return c.nonceManager
}

// CreateMarketOrder creates a market order with minimal parameters
func (c *SignerClient) CreateMarketOrder(ctx context.Context, marketIndex int16, size int64, isBuy bool, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	// For market orders, use max/min price depending on side
	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0, // Auto-generate
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.MarketOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      0,
	}

	return c.GetCreateOrderTransaction(ctx, req, opts)
}

// CreateMarketOrderWithSlippage creates a market order with slippage protection
func (c *SignerClient) CreateMarketOrderWithSlippage(ctx context.Context, marketIndex int16, size int64, isBuy bool, slippageBps int, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	// Fetch current market price from order book details
	orderBookDetails, err := c.fullHTTP.Order().GetOrderBookDetails(ctx, &marketIndex, api.MarketFilterAll)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	// Find the matching perps order book
	var referencePrice string
	for _, ob := range orderBookDetails.PerpsOrderBooks {
		if ob.MarketIndex == marketIndex {
			if isBuy && len(ob.Asks) > 0 {
				referencePrice = ob.Asks[0].Price
			} else if !isBuy && len(ob.Bids) > 0 {
				referencePrice = ob.Bids[0].Price
			}
			break
		}
	}

	if referencePrice == "" {
		return nil, fmt.Errorf("no liquidity in order book for market %d", marketIndex)
	}

	price, err := calculateSlippagePrice(referencePrice, slippageBps, isBuy)
	if err != nil {
		return nil, err
	}

	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.LimitOrder, // Use limit with IOC for slippage protection
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      0,
	}

	return c.GetCreateOrderTransaction(ctx, req, opts)
}

// CreateLimitOrder creates a limit order
func (c *SignerClient) CreateLimitOrder(ctx context.Context, marketIndex int16, size int64, price uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.LimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(ctx, req, opts)
}

// CreateTakeProfitOrder creates a take-profit market order
func (c *SignerClient) CreateTakeProfitOrder(ctx context.Context, marketIndex int16, size int64, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	// For TP market orders, use limit price that will fill immediately once triggered
	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.TakeProfitOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(ctx, req, opts)
}

// CreateTakeProfitLimitOrder creates a take-profit limit order
func (c *SignerClient) CreateTakeProfitLimitOrder(ctx context.Context, marketIndex int16, size int64, price uint32, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.TakeProfitLimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(ctx, req, opts)
}

// CreateStopLossOrder creates a stop-loss market order
func (c *SignerClient) CreateStopLossOrder(ctx context.Context, marketIndex int16, size int64, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.StopLossOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(ctx, req, opts)
}

// CreateStopLossLimitOrder creates a stop-loss limit order
func (c *SignerClient) CreateStopLossLimitOrder(ctx context.Context, marketIndex int16, size int64, price uint32, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.StopLossLimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(ctx, req, opts)
}

// TriggerOrderParams contains parameters for TP/SL trigger orders
type TriggerOrderParams struct {
	TriggerPrice uint32 // Price that triggers the order (6 decimals)
	Expiry       int64  // Order expiry in milliseconds
}

// CreateMarketOrderWithTPSL creates a market order with optional take-profit and/or stop-loss
// as a grouped order. This ensures atomic execution with a single nonce.
//
// Parameters:
//   - marketIndex: The market to trade on
//   - size: Order size in base units (8 decimals)
//   - isBuy: true for buy/long, false for sell/short
//   - takeProfit: Take-profit parameters (nil to skip)
//   - stopLoss: Stop-loss parameters (nil to skip)
//   - opts: Transaction options (nonce, expiry, etc.)
//
// Grouping logic:
//   - Both TP and SL: OTOCO (GroupingType 3) - parent triggers both, children are OCO
//   - Only TP or SL: OTO (GroupingType 1) - parent triggers child
//   - Neither: Returns error (use CreateMarketOrder instead)
func (c *SignerClient) CreateMarketOrderWithTPSL(
	ctx context.Context,
	marketIndex int16,
	size int64,
	isBuy bool,
	takeProfit *TriggerOrderParams,
	stopLoss *TriggerOrderParams,
	opts *types.TransactOpts,
) (*txtypes.L2CreateGroupedOrdersTxInfo, error) {
	if takeProfit == nil && stopLoss == nil {
		return nil, fmt.Errorf("at least one of takeProfit or stopLoss must be provided; use CreateMarketOrder for orders without TP/SL")
	}

	// Determine grouping type
	var groupingType uint8
	if takeProfit != nil && stopLoss != nil {
		groupingType = txtypes.GroupingType_OneTriggersAOneCancelsTheOther // OTOCO
	} else {
		groupingType = txtypes.GroupingType_OneTriggersTheOther // OTO
	}

	// Parent order direction
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	// Parent market order price (will fill at market)
	var parentPrice uint32
	if isBuy {
		parentPrice = txtypes.MaxOrderPrice
	} else {
		parentPrice = txtypes.MinOrderPrice
	}

	// Child orders are opposite direction (to close the position)
	childIsAsk := uint8(1)
	if !isBuy {
		childIsAsk = 0
	}

	// Child order price (will fill at market once triggered)
	var childPrice uint32
	if childIsAsk == 1 {
		childPrice = txtypes.MinOrderPrice // Selling, use min price
	} else {
		childPrice = txtypes.MaxOrderPrice // Buying, use max price
	}

	// Build orders slice
	orders := []*types.CreateOrderTxReq{
		// Order 0: Parent market order
		{
			MarketIndex:      marketIndex,
			ClientOrderIndex: 0,
			BaseAmount:       size,
			Price:            parentPrice,
			IsAsk:            isAsk,
			Type:             txtypes.MarketOrder,
			TimeInForce:      txtypes.ImmediateOrCancel,
			ReduceOnly:       0,
			TriggerPrice:     0,
			OrderExpiry:      0,
		},
	}

	// Add TP order if provided
	if takeProfit != nil {
		orders = append(orders, &types.CreateOrderTxReq{
			MarketIndex:      marketIndex,
			ClientOrderIndex: 0,
			BaseAmount:       0, // 0 = use position size
			Price:            childPrice,
			IsAsk:            childIsAsk,
			Type:             txtypes.TakeProfitOrder,
			TimeInForce:      txtypes.ImmediateOrCancel,
			ReduceOnly:       1,
			TriggerPrice:     takeProfit.TriggerPrice,
			OrderExpiry:      takeProfit.Expiry,
		})
	}

	// Add SL order if provided
	if stopLoss != nil {
		orders = append(orders, &types.CreateOrderTxReq{
			MarketIndex:      marketIndex,
			ClientOrderIndex: 0,
			BaseAmount:       0, // 0 = use position size
			Price:            childPrice,
			IsAsk:            childIsAsk,
			Type:             txtypes.StopLossOrder,
			TimeInForce:      txtypes.ImmediateOrCancel,
			ReduceOnly:       1,
			TriggerPrice:     stopLoss.TriggerPrice,
			OrderExpiry:      stopLoss.Expiry,
		})
	}

	req := &types.CreateGroupedOrdersTxReq{
		GroupingType: groupingType,
		Orders:       orders,
	}

	return c.GetCreateGroupedOrdersTransaction(ctx, req, opts)
}

// SendAndSubmit signs a transaction and submits it to the API
func (c *SignerClient) SendAndSubmit(ctx context.Context, txInfo txtypes.TxInfo) (*api.RespSendTx, error) {
	txInfoJSON, err := txInfo.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tx info: %w", err)
	}

	// Use SendTxWithIndices to pass account_index and api_key_index (matching TS SDK)
	accountIndex := c.GetAccountIndex()
	apiKeyIndex := c.GetApiKeyIndex()
	resp, err := c.fullHTTP.Transaction().SendTxWithIndices(
		ctx,
		txInfo.GetTxType(),
		txInfoJSON,
		nil,
		&accountIndex,
		&apiKeyIndex,
		"",
	)
	if err != nil {
		// Acknowledge failure for nonce recovery
		if optManager, ok := c.nonceManager.(*nonce.OptimisticManager); ok {
			// Extract nonce from tx info if possible
			optManager.AcknowledgeFailure(c.GetAccountIndex(), c.GetApiKeyIndex(), -1)
		}
		return nil, err
	}

	return resp, nil
}

// SendTxBatch submits multiple transactions with auth token
func (c *SignerClient) SendTxBatch(ctx context.Context, txInfos []txtypes.TxInfo) (*api.RespSendTxBatch, error) {
	txTypes := make([]uint8, len(txInfos))
	txInfoJSONs := make([]string, len(txInfos))

	for i, tx := range txInfos {
		txTypes[i] = tx.GetTxType()
		jsonStr, err := tx.GetTxInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize tx %d: %w", i, err)
		}
		txInfoJSONs[i] = jsonStr
	}

	// Generate auth token for the request
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	return c.fullHTTP.Transaction().SendTxBatchWithAuth(ctx, txTypes, txInfoJSONs, authToken)
}

// Transfer creates, signs (L1 + L2), and submits a transfer transaction.
// This is a convenience method that handles L1 signing internally.
// ethPrivateKey is the Ethereum private key (hex-encoded) for L1 signature.
func (c *SignerClient) Transfer(ctx context.Context, ethPrivateKey string, req *types.TransferTxReq, opts *types.TransactOpts) (*api.RespSendTx, error) {
	// Create L1 signer
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}

	// Create the transfer transaction (L2 signed)
	txInfo, err := c.GetTransferTransaction(ctx, req, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer transaction: %w", err)
	}

	// Sign with Ethereum key (L1 signature)
	l1Sig, err := l1Signer.Sign(txInfo.GetL1SignatureBody(c.GetChainId()))
	if err != nil {
		return nil, fmt.Errorf("failed to sign L1 message: %w", err)
	}
	txInfo.SetL1Sig(l1Sig)

	// Submit
	return c.SendAndSubmit(ctx, txInfo)
}

// TransferSameMasterAccount creates, L2-signs, and submits a transfer between
// two accounts that share the same master account. Unlike Transfer, no L1
// (Ethereum) signature is required: the shared master account already
// authorizes the move, so there is no cross-owner approval to sign. Mirrors
// lighter-python's sign_transfer_same_master_account, which submits the same
// SignTransfer output without an eth key.
func (c *SignerClient) TransferSameMasterAccount(ctx context.Context, req *types.TransferTxReq, opts *types.TransactOpts) (*api.RespSendTx, error) {
	txInfo, err := c.GetTransferTransaction(ctx, req, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer transaction: %w", err)
	}
	return c.SendAndSubmit(ctx, txInfo)
}

// FastWithdraw performs a fast withdrawal from Lighter L2 to an L1 address.
// This transfers USDC to a pool account which then sends funds to the destination L1 address.
// ethPrivateKey is the Ethereum private key (hex-encoded) for L1 signature.
// toL1Address is the destination Ethereum address on L1.
// amount is the USDC amount in base units (6 decimals, e.g. 5_000_000 for 5 USDC).
func (c *SignerClient) FastWithdraw(ctx context.Context, ethPrivateKey string, toL1Address string, amount int64) (*api.RespFastWithdraw, error) {
	// Get auth token
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Get fast withdraw pool info
	poolInfo, err := c.fullHTTP.Bridge().GetFastWithdrawInfo(ctx, c.GetAccountIndex(), authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get fast withdraw pool info: %w", err)
	}

	// Get transfer fee (requires auth)
	feeInfo, err := c.fullHTTP.Transaction().GetTransferFeeInfo(ctx, c.GetAccountIndex(), &poolInfo.ToAccountIndex, authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer fee: %w", err)
	}

	// Build memo: 20-byte address + 12 zero bytes
	addrHex := toL1Address
	if len(addrHex) >= 2 && addrHex[:2] == "0x" {
		addrHex = addrHex[2:]
	}
	if len(addrHex) != 40 {
		return nil, fmt.Errorf("invalid L1 address length: expected 40 hex chars, got %d", len(addrHex))
	}

	var memo [32]byte
	addrBytes := make([]byte, 20)
	for i := 0; i < 20; i++ {
		var b byte
		_, err := fmt.Sscanf(addrHex[i*2:i*2+2], "%02x", &b)
		if err != nil {
			return nil, fmt.Errorf("failed to parse address byte: %w", err)
		}
		addrBytes[i] = b
	}
	copy(memo[:20], addrBytes)
	// Last 12 bytes are already zero

	// Create L1 signer
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}

	// Create transfer request
	req := &types.TransferTxReq{
		ToAccountIndex: poolInfo.ToAccountIndex,
		AssetIndex:     3, // USDC
		FromRouteType:  0, // Perps
		ToRouteType:    0, // Perps
		Amount:         amount,
		USDCFee:        feeInfo.TransferFeeUsdc,
		Memo:           memo,
	}

	// Create the transfer transaction (L2 signed)
	txInfo, err := c.GetTransferTransaction(ctx, req, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer transaction: %w", err)
	}

	// Sign with Ethereum key (L1 signature)
	l1Sig, err := l1Signer.Sign(txInfo.GetL1SignatureBody(c.GetChainId()))
	if err != nil {
		return nil, fmt.Errorf("failed to sign L1 message: %w", err)
	}
	txInfo.SetL1Sig(l1Sig)

	// Get tx info JSON
	txInfoJSON, err := txInfo.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tx info: %w", err)
	}

	// Submit to fast withdraw endpoint
	return c.fullHTTP.Bridge().SubmitFastWithdraw(ctx, txInfoJSON, toL1Address, authToken)
}

// ChangePubKey creates, signs (L1 + L2), and submits a change pub key transaction.
// This is a convenience method that handles L1 signing internally.
// ethPrivateKey is the Ethereum private key (hex-encoded) for L1 signature.
func (c *SignerClient) ChangePubKey(ctx context.Context, ethPrivateKey string, req *types.ChangePubKeyReq, opts *types.TransactOpts) (*api.RespSendTx, error) {
	// Create L1 signer
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}

	// Create the change pub key transaction (L2 signed)
	txInfo, err := c.GetChangePubKeyTransaction(ctx, req, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create change pub key transaction: %w", err)
	}

	// Sign with Ethereum key (L1 signature)
	l1Sig, err := l1Signer.Sign(txInfo.GetL1SignatureBody())
	if err != nil {
		return nil, fmt.Errorf("failed to sign L1 message: %w", err)
	}
	txInfo.SetL1Sig(l1Sig)

	// Submit
	txInfoJSON, err := txInfo.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tx info: %w", err)
	}
	return c.fullHTTP.Transaction().SendTx(ctx, txInfo.GetTxType(), txInfoJSON, nil)
}

// ApproveIntegrator signs and submits an ApproveIntegrator transaction, authorizing
// the given integrator account to charge taker/maker fees (up to the supplied maxes)
// on this account's orders. Like ChangePubKey, it requires both the L2 (API key) and
// L1 (Ethereum) signatures: GetApproveIntegratorTx produces the Schnorr L2 signature,
// and the ethPrivateKey produces the EIP-191 L1 signature over the tx's L1 signature body.
func (c *SignerClient) ApproveIntegrator(ctx context.Context, ethPrivateKey string, req *types.ApproveIntegratorTxReq, opts *types.TransactOpts) (*api.RespSendTx, error) {
	// Create L1 signer
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}

	// Create the approve integrator transaction (L2 signed)
	txInfo, err := c.GetApproveIntegratorTx(ctx, req, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create approve integrator transaction: %w", err)
	}

	// Sign with Ethereum key (L1 signature)
	l1Sig, err := l1Signer.Sign(txInfo.GetL1SignatureBody(c.GetChainId()))
	if err != nil {
		return nil, fmt.Errorf("failed to sign L1 message: %w", err)
	}
	txInfo.L1Sig = l1Sig

	// Submit
	txInfoJSON, err := txInfo.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tx info: %w", err)
	}
	return c.fullHTTP.Transaction().SendTx(ctx, txInfo.GetTxType(), txInfoJSON, nil)
}

// ApproveIntegratorSameMasterAccount creates, L2-signs, and submits an
// approve-integrator transaction for an integrator account that shares the
// same master account. Unlike ApproveIntegrator, no L1 (Ethereum) signature is
// required: the shared master account already authorizes the approval. Mirrors
// lighter-python's sign_approve_integrator_same_master_account, which submits
// the same SignApproveIntegrator output without an eth key.
func (c *SignerClient) ApproveIntegratorSameMasterAccount(ctx context.Context, req *types.ApproveIntegratorTxReq, opts *types.TransactOpts) (*api.RespSendTx, error) {
	txInfo, err := c.GetApproveIntegratorTx(ctx, req, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create approve integrator transaction: %w", err)
	}

	txInfoJSON, err := txInfo.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tx info: %w", err)
	}
	return c.fullHTTP.Transaction().SendTx(ctx, txInfo.GetTxType(), txInfoJSON, nil)
}

// GetOpenOrders retrieves open orders for the account
func (c *SignerClient) GetOpenOrders(ctx context.Context, marketID *int16) (*api.Orders, error) {
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, err
	}
	return c.fullHTTP.Order().GetActiveOrders(ctx, c.GetAccountIndex(), marketID, authToken)
}

// CancelAllOrders cancels all open orders
func (c *SignerClient) CancelAllOrders(ctx context.Context, opts *types.TransactOpts) (*txtypes.L2CancelAllOrdersTxInfo, error) {
	req := &types.CancelAllOrdersTxReq{
		TimeInForce: txtypes.ImmediateCancelAll,
		Time:        0,
	}
	return c.GetCancelAllOrdersTransaction(ctx, req, opts)
}

// GetPositions retrieves current positions
func (c *SignerClient) GetPositions(ctx context.Context) (*api.DetailedAccounts, error) {
	return c.fullHTTP.Account().GetAccount(ctx, api.QueryByIndex, fmt.Sprintf("%d", c.GetAccountIndex()))
}

// Helper to get auth token
func (c *SignerClient) getAuthToken() (string, error) {
	deadline := time.Now().Add(8 * time.Hour)
	authInfo, err := c.GetAuthToken(deadline)
	if err != nil {
		return "", err
	}

	// The auth token is returned as a string directly
	return authInfo, nil
}

// calculateSlippagePrice calculates price with slippage
func calculateSlippagePrice(priceStr string, slippageBps int, isBuy bool) (uint32, error) {
	// Parse price as integer (prices are typically scaled integers)
	var price int64
	if _, err := fmt.Sscanf(priceStr, "%d", &price); err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	// Calculate slippage adjustment
	adjustment := (price * int64(slippageBps)) / 10000

	if isBuy {
		price += adjustment
		if price > int64(txtypes.MaxOrderPrice) {
			price = int64(txtypes.MaxOrderPrice)
		}
	} else {
		price -= adjustment
		if price < int64(txtypes.MinOrderPrice) {
			price = int64(txtypes.MinOrderPrice)
		}
	}

	return uint32(price), nil
}
