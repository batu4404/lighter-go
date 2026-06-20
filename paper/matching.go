package paper

import (
	"fmt"
	"math"
)

// SimulateMatch walks the resting book and produces taker fills for the
// request, returning the fills and the unfilled remaining size. Levels with
// unparseable or non-positive price/size are skipped; for IOC orders matching
// stops once the level price crosses the limit.
func SimulateMatch(req OrderRequest, asks, bids []OrderBookLevel, config MarketConfig) ([]Fill, float64) {
	remaining := req.BaseAmount
	fills := []Fill{}

	levels := asks
	if req.Side == SideSell {
		levels = bids
	}

	for _, level := range levels {
		if remaining <= 0 {
			break
		}
		price, okP := level.priceFloat()
		size, okS := level.sizeFloat()
		if !okP || !okS {
			continue
		}
		if price <= 0 || size <= 0 {
			continue
		}
		if req.OrderType == OrderIOC {
			if req.Side == SideBuy && price > req.Price {
				break
			}
			if req.Side == SideSell && price < req.Price {
				break
			}
		}
		fillSize := math.Min(remaining, size)
		fills = append(fills, Fill{
			Price: price,
			Size:  fillSize,
			Fee:   fillSize * price * config.TakerFee,
		})
		remaining -= fillSize
	}

	return fills, remaining
}

// ValidateOrder checks size/price against the market's decimal precision.
func ValidateOrder(req OrderRequest, config MarketConfig) error {
	if req.BaseAmount <= 0 {
		return &ParamError{Msg: fmt.Sprintf("base amount must be positive, got %v", req.BaseAmount)}
	}
	if !fitsDecimals(req.BaseAmount, config.SizeDecimals) {
		return &ParamError{Msg: fmt.Sprintf("base amount %v exceeds %d size decimals for %s", req.BaseAmount, config.SizeDecimals, config.Symbol)}
	}
	if req.OrderType == OrderIOC {
		if req.Price <= 0 {
			return &ParamError{Msg: fmt.Sprintf("IoC order requires positive price, got %v", req.Price)}
		}
		if !fitsDecimals(req.Price, config.PriceDecimals) {
			return &ParamError{Msg: fmt.Sprintf("price %v exceeds %d price decimals for %s", req.Price, config.PriceDecimals, config.Symbol)}
		}
	}
	return nil
}

func fitsDecimals(value float64, decimals int) bool {
	multiplier := math.Pow(10, float64(decimals))
	rounded := math.Round(value*multiplier) / multiplier
	return math.Abs(rounded-value) <= 1e-12
}
