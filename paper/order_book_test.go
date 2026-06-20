package paper

import (
	"reflect"
	"testing"
)

func ptr(i int) *int { return &i }

func makeRuntime() *InMemoryOrderBook {
	return NewOrderBook(
		levels(lvl("3000.00", "1.0"), lvl("3001.00", "2.0")),
		levels(lvl("2999.00", "1.5"), lvl("2998.00", "2.5")),
	)
}

func assertPrices(t *testing.T, got []OrderBookLevel, want []string, msg string) {
	t.Helper()
	if !equalStrings(prices(got), want) {
		t.Errorf("%s: prices = %v, want %v", msg, prices(got), want)
	}
}

func TestOrderBook_ApplySnapshotNormalizesAndSorts(t *testing.T) {
	b := &InMemoryOrderBook{}
	b.ApplySnapshot(BookData{
		Asks: levels(lvl("3001.00", "2.0"), lvl("3000.00", "1.0")),
		Bids: levels(lvl("2999.00", "1.5"), lvl("2998.00", "1.0")),
	})
	assertPrices(t, b.Asks, []string{"3000.00", "3001.00"}, "asks")
	if !equalStrings([]string{b.Asks[0].Size, b.Asks[1].Size}, []string{"1.0", "2.0"}) {
		t.Errorf("ask sizes = %v", []string{b.Asks[0].Size, b.Asks[1].Size})
	}
	assertPrices(t, b.Bids, []string{"2999.00", "2998.00"}, "bids")
	if b.Offset != nil {
		t.Errorf("offset = %v, want nil", b.Offset)
	}
	if b.BestAsk() == nil || b.BestAsk().Price != "3000.00" {
		t.Errorf("best ask = %v", b.BestAsk())
	}
	if b.BestBid() == nil || b.BestBid().Price != "2999.00" {
		t.Errorf("best bid = %v", b.BestBid())
	}
	mid, ok := b.MidPrice()
	if !ok || mid != 2999.5 {
		t.Errorf("mid = %v (ok=%v), want 2999.5", mid, ok)
	}
}

func TestOrderBook_DeltaRemovesForZeroString(t *testing.T) {
	b := makeRuntime()
	b.ApplyDelta(BookData{Asks: levels(lvl("3000.00", "0"))})
	assertPrices(t, b.Asks, []string{"3001.00"}, "asks")
}

func TestOrderBook_DeltaRemovesForZeroPointZero(t *testing.T) {
	b := makeRuntime()
	b.ApplyDelta(BookData{Asks: levels(lvl("3000.00", "0.0"))})
	assertPrices(t, b.Asks, []string{"3001.00"}, "asks")
}

func TestOrderBook_DeltaRemovesForZeroPointZeroZeroZero(t *testing.T) {
	b := makeRuntime()
	b.ApplyDelta(BookData{Asks: levels(lvl("3000.00", "0.0000"))})
	assertPrices(t, b.Asks, []string{"3001.00"}, "asks")
}

func TestOrderBook_DeltaIgnoresTombstoneForMissingLevel(t *testing.T) {
	b := makeRuntime()
	b.ApplyDelta(BookData{Asks: levels(lvl("9999.00", "0.0000"))})
	assertPrices(t, b.Asks, []string{"3000.00", "3001.00"}, "asks")
}

func TestOrderBook_DeltaReplacesExistingSize(t *testing.T) {
	b := makeRuntime()
	b.ApplyDelta(BookData{Asks: levels(lvl("3000.00", "2.5"))})
	if len(b.Asks) != 2 {
		t.Fatalf("asks len = %d, want 2", len(b.Asks))
	}
	if b.Asks[0].Price != "3000.00" || b.Asks[0].Size != "2.5" {
		t.Errorf("asks[0] = %v, want {3000.00 2.5}", b.Asks[0])
	}
}

func TestOrderBook_DeltaAcceptsDepthAndPriceLevels(t *testing.T) {
	b := makeRuntime()
	b.ApplyDelta(BookData{
		Asks:   levels(lvl("3001.00", "0.0000"), lvl("2999.50", "0.7")),
		Bids:   levels(lvl("2999.00", "2.25"), lvl("3000.00", "0.4")),
		Offset: ptr(43),
	})
	wantAsks := []OrderBookLevel{{"2999.50", "0.7"}, {"3000.00", "1.0"}}
	if !reflect.DeepEqual(b.Asks, wantAsks) {
		t.Errorf("asks = %v, want %v", b.Asks, wantAsks)
	}
	wantBids := []OrderBookLevel{{"3000.00", "0.4"}, {"2999.00", "2.25"}, {"2998.00", "2.5"}}
	if !reflect.DeepEqual(b.Bids, wantBids) {
		t.Errorf("bids = %v, want %v", b.Bids, wantBids)
	}
	if b.Offset == nil || *b.Offset != 43 {
		t.Errorf("offset = %v, want 43", b.Offset)
	}
}

