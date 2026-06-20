package client

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xJord4n/lighter-go/types/apimodels"
)

func ask(price, rem string) apimodels.SimpleOrder {
	return apimodels.SimpleOrder{Price: price, RemainingBaseAmount: rem}
}

func TestStripDecimalToInt(t *testing.T) {
	cases := map[string]int64{
		"1691.51": 169151,
		"3.5754":  35754,
		"100":     100,
		"0.0001":  1,
	}
	for in, want := range cases {
		got, err := stripDecimalToInt(in)
		if err != nil {
			t.Fatalf("stripDecimalToInt(%q) error: %v", in, err)
		}
		if got.Int64() != want {
			t.Errorf("stripDecimalToInt(%q) = %d, want %d", in, got.Int64(), want)
		}
	}
	if _, err := stripDecimalToInt("  "); err == nil {
		t.Error("expected error for blank string")
	}
}

func TestRatRound(t *testing.T) {
	cases := []struct {
		num, den, want int64
	}{
		{10, 3, 3},  // 3.33 -> 3
		{11, 3, 4},  // 3.67 -> 4
		{5, 2, 3},   // 2.5 -> 3 (half away from zero)
		{-5, 2, -3}, // -2.5 -> -3
		{6, 2, 3},   // exact
	}
	for _, c := range cases {
		got := ratRound(big.NewRat(c.num, c.den))
		if got.Int64() != c.want {
			t.Errorf("ratRound(%d/%d) = %d, want %d", c.num, c.den, got.Int64(), c.want)
		}
	}
}

// TestGetPotentialExecutionPrice_Base verifies the VWAP price-walk for a
// base-sized buy against a known book, with hand-derived expected values.
func TestGetPotentialExecutionPrice_Base(t *testing.T) {
	c := &SignerClient{}
	ob := &apimodels.OrderBookOrders{
		Asks: []apimodels.SimpleOrder{
			ask("10.0", "2.0"), // price 100, rem 20
			ask("11.0", "5.0"), // price 110, rem 50
		},
	}
	// Buy 30 base units: 20 @100 (=2000) + 10 @110 (=1100) = 3100 over 30 -> 103.33
	vwap, matched, err := c.GetPotentialExecutionPrice(context.Background(), 0, big.NewInt(30), false, true, ob)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if matched.Int64() != 30 {
		t.Errorf("matched = %d, want 30", matched.Int64())
	}
	wantVWAP := big.NewRat(3100, 30)
	if vwap.Cmp(wantVWAP) != 0 {
		t.Errorf("vwap = %s, want %s", vwap.RatString(), wantVWAP.RatString())
	}
}

// TestGetPotentialExecutionPrice_Quote verifies the VWAP walk when sizing by
// quote (USD) amount.
func TestGetPotentialExecutionPrice_Quote(t *testing.T) {
	c := &SignerClient{}
	ob := &apimodels.OrderBookOrders{
		Asks: []apimodels.SimpleOrder{
			ask("10.0", "2.0"), // price 100, rem 20  -> level value 2000
			ask("11.0", "5.0"), // price 110, rem 50
		},
	}
	// Spend quote=2500: take all of level1 (100*20=2000, size 20), then from
	// level2 floor((2500-2000)/110)=4 units -> +440 -> matchedUSD=2440, size=24.
	vwap, matchedUSD, err := c.GetPotentialExecutionPrice(context.Background(), 0, big.NewInt(2500), false, false, ob)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if matchedUSD.Int64() != 2440 {
		t.Errorf("matchedUSD = %d, want 2440", matchedUSD.Int64())
	}
	wantVWAP := big.NewRat(2440, 24)
	if vwap.Cmp(wantVWAP) != 0 {
		t.Errorf("vwap = %s, want %s", vwap.RatString(), wantVWAP.RatString())
	}
}

func TestGetBestPrice(t *testing.T) {
	c := &SignerClient{}
	ob := &apimodels.OrderBookOrders{
		Asks: []apimodels.SimpleOrder{ask("12.34", "1.0")},
		Bids: []apimodels.SimpleOrder{ask("12.30", "1.0")},
	}
	buy, err := c.GetBestPrice(context.Background(), 0, false, ob) // buying -> best ask
	if err != nil || buy.Int64() != 1234 {
		t.Errorf("buy best price = %v (err %v), want 1234", buy, err)
	}
	sell, err := c.GetBestPrice(context.Background(), 0, true, ob) // selling -> best bid
	if err != nil || sell.Int64() != 1230 {
		t.Errorf("sell best price = %v (err %v), want 1230", sell, err)
	}
	empty := &apimodels.OrderBookOrders{}
	if _, err := c.GetBestPrice(context.Background(), 0, false, empty); err == nil {
		t.Error("expected error on empty book")
	}
}

func TestAcceptableExecutionPrice(t *testing.T) {
	// ideal 1000, 1% slippage. Buy (bid): allow up to 1010. Sell (ask): down to 990.
	if got := acceptableExecutionPrice(big.NewInt(1000), 0.01, false); got.Int64() != 1010 {
		t.Errorf("buy acceptable = %d, want 1010", got.Int64())
	}
	if got := acceptableExecutionPrice(big.NewInt(1000), 0.01, true); got.Int64() != 990 {
		t.Errorf("sell acceptable = %d, want 990", got.Int64())
	}
}
