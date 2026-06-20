package client

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/0xJord4n/lighter-go/types"
	"github.com/0xJord4n/lighter-go/types/apimodels"
	"github.com/0xJord4n/lighter-go/types/txtypes"
)

// defaultIOCExpiry mirrors lighter-python's DEFAULT_IOC_EXPIRY (0): market /
// IOC orders carry no expiry.
const defaultIOCExpiry int64 = 0

// stripDecimalToInt converts a decimal price/amount string into a scaled
// integer by removing the decimal point, e.g. "1691.51" -> 169151. This
// mirrors lighter-python's `int(s.replace(".", ""))`: the API encodes prices
// and base amounts as fixed-point decimal strings, and stripping the point
// yields the integer the matching engine works in. All values on one side of a
// market share the same decimal precision, so ratios between them are
// preserved.
func stripDecimalToInt(s string) (*big.Int, error) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(s), ".", "")
	if cleaned == "" {
		return nil, fmt.Errorf("empty numeric string")
	}
	v, ok := new(big.Int).SetString(cleaned, 10)
	if !ok {
		return nil, fmt.Errorf("not a numeric string: %q", s)
	}
	return v, nil
}

// GetBestPrice returns the top-of-book price as a decimal-stripped integer:
// the best ask when buying (isAsk == false) or the best bid when selling
// (isAsk == true). Mirrors lighter-python's get_best_price. Pass a non-nil
// obOrders to reuse a snapshot and avoid an extra API call.
func (c *SignerClient) GetBestPrice(ctx context.Context, marketIndex int16, isAsk bool, obOrders *apimodels.OrderBookOrders) (*big.Int, error) {
	if obOrders == nil {
		var err error
		obOrders, err = c.fullHTTP.Order().GetOrderBookOrders(ctx, marketIndex, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to get order book orders: %w", err)
		}
	}
	var levels []apimodels.SimpleOrder
	if isAsk {
		levels = obOrders.Bids
	} else {
		levels = obOrders.Asks
	}
	if len(levels) == 0 {
		return nil, fmt.Errorf("no liquidity on the relevant side of market %d", marketIndex)
	}
	return stripDecimalToInt(levels[0].Price)
}

// GetPotentialExecutionPrice walks the order book to estimate the volume-
// weighted execution price for filling `amount`, returning that price (as a
// big.Rat, matching lighter-python's Fraction arithmetic) and the amount that
// could be matched. When isAmountBase is true, `amount` and the returned
// matched value are in base units; otherwise they are in quote (USD) units.
// Mirrors lighter-python's get_potential_execution_price.
func (c *SignerClient) GetPotentialExecutionPrice(ctx context.Context, marketIndex int16, amount *big.Int, isAsk, isAmountBase bool, obOrders *apimodels.OrderBookOrders) (price *big.Rat, matched *big.Int, err error) {
	if obOrders == nil {
		obOrders, err = c.fullHTTP.Order().GetOrderBookOrders(ctx, marketIndex, 100)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get order book orders: %w", err)
		}
	}
	var levels []apimodels.SimpleOrder
	if isAsk {
		levels = obOrders.Bids
	} else {
		levels = obOrders.Asks
	}

	matchedUSD := new(big.Int)
	matchedSize := new(big.Int)
	for _, order := range levels {
		// Stop once the requested amount is fully matched.
		if isAmountBase && matchedSize.Cmp(amount) == 0 {
			break
		}
		if !isAmountBase && matchedUSD.Cmp(amount) == 0 {
			break
		}

		curPrice, perr := stripDecimalToInt(order.Price)
		if perr != nil {
			return nil, nil, perr
		}
		curSize, serr := stripDecimalToInt(order.RemainingBaseAmount)
		if serr != nil {
			return nil, nil, serr
		}
		if curPrice.Sign() == 0 {
			continue
		}

		// max base size we still want from this level
		var maxPossible *big.Int
		if isAmountBase {
			maxPossible = new(big.Int).Sub(amount, matchedSize)
		} else {
			// remaining quote / price, floored (Fraction in python, then min)
			remainingQuote := new(big.Int).Sub(amount, matchedUSD)
			maxPossible = new(big.Int).Quo(remainingQuote, curPrice)
		}

		toUse := curSize
		if maxPossible.Cmp(curSize) < 0 {
			toUse = maxPossible
		}
		if toUse.Sign() <= 0 {
			break
		}

		matchedUSD.Add(matchedUSD, new(big.Int).Mul(curPrice, toUse))
		matchedSize.Add(matchedSize, toUse)
	}

	if matchedSize.Sign() == 0 {
		return nil, nil, fmt.Errorf("no liquidity matched for market %d", marketIndex)
	}

	price = new(big.Rat).SetFrac(matchedUSD, matchedSize)
	if isAmountBase {
		matched = matchedSize
	} else {
		matched = matchedUSD
	}
	return price, matched, nil
}

