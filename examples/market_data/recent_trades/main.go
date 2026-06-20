// Example: Getting recent trades
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0xJord4n/lighter-go/examples"
)

func main() {
	network := examples.GetNetwork()
	httpClient := examples.CreateHTTPClient()

	fmt.Printf("Connected to %s\n", network.String())

	// Get recent trades for market 0 (ETH-USD)
	marketIndex := int16(0)
	limit := 20

	fmt.Printf("Fetching last %d trades for market %d...\n\n", limit, marketIndex)

	trades, err := httpClient.Order().GetRecentTrades(context.Background(), marketIndex, limit)
	if err != nil {
		log.Fatalf("Failed to get recent trades: %v", err)
	}

	fmt.Printf("Retrieved %d trades\n", len(trades.Trades))
	fmt.Println()

	fmt.Printf("%-20s %-6s %-15s %-15s\n", "Time", "Side", "Price", "Size")
	fmt.Println("--------------------------------------------------------")

	for _, trade := range trades.Trades {
		ts := time.UnixMilli(trade.Timestamp)
		fmt.Printf("%-20s %-6s %-15s %-15s\n",
			ts.Format("2006-01-02 15:04:05"),
			trade.Side,
			trade.Price,
			trade.Size)
	}

	fmt.Println()
	fmt.Printf("Total trades shown: %d\n", len(trades.Trades))
}
