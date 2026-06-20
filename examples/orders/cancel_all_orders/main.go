// Example: Cancelling all orders
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/0xJord4n/lighter-go/types/txtypes"
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

	// Cancel all GTT (Good-Till-Time) orders
	// TimeInForce options:
	//   - txtypes.GoodTillTime (1): Cancel GTT orders
	//   - txtypes.ImmediateOrCancel (0): Cancel IOC orders
	//   - txtypes.PostOnly (2): Cancel post-only orders
	req := &types.CancelAllOrdersTxReq{
		TimeInForce: txtypes.GoodTillTime,
		Time:        time.Now().UnixMilli(),
	}

	txInfo, err := signerClient.GetCancelAllOrdersTransaction(context.Background(), req, nil)
	if err != nil {
		log.Fatalf("Failed to create cancel all orders transaction: %v", err)
	}

	fmt.Println("Cancel All Orders Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Time In Force: GoodTilTime\n")
	fmt.Printf("  Cancels all matching orders\n")

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
