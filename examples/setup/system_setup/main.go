// Example: Setting up a Lighter account with API keys
//
// Prerequisites:
// 1. Have an Ethereum wallet with some ETH
// 2. Deposit USDC to Lighter via https://lighter.xyz (minimum 5 USDC)
// 3. Set LIGHTER_ETH_PRIVATE_KEY environment variable
//
// This script will:
// 1. Find your account by L1 address
// 2. Generate new API keys
// 3. Register them on your account
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/signer"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// APIKeyConfig stores the generated API keys
type APIKeyConfig struct {
	AccountIndex int64              `json:"account_index"`
	L1Address    string             `json:"l1_address"`
	APIKeys      map[int]APIKeyPair `json:"api_keys"`
}

type APIKeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func main() {
	// Ethereum private key - this identifies your account
	ethPrivateKey := examples.GetEthPrivateKey()
	if ethPrivateKey == "" {
		log.Fatal("LIGHTER_ETH_PRIVATE_KEY environment variable not set")
	}

	network := examples.GetNetwork()
	httpClient := examples.CreateHTTPClient()

	fmt.Printf("Connected to %s (chain ID: %d)\n", network.String(), network.ChainID())

	// Create L1 signer to get our address
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create L1 signer: %v", err)
	}
	l1Address := l1Signer.Address()
	fmt.Printf("L1 Address: %s\n", l1Address)

	// Find account by L1 address
	accounts, err := httpClient.Account().GetAccountsByL1Address(context.Background(), l1Address)
	if err != nil {
		log.Fatalf("Failed to get accounts: %v", err)
	}

	if len(accounts.SubAccounts) == 0 {
		log.Fatal("No account found for this L1 address. Please deposit USDC first at https://lighter.xyz")
	}

	// Use the master account (first one)
	account := accounts.SubAccounts[0]
	accountIndex := account.Index
	fmt.Printf("Found account index: %d\n", accountIndex)

	// Generate API keys for slots 0, 1, 2
	numKeys := 3
	config := APIKeyConfig{
		AccountIndex: accountIndex,
		L1Address:    l1Address,
		APIKeys:      make(map[int]APIKeyPair),
	}

	fmt.Println("\nGenerating API keys...")
	for i := 0; i < numKeys; i++ {
		privateKey, publicKey, err := client.GenerateAPIKey()
		if err != nil {
			log.Fatalf("Failed to generate API key %d: %v", i, err)
		}
		config.APIKeys[i] = APIKeyPair{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		}
		fmt.Printf("  Key %d: %s\n", i, publicKey)
	}

	// Register the first key (slot 0) so we can use it to register others
	fmt.Println("\nRegistering API key 0...")

	// For the first key, we need an existing API key to sign
	// If this is a fresh account, we need to use a bootstrap process
	// For now, assume we have at least one key registered

	// Try to use the first generated key
	firstKey := config.APIKeys[0]
	signerClient, err := client.NewSignerClient(
		httpClient,
		firstKey.PrivateKey,
		network.ChainID(),
		0, // API key index 0
		accountIndex,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	// Register remaining keys
	for i := 1; i < numKeys; i++ {
		fmt.Printf("Registering API key %d...\n", i)

		keyPair := config.APIKeys[i]
		pubKeyBytes, err := hexutil.Decode(keyPair.PublicKey)
		if err != nil {
			log.Fatalf("Failed to decode public key: %v", err)
		}

		var pubKey [40]byte
		copy(pubKey[:], pubKeyBytes)

		// Create a new signer client for this specific API key index
		keyClient, err := client.NewSignerClient(
			httpClient,
			firstKey.PrivateKey, // Use first key to sign
			network.ChainID(),
			uint8(i), // Target API key index
			accountIndex,
			nil,
		)
		if err != nil {
			log.Fatalf("Failed to create key client: %v", err)
		}

		req := &types.ChangePubKeyReq{
			PubKey: pubKey,
		}

		resp, err := keyClient.ChangePubKey(context.Background(), ethPrivateKey, req, nil)
		if err != nil {
			log.Fatalf("Failed to register API key %d: %v", i, err)
		}
		fmt.Printf("  TX Hash: %s\n", resp.TxHash)

		// Wait for propagation
		time.Sleep(2 * time.Second)
	}

	// Save config to file
	configFile := "lighter_api_keys.json"
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, configJSON, 0600); err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}

	fmt.Printf("\nSetup complete! API keys saved to %s\n", configFile)
	fmt.Println("\nIMPORTANT: Keep this file secure - it contains your private keys!")
	fmt.Println("\nTo use the SDK, set:")
	fmt.Printf("  export LIGHTER_PRIVATE_KEY=%s\n", config.APIKeys[0].PrivateKey)

	// Verify the client works
	fmt.Println("\nVerifying setup...")
	positions, err := signerClient.GetPositions(context.Background())
	if err != nil {
		log.Printf("Warning: Could not verify setup: %v", err)
	} else {
		fmt.Printf("Account verified! Found %d account(s)\n", len(positions.Accounts))
	}
}
