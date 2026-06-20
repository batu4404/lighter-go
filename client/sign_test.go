package client

import (
	"context"
	"testing"

	"github.com/0xJord4n/lighter-go/types"
	"github.com/0xJord4n/lighter-go/types/txtypes"
)

const (
	testChainID      uint32 = 304
	testAccountIndex int64  = 1
	testAPIKeyIndex  uint8  = 0
	testNonce        int64  = 42
)

func newTestClient(t *testing.T, privateKey string) *TxClient {
	t.Helper()
	c, err := NewTxClient(nil, privateKey, testAccountIndex, testAPIKeyIndex, testChainID)
	if err != nil {
		t.Fatalf("NewTxClient failed: %v", err)
	}
	return c
}

func opsWithSkipNonce(skipNonce uint8, nonce int64) *types.TransactOpts {
	attr := &types.L2TxAttributes{}
	if skipNonce == 1 {
		sn := skipNonce
		attr.SkipNonce = &sn
	}
	n := nonce
	return &types.TransactOpts{
		Nonce:        &n,
		TxAttributes: attr,
	}
}

func assertSkipNonce(t *testing.T, name string, attrs txtypes.L2TxAttributes, expectedSet bool) {
	t.Helper()
	if expectedSet {
		if attrs == nil {
			t.Errorf("%s: attributes map should be non-nil when skipNonce=1", name)
			return
		}
		att, ok := attrs[txtypes.AttributeTypeSkipTxNonce]
		if !ok {
			t.Errorf("%s: SkipTxNonce key missing", name)
			return
		}
		if att != 1 {
			t.Errorf("%s: SkipTxNonce = %d, want 1", name, att)
		}
		return
	}
	if attrs != nil {
		t.Errorf("%s: attributes map should be nil when skipNonce=0, got %v", name, attrs)
	}
}

func TestGenerateAPIKey(t *testing.T) {
	priv, pub, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	if len(priv) < 3 || priv[:2] != "0x" {
		t.Errorf("privateKey should be a 0x-prefixed hex string, got %q", priv)
	}
	if len(pub) < 3 || pub[:2] != "0x" {
		t.Errorf("publicKey should be a 0x-prefixed hex string, got %q", pub)
	}
}

func TestCreateClient(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)
	if c.GetChainId() != testChainID {
		t.Errorf("chainId = %d, want %d", c.GetChainId(), testChainID)
	}
	if c.GetAccountIndex() != testAccountIndex {
		t.Errorf("accountIndex = %d, want %d", c.GetAccountIndex(), testAccountIndex)
	}
	if c.GetApiKeyIndex() != testAPIKeyIndex {
		t.Errorf("apiKeyIndex = %d, want %d", c.GetApiKeyIndex(), testAPIKeyIndex)
	}
}

func TestSignCancelOrder(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)
	req := &types.CancelOrderTxReq{MarketIndex: 0, Index: 12345}

	// skipNonce = 1
	tx1, err := c.GetCancelOrderTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCancelOrderTransaction (skipNonce=1) failed: %v", err)
	}
	if tx1.GetTxType() != txtypes.TxTypeL2CancelOrder {
		t.Errorf("txType = %d, want %d", tx1.GetTxType(), txtypes.TxTypeL2CancelOrder)
	}
	if tx1.AccountIndex != testAccountIndex {
		t.Errorf("AccountIndex = %d, want %d", tx1.AccountIndex, testAccountIndex)
	}
	if tx1.ApiKeyIndex != testAPIKeyIndex {
		t.Errorf("ApiKeyIndex = %d, want %d", tx1.ApiKeyIndex, testAPIKeyIndex)
	}
	if tx1.MarketIndex != 0 {
		t.Errorf("MarketIndex = %d, want 0", tx1.MarketIndex)
	}
	if tx1.Index != 12345 {
		t.Errorf("Index = %d, want 12345", tx1.Index)
	}
	if tx1.Nonce != testNonce {
		t.Errorf("Nonce = %d, want %d", tx1.Nonce, testNonce)
	}
	if len(tx1.Sig) == 0 {
		t.Error("Sig should not be empty")
	}
	assertSkipNonce(t, "SignCancelOrder(skipNonce=1)", tx1.L2TxAttributes, true)

	tx2, err := c.GetCancelOrderTransaction(context.Background(), req, opsWithSkipNonce(0, testNonce))
	if err != nil {
		t.Fatalf("GetCancelOrderTransaction (skipNonce=0) failed: %v", err)
	}
	assertSkipNonce(t, "SignCancelOrder(skipNonce=0)", tx2.L2TxAttributes, false)

	if tx1.SignedHash == tx2.SignedHash {
		t.Error("skipNonce flag should affect the signed tx hash, but hashes are identical")
	}
}

