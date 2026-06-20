// Example: Updating a public pool
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

	// Update public pool settings
	// Status: 0 = Active, 1 = Paused, 2 = Closed
	req := &types.UpdatePublicPoolTxReq{
		PublicPoolIndex:      1,   // Pool index to update
		Status:               0,   // Active
		OperatorFee:          500, // 5% operator fee
		MinOperatorShareRate: 50,  // Minimum operator share rate
	}

	txInfo, err := signerClient.GetUpdatePublicPoolTransaction(context.Background(), req, nil)
	if err != nil {
		log.Fatalf("Failed to create update public pool transaction: %v", err)
	}

	fmt.Println("Update Public Pool Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Pool Index: %d\n", req.PublicPoolIndex)
	fmt.Printf("  Status: %s\n", map[uint8]string{0: "Active", 1: "Paused", 2: "Closed"}[req.Status])
	fmt.Printf("  Operator Fee: %.1f%%\n", float64(req.OperatorFee)/100)

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
