// Example: Creating a market order
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types"
)

func main() {
	// Get configuration from environment
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
	fmt.Printf("Connected to %s (chain ID: %d)\n", network.String(), network.ChainID())

	// Create a market buy order
	// Parameters: marketIndex, size, isBuy, opts
	marketIndex := int16(0) // ETH-USD perp
	size := int64(1000000)  // 0.01 ETH (scaled)
	isBuy := true

	opts := &types.TransactOpts{
		Nonce: types.NewInt64(-1), // Auto-fetch nonce
	}

	txInfo, err := signerClient.CreateMarketOrder(context.Background(), marketIndex, size, isBuy, opts)
	if err != nil {
		log.Fatalf("Failed to create market order: %v", err)
	}

	fmt.Printf("Market order created!\n")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Market: %d\n", marketIndex)
	fmt.Printf("  Side: %s\n", map[bool]string{true: "BUY", false: "SELL"}[isBuy])
	fmt.Printf("  Size: %d\n", size)

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit order: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
