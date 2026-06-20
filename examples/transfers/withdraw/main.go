// Example: Withdrawing funds from Lighter
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

	// Create a withdrawal request
	req := &types.WithdrawTxReq{
		AssetIndex: 0,         // 0 = USDC
		RouteType:  0,         // Default route
		Amount:     50_000000, // 50 USDC (6 decimals)
	}

	txInfo, err := signerClient.GetWithdrawTransaction(context.Background(), req, nil)
	if err != nil {
		log.Fatalf("Failed to create withdraw transaction: %v", err)
	}

	fmt.Println("Withdraw Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Asset: USDC\n")
	fmt.Printf("  Amount: %.2f USDC\n", float64(req.Amount)/1_000000)
	fmt.Println()
	fmt.Println("NOTE: Withdrawals are processed after L1 confirmation.")

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
