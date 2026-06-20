// Example: Creating grouped orders (OCO, bracket orders)
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

	marketIndex := int16(0)
	expiry := time.Now().Add(24 * time.Hour).UnixMilli()

	// Create a bracket order: Take Profit + Stop Loss
	// When one triggers, the other is cancelled
	orders := []*types.CreateOrderTxReq{
		{
			MarketIndex:      marketIndex,
			ClientOrderIndex: 1,
			BaseAmount:       1000000,     // 0.01 ETH
			Price:            3500_000000, // Take profit at $3500
			IsAsk:            1,           // Sell
			Type:             txtypes.LimitOrder,
			TimeInForce:      txtypes.GoodTillTime,
			ReduceOnly:       1, // Reduce only
			TriggerPrice:     0,
			OrderExpiry:      expiry,
		},
		{
			MarketIndex:      marketIndex,
			ClientOrderIndex: 2,
			BaseAmount:       1000000,     // 0.01 ETH
			Price:            2800_000000, // Stop loss at $2800
			IsAsk:            1,           // Sell
			Type:             txtypes.StopLossOrder,
			TimeInForce:      txtypes.GoodTillTime,
			ReduceOnly:       1,
			TriggerPrice:     2800_000000,
			OrderExpiry:      expiry,
		},
	}

	req := &types.CreateGroupedOrdersTxReq{
		GroupingType: txtypes.GroupingType_OneCancelsTheOther, // One-Cancels-Other
		Orders:       orders,
	}

	txInfo, err := signerClient.GetCreateGroupedOrdersTransaction(context.Background(), req, nil)
	if err != nil {
		log.Fatalf("Failed to create grouped orders: %v", err)
	}

	fmt.Println("Grouped Orders (OCO) Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Grouping Type: OCO (One-Cancels-Other)\n")
	fmt.Printf("  Order 1: Take Profit @ $3500\n")
	fmt.Printf("  Order 2: Stop Loss @ $2800\n")

	// Submit to API
	resp, err := signerClient.SendAndSubmit(context.Background(), txInfo)
	if err != nil {
		log.Fatalf("Failed to submit orders: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
