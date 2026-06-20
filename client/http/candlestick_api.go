package http

import (
	"context"

	core "github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/types/api"
)

type candlestickAPIImpl struct {
	client *client
}

// Ensure candlestickAPIImpl implements CandlestickAPI
var _ core.CandlestickAPI = (*candlestickAPIImpl)(nil)

func (c *candlestickAPIImpl) GetCandlesticks(ctx context.Context, marketID int16, resolution api.CandlestickResolution, timestamps api.TimestampRange, countBack int) (*api.Candlesticks, error) {
	result := &api.Candlesticks{}
	params := map[string]any{
		"market_id":       marketID,
		"resolution":      string(resolution),
		"start_timestamp": timestamps.StartTimestamp,
		"end_timestamp":   timestamps.EndTimestamp,
		"count_back":      countBack,
	}
	// Use /api/v1/candles endpoint (candlesticks is blocked/deprecated)
	err := c.client.getAndParseL2HTTPResponse(ctx, "api/v1/candles", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *candlestickAPIImpl) GetFundings(ctx context.Context, marketID int16, resolution api.FundingResolution, timestamps api.TimestampRange, countBack int) (*api.Fundings, error) {
	result := &api.Fundings{}
	params := map[string]any{
		"market_id":       marketID,
		"resolution":      string(resolution),
		"start_timestamp": timestamps.StartTimestamp,
		"end_timestamp":   timestamps.EndTimestamp,
		"count_back":      countBack,
	}
	err := c.client.getAndParseL2HTTPResponse(ctx, "api/v1/fundings", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *candlestickAPIImpl) GetFundingRates(ctx context.Context) (*api.FundingRates, error) {
	result := &api.FundingRates{}
	err := c.client.getAndParseL2HTTPResponse(ctx, "api/v1/funding-rates", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
