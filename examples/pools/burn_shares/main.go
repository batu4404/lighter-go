// Example: Burning shares in a public pool (withdrawal)
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

	// Burn shares in a public pool (withdraw)
	req := &types.BurnSharesTxReq{
		PublicPoolIndex: 1,     // Pool to burn shares from
		ShareAmount:     50000, // Number of shares to burn
	}

	txInfo, err := signerClient.GetBurnSharesTransaction(context.Background(), req, nil)
	if err != nil {
		log.Fatalf("Failed to create burn shares transaction: %v", err)
	}

	fmt.Println("Burn Shares Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Pool Index: %d\n", req.PublicPoolIndex)
	fmt.Printf("  Shares to Burn: %d\n", req.ShareAmount)
	fmt.Println()
	fmt.Println("Burning shares withdraws your proportional share")
	fmt.Println("of the pool's assets back to your account.")

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
