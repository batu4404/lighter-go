// Example: Creating a limit order
package main

import (
	"context"
	"fmt"
	"log"
	"time"

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
	fmt.Printf("Connected to %s (chain ID: %d)\n", network.String(), network.ChainID())

	// Create a limit buy order
	marketIndex := int16(0)      // ETH-USD perp
	size := int64(1000000)       // 0.01 ETH (scaled)
	price := uint32(3000_000000) // $3000 (scaled) - price is in 6 decimals
	isBuy := true
	expiry := time.Now().Add(24 * time.Hour).UnixMilli()

	opts := &types.TransactOpts{
		Nonce: types.NewInt64(-1),
	}

	txInfo, err := signerClient.CreateLimitOrder(context.Background(), marketIndex, size, price, isBuy, expiry, opts)
	if err != nil {
		log.Fatalf("Failed to create limit order: %v", err)
	}

	fmt.Printf("Limit order created!\n")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Market: %d\n", marketIndex)
	fmt.Printf("  Side: %s\n", map[bool]string{true: "BUY", false: "SELL"}[isBuy])
	fmt.Printf("  Size: %d\n", size)
	fmt.Printf("  Price: %d\n", price)
	fmt.Printf("  Expiry: %s\n", time.UnixMilli(expiry).Format(time.RFC3339))

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit order: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
