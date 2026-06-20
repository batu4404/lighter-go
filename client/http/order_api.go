package http

import (
	"context"

	core "github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/types/api"
	"github.com/0xJord4n/lighter-go/types/apimodels"
)

type orderAPIImpl struct {
	client *client
}

// Ensure orderAPIImpl implements OrderAPI
var _ core.OrderAPI = (*orderAPIImpl)(nil)

func (o *orderAPIImpl) GetActiveOrders(ctx context.Context, accountIndex int64, marketID *int16, auth string) (*api.Orders, error) {
	result := &api.Orders{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if marketID != nil {
		params["market_id"] = *marketID
	}
	if auth != "" {
		params["auth"] = auth
	}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/accountActiveOrders", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *orderAPIImpl) GetInactiveOrders(ctx context.Context, accountIndex int64, marketID *int16, opts *core.InactiveOrdersOpts) (*api.Orders, error) {
	result := &api.Orders{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if marketID != nil {
		params["market_id"] = *marketID
	}
	if opts != nil {
		if opts.Limit > 0 {
			params["limit"] = opts.Limit
		}
		if opts.Cursor != "" {
			params["cursor"] = opts.Cursor
		}
	}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/accountInactiveOrders", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *orderAPIImpl) GetOrderBooks(ctx context.Context, marketID *int16, filter api.MarketFilter) (*api.OrderBooks, error) {
	result := &api.OrderBooks{}
	params := map[string]any{}
	if marketID != nil {
		params["market_id"] = *marketID
	}
	if filter != "" {
		params["filter"] = string(filter)
	}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/orderBooks", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *orderAPIImpl) GetOrderBookDetails(ctx context.Context, marketID *int16, filter api.MarketFilter) (*api.OrderBookDetails, error) {
	result := &api.OrderBookDetails{}
	params := map[string]any{}
	if marketID != nil {
		params["market_id"] = *marketID
	}
	if filter != "" {
		params["filter"] = string(filter)
	}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/orderBookDetails", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *orderAPIImpl) GetOrderBookOrders(ctx context.Context, marketID int16, limit int) (*apimodels.OrderBookOrders, error) {
	result := &apimodels.OrderBookOrders{}
	params := map[string]any{
		"market_id": marketID,
		"limit":     limit,
	}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/orderBookOrders", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *orderAPIImpl) GetRecentTrades(ctx context.Context, marketID int16, limit int) (*api.Trades, error) {
	result := &api.Trades{}
	params := map[string]any{
		"market_id": marketID,
		"limit":     limit,
	}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/recentTrades", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *orderAPIImpl) GetTrades(ctx context.Context, marketID int16, accountIndex *int64, opts *core.TradesOpts) (*api.Trades, error) {
	result := &api.Trades{}
	params := map[string]any{
		"market_id": marketID,
	}
	if accountIndex != nil {
		params["account_index"] = *accountIndex
	}
	if opts != nil {
		if opts.Limit > 0 {
			params["limit"] = opts.Limit
		}
		if opts.Cursor != "" {
			params["cursor"] = opts.Cursor
		}
		if opts.SortBy != "" {
			params["sort_by"] = opts.SortBy
		}
		if opts.SortOrder != "" {
			// The /trades endpoint names this query param "sort_dir" (not
			// "sort_order"); see lighter-python order_api.py.
			params["sort_dir"] = opts.SortOrder
		}
	}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/trades", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *orderAPIImpl) GetAssetDetails(ctx context.Context, assetID *int16) (*api.AssetDetails, error) {
	result := &api.AssetDetails{}
	params := map[string]any{}
	if assetID != nil {
		params["asset_id"] = *assetID
	}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/assetDetails", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *orderAPIImpl) GetExchangeStats(ctx context.Context) (*api.ExchangeStats, error) {
	result := &api.ExchangeStats{}
	err := o.client.getAndParseL2HTTPResponse(ctx, "api/v1/exchangeStats", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