func TestSignCancelAllOrders(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)
	req := &types.CancelAllOrdersTxReq{TimeInForce: 0, Time: 0}

	tx, err := c.GetCancelAllOrdersTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCancelAllOrdersTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CancelAllOrders {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CancelAllOrders)
	}
	if tx.TimeInForce != 0 {
		t.Errorf("TimeInForce = %d, want 0", tx.TimeInForce)
	}
	if tx.Time != 0 {
		t.Errorf("Time = %d, want 0", tx.Time)
	}
	assertSkipNonce(t, "SignCancelAllOrders(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignCreateOrder(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)

	req := &types.CreateOrderTxReq{
		MarketIndex:      0,
		ClientOrderIndex: 1,
		BaseAmount:       1000,
		Price:            50000,
		IsAsk:            0,
		Type:             0,
		TimeInForce:      0,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      0,
	}
	tx, err := c.GetCreateOrderTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCreateOrderTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CreateOrder {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CreateOrder)
	}
	if tx.MarketIndex != 0 {
		t.Errorf("MarketIndex = %d, want 0", tx.MarketIndex)
	}
	if tx.ClientOrderIndex != 1 {
		t.Errorf("ClientOrderIndex = %d, want 1", tx.ClientOrderIndex)
	}
	if tx.BaseAmount != 1000 {
		t.Errorf("BaseAmount = %d, want 1000", tx.BaseAmount)
	}
	if tx.Price != 50000 {
		t.Errorf("Price = %d, want 50000", tx.Price)
	}
	assertSkipNonce(t, "SignCreateOrder(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignCreateSubAccount(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)

	tx, err := c.GetCreateSubAccountTransaction(context.Background(), opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCreateSubAccountTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CreateSubAccount {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CreateSubAccount)
	}
	if tx.AccountIndex != testAccountIndex {
		t.Errorf("AccountIndex = %d, want %d", tx.AccountIndex, testAccountIndex)
	}
	if tx.Nonce != testNonce {
		t.Errorf("Nonce = %d, want %d", tx.Nonce, testNonce)
	}
	assertSkipNonce(t, "SignCreateSubAccount(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignUpdateLeverage(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)

	req := &types.UpdateLeverageTxReq{
		MarketIndex:           0,
		InitialMarginFraction: 100,
		MarginMode:            0,
	}
	tx, err := c.GetUpdateLeverageTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetUpdateLeverageTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2UpdateLeverage {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2UpdateLeverage)
	}
	if tx.MarketIndex != 0 {
		t.Errorf("MarketIndex = %d, want 0", tx.MarketIndex)
	}
	if tx.InitialMarginFraction != 100 {
		t.Errorf("InitialMarginFraction = %d, want 100", tx.InitialMarginFraction)
	}
	assertSkipNonce(t, "SignUpdateLeverage(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignCreateGroupedOrders(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)

	const expiry int64 = 7 * 24 * 60 * 60 * 1000

	req := &types.CreateGroupedOrdersTxReq{
		GroupingType: 1,
		Orders: []*types.CreateOrderTxReq{
			{
				MarketIndex: 0, ClientOrderIndex: 0, BaseAmount: 1000, Price: 50000,
				IsAsk: 0, Type: 0, TimeInForce: 1, ReduceOnly: 0, TriggerPrice: 0, OrderExpiry: expiry,
			},
			{
				MarketIndex: 0, ClientOrderIndex: 0, BaseAmount: 0, Price: 51000,
				IsAsk: 1, Type: 4, TimeInForce: 0, ReduceOnly: 1, TriggerPrice: 49000, OrderExpiry: expiry,
			},
		},
	}

	tx, err := c.GetCreateGroupedOrdersTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCreateGroupedOrdersTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CreateGroupedOrders {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CreateGroupedOrders)
	}
	if tx.GroupingType != 1 {
		t.Errorf("GroupingType = %d, want 1", tx.GroupingType)
	}
	if len(tx.Orders) != 2 {
		t.Fatalf("Orders len = %d, want 2", len(tx.Orders))
	}
	if tx.Orders[0].TimeInForce != 1 {
		t.Errorf("Orders[0].TimeInForce = %d, want 1", tx.Orders[0].TimeInForce)
	}
	if tx.Orders[1].Type != 4 {
		t.Errorf("Orders[1].Type = %d, want 4", tx.Orders[1].Type)
	}
	if tx.Orders[1].TriggerPrice != 49000 {
		t.Errorf("Orders[1].TriggerPrice = %d, want 49000", tx.Orders[1].TriggerPrice)
	}
	assertSkipNonce(t, "SignCreateGroupedOrders(skipNonce=1)", tx.L2TxAttributes, true)
}

