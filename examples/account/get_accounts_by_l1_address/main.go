// Example: Getting accounts by L1 (Ethereum) address
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/0xJord4n/lighter-go/examples"
)

func main() {
	network := examples.GetNetwork()
	httpClient := examples.CreateHTTPClient()

	fmt.Printf("Connected to %s\n", network.String())

	// Get L1 address from command line or use a default
	l1Address := "0x742d35Cc6634C0532925a3b844Bc9e7595f5bE91" // Replace with your L1 address
	if len(os.Args) > 1 {
		l1Address = os.Args[1]
	}

	fmt.Printf("Looking up accounts for L1 address: %s\n\n", l1Address)

	accounts, err := httpClient.Account().GetAccountsByL1Address(context.Background(), l1Address)
	if err != nil {
		log.Fatalf("Failed to get accounts: %v", err)
	}

	if len(accounts.SubAccounts) == 0 {
		fmt.Println("No accounts found for this L1 address")
		return
	}

	fmt.Printf("Found %d account(s):\n\n", len(accounts.SubAccounts))
	fmt.Printf("Master Account: %d\n\n", accounts.MasterAccount)
	for _, acc := range accounts.SubAccounts {
		fmt.Printf("Account Index: %d\n", acc.Index)
		fmt.Printf("  Master Index: %d\n", acc.MasterIndex)
		fmt.Printf("  L1 Address:   %s\n", acc.L1Address)
		fmt.Println()
	}
}
