// Example: Getting order book data
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types/api"
)

func main() {
	httpClient := examples.CreateHTTPClient()

	network := examples.GetNetwork()
	fmt.Printf("Connected to %s\n\n", network.String())

	// Get order book details for market 0 (ETH-USD)
	marketIndex := int16(0)
	orderBookDetails, err := httpClient.Order().GetOrderBookDetails(context.Background(), &marketIndex, api.MarketFilterAll)
	if err != nil {
		log.Fatalf("Failed to get order book details: %v", err)
	}

	for _, ob := range orderBookDetails.PerpsOrderBooks {
		fmt.Printf("Order Book for Market %d (%s):\n", ob.MarketIndex, ob.MarketSymbol)
		fmt.Printf("  Mark Price: %s\n", ob.MarkPrice)
		fmt.Printf("  Funding Rate: %s\n", ob.FundingRate)

		fmt.Printf("\n  Top 5 Bids:\n")
		for i, bid := range ob.Bids {
			if i >= 5 {
				break
			}
			fmt.Printf("    %s @ %s\n", bid.Size, bid.Price)
		}

		fmt.Printf("\n  Top 5 Asks:\n")
		for i, ask := range ob.Asks {
			if i >= 5 {
				break
			}
			fmt.Printf("    %s @ %s\n", ask.Size, ask.Price)
		}
	}
}
