package paper

import "testing"

func TestApplyFill_OpenNewLong(t *testing.T) {
	a := newAccount(10000.0)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0.5, false)
	pos := a.Positions[0]
	almostEqual(t, pos.Size, 1.0, "size")
	almostEqual(t, pos.EntryQuote, 100.0, "entry_quote")
	almostEqual(t, a.Collateral, 10000.0-0.5, "collateral")
}

func TestApplyFill_OpenNewShort(t *testing.T) {
	a := newAccount(10000.0)
	ApplyFill(a, 0, SideSell, 1.0, 100.0, 0.5, false)
	pos := a.Positions[0]
	almostEqual(t, pos.Size, -1.0, "size")
	almostEqual(t, pos.EntryQuote, 100.0, "entry_quote")
	almostEqual(t, a.Collateral, 10000.0-0.5, "collateral")
}

func TestApplyFill_IncreaseExistingLong(t *testing.T) {
	a := newAccount(10000.0)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0.0, false)
	ApplyFill(a, 0, SideBuy, 1.0, 110.0, 0.0, false)
	pos := a.Positions[0]
	almostEqual(t, pos.Size, 2.0, "size")
	almostEqual(t, pos.EntryQuote, 210.0, "entry_quote")
}

func TestApplyFill_PartialCloseLong(t *testing.T) {
	a := newAccount(10000.0)
	ApplyFill(a, 0, SideBuy, 2.0, 100.0, 0.0, false)
	pnl := ApplyFill(a, 0, SideSell, 1.0, 120.0, 0.0, false)
	pos := a.Positions[0]
	almostEqual(t, pnl, 20.0, "pnl")
	almostEqual(t, pos.EntryQuote, 100.0, "entry_quote")
	almostEqual(t, a.Collateral, 10000.0+20.0, "collateral")
}

func TestApplyFill_PartialCloseShort(t *testing.T) {
	a := newAccount(10000.0)
	ApplyFill(a, 0, SideSell, 2.0, 100.0, 0.0, false)
	pnl := ApplyFill(a, 0, SideBuy, 1.0, 80.0, 0.0, false)
	almostEqual(t, pnl, 20.0, "pnl")
}

func TestApplyFill_FlipLongToShort(t *testing.T) {
	a := newAccount(10000.0)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0.0, false)
	pnl := ApplyFill(a, 0, SideSell, 3.0, 120.0, 0.0, false)
	pos := a.Positions[0]
	almostEqual(t, pnl, 20.0, "pnl")
	almostEqual(t, pos.Size, -2.0, "size")
	almostEqual(t, pos.EntryQuote, 240.0, "entry_quote")
}

func TestApplyFill_FullCloseRemovesPosition(t *testing.T) {
	a := newAccount(10000.0)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0.0, false)
	pnl := ApplyFill(a, 0, SideSell, 1.0, 110.0, 0.0, false)
	if _, ok := a.Positions[0]; ok {
		t.Error("position should be removed after full close")
	}
	almostEqual(t, pnl, 10.0, "pnl")
}

func TestComputeUnrealizedPnl_Long(t *testing.T) {
	pos := &Position{MarketID: 0, Size: 1.0, EntryQuote: 100.0}
	almostEqual(t, ComputeUnrealizedPnl(pos, 120.0), 20.0, "long +")
	almostEqual(t, ComputeUnrealizedPnl(pos, 80.0), -20.0, "long -")
}

func TestComputeUnrealizedPnl_Short(t *testing.T) {
	pos := &Position{MarketID: 0, Size: -1.0, EntryQuote: 100.0}
	almostEqual(t, ComputeUnrealizedPnl(pos, 80.0), 20.0, "short +")
	almostEqual(t, ComputeUnrealizedPnl(pos, 120.0), -20.0, "short -")
}

func TestComputeTotalAccountValue(t *testing.T) {
	a := newAccount(1000.0)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0.0, false)
	ApplyFill(a, 1, SideSell, 1.0, 200.0, 0.0, false)
	tav := ComputeTotalAccountValue(a, map[int]float64{0: 110.0, 1: 190.0})
	almostEqual(t, tav, 1020.0, "tav")
}
