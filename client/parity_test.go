package client

import (
	"context"
	"testing"

	"github.com/0xJord4n/lighter-go/types"
)

// goldenExpiredAt is a fixed expiry so the signed tx hash is deterministic.
const goldenExpiredAt int64 = 1700000000000

// goldenOps pins every input that feeds the tx hash (nonce, expiry, account,
// api key). Combined with the fixed testChainID this makes GetTxHash()
// fully deterministic and independent of the (randomly generated) signing key.
func goldenOps() *types.TransactOpts {
	n := testNonce
	acc := testAccountIndex
	ak := testAPIKeyIndex
	return &types.TransactOpts{
		FromAccountIndex: &acc,
		ApiKeyIndex:      &ak,
		Nonce:            &n,
		ExpiredAt:        goldenExpiredAt,
	}
}

func goldenSkipOps() *types.TransactOpts {
	ops := goldenOps()
	sn := uint8(1)
	ops.TxAttributes = &types.L2TxAttributes{SkipNonce: &sn}
	return ops
}

// TestGoldenTxHashes pins every signing input so each tx hash is fully
// deterministic, then asserts the known-good value. These vectors guard the
// poseidon hashing + field ordering against regressions: any change to how a
// transaction is hashed (or to the crypto dependency) breaks them loudly.
// They are independent of the signing key.
//
// Note: these lock THIS SDK's hashing behaviour. Byte-for-byte equality with
// the lighter-python v1.1.0 signer is verified separately by tools/parity_check.py
// (the C ABI sets the 10-minute expiry internally, so the pinned ExpiredAt used
// here cannot be reproduced through that interface).
func TestGoldenTxHashes(t *testing.T) {
	c := newSignTestClient(t)

	want := map[string]string{
		"CreateOrder":              "25647bf44b00936df7eeb383d0b2c594e975d9dcf031f1d97cfce6b3c1434ddf9e460a259a9ac097",
		"CreateOrder+skip":         "6a7b3aef6912b2f8af3f8ca347412d04fc43b042d15648b3512c42011fa25c80fb8fadf016980034",
		"CancelOrder":              "29ff8c9294e7a7a3c732a177bed9f906ad10aab5031c9af2755389c3e1694b0c2dd588134e61288b",
		"CancelAllOrders":          "85190e6733830056cabbc6eb98013d7874795e71bbf0c010c8a4c6db20d8660aa8d15eb0e645d64c",
		"Withdraw":                 "811f1f17e2341ddfb5df31b5e202d97adaf8f5613352ec6086bcd877fef23315ae5a65eef8bfb332",
		"CreateSubAccount":         "5b247ef7aad88a3350d26143468c09db3cd22d560952e6c371026610e8c6dff7dd288d073fd69014",
		"ModifyOrder":              "e25e5392d59604c451d516f8d9784d624cde3187a63ecdc84ff455733d36857030b2683b049b2b72",
		"Transfer":                 "051512596b943f72bbcd8279dfa8844e3458382fb513a2d23ff9b93b23c00691ed2a79d56c6caee3",
		"CreatePublicPool":         "74786ecb2651671b1c732c3abb16f46a42c26a8b8ca5dd20c19ab87c1e62aef3ca3aa6dbb391c851",
		"UpdatePublicPool":         "58268ec07fcdbf85dcd4dc51fef68006a69dc5e39ffa5cb5cec16c251ea273ac269533813619d2d8",
		"MintShares":               "0910e238a63e94ccbd3dd3000b07d9c76453993522f0c6fd8eca25c3939942c2a4e26d24dabbe637",
		"BurnShares":               "f08b4a731c853e4c76389fe6a9226dc6295d5df06011dbc0cb4d7717353540325a9cd7c6a9d92c36",
		"UpdateLeverage":           "980f3ec68b85468f62d46ae99657e1859f1a5cfa06f74733bd3057b0052086f4e0ff05ac7d81ffac",
		"UpdateMargin":             "b5d450801445aa2d733c09904ea69867fcdcb14370543bfef790f170fb6dd33443d860e976a293e0",
		"StakeAssets":              "cf493e7ae356919f60d24bb3941ef4536c4fd12d5236028efde767a8d1443b68cbd5f6554a460ac3",
		"UnstakeAssets":            "53b449ced4ebfbeb436a9a020e5d85d96d540e2a0ab3d283eefd10b1162e62d84367b4d8af8f2f51",
		"ApproveIntegrator":        "784d0469d5a3dda668ccfddfd5c5348a3b2611d9f1141e1e3e0520cea7a3e844252fd7e47b914656",
		"UpdateAccountConfig":      "528905c50873162b584068eb23adbea2b51b4cbca1ce4e16c5da5ac2f180d3f8cb57b975bd1efaca",
		"UpdateAccountAssetConfig": "c325529672d0d3f702f558a9439c20e6d65421d341824a9b004945d8f68c063d1680285cd7aff98d",
	}

	cases := []struct {
		name string
		hash func() (string, error)
	}{
		{"CreateOrder", func() (string, error) {
			tx, err := c.GetCreateOrderTransaction(context.Background(), &types.CreateOrderTxReq{
				MarketIndex: 1, ClientOrderIndex: 0, BaseAmount: 1000, Price: 350000,
				IsAsk: 0, Type: 0, TimeInForce: 0, ReduceOnly: 0, TriggerPrice: 0, OrderExpiry: 0,
			}, goldenOps())
			return hashOf(tx, err)
		}},
		{"CreateOrder+skip", func() (string, error) {
			tx, err := c.GetCreateOrderTransaction(context.Background(), &types.CreateOrderTxReq{
				MarketIndex: 1, ClientOrderIndex: 0, BaseAmount: 1000, Price: 350000,
				IsAsk: 0, Type: 0, TimeInForce: 0, ReduceOnly: 0, TriggerPrice: 0, OrderExpiry: 0,
			}, goldenSkipOps())
			return hashOf(tx, err)
		}},
		{"CancelOrder", func() (string, error) {
			tx, err := c.GetCancelOrderTransaction(context.Background(), &types.CancelOrderTxReq{MarketIndex: 1, Index: 7}, goldenOps())
			return hashOf(tx, err)
		}},
		{"CancelAllOrders", func() (string, error) {
			tx, err := c.GetCancelAllOrdersTransaction(context.Background(), &types.CancelAllOrdersTxReq{TimeInForce: 0, Time: 0}, goldenOps())
			return hashOf(tx, err)
		}},
		{"Withdraw", func() (string, error) {
			tx, err := c.GetWithdrawTransaction(context.Background(), &types.WithdrawTxReq{AssetIndex: 3, RouteType: 0, Amount: 5_000_000}, goldenOps())
			return hashOf(tx, err)
		}},
		{"CreateSubAccount", func() (string, error) {
			tx, err := c.GetCreateSubAccountTransaction(context.Background(), goldenOps())
			return hashOf(tx, err)
		}},
		{"ModifyOrder", func() (string, error) {
			tx, err := c.GetModifyOrderTransaction(context.Background(), &types.ModifyOrderTxReq{MarketIndex: 1, Index: 7, BaseAmount: 2000, Price: 360000, TriggerPrice: 0}, goldenOps())
			return hashOf(tx, err)
		}},
		{"Transfer", func() (string, error) {
			tx, err := c.GetTransferTransaction(context.Background(), &types.TransferTxReq{ToAccountIndex: 2, AssetIndex: 3, FromRouteType: 0, ToRouteType: 0, Amount: 1_000_000, USDCFee: 100}, goldenOps())
			return hashOf(tx, err)
		}},
		{"CreatePublicPool", func() (string, error) {
			tx, err := c.GetCreatePublicPoolTransaction(context.Background(), &types.CreatePublicPoolTxReq{OperatorFee: 100, InitialTotalShares: 1000, MinOperatorShareRate: 50}, goldenOps())
			return hashOf(tx, err)
		}},
		{"UpdatePublicPool", func() (string, error) {
			tx, err := c.GetUpdatePublicPoolTransaction(context.Background(), &types.UpdatePublicPoolTxReq{PublicPoolIndex: minSubAccountIndex, Status: 1, OperatorFee: 100, MinOperatorShareRate: 50}, goldenOps())
			return hashOf(tx, err)
		}},
		{"MintShares", func() (string, error) {
			tx, err := c.GetMintSharesTransaction(context.Background(), &types.MintSharesTxReq{PublicPoolIndex: minSubAccountIndex, ShareAmount: 500}, goldenOps())
			return hashOf(tx, err)
		}},
		{"BurnShares", func() (string, error) {
			tx, err := c.GetBurnSharesTransaction(context.Background(), &types.BurnSharesTxReq{PublicPoolIndex: minSubAccountIndex, ShareAmount: 500}, goldenOps())
			return hashOf(tx, err)
		}},
		{"UpdateLeverage", func() (string, error) {
			tx, err := c.GetUpdateLeverageTransaction(context.Background(), &types.UpdateLeverageTxReq{MarketIndex: 1, InitialMarginFraction: 2000, MarginMode: 0}, goldenOps())
			return hashOf(tx, err)
		}},
		{"UpdateMargin", func() (string, error) {
			tx, err := c.GetUpdateMarginTransaction(context.Background(), &types.UpdateMarginTxReq{MarketIndex: 1, USDCAmount: 1_000_000, Direction: 0}, goldenOps())
			return hashOf(tx, err)
		}},
		{"StakeAssets", func() (string, error) {
			tx, err := c.GetStakeAssetsTransaction(context.Background(), &types.StakeAssetsTxReq{StakingPoolIndex: minSubAccountIndex, ShareAmount: 500}, goldenOps())
			return hashOf(tx, err)
		}},
		{"UnstakeAssets", func() (string, error) {
			tx, err := c.GetUnstakeAssetsTransaction(context.Background(), &types.UnstakeAssetsTxReq{StakingPoolIndex: minSubAccountIndex, ShareAmount: 500}, goldenOps())
			return hashOf(tx, err)
		}},
		{"ApproveIntegrator", func() (string, error) {
			tx, err := c.GetApproveIntegratorTx(context.Background(), &types.ApproveIntegratorTxReq{
				IntegratorAccountIndex: 5, MaxPerpsTakerFee: 10, MaxPerpsMakerFee: 20,
				MaxSpotTakerFee: 30, MaxSpotMakerFee: 40, ApprovalExpiry: 999,
			}, goldenOps())
			return hashOf(tx, err)
		}},
		{"UpdateAccountConfig", func() (string, error) {
			tx, err := c.GetUpdateAccountConfigTransaction(context.Background(), &types.UpdateAccountConfigTxReq{AccountTradingMode: 1}, goldenOps())
			return hashOf(tx, err)
		}},
		{"UpdateAccountAssetConfig", func() (string, error) {
			tx, err := c.GetUpdateAccountAssetConfigTransaction(context.Background(), &types.UpdateAccountAssetConfigTxReq{AssetIndex: 3, AssetMarginMode: 1}, goldenOps())
			return hashOf(tx, err)
		}},
	}

	for _, tc := range cases {
		got, err := tc.hash()
		if err != nil {
			t.Errorf("%s: build/sign error: %v", tc.name, err)
			continue
		}
		if got == "" {
			t.Errorf("%s: empty tx hash", tc.name)
			continue
		}
		w, ok := want[tc.name]
		if !ok || w == "" {
			t.Errorf("%s: GOLDEN %q", tc.name, got) // capture mode
			continue
		}
		if got != w {
			t.Errorf("%s: tx hash = %s, want %s", tc.name, got, w)
		}
	}
}

type hasher interface{ GetTxHash() string }

func hashOf(tx hasher, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return tx.GetTxHash(), nil
}