// minSubAccountIndex (1<<47) is the lowest valid public-pool / staking-pool index.
const minSubAccountIndex int64 = 140737488355328

// newSignTestClient is a convenience wrapper that creates a client with a fresh key.
func newSignTestClient(t *testing.T) *TxClient {
	t.Helper()
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	return newTestClient(t, priv)
}

func TestSignWithdraw(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.WithdrawTxReq{AssetIndex: 3, RouteType: 0, Amount: 5_000_000}

	tx, err := c.GetWithdrawTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetWithdrawTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2Withdraw {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2Withdraw)
	}
	if tx.Amount != 5_000_000 {
		t.Errorf("Amount = %d, want 5000000", tx.Amount)
	}
	if len(tx.Sig) == 0 {
		t.Error("Sig should not be empty")
	}
	assertSkipNonce(t, "SignWithdraw(skipNonce=1)", tx.L2TxAttributes, true)

	tx0, err := c.GetWithdrawTransaction(context.Background(), req, opsWithSkipNonce(0, testNonce))
	if err != nil {
		t.Fatalf("GetWithdrawTransaction (skipNonce=0) failed: %v", err)
	}
	if tx.SignedHash == tx0.SignedHash {
		t.Error("skipNonce flag should change the signed hash")
	}
}

func TestSignModifyOrder(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.ModifyOrderTxReq{MarketIndex: 0, Index: 12345, BaseAmount: 2000, Price: 51000, TriggerPrice: 0}

	tx, err := c.GetModifyOrderTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetModifyOrderTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2ModifyOrder {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2ModifyOrder)
	}
	if tx.Index != 12345 {
		t.Errorf("Index = %d, want 12345", tx.Index)
	}
	if tx.BaseAmount != 2000 {
		t.Errorf("BaseAmount = %d, want 2000", tx.BaseAmount)
	}
	assertSkipNonce(t, "SignModifyOrder(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignTransfer(t *testing.T) {
	c := newSignTestClient(t)
	var memo [32]byte
	req := &types.TransferTxReq{
		ToAccountIndex: 2, AssetIndex: 3, FromRouteType: 0, ToRouteType: 0,
		Amount: 1_000_000, USDCFee: 100, Memo: memo,
	}

	tx, err := c.GetTransferTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetTransferTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2Transfer {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2Transfer)
	}
	if tx.ToAccountIndex != 2 {
		t.Errorf("ToAccountIndex = %d, want 2", tx.ToAccountIndex)
	}
	if tx.Amount != 1_000_000 {
		t.Errorf("Amount = %d, want 1000000", tx.Amount)
	}
	// L1 signature body should be deterministic and non-empty.
	if body := tx.GetL1SignatureBody(c.GetChainId()); body == "" {
		t.Error("GetL1SignatureBody returned empty string")
	}
	tx.SetL1Sig("0xdeadbeef")
	if tx.L1Sig != "0xdeadbeef" {
		t.Errorf("SetL1Sig did not set L1Sig, got %q", tx.L1Sig)
	}
	assertSkipNonce(t, "SignTransfer(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignChangePubKey(t *testing.T) {
	c := newSignTestClient(t)
	var pk [40]byte
	for i := range pk {
		pk[i] = byte(i + 1)
	}
	req := &types.ChangePubKeyReq{PubKey: pk}

	tx, err := c.GetChangePubKeyTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetChangePubKeyTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2ChangePubKey {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2ChangePubKey)
	}
	if body := tx.GetL1SignatureBody(); body == "" {
		t.Error("GetL1SignatureBody returned empty string")
	}
	tx.SetL1Sig("0xfeed")
	if tx.L1Sig != "0xfeed" {
		t.Errorf("SetL1Sig did not set L1Sig, got %q", tx.L1Sig)
	}
	assertSkipNonce(t, "SignChangePubKey(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignCreatePublicPool(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.CreatePublicPoolTxReq{OperatorFee: 100, InitialTotalShares: 1000, MinOperatorShareRate: 50}

	tx, err := c.GetCreatePublicPoolTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCreatePublicPoolTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CreatePublicPool {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CreatePublicPool)
	}
	if tx.OperatorFee != 100 {
		t.Errorf("OperatorFee = %d, want 100", tx.OperatorFee)
	}
	assertSkipNonce(t, "SignCreatePublicPool(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignUpdatePublicPool(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.UpdatePublicPoolTxReq{PublicPoolIndex: minSubAccountIndex, Status: 1, OperatorFee: 100, MinOperatorShareRate: 50}

	tx, err := c.GetUpdatePublicPoolTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetUpdatePublicPoolTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2UpdatePublicPool {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2UpdatePublicPool)
	}
	if tx.PublicPoolIndex != minSubAccountIndex {
		t.Errorf("PublicPoolIndex = %d, want %d", tx.PublicPoolIndex, minSubAccountIndex)
	}
	assertSkipNonce(t, "SignUpdatePublicPool(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignMintShares(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.MintSharesTxReq{PublicPoolIndex: minSubAccountIndex, ShareAmount: 500}

	tx, err := c.GetMintSharesTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetMintSharesTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2MintShares {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2MintShares)
	}
	if tx.ShareAmount != 500 {
		t.Errorf("ShareAmount = %d, want 500", tx.ShareAmount)
	}
	assertSkipNonce(t, "SignMintShares(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignBurnShares(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.BurnSharesTxReq{PublicPoolIndex: minSubAccountIndex, ShareAmount: 500}

	tx, err := c.GetBurnSharesTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetBurnSharesTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2BurnShares {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2BurnShares)
	}
	assertSkipNonce(t, "SignBurnShares(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignUpdateMargin(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.UpdateMarginTxReq{MarketIndex: 0, USDCAmount: 1_000_000, Direction: 0}

	tx, err := c.GetUpdateMarginTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetUpdateMarginTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2UpdateMargin {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2UpdateMargin)
	}
	if tx.USDCAmount != 1_000_000 {
		t.Errorf("USDCAmount = %d, want 1000000", tx.USDCAmount)
	}
	assertSkipNonce(t, "SignUpdateMargin(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignStakeAssets(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.StakeAssetsTxReq{StakingPoolIndex: minSubAccountIndex, ShareAmount: 500}

	tx, err := c.GetStakeAssetsTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetStakeAssetsTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2StakeAssets {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2StakeAssets)
	}
	if tx.StakingPoolIndex != minSubAccountIndex {
		t.Errorf("StakingPoolIndex = %d, want %d", tx.StakingPoolIndex, minSubAccountIndex)
	}
	assertSkipNonce(t, "SignStakeAssets(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignUnstakeAssets(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.UnstakeAssetsTxReq{StakingPoolIndex: minSubAccountIndex, ShareAmount: 500}

	tx, err := c.GetUnstakeAssetsTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetUnstakeAssetsTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2UnstakeAssets {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2UnstakeAssets)
	}
	assertSkipNonce(t, "SignUnstakeAssets(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignApproveIntegrator(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.ApproveIntegratorTxReq{
		IntegratorAccountIndex: 5, MaxPerpsTakerFee: 10, MaxPerpsMakerFee: 20,
		MaxSpotTakerFee: 30, MaxSpotMakerFee: 40, ApprovalExpiry: 999,
	}

	tx, err := c.GetApproveIntegratorTx(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetApproveIntegratorTx failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2ApproveIntegrator {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2ApproveIntegrator)
	}
	if tx.IntegratorAccountIndex != 5 {
		t.Errorf("IntegratorAccountIndex = %d, want 5", tx.IntegratorAccountIndex)
	}
	if body := tx.GetL1SignatureBody(c.GetChainId()); body == "" {
		t.Error("GetL1SignatureBody returned empty string")
	}
	assertSkipNonce(t, "SignApproveIntegrator(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignUpdateAccountConfig(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.UpdateAccountConfigTxReq{AccountTradingMode: 1}

	tx, err := c.GetUpdateAccountConfigTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetUpdateAccountConfigTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2UpdateAccountConfig {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2UpdateAccountConfig)
	}
	if tx.AccountTradingMode != 1 {
		t.Errorf("AccountTradingMode = %d, want 1", tx.AccountTradingMode)
	}
	assertSkipNonce(t, "SignUpdateAccountConfig(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignUpdateAccountAssetConfig(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.UpdateAccountAssetConfigTxReq{AssetIndex: 3, AssetMarginMode: 1}

	tx, err := c.GetUpdateAccountAssetConfigTransaction(context.Background(), req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetUpdateAccountAssetConfigTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2UpdateAccountAssetConfig {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2UpdateAccountAssetConfig)
	}
	if tx.AssetIndex != 3 {
		t.Errorf("AssetIndex = %d, want 3", tx.AssetIndex)
	}
	assertSkipNonce(t, "SignUpdateAccountAssetConfig(skipNonce=1)", tx.L2TxAttributes, true)
}