// acceptableExecutionPrice computes round(ideal * (1 +/- maxSlippage)) using
// rationals so the rounding matches lighter-python's behaviour. maxSlippage is
// a fraction (0.01 == 1%). For asks the bound is below ideal; for bids, above.
func acceptableExecutionPrice(ideal *big.Int, maxSlippage float64, isAsk bool) *big.Int {
	slip := new(big.Rat).SetFloat64(maxSlippage)
	if slip == nil {
		slip = new(big.Rat)
	}
	if isAsk {
		slip.Neg(slip)
	}
	factor := new(big.Rat).Add(big.NewRat(1, 1), slip)
	val := new(big.Rat).Mul(new(big.Rat).SetInt(ideal), factor)
	return ratRound(val)
}

// ratRound rounds a big.Rat to the nearest integer (half away from zero),
// matching python's round() closely enough for price bounds.
func ratRound(r *big.Rat) *big.Int {
	num := new(big.Int).Set(r.Num())
	den := new(big.Int).Set(r.Denom())
	// q = num/den, rem
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	if rem.Sign() == 0 {
		return q
	}
	twiceRem := new(big.Int).Mul(new(big.Int).Abs(rem), big.NewInt(2))
	if twiceRem.Cmp(new(big.Int).Abs(den)) >= 0 {
		if r.Sign() < 0 {
			q.Sub(q, big.NewInt(1))
		} else {
			q.Add(q, big.NewInt(1))
		}
	}
	return q
}

// CreateMarketOrderQuoteAmount builds a market order sized by a quote (USDC)
// amount rather than a base amount. It walks the order book to estimate the
// fill, rejects the order if the volume-weighted price exceeds maxSlippage from
// the top of book or if there is insufficient liquidity, then returns a market
// order whose limit price is the slippage bound. quoteAmount is in USDC (it is
// scaled by 1e6 internally); maxSlippage is a fraction (0.01 == 1%). Mirrors
// lighter-python's create_market_order_quote_amount. Returns the built tx; the
// caller submits via SendAndSubmit.
func (c *SignerClient) CreateMarketOrderQuoteAmount(ctx context.Context, marketIndex int16, clientOrderIndex int64, quoteAmount float64, maxSlippage float64, isAsk, reduceOnly bool, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	// quote_amount = int(quote_amount * 1e6)
	quoteScaled := ratRound(new(big.Rat).Mul(new(big.Rat).SetFloat64(quoteAmount), big.NewRat(1_000_000, 1)))

	obOrders, err := c.fullHTTP.Order().GetOrderBookOrders(ctx, marketIndex, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book orders: %w", err)
	}

	idealPrice, err := c.GetBestPrice(ctx, marketIndex, isAsk, obOrders)
	if err != nil {
		return nil, err
	}
	acceptable := acceptableExecutionPrice(idealPrice, maxSlippage, isAsk)

	vwap, matchedUSD, err := c.GetPotentialExecutionPrice(ctx, marketIndex, quoteScaled, isAsk, false, obOrders)
	if err != nil {
		return nil, err
	}

	acceptableRat := new(big.Rat).SetInt(acceptable)
	if (isAsk && vwap.Cmp(acceptableRat) < 0) || (!isAsk && vwap.Cmp(acceptableRat) > 0) {
		return nil, fmt.Errorf("excessive slippage")
	}
	if matchedUSD.Cmp(quoteScaled) < 0 {
		return nil, fmt.Errorf("cannot be sure slippage will be acceptable due to the high size")
	}

	// base_amount = int(quote_amount / potential_execution_price)
	baseRat := new(big.Rat).Quo(new(big.Rat).SetInt(quoteScaled), vwap)
	baseAmount := new(big.Int).Quo(baseRat.Num(), baseRat.Denom()) // floor like python int()

	return c.buildSlippageMarketOrder(ctx, marketIndex, clientOrderIndex, baseAmount, acceptable, isAsk, reduceOnly, opts)
}

