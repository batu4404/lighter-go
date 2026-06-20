package paper

import (
	"math"
	"testing"
)

const epsilon = 1e-7

func almostEqual(t *testing.T, got, want float64, msg string) {
	t.Helper()
	if math.Abs(got-want) > epsilon {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

func almostEqualPlaces(t *testing.T, got, want float64, places int, msg string) {
	t.Helper()
	tol := math.Pow(10, -float64(places)) / 2
	if math.Abs(got-want) > tol {
		t.Errorf("%s: got %v, want %v (within %d places)", msg, got, want, places)
	}
}

// cfg mirrors test/paper_client/helpers.cfg.
func cfg(marketID, imf, mmf, comf int) MarketConfig {
	return MarketConfig{
		MarketID:                     marketID,
		Symbol:                       "MKT",
		SizeDecimals:                 4,
		PriceDecimals:                2,
		DefaultInitialMarginFraction: imf,
		MinInitialMarginFraction:     imf,
		MaintenanceMarginFraction:    mmf,
		CloseoutMarginFraction:       comf,
		TakerFee:                     0.0005,
		MakerFee:                     0.0002,
		MinBaseAmount:                0.001,
		MinQuoteAmount:               1.0,
		LastTradePrice:               100.0,
	}
}

// defaultCfg is cfg() with the python defaults (imf=1000, mmf=500, comf=250).
func defaultCfg() MarketConfig { return cfg(0, 1000, 500, 250) }

// levels builds an ordered []OrderBookLevel from (price, size) string pairs.
func levels(pairs ...[2]string) []OrderBookLevel {
	out := make([]OrderBookLevel, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, OrderBookLevel{Price: p[0], Size: p[1]})
	}
	return out
}

func lvl(price, size string) [2]string { return [2]string{price, size} }

// prices extracts the price strings from a level slice.
func prices(ls []OrderBookLevel) []string {
	out := make([]string, len(ls))
	for i, l := range ls {
		out[i] = l.Price
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// fakeOrderSource mirrors test/paper_client/helpers.FakeOrderApi.
type fakeOrderSource struct {
	details map[int]MarketDetail
	books   map[int]BookData
}

func newFakeOrderSource() *fakeOrderSource {
	return &fakeOrderSource{
		details: map[int]MarketDetail{
			0: defaultDetail(0, "ETH", 3000.0),
			1: defaultDetail(1, "BTC", 3000.0),
		},
		books: map[int]BookData{},
	}
}

func (f *fakeOrderSource) OrderBookDetails(marketID int) ([]MarketDetail, error) {
	d, ok := f.details[marketID]
	if !ok {
		return nil, &ParamError{Msg: "no detail"}
	}
	return []MarketDetail{d}, nil
}

func (f *fakeOrderSource) OrderBookOrders(marketID, limit int) (BookData, error) {
	b, ok := f.books[marketID]
	if !ok {
		return BookData{}, &ParamError{Msg: "no book"}
	}
	return b, nil
}

// defaultDetail mirrors helpers.default_detail.
func defaultDetail(marketID int, symbol string, lastTradePrice float64) MarketDetail {
	return MarketDetail{
		MarketID:                     marketID,
		Symbol:                       symbol,
		MarketType:                   "perp",
		SizeDecimals:                 4,
		PriceDecimals:                2,
		DefaultInitialMarginFraction: 1000,
		MinInitialMarginFraction:     500,
		MaintenanceMarginFraction:    50,
		CloseoutMarginFraction:       25,
		MinBaseAmount:                0.001,
		MinQuoteAmount:               1.0,
		LastTradePrice:               lastTradePrice,
	}
}

// book mirrors helpers.book: builds a snapshot BookData from (price,size) pairs.
func book(asks, bids [][2]string) BookData {
	return BookData{Asks: levels(asks...), Bids: levels(bids...)}
}
