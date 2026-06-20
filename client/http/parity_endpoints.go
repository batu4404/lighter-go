package http

import (
	"context"

	"github.com/0xJord4n/lighter-go/types/apimodels"
)

// This file adds the Lighter REST endpoints that were missing from the
// hand-written client, bringing HTTP coverage to full parity with the
// Python SDK (lighter-python v1.1.1). Responses use the field-faithful
// generated types in types/apimodels. The signer-coupled SendTx path in
// transaction_api.go is intentionally left untouched.
//
// Endpoint shapes (method, params, body, response) mirror openapi.json
// exactly. GET endpoints pass query params; POST endpoints that the Python
// SDK sends as x-www-form-urlencoded are sent the same way here; JSON-body
// endpoints use postAndParseL2HTTPResponse.

// ---- RFQ ----

// CreateRFQ POST /api/v1/rfq/create (x-www-form-urlencoded, auth required).
func (c *client) CreateRFQ(ctx context.Context, params map[string]any, auth string) (*apimodels.RespCreateRFQ, error) {
	result := &apimodels.RespCreateRFQ{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/rfq/create", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetRFQ GET /api/v1/rfq/get (auth required).
func (c *client) GetRFQ(ctx context.Context, rfqID int64, auth string) (*apimodels.RespGetRFQ, error) {
	result := &apimodels.RespGetRFQ{}
	err := c.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/rfq/get", map[string]any{"rfq_id": rfqID}, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListRFQs GET /api/v1/rfq/list (auth required). Optional filters go in params.
func (c *client) ListRFQs(ctx context.Context, params map[string]any, auth string) (*apimodels.RespListRFQs, error) {
	result := &apimodels.RespListRFQs{}
	err := c.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/rfq/list", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RespondToRFQ POST /api/v1/rfq/respond (x-www-form-urlencoded, auth required).
func (c *client) RespondToRFQ(ctx context.Context, params map[string]any, auth string) (*apimodels.RespRespondToRFQ, error) {
	result := &apimodels.RespRespondToRFQ{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/rfq/respond", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateRFQ POST /api/v1/rfq/update (x-www-form-urlencoded, auth required).
func (c *client) UpdateRFQ(ctx context.Context, params map[string]any, auth string) (*apimodels.RespUpdateRFQ, error) {
	result := &apimodels.RespUpdateRFQ{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/rfq/update", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---- Referral (full) ----

// CreateReferralCode POST /api/v1/referral/create (x-www-form-urlencoded).
func (c *client) CreateReferralCode(ctx context.Context, params map[string]any, auth string) (*apimodels.ReferralCode, error) {
	result := &apimodels.ReferralCode{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/referral/create", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetReferralCode GET /api/v1/referral/get.
func (c *client) GetReferralCode(ctx context.Context, accountIndex int64, auth string) (*apimodels.ReferralCode, error) {
	result := &apimodels.ReferralCode{}
	err := c.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/referral/get", map[string]any{"account_index": accountIndex}, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetReferralPoints GET /api/v1/referral/points (auth required).
func (c *client) GetReferralPoints(ctx context.Context, accountIndex int64, auth string) (*apimodels.ReferralPoints, error) {
	result := &apimodels.ReferralPoints{}
	err := c.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/referral/points", map[string]any{"account_index": accountIndex}, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateReferralCode POST /api/v1/referral/update (x-www-form-urlencoded).
func (c *client) UpdateReferralCode(ctx context.Context, params map[string]any, auth string) (*apimodels.RespUpdateReferralCode, error) {
	result := &apimodels.RespUpdateReferralCode{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/referral/update", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateReferralKickback POST /api/v1/referral/kickback/update (x-www-form-urlencoded).
func (c *client) UpdateReferralKickback(ctx context.Context, params map[string]any, auth string) (*apimodels.RespUpdateKickback, error) {
	result := &apimodels.RespUpdateKickback{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/referral/kickback/update", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetUserReferrals GET /api/v1/referral/userReferrals. l1Address required; optional filters in params.
func (c *client) GetUserReferrals(ctx context.Context, l1Address string, params map[string]any, auth string) (*apimodels.UserReferrals, error) {
	if params == nil {
		params = map[string]any{}
	}
	params["l1_address"] = l1Address
	result := &apimodels.UserReferrals{}
	err := c.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/referral/userReferrals", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---- API tokens ----

// GetApiTokens GET /api/v1/tokens.
func (c *client) GetApiTokens(ctx context.Context, accountIndex int64, auth string) (*apimodels.RespGetApiTokens, error) {
	result := &apimodels.RespGetApiTokens{}
	err := c.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/tokens", map[string]any{"account_index": accountIndex}, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateApiToken POST /api/v1/tokens/create (x-www-form-urlencoded).
func (c *client) CreateApiToken(ctx context.Context, params map[string]any, auth string) (*apimodels.RespPostApiToken, error) {
	result := &apimodels.RespPostApiToken{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/tokens/create", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RevokeApiToken POST /api/v1/tokens/revoke (x-www-form-urlencoded).
func (c *client) RevokeApiToken(ctx context.Context, params map[string]any, auth string) (*apimodels.RespRevokeApiToken, error) {
	result := &apimodels.RespRevokeApiToken{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/tokens/revoke", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTokenList GET /api/v1/tokenlist.
func (c *client) GetTokenList(ctx context.Context) (*apimodels.TokenList, error) {
	result := &apimodels.TokenList{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/tokenlist", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---- Leases ----

// GetLeases GET /api/v1/leases. accountIndex required; optional filters in params.
func (c *client) GetLeases(ctx context.Context, accountIndex int64, params map[string]any, auth string) (*apimodels.RespGetLeases, error) {
	if params == nil {
		params = map[string]any{}
	}
	params["account_index"] = accountIndex
	result := &apimodels.RespGetLeases{}
	err := c.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/leases", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetLeaseOptions GET /api/v1/leaseOptions.
func (c *client) GetLeaseOptions(ctx context.Context) (*apimodels.RespGetLeaseOptions, error) {
	result := &apimodels.RespGetLeaseOptions{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/leaseOptions", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SubmitLitLease POST /api/v1/litLease (x-www-form-urlencoded).
func (c *client) SubmitLitLease(ctx context.Context, params map[string]any, auth string) (*apimodels.TxHash, error) {
	result := &apimodels.TxHash{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/litLease", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---- Maker-only API keys ----

// GetMakerOnlyApiKeys GET /api/v1/getMakerOnlyApiKeys (auth required).
func (c *client) GetMakerOnlyApiKeys(ctx context.Context, accountIndex int64, auth string) (*apimodels.RespGetMakerOnlyApiKeys, error) {
	result := &apimodels.RespGetMakerOnlyApiKeys{}
	err := c.getAndParseL2HTTPResponseWithAuth(ctx, "api/v1/getMakerOnlyApiKeys", map[string]any{"account_index": accountIndex}, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SetMakerOnlyApiKeys POST /api/v1/setMakerOnlyApiKeys (x-www-form-urlencoded, auth required).
func (c *client) SetMakerOnlyApiKeys(ctx context.Context, params map[string]any, auth string) (*apimodels.RespSetMakerOnlyApiKeys, error) {
	result := &apimodels.RespSetMakerOnlyApiKeys{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/setMakerOnlyApiKeys", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---- Account metadata / notifications ----

// SetAccountMetadata POST /api/v1/setAccountMetadata (JSON body).
func (c *client) SetAccountMetadata(ctx context.Context, body any) (*apimodels.ResultCode, error) {
	result := &apimodels.ResultCode{}
	err := c.postAndParseL2HTTPResponse(ctx, "api/v1/setAccountMetadata", body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AckNotification POST /api/v1/notification/ack (x-www-form-urlencoded).
func (c *client) AckNotification(ctx context.Context, params map[string]any, auth string) (*apimodels.ResultCode, error) {
	result := &apimodels.ResultCode{}
	err := c.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/notification/ack", params, auth, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---- Deposits ----

// GetLatestDeposit GET /api/v1/deposit/latest.
func (c *client) GetLatestDeposit(ctx context.Context, l1Address string) (*apimodels.Deposit, error) {
	result := &apimodels.Deposit{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/deposit/latest", map[string]any{"l1_address": l1Address}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetDepositNetworks GET /api/v1/deposit/networks.
func (c *client) GetDepositNetworks(ctx context.Context) (*apimodels.BridgeSupportedNetworks, error) {
	result := &apimodels.BridgeSupportedNetworks{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/deposit/networks", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---- Stats / config ----

// GetExchangeMetrics GET /api/v1/exchangeMetrics. period and kind required; optional filter/value in params.
func (c *client) GetExchangeMetrics(ctx context.Context, period, kind string, params map[string]any) (*apimodels.RespGetExchangeMetrics, error) {
	if params == nil {
		params = map[string]any{}
	}
	params["period"] = period
	params["kind"] = kind
	result := &apimodels.RespGetExchangeMetrics{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/exchangeMetrics", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetExecuteStats GET /api/v1/executeStats.
func (c *client) GetExecuteStats(ctx context.Context, period string) (*apimodels.RespGetExecuteStats, error) {
	result := &apimodels.RespGetExecuteStats{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/executeStats", map[string]any{"period": period}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPartnerStats GET /api/v1/partnerStats. accountIndex required; optional timestamps in params.
func (c *client) GetPartnerStats(ctx context.Context, accountIndex int64, params map[string]any) (*apimodels.PartnerStats, error) {
	if params == nil {
		params = map[string]any{}
	}
	params["account_index"] = accountIndex
	result := &apimodels.PartnerStats{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/partnerStats", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetSyntheticSpotInfo GET /api/v1/syntheticSpotInfo.
func (c *client) GetSyntheticSpotInfo(ctx context.Context, symbol string) (*apimodels.RespSyntheticSpotInfo, error) {
	result := &apimodels.RespSyntheticSpotInfo{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/syntheticSpotInfo", map[string]any{"symbol": symbol}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetSystemConfig GET /api/v1/systemConfig.
func (c *client) GetSystemConfig(ctx context.Context) (*apimodels.SystemConfig, error) {
	result := &apimodels.SystemConfig{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/systemConfig", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---- Root / L1 ----

// GetRootStatus GET / (API root health/status).
func (c *client) GetRootStatus(ctx context.Context) (*apimodels.Status, error) {
	result := &apimodels.Status{}
	err := c.getAndParseL2HTTPResponse(ctx, "", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetLayer1BasicInfo GET /api/v1/layer1BasicInfo.
func (c *client) GetLayer1BasicInfo(ctx context.Context) (*apimodels.Layer1BasicInfo, error) {
	result := &apimodels.Layer1BasicInfo{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/layer1BasicInfo", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
