package http

import (
	"context"

	core "github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/types/api"
)

type bridgeAPIImpl struct {
	client *client
}

// Ensure bridgeAPIImpl implements BridgeAPI
var _ core.BridgeAPI = (*bridgeAPIImpl)(nil)

func (b *bridgeAPIImpl) GetFastBridgeInfo(ctx context.Context) (*api.RespGetFastBridgeInfo, error) {
	result := &api.RespGetFastBridgeInfo{}
	err := b.client.getAndParseL2HTTPResponse(ctx, "api/v1/fastbridge/info", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *bridgeAPIImpl) CreateIntentAddress(ctx context.Context, chainID int64, fromAddr string, amount string, isExternalDeposit bool) (*api.RespCreateIntentAddress, error) {
	result := &api.RespCreateIntentAddress{}
	err := b.client.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/createIntentAddress", map[string]any{
		"chain_id":            chainID,
		"from_addr":           fromAddr,
		"amount":              amount,
		"is_external_deposit": isExternalDeposit,
	}, "", result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *bridgeAPIImpl) GetFastWithdrawInfo(ctx context.Context, accountIndex int64, auth string) (*api.FastWithdrawPoolInfo, error) {
	result := &api.FastWithdrawPoolInfo{}
	err := b.client.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/fastwithdraw/info", map[string]any{
		"account_index": accountIndex,
	}, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *bridgeAPIImpl) SubmitFastWithdraw(ctx context.Context, txInfo string, toAddress string, auth string) (*api.RespFastWithdraw, error) {
	result := &api.RespFastWithdraw{}
	err := b.client.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/fastwithdraw", map[string]any{
		"tx_info":    txInfo,
		"to_address": toAddress,
	}, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
