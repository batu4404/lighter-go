// Example: Getting current funding rates
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
	fmt.Println("Fetching current funding rates for all markets...")

	fundingRates, err := httpClient.Candlestick().GetFundingRates(context.Background())
	if err != nil {
		log.Fatalf("Failed to get funding rates: %v", err)
	}

	fmt.Printf("Retrieved funding rates for %d markets\n", len(fundingRates.FundingRates))
	fmt.Println()

	fmt.Printf("%-10s %-15s %-15s %-15s %-20s\n", "Market", "Funding Rate", "Mark Price", "Index Price", "Next Funding")
	fmt.Println("------------------------------------------------------------------------------")

	for _, fr := range fundingRates.FundingRates {
		nextFunding := time.UnixMilli(fr.NextFundingTime)
		fmt.Printf("%-10d %-15s %-15s %-15s %-20s\n",
			fr.MarketIndex,
			fr.FundingRate,
			fr.MarkPrice,
			fr.IndexPrice,
			nextFunding.Format("2006-01-02 15:04"))
	}

	fmt.Println()
	fmt.Println("Note: Funding rates are charged/credited every hour")
	fmt.Println("Positive rate = longs pay shorts, Negative rate = shorts pay longs")
}
