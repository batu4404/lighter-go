package paper

import "testing"

func wantErr(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error, got nil", msg)
	}
}

func TestValidate_RejectsZeroSize(t *testing.T) {
	req := OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 0}
	wantErr(t, ValidateOrder(req, defaultCfg()), "zero size")
}

func TestValidate_RejectsBadSizeDecimals(t *testing.T) {
	c := defaultCfg()
	c.SizeDecimals = 2
	req := OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 0.001}
	wantErr(t, ValidateOrder(req, c), "bad size decimals")
}

func TestValidate_RejectsIocZeroPrice(t *testing.T) {
	req := OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 0.01, Price: 0, OrderType: OrderIOC}
	wantErr(t, ValidateOrder(req, defaultCfg()), "ioc zero price")
}

func TestValidate_RejectsIocBadPriceDecimals(t *testing.T) {
	c := defaultCfg()
	c.PriceDecimals = 2
	req := OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 0.01, Price: 0.001, OrderType: OrderIOC}
	wantErr(t, ValidateOrder(req, c), "ioc bad price decimals")
}

func TestSimulate_MarketBuyAgainstEmptyBook(t *testing.T) {
	req := OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 5.0}
	fills, remaining := SimulateMatch(req, nil, nil, defaultCfg())
	if len(fills) != 0 {
		t.Errorf("fills = %d, want 0", len(fills))
	}
	almostEqual(t, remaining, 5.0, "remaining")
}

func TestSimulate_MarketBuyPartialLiquidity(t *testing.T) {
	req := OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 10.0}
	asks := levels(lvl("3000", "3"))
	fills, remaining := SimulateMatch(req, asks, nil, defaultCfg())
	if len(fills) != 1 {
		t.Fatalf("fills = %d, want 1", len(fills))
	}
	almostEqual(t, fills[0].Size, 3.0, "fill size")
	almostEqual(t, remaining, 7.0, "remaining")
}

func TestSimulate_SellIocStopsBelowLimit(t *testing.T) {
	req := OrderRequest{MarketID: 0, Side: SideSell, BaseAmount: 5.0, Price: 2995, OrderType: OrderIOC}
	bids := levels(lvl("2999", "2"), lvl("2998", "2"), lvl("2990", "2"))
	fills, _ := SimulateMatch(req, nil, bids, defaultCfg())
	if len(fills) != 2 {
		t.Fatalf("fills = %d, want 2", len(fills))
	}
	almostEqual(t, fills[0].Price, 2999, "fill 0 price")
	almostEqual(t, fills[1].Price, 2998, "fill 1 price")
}

func TestSimulate_MalformedLevelSkipped(t *testing.T) {
	req := OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 5.0}
	asks := levels(lvl("bad", "1"), lvl("3000", "5"))
	fills, remaining := SimulateMatch(req, asks, nil, defaultCfg())
	if len(fills) != 1 {
		t.Fatalf("fills = %d, want 1", len(fills))
	}
	almostEqual(t, fills[0].Price, 3000, "fill price")
	almostEqual(t, remaining, 0.0, "remaining")
}
