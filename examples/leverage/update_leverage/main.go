// Example: Updating leverage for a market
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

	// Update leverage settings for a market
	// InitialMarginFraction is scaled (e.g., 1000 = 10% = 10x leverage)
	// MarginMode: 0 = Cross, 1 = Isolated
	req := &types.UpdateLeverageTxReq{
		MarketIndex:           0,   // ETH-USD perp
		InitialMarginFraction: 500, // 5% initial margin = 20x leverage
		MarginMode:            0,   // Cross margin
	}

	txInfo, err := signerClient.GetUpdateLeverageTransaction(context.Background(), req, nil)
	if err != nil {
		log.Fatalf("Failed to create update leverage transaction: %v", err)
	}

	fmt.Println("Update Leverage Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Market Index: %d\n", req.MarketIndex)
	fmt.Printf("  Initial Margin Fraction: %d (%.1f%% = %.0fx leverage)\n",
		req.InitialMarginFraction,
		float64(req.InitialMarginFraction)/100,
		10000/float64(req.InitialMarginFraction))
	fmt.Printf("  Margin Mode: %s\n", map[uint8]string{0: "Cross", 1: "Isolated"}[req.MarginMode])

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
