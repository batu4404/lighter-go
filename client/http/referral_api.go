package http

import (
	"context"

	core "github.com/0xJord4n/lighter-go/client"
)

type referralAPIImpl struct {
	client *client
}

// Ensure referralAPIImpl implements ReferralAPI
var _ core.ReferralAPI = (*referralAPIImpl)(nil)

// ReferralUse assigns a referral code to the account identified by l1Address.
//
// Per Lighter's OpenAPI (ReqUseReferralCode), the body is
// application/x-www-form-urlencoded with l1_address and referral_code required.
// `x` is the account owner's X/Twitter handle: the spec marks it required but the
// Lighter team confirmed that is a typo and it is optional, so it (like
// auth/discord/telegram/signature) is only sent when non-empty. The auth token is
// sent both as the `auth` form field and the Authorization header.
func (r *referralAPIImpl) ReferralUse(ctx context.Context, l1Address, referralCode, x, auth string, opts *core.ReferralUseOpts) error {
	params := map[string]any{
		"l1_address":    l1Address,
		"referral_code": referralCode,
	}
	if x != "" {
		params["x"] = x
	}
	if auth != "" {
		params["auth"] = auth
	}
	if opts != nil {
		if opts.Discord != "" {
			params["discord"] = opts.Discord
		}
		if opts.Telegram != "" {
			params["telegram"] = opts.Telegram
		}
		if opts.Signature != "" {
			params["signature"] = opts.Signature
		}
	}

	// postURLEncodedL2HTTPResponseWithAuth parses the ResultCode and returns an
	// error for any non-OK code, so a nil error means the code was accepted.
	result := &ResultCode{}
	return r.client.postURLEncodedL2HTTPResponseWithAuth(ctx, "api/v1/referral/use", params, auth, result)
}
