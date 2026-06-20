package paper

import "testing"

// These mirror test/paper_client/test_live.py. The Python tests drive a real
// WebSocket server feeding the client's listener; here we feed the same
// snapshots and deltas directly through ApplyLiveSnapshot / ApplyLiveDelta,
// which is exactly what a live listener calls internally. The engine behaviour
// under test (tombstone sorting, mark-driven liquidation, cross-margin cascade)
// is identical.

func snapshot(asks, bids [][2]string, offset int) BookData {
	d := book(asks, bids)
	d.Offset = ptr(offset)
	return d
}

func TestLive_TrackMarketAppliesSnapshot(t *testing.T) {
	src := newFakeOrderSource()
	c, err := NewClient(src, 5000.0)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := c.TrackMarket(0); err != nil {
		t.Fatalf("TrackMarket: %v", err)
	}
	c.ApplyLiveSnapshot(0, snapshot(
		[][2]string{lvl("3002.00", "2.0"), lvl("3000.00", "1.0")},
		[][2]string{lvl("2998.00", "1.5"), lvl("2999.00", "1.0")},
		11,
	))

	b := c.orderBooks[0]
	assertPrices(t, b.Asks, []string{"3000.00", "3002.00"}, "asks")
	assertPrices(t, b.Bids, []string{"2999.00", "2998.00"}, "bids")
	if b.Offset == nil || *b.Offset != 11 {
		t.Errorf("offset = %v, want 11", b.Offset)
	}
}

func TestLive_UpdatesTombstonesSortingAndOffset(t *testing.T) {
	src := newFakeOrderSource()
	c, _ := NewClient(src, 5000.0)
	if err := c.TrackMarket(0); err != nil {
		t.Fatalf("TrackMarket: %v", err)
	}
	c.ApplyLiveSnapshot(0, snapshot(
		[][2]string{lvl("3000.00", "1.0"), lvl("3002.00", "1.0")},
		[][2]string{lvl("2999.00", "1.0"), lvl("2998.00", "1.0")},
		11,
	))
	c.ApplyLiveDelta(0, snapshot(
		[][2]string{lvl("3000.00", "0.0000"), lvl("2997.00", "0.5")},
		[][2]string{lvl("2998.00", "0.0"), lvl("3001.00", "0.25")},
		12,
	))

	b := c.orderBooks[0]
	wantAsks := []OrderBookLevel{{"2997.00", "0.5"}, {"3002.00", "1.0"}}
	if prices(b.Asks)[0] != "2997.00" || prices(b.Asks)[1] != "3002.00" {
		t.Errorf("asks = %v, want %v", b.Asks, wantAsks)
	}
	wantBids := []OrderBookLevel{{"3001.00", "0.25"}, {"2999.00", "1.0"}}
	if prices(b.Bids)[0] != "3001.00" || prices(b.Bids)[1] != "2999.00" {
		t.Errorf("bids = %v, want %v", b.Bids, wantBids)
	}
	if b.Offset == nil || *b.Offset != 12 {
		t.Errorf("offset = %v, want 12", b.Offset)
	}
}

func TestLive_UpdateLiquidatesPosition(t *testing.T) {
	src := newFakeOrderSource()
	d0 := defaultDetail(0, "ETH", 3000.0)
	d0.MaintenanceMarginFraction = 500
	d0.CloseoutMarginFraction = 250
	src.details[0] = d0
	c, _ := NewClient(src, 5.0)
	if err := c.TrackMarket(0); err != nil {
		t.Fatalf("TrackMarket: %v", err)
	}
	c.ApplyLiveSnapshot(0, snapshot([][2]string{lvl("100.00", "10.0")}, [][2]string{lvl("99.00", "10.0")}, 1))
	ApplyFill(c.account, 0, SideBuy, 1.0, 100.0, 0, false)

	c.ApplyLiveDelta(0, snapshot(
		[][2]string{lvl("100.00", "0"), lvl("10.00", "10.0")},
		[][2]string{lvl("99.00", "0.0000"), lvl("9.00", "10.0")},
		2,
	))

	if c.GetPosition(0) != nil {
		t.Error("position should be liquidated")
	}
	trades := c.GetTrades()
	if !trades[len(trades)-1].IsLiquidation {
		t.Error("last trade should be a liquidation")
	}
}

