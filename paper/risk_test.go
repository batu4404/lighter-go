package paper

import (
	"math"
	"testing"
)

func TestRisk_MarginRequirements(t *testing.T) {
	a := newAccount(1000)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0, false)
	c := map[int]MarketConfig{0: cfg(0, 1000, 500, 250)}
	mp := map[int]float64{0: 100.0}
	almostEqual(t, ComputeInitialMarginRequirement(a, mp, c), 10.0, "imr")
	almostEqual(t, ComputeMaintenanceMarginRequirement(a, mp, c), 5.0, "mmr")
	almostEqual(t, ComputeCloseoutMarginRequirement(a, mp, c), 2.5, "comr")
}

func TestRisk_HealthHealthy(t *testing.T) {
	a := newAccount(1000)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0, false)
	h := ComputeHealth(a, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	if h.Status != StatusHealthy {
		t.Errorf("status = %v, want healthy", h.Status)
	}
}

func TestRisk_HealthPreLiquidation(t *testing.T) {
	a := newAccount(9)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0, false)
	h := ComputeHealth(a, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	if h.Status != StatusPreLiquidation {
		t.Errorf("status = %v, want pre-liquidation", h.Status)
	}
}

func TestRisk_HealthBankruptcy(t *testing.T) {
	a := newAccount(1)
	ApplyFill(a, 0, SideBuy, 1.0, 200.0, 0, false)
	h := ComputeHealth(a, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	if h.Status != StatusBankruptcy {
		t.Errorf("status = %v, want bankruptcy", h.Status)
	}
}

func TestRisk_MarginUsageInfWhenUnderwaterWithPosition(t *testing.T) {
	a := newAccount(1)
	ApplyFill(a, 0, SideBuy, 1.0, 200.0, 0, false)
	h := ComputeHealth(a, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	if !math.IsInf(h.MarginUsage, 1) {
		t.Errorf("margin usage = %v, want +Inf", h.MarginUsage)
	}
	almostEqual(t, h.Leverage, 0.0, "leverage")
}

func TestRisk_MarginUsageZeroWhenNoPositions(t *testing.T) {
	a := newAccount(1000)
	h := ComputeHealth(a, map[int]float64{}, map[int]MarketConfig{})
	if h.Status != StatusHealthy {
		t.Errorf("status = %v, want healthy", h.Status)
	}
	almostEqual(t, h.MarginUsage, 0.0, "margin usage")

	a2 := newAccount(1)
	ApplyFill(a2, 0, SideBuy, 1.0, 200.0, 10, false)
	ApplyFill(a2, 0, SideSell, 1.0, 100.0, 10, false)
	h2 := ComputeHealth(a2, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	almostEqual(t, h2.MarginUsage, 0.0, "margin usage after wipe")
}

func TestRisk_LiquidationPriceLong(t *testing.T) {
	a := newAccount(15)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0, false)
	lp := ComputeLiquidationPrice(a, 0, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	if !(lp > 0.0 && lp < 100.0) {
		t.Errorf("liq price = %v, want in (0,100)", lp)
	}
}

func TestRisk_LiquidationPriceShort(t *testing.T) {
	a := newAccount(100)
	ApplyFill(a, 0, SideSell, 1.0, 100.0, 0, false)
	lp := ComputeLiquidationPrice(a, 0, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	if lp <= 100.0 {
		t.Errorf("liq price = %v, want > 100", lp)
	}
}

func TestRisk_LiquidationPriceClampedNegative(t *testing.T) {
	a := newAccount(10_000_000)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0, false)
	lp := ComputeLiquidationPrice(a, 0, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	almostEqual(t, lp, 0.0, "liq price")
}

func TestRisk_LiquidationPriceCappedAtMark(t *testing.T) {
	a := newAccount(2)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0, false)
	lp := ComputeLiquidationPrice(a, 0, map[int]float64{0: 100.0}, map[int]MarketConfig{0: defaultCfg()})
	almostEqual(t, lp, 100.0, "liq price")
}

func TestRisk_LiquidationPriceNoPosition(t *testing.T) {
	a := newAccount(1000)
	lp := ComputeLiquidationPrice(a, 99, map[int]float64{99: 100.0}, map[int]MarketConfig{99: cfg(99, 1000, 500, 250)})
	almostEqual(t, lp, 0.0, "liq price")
}

func TestRisk_CheckAndLiquidateShort(t *testing.T) {
	a := newAccount(5)
	ApplyFill(a, 0, SideSell, 1.0, 100.0, 0, false)
	liquidated := CheckAndLiquidate(a, map[int]float64{0: 200.0}, map[int]MarketConfig{0: defaultCfg()})
	if len(liquidated) != 1 || liquidated[0] != 0 {
		t.Errorf("liquidated = %v, want [0]", liquidated)
	}
	if _, ok := a.Positions[0]; ok {
		t.Error("position should be removed")
	}
	if !a.Trades[len(a.Trades)-1].IsLiquidation {
		t.Error("last trade should be a liquidation")
	}
}

func TestRisk_LiquidationScenario(t *testing.T) {
	a := newAccount(1000)
	c := map[int]MarketConfig{0: cfg(0, 1000, 500, 250)}
	ApplyFill(a, 0, SideBuy, 0.05, 100_000.0, 0, false)

	h0 := ComputeHealth(a, map[int]float64{0: 100_000.0}, c)
	if h0.Status != StatusHealthy || h0.HasBeenLiquidated {
		t.Errorf("T0 health = %+v", h0)
	}
	almostEqual(t, h0.TotalAccountValue, 1000.0, "T0 tav")
	almostEqual(t, h0.MarginUsage, 50.0, "T0 margin usage")

	h1 := ComputeHealth(a, map[int]float64{0: 92_000.0}, c)
	if h1.Status != StatusHealthy {
		t.Errorf("T1 status = %v", h1.Status)
	}

	h2 := ComputeHealth(a, map[int]float64{0: 88_000.0}, c)
	if h2.Status != StatusPreLiquidation {
		t.Errorf("T2 status = %v, want pre-liquidation", h2.Status)
	}
	if h2.MarginUsage <= 100.0 {
		t.Errorf("T2 margin usage = %v, want > 100", h2.MarginUsage)
	}

	liquidated := CheckAndLiquidate(a, map[int]float64{0: 84_000.0}, c)
	if len(liquidated) != 1 || liquidated[0] != 0 {
		t.Errorf("liquidated = %v, want [0]", liquidated)
	}
	if _, ok := a.Positions[0]; ok {
		t.Error("position should be gone")
	}
	if !a.Trades[len(a.Trades)-1].IsLiquidation {
		t.Error("last trade should be liquidation")
	}

	h3 := ComputeHealth(a, map[int]float64{0: 84_000.0}, c)
	if h3.Status != StatusHealthy || !h3.HasBeenLiquidated {
		t.Errorf("T3 health = %+v", h3)
	}
	almostEqual(t, h3.InitialMarginRequirement, 0.0, "T3 imr")
	almostEqual(t, h3.MaintenanceMarginRequirement, 0.0, "T3 mmr")
	almostEqual(t, h3.MarginUsage, 0.0, "T3 margin usage")
	if h3.TotalAccountValue >= 1000.0 {
		t.Errorf("T3 tav = %v, want < 1000", h3.TotalAccountValue)
	}
}

func TestRisk_UpdatePositionMetrics(t *testing.T) {
	a := newAccount(20)
	ApplyFill(a, 0, SideBuy, 1.0, 100.0, 0, false)
	UpdatePositionMetrics(a, map[int]float64{0: 120.0}, map[int]MarketConfig{0: defaultCfg()})
	pos := a.Positions[0]
	almostEqual(t, pos.MarkPrice, 120.0, "mark price")
	almostEqual(t, pos.UnrealizedPnl, 20.0, "unrealized pnl")
	if pos.LiquidationPrice <= 0.0 {
		t.Errorf("liq price = %v, want > 0", pos.LiquidationPrice)
	}
}
