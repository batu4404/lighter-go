// Example: Streaming account open orders via WebSocket
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/0xJord4n/lighter-go/client/ws"
	"github.com/0xJord4n/lighter-go/examples"
)

func main() {
	privateKey := examples.GetPrivateKey()
	if privateKey == "" {
		log.Fatal("LIGHTER_PRIVATE_KEY environment variable not set")
	}

	// Get account index from args or use default
	accountIndex := int64(35928) // Default test account
	if len(os.Args) > 1 {
		idx, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err == nil {
			accountIndex = idx
		}
	}

	network := examples.GetNetwork()
	fmt.Printf("Connecting to %s WebSocket...\n", network.String())

	// Create signer client to get auth token
	apiKeyIndex := uint8(0)
	signerClient, err := examples.CreateSignerClient(privateKey, apiKeyIndex, accountIndex)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	// Get auth token (valid for 8 hours)
	authToken, err := signerClient.GetAuthToken(time.Now().Add(8 * time.Hour))
	if err != nil {
		log.Fatalf("Failed to get auth token: %v", err)
	}
	fmt.Printf("Auth token obtained for account %d\n", accountIndex)

	// Create WebSocket client with callbacks
	opts := ws.DefaultOptions().
		WithOnConnect(func() {
			fmt.Println("Connected to WebSocket!")
		}).
		WithOnDisconnect(func(err error) {
			fmt.Printf("Disconnected: %v\n", err)
		}).
		WithOnAccountAllOrdersUpdate(func(update *ws.AccountAllOrdersUpdate) {
			fmt.Printf("\n=== Orders Update for Account %d ===\n", update.AccountIndex)
			for marketIdx, orders := range update.Orders {
				fmt.Printf("Market %s: %d orders\n", marketIdx, len(orders))
				for _, order := range orders {
					side := "BUY"
					if order.IsAsk {
						side = "SELL"
					}
					fmt.Printf("  [%s] %s @ %s | Remaining: %s | Status: %s | Type: %s\n",
						order.OrderID,
						side,
						order.Price,
						order.RemainingBaseAmount,
						order.Status,
						order.Type,
					)
				}
			}
		})

	client := ws.NewClient(network.WSURL(), opts)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Connect
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close() //nolint:errcheck // Cleanup on exit

	// Subscribe to account_all_orders channel
	fmt.Printf("Subscribing to account_all_orders for account %d...\n", accountIndex)
	if err := client.SubscribeAccountAllOrders(accountIndex, authToken); err != nil {
		log.Fatalf("Failed to subscribe to account orders: %v", err)
	}
	fmt.Println("Subscribed! Waiting for orders...")

	// Process updates (the callback will handle order updates)
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-client.AccountUpdates():
			// Generic account updates will also appear here
			if update.Channel == string(ws.ChannelAccountAllOrders) {
				// Already handled by callback, but you can also process here
				fmt.Printf("Received account update from channel: %s\n", update.Channel)
			}
		case err := <-client.Errors():
			fmt.Printf("Error: %v\n", err)
		}
	}
}
