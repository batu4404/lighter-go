// Example: Getting candlestick (OHLCV) data
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types/api"
)

func main() {
	network := examples.GetNetwork()
	httpClient := examples.CreateHTTPClient()

	fmt.Printf("Connected to %s\n", network.String())

	// Get candlestick data for market 0 (ETH-USD)
	marketIndex := int16(0)
	resolution := api.Resolution1h // 1 hour candles
	countBack := 24                // Last 24 candles

	// TimestampRange is optional - use empty for recent data
	timestamps := api.TimestampRange{
		StartTimestamp: 0,
		EndTimestamp:   0,
	}

	fmt.Printf("Fetching last %d 1h candles for market %d...\n\n", countBack, marketIndex)

	candles, err := httpClient.Candlestick().GetCandlesticks(context.Background(), marketIndex, resolution, timestamps, countBack)
	if err != nil {
		log.Fatalf("Failed to get candlesticks: %v", err)
	}

	fmt.Printf("Retrieved %d candles\n", len(candles.Candlesticks))
	fmt.Println()

	fmt.Printf("%-20s %-12s %-12s %-12s %-12s %-15s\n",
		"Time", "Open", "High", "Low", "Close", "Volume")
	fmt.Println("--------------------------------------------------------------------------------------")

	for _, candle := range candles.Candlesticks {
		ts := time.UnixMilli(candle.Timestamp)
		fmt.Printf("%-20s %-12.6f %-12.6f %-12.6f %-12.6f %-15.6f\n",
			ts.Format("2006-01-02 15:04"),
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume)
	}

	// Demonstrate available resolutions
	fmt.Println("\nAvailable resolutions: 1m, 5m, 15m, 30m, 1h, 4h, 1D, 1W")
}