func TestOrderBook_DeltaInsertsNewLevelsAndKeepsSorted(t *testing.T) {
	b := makeRuntime()
	b.ApplyDelta(BookData{
		Asks:   levels(lvl("2999.50", "0.7"), lvl("3005.00", "0.8")),
		Bids:   levels(lvl("2999.50", "0.6"), lvl("2997.50", "0.4")),
		Offset: ptr(42),
	})
	assertPrices(t, b.Asks, []string{"2999.50", "3000.00", "3001.00", "3005.00"}, "asks")
	assertPrices(t, b.Bids, []string{"2999.50", "2999.00", "2998.00", "2997.50"}, "bids")
	if b.Offset == nil || *b.Offset != 42 {
		t.Errorf("offset = %v, want 42", b.Offset)
	}
}

func TestOrderBook_MidPriceNoneWhenSideMissing(t *testing.T) {
	if _, ok := (&InMemoryOrderBook{}).MidPrice(); ok {
		t.Error("empty book should have no mid price")
	}
	asksOnly := NewOrderBook(levels(lvl("3000.00", "1.0")), nil)
	if _, ok := asksOnly.MidPrice(); ok {
		t.Error("asks-only book should have no mid price")
	}
	bidsOnly := NewOrderBook(nil, levels(lvl("2999.00", "1.0")))
	if _, ok := bidsOnly.MidPrice(); ok {
		t.Error("bids-only book should have no mid price")
	}
}

func TestOrderBook_SnapshotFiltersZeroSizeLevels(t *testing.T) {
	b := &InMemoryOrderBook{}
	b.ApplySnapshot(BookData{
		Asks:   levels(lvl("3000.00", "0"), lvl("3001.00", "1.25")),
		Bids:   levels(lvl("2999.00", "0.0000"), lvl("2998.00", "2.5")),
		Offset: ptr(11),
	})
	if !reflect.DeepEqual(b.Asks, []OrderBookLevel{{"3001.00", "1.25"}}) {
		t.Errorf("asks = %v", b.Asks)
	}
	if !reflect.DeepEqual(b.Bids, []OrderBookLevel{{"2998.00", "2.5"}}) {
		t.Errorf("bids = %v", b.Bids)
	}
	if b.Offset == nil || *b.Offset != 11 {
		t.Errorf("offset = %v, want 11", b.Offset)
	}
}

func TestOrderBook_DeltaPreservesBookAndUpdatesOffset(t *testing.T) {
	b := makeRuntime()
	b.ApplyDelta(BookData{Offset: ptr(77)})
	wantAsks := []OrderBookLevel{{"3000.00", "1.0"}, {"3001.00", "2.0"}}
	wantBids := []OrderBookLevel{{"2999.00", "1.5"}, {"2998.00", "2.5"}}
	if !reflect.DeepEqual(b.Asks, wantAsks) || !reflect.DeepEqual(b.Bids, wantBids) {
		t.Errorf("book changed unexpectedly: asks=%v bids=%v", b.Asks, b.Bids)
	}
	if b.Offset == nil || *b.Offset != 77 {
		t.Errorf("offset = %v, want 77", b.Offset)
	}
}

func TestOrderBook_LevelFromMapRequiresPriceAndSize(t *testing.T) {
	if _, err := LevelFromMap(map[string]string{"size": "1.0"}); err == nil {
		t.Error("missing price should error")
	}
	if _, err := LevelFromMap(map[string]string{"price": "2999.00"}); err == nil {
		t.Error("missing size should error")
	}
	l, err := LevelFromMap(map[string]string{"price": "1", "remaining_base_amount": "2"})
	if err != nil || l.Size != "2" {
		t.Errorf("remaining_base_amount alias failed: %v %v", l, err)
	}
}

func TestOrderBook_AsksStayAscendingAfterMixedUpdates(t *testing.T) {
	b := NewOrderBook(
		levels(lvl("2111.04", "304.1013"), lvl("2111.36", "474.9197"), lvl("6666.00", "1.8000")),
		nil,
	)
	b.ApplyDelta(BookData{Asks: levels(lvl("6666.01", "0.3000"), lvl("4050.00", "0.3000"))})
	assertPrices(t, b.Asks, []string{"2111.04", "2111.36", "4050.00", "6666.00", "6666.01"}, "asks")
}

func TestOrderBook_BidsStayDescendingAfterMixedUpdates(t *testing.T) {
	b := NewOrderBook(
		nil,
		levels(lvl("2103.36", "451.4887"), lvl("2102.72", "474.9231")),
	)
	b.ApplyDelta(BookData{Bids: levels(lvl("1893.31", "0.0053"), lvl("2103.00", "0.0100"))})
	assertPrices(t, b.Bids, []string{"2103.36", "2103.00", "2102.72", "1893.31"}, "bids")
}