// TestSignCreateOrderWithIntegratorFees verifies the integrator-fee attributes
// are carried into the signed hash.
func TestSignCreateOrderWithIntegratorFees(t *testing.T) {
	c := newSignTestClient(t)
	req := &types.CreateOrderTxReq{
		MarketIndex: 0, ClientOrderIndex: 1, BaseAmount: 1000, Price: 50000,
		IsAsk: 0, Type: 0, TimeInForce: 0, ReduceOnly: 0, TriggerPrice: 0, OrderExpiry: 0,
	}

	n := testNonce
	integratorIdx := int64(2)
	takerFee := uint32(10)
	makerFee := uint32(20)
	opsWithFees := &types.TransactOpts{
		Nonce: &n,
		TxAttributes: &types.L2TxAttributes{
			IntegratorAccountIndex: &integratorIdx,
			IntegratorTakerFee:     &takerFee,
			IntegratorMakerFee:     &makerFee,
		},
	}
	withFees, err := c.GetCreateOrderTransaction(context.Background(), req, opsWithFees)
	if err != nil {
		t.Fatalf("GetCreateOrderTransaction (integrator fees) failed: %v", err)
	}
	attrs := withFees.L2TxAttributes
	if attrs[txtypes.AttributeTypeIntegratorAccountIndex] != int(integratorIdx) {
		t.Errorf("IntegratorAccountIndex attr = %d, want %d", attrs[txtypes.AttributeTypeIntegratorAccountIndex], integratorIdx)
	}
	if attrs[txtypes.AttributeTypeIntegratorTakerFee] != int(takerFee) {
		t.Errorf("IntegratorTakerFee attr = %d, want %d", attrs[txtypes.AttributeTypeIntegratorTakerFee], takerFee)
	}
	if attrs[txtypes.AttributeTypeIntegratorMakerFee] != int(makerFee) {
		t.Errorf("IntegratorMakerFee attr = %d, want %d", attrs[txtypes.AttributeTypeIntegratorMakerFee], makerFee)
	}

	noFees, err := c.GetCreateOrderTransaction(context.Background(), req, &types.TransactOpts{Nonce: &n})
	if err != nil {
		t.Fatalf("GetCreateOrderTransaction (no fees) failed: %v", err)
	}
	if withFees.SignedHash == noFees.SignedHash {
		t.Error("integrator fees should change the signed hash, but hashes are identical")
	}
}