func TestLive_MultiMarketTradingAndMetrics(t *testing.T) {
	src := newFakeOrderSource()
	d1 := defaultDetail(1, "BTC", 60000.0)
	d1.MinBaseAmount = 0.0001
	src.details[1] = d1
	c, _ := NewClient(src, 10000.0)
	if err := c.TrackMarket(0); err != nil {
		t.Fatalf("TrackMarket 0: %v", err)
	}
	if err := c.TrackMarket(1); err != nil {
		t.Fatalf("TrackMarket 1: %v", err)
	}
	c.ApplyLiveSnapshot(0, snapshot([][2]string{lvl("3000.00", "5.0")}, [][2]string{lvl("2999.00", "5.0")}, 1))
	c.ApplyLiveSnapshot(1, snapshot([][2]string{lvl("60000.00", "1.0")}, [][2]string{lvl("59999.00", "1.0")}, 1))

	if _, err := c.CreatePaperOrder(OrderRequest{MarketID: 0, Side: SideBuy, BaseAmount: 1.0}); err != nil {
		t.Fatalf("order 0: %v", err)
	}
	if _, err := c.CreatePaperOrder(OrderRequest{MarketID: 1, Side: SideBuy, BaseAmount: 0.01}); err != nil {
		t.Fatalf("order 1: %v", err)
	}
	almostEqual(t, c.GetPosition(0).Size, 1.0, "pos 0")
	almostEqual(t, c.GetPosition(1).Size, 0.01, "pos 1")

	c.ApplyLiveDelta(0, snapshot(
		[][2]string{lvl("3000.00", "0"), lvl("3100.00", "5.0")},
		[][2]string{lvl("2999.00", "0"), lvl("3099.00", "5.0")},
		2,
	))
	pos0 := c.GetPosition(0)
	almostEqual(t, pos0.MarkPrice, 3099.5, "mark after delta 1")
	almostEqual(t, pos0.UnrealizedPnl, 99.5, "upnl after delta 1")

	c.ApplyLiveDelta(0, snapshot(
		[][2]string{lvl("3100.00", "0"), lvl("3200.00", "5.0")},
		[][2]string{lvl("3099.00", "0"), lvl("3199.00", "5.0")},
		3,
	))
	pos0 = c.GetPosition(0)
	almostEqual(t, pos0.MarkPrice, 3199.5, "mark after delta 2")
	almostEqual(t, pos0.UnrealizedPnl, 199.5, "upnl after delta 2")

	h := c.GetHealth()
	if h.Status != StatusHealthy {
		t.Errorf("status = %v, want healthy", h.Status)
	}
	if h.Leverage <= 0 {
		t.Errorf("leverage = %v, want > 0", h.Leverage)
	}
}

func TestLive_LiquidationStateInspection(t *testing.T) {
	src := newFakeOrderSource()
	d0 := defaultDetail(0, "ETH", 100.0)
	d0.MaintenanceMarginFraction = 500
	d0.CloseoutMarginFraction = 250
	src.details[0] = d0
	src.details[1] = defaultDetail(1, "ALT", 10.0)
	c, _ := NewClient(src, 20.0)
	if err := c.TrackMarket(0); err != nil {
		t.Fatalf("TrackMarket 0: %v", err)
	}
	if err := c.TrackMarket(1); err != nil {
		t.Fatalf("TrackMarket 1: %v", err)
	}
	c.ApplyLiveSnapshot(0, snapshot([][2]string{lvl("100.00", "10.0")}, [][2]string{lvl("99.00", "10.0")}, 1))
	c.ApplyLiveSnapshot(1, snapshot([][2]string{lvl("10.00", "10.0")}, [][2]string{lvl("9.99", "10.0")}, 1))

	ApplyFill(c.account, 0, SideBuy, 1.0, 100.0, 0, false)
	ApplyFill(c.account, 1, SideBuy, 1.0, 10.0, 0, false)
	collateralBefore := c.GetCollateral()

	c.ApplyLiveDelta(0, snapshot(
		[][2]string{lvl("100.00", "0"), lvl("84.00", "10.0")},
		[][2]string{lvl("99.00", "0"), lvl("82.00", "10.0")},
		2,
	))

	if c.GetPosition(0) != nil || c.GetPosition(1) != nil {
		t.Error("both positions should be liquidated in the cross-margin cascade")
	}
	if c.GetCollateral() >= collateralBefore {
		t.Errorf("collateral = %v, want < %v", c.GetCollateral(), collateralBefore)
	}

	var liqTrades []Trade
	markets := map[int]bool{}
	for _, tr := range c.GetTrades() {
		if tr.IsLiquidation {
			liqTrades = append(liqTrades, tr)
			markets[tr.MarketID] = true
		}
	}
	if len(liqTrades) != 2 {
		t.Fatalf("liquidation trades = %d, want 2", len(liqTrades))
	}
	if !markets[0] || !markets[1] {
		t.Errorf("liquidated markets = %v, want {0,1}", markets)
	}
	for _, tr := range liqTrades {
		if tr.MarketID == 0 && tr.RealizedPnl >= -10 {
			t.Errorf("market 0 liq realized pnl = %v, want < -10", tr.RealizedPnl)
		}
	}

	h := c.GetHealth()
	if h.TotalAccountValue <= 0 {
		t.Errorf("tav = %v, want > 0", h.TotalAccountValue)
	}
	almostEqual(t, h.Leverage, 0.0, "leverage")
}
