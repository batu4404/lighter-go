// Example: Cancelling orders
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types"
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

	// First, get active orders to see what we can cancel
	orders, err := signerClient.GetOpenOrders(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to get open orders: %v", err)
	}

	fmt.Printf("Found %d open orders\n", len(orders.Orders))

	if len(orders.Orders) == 0 {
		fmt.Println("No orders to cancel")
		return
	}

	// List all open orders
	for _, order := range orders.Orders {
		fmt.Printf("  Order %d: %s side, price=%s, size=%s, filled=%s\n",
			order.Index,
			order.Side,
			order.Price,
			order.Size,
			order.FilledSize,
		)
	}

	// Cancel all orders
	nonce := int64(-1)
	opts := &types.TransactOpts{
		Nonce: &nonce,
	}

	txInfo, err := signerClient.CancelAllOrders(context.Background(), opts)
	if err != nil {
		log.Fatalf("Failed to create cancel all orders transaction: %v", err)
	}

	fmt.Printf("\nCancel all orders transaction created!\n")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit cancel all: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
	fmt.Printf("  All orders cancelled!\n")
}
