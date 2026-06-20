package paper

import "math"

func marginRequirement(account *Account, markPrices map[int]float64, configs map[int]MarketConfig, fraction func(MarketConfig) int) float64 {
	total := 0.0
	for marketID, position := range account.Positions {
		mark, okM := markPrices[marketID]
		config, okC := configs[marketID]
		if !okM || !okC {
			continue
		}
		total += math.Abs(position.Size) * mark * (float64(fraction(config)) / 10_000)
	}
	return total
}

// ComputeInitialMarginRequirement is the cross IMR over open positions.
func ComputeInitialMarginRequirement(account *Account, markPrices map[int]float64, configs map[int]MarketConfig) float64 {
	return marginRequirement(account, markPrices, configs, func(c MarketConfig) int { return c.DefaultInitialMarginFraction })
}

// ComputeMaintenanceMarginRequirement is the cross MMR over open positions.
func ComputeMaintenanceMarginRequirement(account *Account, markPrices map[int]float64, configs map[int]MarketConfig) float64 {
	return marginRequirement(account, markPrices, configs, func(c MarketConfig) int { return c.MaintenanceMarginFraction })
}

// ComputeCloseoutMarginRequirement is the cross COMR over open positions.
func ComputeCloseoutMarginRequirement(account *Account, markPrices map[int]float64, configs map[int]MarketConfig) float64 {
	return marginRequirement(account, markPrices, configs, func(c MarketConfig) int { return c.CloseoutMarginFraction })
}

// ComputeHealth evaluates cross-margin account health.
func ComputeHealth(account *Account, markPrices map[int]float64, configs map[int]MarketConfig) AccountHealth {
	tav := ComputeTotalAccountValue(account, markPrices)
	imr := ComputeInitialMarginRequirement(account, markPrices, configs)
	mmr := ComputeMaintenanceMarginRequirement(account, markPrices, configs)

	var status HealthStatus
	switch {
	case tav < 0:
		status = StatusBankruptcy
	case tav < imr:
		status = StatusPreLiquidation
	default:
		status = StatusHealthy
	}

	var marginUsage float64
	switch {
	case imr == 0:
		marginUsage = 0
	case tav > 0:
		marginUsage = imr / tav * 100
	default:
		marginUsage = math.Inf(1)
	}

	totalNotional := 0.0
	for marketID, position := range account.Positions {
		if mark, ok := markPrices[marketID]; ok {
			totalNotional += math.Abs(position.Size) * mark
		}
	}
	leverage := 0.0
	if tav > 0 {
		leverage = totalNotional / tav
	}

	return AccountHealth{
		Status:                       status,
		TotalAccountValue:            tav,
		InitialMarginRequirement:     imr,
		MaintenanceMarginRequirement: mmr,
		MarginUsage:                  marginUsage,
		Leverage:                     leverage,
		HasBeenLiquidated:            account.HasBeenLiquidated,
	}
}

// ComputeLiquidationPrice returns the mark price at which the position would be
// liquidated under cross margin, clamped to [0, mark] on the losing side.
func ComputeLiquidationPrice(account *Account, marketID int, markPrices map[int]float64, configs map[int]MarketConfig) float64 {
	position := account.Positions[marketID]
	config, okC := configs[marketID]
	mark, okM := markPrices[marketID]
	if position == nil || position.Size == 0 || !okC || !okM {
		return 0
	}

	tav := ComputeTotalAccountValue(account, markPrices)
	crossMMR := ComputeMaintenanceMarginRequirement(account, markPrices, configs)
	absSize := math.Abs(position.Size)
	mmFraction := float64(config.MaintenanceMarginFraction) / 10_000
	positionSign := 1.0
	if position.Size < 0 {
		positionSign = -1.0
	}
	denominator := absSize * (mmFraction - positionSign)
	if denominator == 0 {
		return 0
	}

	liquidationPrice := mark + (tav-crossMMR)/denominator
	if liquidationPrice < 0 {
		return 0
	}
	if position.Size > 0 && liquidationPrice >= mark {
		return mark
	}
	if position.Size < 0 && liquidationPrice <= mark {
		return mark
	}
	return liquidationPrice
}

// CheckAndLiquidate detects every position whose mark has crossed its
// liquidation price (in a single pass over the pre-liquidation state) and
// closes each at its mark price. Returns the liquidated market IDs.
func CheckAndLiquidate(account *Account, markPrices map[int]float64, configs map[int]MarketConfig) []int {
	liquidated := []int{}
	for marketID, position := range account.Positions {
		liquidationPrice := ComputeLiquidationPrice(account, marketID, markPrices, configs)
		if liquidationPrice == 0 {
			continue
		}
		mark, ok := markPrices[marketID]
		if !ok {
			continue
		}
		if position.Size > 0 && mark <= liquidationPrice {
			liquidated = append(liquidated, marketID)
		} else if position.Size < 0 && mark >= liquidationPrice {
			liquidated = append(liquidated, marketID)
		}
	}

	for _, marketID := range liquidated {
		position := account.Positions[marketID]
		if position == nil {
			continue
		}
		mark := markPrices[marketID]
		closeSide := SideBuy
		if position.Size > 0 {
			closeSide = SideSell
		}
		ApplyFill(account, marketID, closeSide, math.Abs(position.Size), mark, 0, true)
	}

	return liquidated
}

// UpdatePositionMetrics refreshes mark price, unrealized PnL and liquidation
// price on each open position.
func UpdatePositionMetrics(account *Account, markPrices map[int]float64, configs map[int]MarketConfig) {
	for marketID, position := range account.Positions {
		mark, ok := markPrices[marketID]
		if !ok {
			continue
		}
		position.MarkPrice = mark
		position.UnrealizedPnl = ComputeUnrealizedPnl(position, mark)
		position.LiquidationPrice = ComputeLiquidationPrice(account, marketID, markPrices, configs)
	}
}
