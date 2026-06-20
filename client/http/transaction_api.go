package http

import (
	"context"
	"fmt"
	"strings"

	core "github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/types/api"
)

type transactionAPIImpl struct {
	client *client
}

// Ensure transactionAPIImpl implements TransactionAPI
var _ core.TransactionAPI = (*transactionAPIImpl)(nil)

func (t *transactionAPIImpl) SendTx(ctx context.Context, txType uint8, txInfo string, priceProtection *api.PriceProtection) (*api.RespSendTx, error) {
	return t.SendTxWithIndices(ctx, txType, txInfo, priceProtection, nil, nil, "")
}

// SendTxWithIndices sends a transaction with optional account_index, api_key_index, and auth parameters
func (t *transactionAPIImpl) SendTxWithIndices(ctx context.Context, txType uint8, txInfo string, priceProtection *api.PriceProtection, accountIndex *int64, apiKeyIndex *uint8, auth string) (*api.RespSendTx, error) {
	result := &api.RespSendTx{}
	// Lighter expects application/x-www-form-urlencoded for sendTx (matches the
	// Python SDK and openapi.json: tx_type/tx_info are form fields, not multipart
	// or JSON). auth, when present, is sent as a form field, not a header.
	params := map[string]any{
		"tx_type": txType,
		"tx_info": txInfo,
	}
	if priceProtection != nil {
		params["price_protection"] = true
	}
	if accountIndex != nil {
		params["account_index"] = *accountIndex
	}
	if apiKeyIndex != nil {
		params["api_key_index"] = *apiKeyIndex
	}
	if auth != "" {
		params["auth"] = auth
	}
	err := t.client.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/sendTx", params, "", result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) SendTxBatch(ctx context.Context, txTypes []uint8, txInfos []string) (*api.RespSendTxBatch, error) {
	return t.SendTxBatchWithAuth(ctx, txTypes, txInfos, "")
}

func (t *transactionAPIImpl) SendTxBatchWithAuth(ctx context.Context, txTypes []uint8, txInfos []string, auth string) (*api.RespSendTxBatch, error) {
	result := &api.RespSendTxBatch{}
	body := map[string]any{
		"tx_types": txTypes,
		"tx_infos": txInfos,
	}
	err := t.client.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/sendTxBatch", body, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTx(ctx context.Context, by api.QueryBy, value string) (*api.EnrichedTx, error) {
	result := &api.EnrichedTx{}
	err := t.client.getAndParseL2HTTPResponse(ctx, "api/v1/tx", map[string]any{
		"by":    string(by),
		"value": value,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTxs(ctx context.Context, index *int64, limit int) (*api.Txs, error) {
	result := &api.Txs{}
	params := map[string]any{
		"limit": limit,
	}
	if index != nil {
		params["index"] = *index
	}
	err := t.client.getAndParseL2HTTPResponse(ctx, "api/v1/txs", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetAccountTxs(ctx context.Context, by api.QueryBy, value string, limit int, types []api.TxType) (*api.Txs, error) {
	result := &api.Txs{}
	params := map[string]any{
		"by":    string(by),
		"value": value,
		"limit": limit,
	}
	if len(types) > 0 {
		typeStrs := make([]string, len(types))
		for i, t := range types {
			typeStrs[i] = fmt.Sprintf("%d", t)
		}
		params["types"] = strings.Join(typeStrs, ",")
	}
	err := t.client.getAndParseL2HTTPResponse(ctx, "api/v1/accountTxs", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTxFromL1TxHash(ctx context.Context, hash string) (*api.EnrichedTx, error) {
	result := &api.EnrichedTx{}
	err := t.client.getAndParseL2HTTPResponse(ctx, "api/v1/txFromL1TxHash", map[string]any{
		"hash": hash,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetDepositHistory(ctx context.Context, accountIndex int64, l1Address string, filter string, cursor string) (*api.DepositHistory, error) {
	result := &api.DepositHistory{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if l1Address != "" {
		params["l1_address"] = l1Address
	}
	if filter != "" {
		params["filter"] = filter
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	err := t.client.getAndParseL2HTTPResponse(ctx, "api/v1/deposit/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetWithdrawHistory(ctx context.Context, accountIndex int64, filter string, cursor string) (*api.WithdrawHistory, error) {
	result := &api.WithdrawHistory{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if filter != "" {
		params["filter"] = filter
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	err := t.client.getAndParseL2HTTPResponse(ctx, "api/v1/withdraw/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTransferHistory(ctx context.Context, accountIndex int64, cursor string) (*api.TransferHistory, error) {
	result := &api.TransferHistory{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	err := t.client.getAndParseL2HTTPResponse(ctx, "api/v1/transfer/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTransferFeeInfo(ctx context.Context, accountIndex int64, toAccountIndex *int64, auth string) (*api.TransferFeeInfo, error) {
	result := &api.TransferFeeInfo{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if toAccountIndex != nil {
		params["to_account_index"] = *toAccountIndex
	}
	var err error
	if auth != "" {
		err = t.client.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/transferFeeInfo", params, auth, result)
	} else {
		err = t.client.getAndParseL2HTTPResponse(ctx, "api/v1/transferFeeInfo", params, result)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetWithdrawalDelay(ctx context.Context) (*api.RespWithdrawalDelay, error) {
	result := &api.RespWithdrawalDelay{}
	err := t.client.getAndParseL2HTTPResponse(ctx, "api/v1/withdrawalDelay", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
