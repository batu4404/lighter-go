package paper

import "testing"

// newTrackedClient builds a client over a fake source and tracks the given markets.
func newTrackedClient(t *testing.T, collateral float64, src *fakeOrderSource, opts []Option, markets ...int) *Client {
	t.Helper()
	c, err := NewClient(src, collateral, opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	for _, m := range markets {
		if err := c.TrackMarketSnapshot(m); err != nil {
			t.Fatalf("TrackMarketSnapshot(%d): %v", m, err)
		}
	}
	return c
}

func TestClient_SnapshotMarketBuyThenSell(t *testing.T) {
	src := newFakeOrderSource()
	src.books[0] = book([][2]string{lvl("3000.00", "0.5"), lvl("3001.00", "2.0")}, [][2]string{lvl("2999.00", "0.5"), lvl("2998.00", "2.0")})
	c := newTrackedClient(t, 5000.0, src, nil, 0)

	buy, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 1.0})
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	almostEqual(t, buy.FilledSize, 1.0, "filled")
	almostEqual(t, buy.Unfilled, 0.0, "unfilled")
	if len(buy.Fills) != 2 {
		t.Errorf("fills = %d, want 2", len(buy.Fills))
	}
	almostEqual(t, buy.AvgPrice, 3000.5, "avg price")
	almostEqual(t, buy.QuoteAmount, 3000.5, "quote")
	almostEqual(t, buy.TotalFee, 0.0, "fee")

	pos := c.GetPosition(0)
	if pos == nil {
		t.Fatal("expected position")
	}
	almostEqual(t, pos.Size, 1.0, "position size")

	sell, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideSell, BaseAmount: 1.0})
	if err != nil {
		t.Fatalf("sell: %v", err)
	}
	almostEqual(t, sell.FilledSize, 1.0, "sell filled")
	if c.GetPosition(0) != nil {
		t.Error("position should be closed")
	}
	if len(c.GetTrades()) != 4 {
		t.Errorf("trades = %d, want 4", len(c.GetTrades()))
	}
}

func TestClient_IocPartiallyFillsUntilLimit(t *testing.T) {
	src := newFakeOrderSource()
	src.books[0] = book([][2]string{lvl("3000.00", "0.5"), lvl("3010.00", "2.0")}, [][2]string{lvl("2999.00", "1.0")})
	c := newTrackedClient(t, 5000.0, src, nil, 0)

	res, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 2.0, Price: 3005.00, OrderType: OrderIOC})
	if err != nil {
		t.Fatalf("order: %v", err)
	}
	almostEqual(t, res.FilledSize, 0.5, "filled")
	almostEqual(t, res.Unfilled, 1.5, "unfilled")
	if len(res.Fills) != 1 {
		t.Errorf("fills = %d, want 1", len(res.Fills))
	}
}

func TestClient_HealthCrossMarginAcrossMarkets(t *testing.T) {
	src := newFakeOrderSource()
	d1 := defaultDetail(1, "BTC", 60000.0)
	d1.MinBaseAmount = 0.0001
	src.details[1] = d1
	src.books[0] = book([][2]string{lvl("3000.00", "10.0")}, [][2]string{lvl("2999.00", "10.0")})
	src.books[1] = book([][2]string{lvl("60000.00", "10.0")}, [][2]string{lvl("59999.00", "10.0")})
	c := newTrackedClient(t, 10000.0, src, nil, 0, 1)

	if _, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 1.0}); err != nil {
		t.Fatalf("order 0: %v", err)
	}
	if _, err := c.CreatePaperOrder(OrderRequest{MarketID: 1, Side: SideBuy, BaseAmount: 0.1}); err != nil {
		t.Fatalf("order 1: %v", err)
	}
	almostEqual(t, c.GetPosition(0).Size, 1.0, "pos 0")
	almostEqual(t, c.GetPosition(1).Size, 0.1, "pos 1")

	h := c.GetHealth()
	if h.Status != StatusHealthy {
		t.Errorf("status = %v, want healthy", h.Status)
	}
	if !(h.Leverage > 0 && h.Leverage < 2) {
		t.Errorf("leverage = %v, want in (0,2)", h.Leverage)
	}
}

