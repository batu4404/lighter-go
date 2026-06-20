// Example: Creating a public pool
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

	// Create a public pool
	// Public pools allow others to deposit and share in trading profits
	req := &types.CreatePublicPoolTxReq{
		OperatorFee:          1000,    // 10% operator fee (scaled by 10000)
		InitialTotalShares:   1000000, // Initial shares to mint
		MinOperatorShareRate: 100,     // Minimum operator share rate
	}

	txInfo, err := signerClient.GetCreatePublicPoolTransaction(context.Background(), req, nil)
	if err != nil {
		log.Fatalf("Failed to create public pool transaction: %v", err)
	}

	fmt.Println("Create Public Pool Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Operator Fee: %.1f%%\n", float64(req.OperatorFee)/100)
	fmt.Printf("  Initial Shares: %d\n", req.InitialTotalShares)
	fmt.Printf("  Min Operator Share Rate: %d\n", req.MinOperatorShareRate)

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
