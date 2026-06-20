// Example: Creating a sub-account
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
	apiKeyIndex := uint8(0)
	accountIndex := int64(1)

	signerClient, err := examples.CreateSignerClient(privateKey, apiKeyIndex, accountIndex)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	network := examples.GetNetwork()
	fmt.Printf("Connected to %s (chain ID: %d)\n", network.String(), network.ChainID())

	// Create a new sub-account
	// Sub-accounts allow you to isolate trading activity
	txInfo, err := signerClient.GetCreateSubAccountTransaction(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to create sub-account transaction: %v", err)
	}

	fmt.Println("Create Sub-Account Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Println()
	fmt.Println("Sub-accounts allow you to:")
	fmt.Println("  - Isolate trading strategies")
	fmt.Println("  - Separate margin requirements")
	fmt.Println("  - Track P&L independently")

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
