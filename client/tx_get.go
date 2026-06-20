package client

import (
	"context"
	"fmt"
	"time"

	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"

	"github.com/0xJord4n/lighter-go/types"
	"github.com/0xJord4n/lighter-go/types/txtypes"
)

func (c *TxClient) GetAuthToken(deadline time.Time) (string, error) {
	return types.ConstructAuthToken(c.keyManager, deadline, &types.TransactOpts{
		ApiKeyIndex:      &c.apiKeyIndex,
		FromAccountIndex: &c.accountIndex,
	})
}

func (c *TxClient) GetChangePubKeyTransaction(ctx context.Context, tx *types.ChangePubKeyReq, ops *types.TransactOpts) (*txtypes.L2ChangePubKeyTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructChangePubKeyTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}

	pk := c.keyManager.PubKeyBytes()
	msgHash, _ := txInfo.Hash(c.chainId)

	if err := schnorr.Validate(pk[:], msgHash, txInfo.Sig); err != nil {
		return nil, fmt.Errorf("failed to validate signature. error: %v", err)
	}

	// Signing a ChangePubKey is a strong signal that this account's
	// server-side pubkey is about to change. Drop cached entries so a
	// subsequent Check refetches from the server. (apiClient is nil when the
	// client was created without an HTTP backend, e.g. offline signing.)
	if c.apiClient != nil {
		c.apiClient.InvalidateApiKeys(c.accountIndex)
	}

	return txInfo, nil
}

func (c *TxClient) GetCreateSubAccountTransaction(ctx context.Context, ops *types.TransactOpts) (*txtypes.L2CreateSubAccountTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructCreateSubAccountTx(c.keyManager, c.chainId, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetCreatePublicPoolTransaction(ctx context.Context, tx *types.CreatePublicPoolTxReq, ops *types.TransactOpts) (*txtypes.L2CreatePublicPoolTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructCreatePublicPoolTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetUpdatePublicPoolTransaction(ctx context.Context, tx *types.UpdatePublicPoolTxReq, ops *types.TransactOpts) (*txtypes.L2UpdatePublicPoolTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructUpdatePublicPoolTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetTransferTransaction(ctx context.Context, tx *types.TransferTxReq, ops *types.TransactOpts) (*txtypes.L2TransferTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructTransferTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetWithdrawTransaction(ctx context.Context, tx *types.WithdrawTxReq, ops *types.TransactOpts) (*txtypes.L2WithdrawTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructWithdrawTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}

	return txInfo, nil
}

func (c *TxClient) GetCreateOrderTransaction(ctx context.Context, tx *types.CreateOrderTxReq, ops *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructCreateOrderTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetCreateGroupedOrdersTransaction(ctx context.Context, tx *types.CreateGroupedOrdersTxReq, ops *types.TransactOpts) (*txtypes.L2CreateGroupedOrdersTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructL2CreateGroupedOrdersTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetCancelOrderTransaction(ctx context.Context, tx *types.CancelOrderTxReq, ops *types.TransactOpts) (*txtypes.L2CancelOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructL2CancelOrderTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetModifyOrderTransaction(ctx context.Context, tx *types.ModifyOrderTxReq, ops *types.TransactOpts) (*txtypes.L2ModifyOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}

	txInfo, err := types.ConstructL2ModifyOrderTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}

	return txInfo, nil
}

func (c *TxClient) GetCancelAllOrdersTransaction(ctx context.Context, tx *types.CancelAllOrdersTxReq, ops *types.TransactOpts) (*txtypes.L2CancelAllOrdersTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructL2CancelAllOrdersTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetMintSharesTransaction(ctx context.Context, tx *types.MintSharesTxReq, ops *types.TransactOpts) (*txtypes.L2MintSharesTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructMintSharesTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetBurnSharesTransaction(ctx context.Context, tx *types.BurnSharesTxReq, ops *types.TransactOpts) (*txtypes.L2BurnSharesTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructBurnSharesTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetUpdateLeverageTransaction(ctx context.Context, tx *types.UpdateLeverageTxReq, ops *types.TransactOpts) (*txtypes.L2UpdateLeverageTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructUpdateLeverageTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetUpdateMarginTransaction(ctx context.Context, tx *types.UpdateMarginTxReq, ops *types.TransactOpts) (*txtypes.L2UpdateMarginTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructUpdateMarginTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetStakeAssetsTransaction(ctx context.Context, tx *types.StakeAssetsTxReq, ops *types.TransactOpts) (*txtypes.L2StakeAssetsTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructStakeAssetsTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetUnstakeAssetsTransaction(ctx context.Context, tx *types.UnstakeAssetsTxReq, ops *types.TransactOpts) (*txtypes.L2UnstakeAssetsTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructUnstakeAssetsTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetApproveIntegratorTx(ctx context.Context, tx *types.ApproveIntegratorTxReq, ops *types.TransactOpts) (*txtypes.L2ApproveIntegratorTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructApproveIntegratorTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}

	pk := c.keyManager.PubKeyBytes()
	msgHash, _ := txInfo.Hash(c.chainId)

	if err := schnorr.Validate(pk[:], msgHash, txInfo.Sig); err != nil {
		return nil, fmt.Errorf("failed to validate signature. error: %v", err)
	}

	return txInfo, nil
}

func (c *TxClient) GetUpdateAccountConfigTransaction(ctx context.Context, tx *types.UpdateAccountConfigTxReq, ops *types.TransactOpts) (*txtypes.L2UpdateAccountConfigTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructUpdateAccountConfigTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetUpdateAccountAssetConfigTransaction(ctx context.Context, tx *types.UpdateAccountAssetConfigTxReq, ops *types.TransactOpts) (*txtypes.L2UpdateAccountAssetConfigTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ctx, ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructUpdateAccountAssetConfigTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}
