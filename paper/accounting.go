package paper

import "math"

// ApplyFill applies a single fill to the account: updates the position size and
// entry quote, realizes PnL on reducing/closing/flipping, adjusts collateral by
// realized PnL minus fee, records a trade, and removes the position if it
// becomes flat. Returns the realized PnL for this fill.
func ApplyFill(account *Account, marketID int, side OrderSide, fillSize, fillPrice, fee float64, isLiquidation bool) float64 {
	position := account.Positions[marketID]
	if position == nil {
		position = &Position{MarketID: marketID}
		account.Positions[marketID] = position
	}

	oldPosition := position.Size
	oldAbsSize := math.Abs(oldPosition)
	oldEntryQuote := position.EntryQuote

	positionDelta := fillSize
	if side == SideSell {
		positionDelta = -fillSize
	}
	newPosition := oldPosition + positionDelta
	newAbsSize := math.Abs(newPosition)

	oldSign := sign(oldPosition)
	newSign := sign(newPosition)
	realizedPnl := 0.0
	newEntryQuote := 0.0

	switch {
	case oldSign == 0:
		newEntryQuote = fillSize * fillPrice
	case oldSign != newSign && newSign != 0:
		realizedPnl = fullClosePnl(oldSign, oldAbsSize, oldEntryQuote, fillPrice)
		newEntryQuote = newAbsSize * fillPrice
	case oldSign != newSign && newSign == 0:
		realizedPnl = fullClosePnl(oldSign, oldAbsSize, oldEntryQuote, fillPrice)
	case newAbsSize > oldAbsSize:
		newEntryQuote = oldEntryQuote + fillSize*fillPrice
	default:
		if oldAbsSize > 0 {
			newEntryQuote = oldEntryQuote * (newAbsSize / oldAbsSize)
			closedSize := oldAbsSize - newAbsSize
			avgEntry := oldEntryQuote / oldAbsSize
			if oldSign > 0 {
				realizedPnl = closedSize*fillPrice - closedSize*avgEntry
			} else {
				realizedPnl = closedSize*avgEntry - closedSize*fillPrice
			}
		}
	}

	position.Size = newPosition
	position.EntryQuote = newEntryQuote
	if newAbsSize > 0 {
		position.AvgEntryPrice = newEntryQuote / newAbsSize
	} else {
		position.AvgEntryPrice = 0
	}
	position.RealizedPnl += realizedPnl

	account.Collateral += realizedPnl
	account.Collateral -= fee
	account.Trades = append(account.Trades, Trade{
		MarketID:      marketID,
		Side:          side,
		Size:          fillSize,
		Price:         fillPrice,
		Fee:           fee,
		RealizedPnl:   realizedPnl,
		IsLiquidation: isLiquidation,
	})
	if isLiquidation {
		account.HasBeenLiquidated = true
	}

	if math.Abs(newPosition) < 1e-12 {
		delete(account.Positions, marketID)
	}

	return realizedPnl
}

// ComputeUnrealizedPnl returns mark-to-market PnL for a position.
func ComputeUnrealizedPnl(position *Position, markPrice float64) float64 {
	if position == nil || position.Size == 0 {
		return 0
	}
	positionValue := math.Abs(position.Size) * markPrice
	if position.Size > 0 {
		return positionValue - position.EntryQuote
	}
	return position.EntryQuote - positionValue
}

// ComputeTotalAccountValue is collateral plus unrealized PnL across positions
// that have a mark price.
func ComputeTotalAccountValue(account *Account, markPrices map[int]float64) float64 {
	total := account.Collateral
	for marketID, position := range account.Positions {
		if mark, ok := markPrices[marketID]; ok {
			total += ComputeUnrealizedPnl(position, mark)
		}
	}
	return total
}

func sign(value float64) int {
	if value > 0 {
		return 1
	}
	if value < 0 {
		return -1
	}
	return 0
}

func fullClosePnl(oldSign int, oldAbsSize, oldEntryQuote, fillPrice float64) float64 {
	if oldSign > 0 {
		return oldAbsSize*fillPrice - oldEntryQuote
	}
	return oldEntryQuote - oldAbsSize*fillPrice
}
