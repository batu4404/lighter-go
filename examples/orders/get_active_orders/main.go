// Example: Getting active orders
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/examples"
)

func main() {
	privateKey := examples.GetPrivateKey()
	if privateKey == "" {
		log.Fatal("LIGHTER_PRIVATE_KEY environment variable not set")
	}

	// Create signer client (uses LIGHTER_NETWORK env var, defaults to mainnet)
	signerClient, err := examples.CreateSignerClient(privateKey, 0, 0)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	network := examples.GetNetwork()
	fmt.Printf("Connected to %s (chain ID: %d)\n\n", network.String(), network.ChainID())

	// Get all active orders across all markets
	fmt.Println("Fetching active orders for all markets...")
	orders, err := signerClient.GetOpenOrders(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to get open orders: %v", err)
	}

	fmt.Printf("Found %d active orders\n\n", len(orders.Orders))

	if len(orders.Orders) == 0 {
		fmt.Println("No active orders")
	}

	for i, order := range orders.Orders {
		fmt.Printf("Order #%d:\n", i+1)
		fmt.Printf("  Index: %d\n", order.Index)
		fmt.Printf("  Market: %d\n", order.MarketIndex)
		fmt.Printf("  Side: %s\n", order.Side)
		fmt.Printf("  Type: %s\n", order.Type)
		fmt.Printf("  Price: %s\n", order.Price)
		fmt.Printf("  Size: %s\n", order.Size)
		fmt.Printf("  Filled Size: %s\n", order.FilledSize)
		fmt.Printf("  Remaining: %s\n", order.RemainingSize)
		fmt.Printf("  Status: %s\n", order.Status)
		fmt.Printf("  Created: %d\n", order.CreatedAt)
		fmt.Println()
	}

	// Get orders for a specific market (ETH-USD, market 0)
	fmt.Println("Fetching active orders for ETH-USD market only...")
	marketIndex := int16(0)
	ethOrders, err := signerClient.GetOpenOrders(context.Background(), &marketIndex)
	if err != nil {
		log.Fatalf("Failed to get ETH-USD orders: %v", err)
	}

	fmt.Printf("Found %d active orders for ETH-USD\n", len(ethOrders.Orders))
}