func TestClient_OrderTriggersCrossMarketLiquidation(t *testing.T) {
	src := newFakeOrderSource()
	d1 := defaultDetail(1, "BTC", 3000.0)
	d1.MaintenanceMarginFraction = 500
	d1.CloseoutMarginFraction = 250
	d1.MinBaseAmount = 0.0001
	src.details[1] = d1
	src.books[0] = book([][2]string{lvl("1.00", "10.0")}, [][2]string{lvl("0.99", "10.0")})
	src.books[1] = book([][2]string{lvl("31.00", "10.0")}, [][2]string{lvl("30.00", "10.0")})
	c := newTrackedClient(t, 350.0, src, nil, 0, 1)

	ApplyFill(c.account, 1, SideBuy, 5.0, 100.0, 0, false)

	if _, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 10.0, Price: 0.50, OrderType: OrderIOC}); err != nil {
		t.Fatalf("order: %v", err)
	}
	if c.GetPosition(0) != nil {
		t.Error("position 0 should be liquidated")
	}
	if c.GetPosition(1) != nil {
		t.Error("position 1 should be liquidated")
	}
}

func TestClient_RequestMarketLiquidationClearsPosition(t *testing.T) {
	src := newFakeOrderSource()
	d0 := defaultDetail(0, "ETH", 3000.0)
	d0.MaintenanceMarginFraction = 500
	d0.CloseoutMarginFraction = 250
	src.details[0] = d0
	src.books[0] = book([][2]string{lvl("10.00", "10.0")}, [][2]string{lvl("9.00", "10.0")})
	c := newTrackedClient(t, 5.0, src, nil, 0)

	ApplyFill(c.account, 0, SideBuy, 1.0, 100.0, 0, false)

	if _, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 0.1}); err != nil {
		t.Fatalf("order: %v", err)
	}
	if c.GetPosition(0) != nil {
		t.Error("position should be liquidated")
	}
	if !c.GetHealth().HasBeenLiquidated {
		t.Error("has_been_liquidated should be true")
	}
}

func TestClient_MarkPriceFallbackToLastTrade(t *testing.T) {
	src := newFakeOrderSource()
	src.books[0] = book(nil, nil)
	c := newTrackedClient(t, 10000.0, src, nil, 0)

	ApplyFill(c.account, 0, SideBuy, 1.0, 3000.0, 0.0, false)

	pv, err := c.GetPortfolioValue()
	if err != nil {
		t.Fatalf("portfolio value: %v", err)
	}
	almostEqual(t, pv, 10000.0, "portfolio value")
}

func TestClient_RefreshOrderBookUpdatesUnrealizedPnl(t *testing.T) {
	src := newFakeOrderSource()
	src.books[0] = book([][2]string{lvl("3000.00", "2.0")}, [][2]string{lvl("2999.00", "2.0")})
	c := newTrackedClient(t, 10000.0, src, nil, 0)

	if _, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 1.0}); err != nil {
		t.Fatalf("order: %v", err)
	}
	src.books[0] = book([][2]string{lvl("3100.00", "2.0")}, [][2]string{lvl("3099.00", "2.0")})
	if err := c.RefreshOrderBook(0); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	pos := c.GetPosition(0)
	if pos == nil {
		t.Fatal("expected position")
	}
	almostEqual(t, pos.UnrealizedPnl, 99.5, "unrealized pnl")
}

func TestClient_PremiumTierAppliesFees(t *testing.T) {
	src := newFakeOrderSource()
	src.books[0] = book([][2]string{lvl("3000.00", "0.5"), lvl("3001.00", "2.0")}, [][2]string{lvl("2999.00", "0.5"), lvl("2998.00", "2.0")})
	c := newTrackedClient(t, 5000.0, src, []Option{WithAccountTier(TierPremium)}, 0)

	res, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 1.0})
	if err != nil {
		t.Fatalf("order: %v", err)
	}
	expected := 0.5*3000*0.000280 + 0.5*3001*0.000280
	almostEqualPlaces(t, res.TotalFee, expected, 8, "total fee")
}

func TestClient_DefaultTierIsStandard(t *testing.T) {
	c, err := NewClient(newFakeOrderSource(), 5000.0)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.AccountTier() != TierStandard {
		t.Errorf("tier = %v, want standard", c.AccountTier())
	}
}

func TestClient_RepeatedOrdersProduceIdenticalFills(t *testing.T) {
	src := newFakeOrderSource()
	src.books[0] = book([][2]string{lvl("3000.00", "0.5"), lvl("3001.00", "2.0")}, [][2]string{lvl("2999.00", "1.0")})
	c := newTrackedClient(t, 50000.0, src, nil, 0)

	r1, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 1.0})
	if err != nil {
		t.Fatalf("order 1: %v", err)
	}
	r2, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 1.0})
	if err != nil {
		t.Fatalf("order 2: %v", err)
	}
	almostEqual(t, r1.FilledSize, r2.FilledSize, "filled size")
	almostEqual(t, r1.AvgPrice, r2.AvgPrice, "avg price")
	almostEqual(t, r1.TotalFee, r2.TotalFee, "total fee")
}