// CreateMarketOrderLimitedSlippage builds a market order for a fixed base
// amount, priced at the slippage bound (round(ideal * (1 +/- maxSlippage))).
// It does not walk the book for fill estimation -- the limit price caps how
// much of the size can execute. Mirrors create_market_order_limited_slippage.
func (c *SignerClient) CreateMarketOrderLimitedSlippage(ctx context.Context, marketIndex int16, clientOrderIndex int64, baseAmount int64, maxSlippage float64, isAsk, reduceOnly bool, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	idealPrice, err := c.GetBestPrice(ctx, marketIndex, isAsk, nil)
	if err != nil {
		return nil, err
	}
	acceptable := acceptableExecutionPrice(idealPrice, maxSlippage, isAsk)
	return c.buildSlippageMarketOrder(ctx, marketIndex, clientOrderIndex, big.NewInt(baseAmount), acceptable, isAsk, reduceOnly, opts)
}

// CreateMarketOrderIfSlippage builds a market order for a fixed base amount
// only if the estimated volume-weighted execution price stays within
// maxSlippage of the top of book and there is enough liquidity; otherwise it
// returns an error. Mirrors create_market_order_if_slippage.
func (c *SignerClient) CreateMarketOrderIfSlippage(ctx context.Context, marketIndex int16, clientOrderIndex int64, baseAmount int64, maxSlippage float64, isAsk, reduceOnly bool, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	obOrders, err := c.fullHTTP.Order().GetOrderBookOrders(ctx, marketIndex, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book orders: %w", err)
	}
	idealPrice, err := c.GetBestPrice(ctx, marketIndex, isAsk, obOrders)
	if err != nil {
		return nil, err
	}
	base := big.NewInt(baseAmount)
	vwap, matchedSize, err := c.GetPotentialExecutionPrice(ctx, marketIndex, base, isAsk, true, obOrders)
	if err != nil {
		return nil, err
	}
	acceptable := acceptableExecutionPrice(idealPrice, maxSlippage, isAsk)
	acceptableRat := new(big.Rat).SetInt(acceptable)
	if (isAsk && vwap.Cmp(acceptableRat) < 0) || (!isAsk && vwap.Cmp(acceptableRat) > 0) {
		return nil, fmt.Errorf("excessive slippage")
	}
	if matchedSize.Cmp(base) < 0 {
		return nil, fmt.Errorf("cannot be sure slippage will be acceptable due to the high size")
	}
	return c.buildSlippageMarketOrder(ctx, marketIndex, clientOrderIndex, base, acceptable, isAsk, reduceOnly, opts)
}

// buildSlippageMarketOrder is the shared tail of the slippage-aware market
// order helpers: an IOC market order whose limit price is the slippage bound.
func (c *SignerClient) buildSlippageMarketOrder(ctx context.Context, marketIndex int16, clientOrderIndex int64, baseAmount, price *big.Int, isAsk, reduceOnly bool, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	if baseAmount.Sign() <= 0 {
		return nil, fmt.Errorf("computed base amount is not positive")
	}
	if price.Sign() <= 0 || !price.IsInt64() || price.Int64() > int64(txtypes.MaxOrderPrice) {
		return nil, fmt.Errorf("computed price %s is out of range", price.String())
	}
	if !baseAmount.IsInt64() {
		return nil, fmt.Errorf("computed base amount %s overflows int64", baseAmount.String())
	}
	ask := uint8(0)
	if isAsk {
		ask = 1
	}
	ro := uint8(0)
	if reduceOnly {
		ro = 1
	}
	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: clientOrderIndex,
		BaseAmount:       baseAmount.Int64(),
		Price:            uint32(price.Int64()),
		IsAsk:            ask,
		Type:             txtypes.MarketOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       ro,
		TriggerPrice:     0,
		OrderExpiry:      defaultIOCExpiry,
	}
	return c.GetCreateOrderTransaction(ctx, req, opts)
}
