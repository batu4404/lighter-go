// Example: Adding or removing margin from a position
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
	apiKeyIndex := uint8(0)
	accountIndex := int64(1)

	signerClient, err := examples.CreateSignerClient(privateKey, apiKeyIndex, accountIndex)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	network := examples.GetNetwork()
	fmt.Printf("Connected to %s (chain ID: %d)\n", network.String(), network.ChainID())

	// Update margin for a position
	// Direction: 0 = Add margin, 1 = Remove margin
	req := &types.UpdateMarginTxReq{
		MarketIndex: 0,         // ETH-USD perp
		USDCAmount:  10_000000, // 10 USDC (6 decimals)
		Direction:   0,         // Add margin
	}

	txInfo, err := signerClient.GetUpdateMarginTransaction(context.Background(), req, nil)
	if err != nil {
		log.Fatalf("Failed to create update margin transaction: %v", err)
	}

	fmt.Println("Update Margin Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Market Index: %d\n", req.MarketIndex)
	fmt.Printf("  Amount: %.2f USDC\n", float64(req.USDCAmount)/1_000000)
	fmt.Printf("  Direction: %s\n", map[uint8]string{0: "Add", 1: "Remove"}[req.Direction])

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
