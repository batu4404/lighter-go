package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/0xJord4n/lighter-go/client"
	lighterhttp "github.com/0xJord4n/lighter-go/client/http"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Network to use (change to client.Testnet for testnet)
var network = client.Mainnet

func main() {
	ethPrivateKey := flag.String("key", "", "Ethereum private key (hex, with or without 0x prefix)")
	flag.Parse()

	if *ethPrivateKey == "" {
		fmt.Println("Usage: go run main.go -key <ethereum_private_key>")
		os.Exit(1)
	}

	// Remove 0x prefix if present
	keyHex := strings.TrimPrefix(*ethPrivateKey, "0x")

	// Derive wallet address from private key
	privateKey, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		fmt.Printf("ERROR: Invalid private key: %v\n", err)
		os.Exit(1)
	}
	walletAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("Wallet address: %s\n", walletAddress.Hex())

	// Create Lighter HTTP client
	lighterClient := lighterhttp.NewFullClientForNetwork(network)
	if lighterClient == nil {
		fmt.Println("ERROR: Failed to create lighter client")
		os.Exit(1)
	}
	fmt.Printf("Connected to Lighter API: %s (chain ID: %d)\n", network.APIURL(), network.ChainID())

	// Look up account by L1 address
	checksummedAddress := common.HexToAddress(walletAddress.Hex()).Hex()
	fmt.Printf("Looking up account for address: %s\n", checksummedAddress)

	accounts, err := lighterClient.Account().GetAccountsByL1Address(context.Background(), checksummedAddress)
	if err != nil {
		fmt.Printf("ERROR: GetAccountsByL1Address failed: %v\n", err)
		fmt.Println("This usually means the user hasn't deposited at least $5 on Lighter yet.")
		os.Exit(1)
	}

	if accounts == nil || len(accounts.SubAccounts) == 0 {
		fmt.Println("ERROR: No Lighter account found for this address")
		fmt.Println("The user needs to deposit at least $5 on Lighter first.")
		os.Exit(1)
	}

	accountIndex := accounts.SubAccounts[0].Index
	fmt.Printf("Found account index: %d\n", accountIndex)

	// Check existing API keys
	existingKeys, err := lighterClient.Account().GetApiKeys(context.Background(), accountIndex, nil)
	if err != nil {
		fmt.Printf("WARNING: Could not fetch existing API keys: %v\n", err)
	} else {
		fmt.Printf("Existing API keys count: %d\n", len(existingKeys.ApiKeys))
		for i, key := range existingKeys.ApiKeys {
			fmt.Printf("  Key %d: index=%d, pubkey=%s\n", i, key.ApiKeyIndex, key.PublicKey)
		}
	}

	// Check nonce for slot 0
	nonce, err := lighterClient.GetNextNonce(context.Background(), accountIndex, 0)
	if err != nil {
		fmt.Printf("WARNING: Could not fetch nonce for slot 0: %v\n", err)
	} else {
		fmt.Printf("Next nonce for slot 0: %d\n", nonce)
	}

	// Generate API key pair
	fmt.Println("\nGenerating new API key pair...")
	apiKeyPrivate, apiKeyPublic, err := client.GenerateAPIKey()
	if err != nil {
		fmt.Printf("ERROR: Failed to generate API key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("API key private: %s\n", apiKeyPrivate)
	fmt.Printf("API key public:  %s\n", apiKeyPublic)

	// Create signer client with the generated API key
	signerClient, err := client.NewSignerClient(
		lighterClient,
		apiKeyPrivate,
		network.ChainID(),
		0, // API key index (slot 0)
		accountIndex,
		nil,
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create signer client: %v\n", err)
		os.Exit(1)
	}

	// Convert public key hex string to [40]byte array
	pubKeyHex := strings.TrimPrefix(apiKeyPublic, "0x")
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		fmt.Printf("ERROR: Failed to decode API key public: %v\n", err)
		os.Exit(1)
	}

	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	// Verify public keys match
	signerPubKey := signerClient.GetKeyManager().PubKeyBytes()
	fmt.Printf("\nPublic key comparison:\n")
	fmt.Printf("  Request pubkey:  0x%x\n", pubKeyBytes)
	fmt.Printf("  Signer pubkey:   0x%x\n", signerPubKey[:])
	fmt.Printf("  Match: %v\n", string(pubKeyBytes) == string(signerPubKey[:]))

	// Call ChangePubKey
	fmt.Println("\nCalling ChangePubKey...")
	resp, err := signerClient.ChangePubKey(context.Background(), *ethPrivateKey, &types.ChangePubKeyReq{
		PubKey: pubKey,
	}, nil)

	if err != nil {
		fmt.Printf("\nERROR: ChangePubKey failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSUCCESS! ChangePubKey response:\n")
	fmt.Printf("  TxHash: %s\n", resp.TxHash)
	fmt.Printf("  SequenceIndex: %d\n", resp.SequenceIndex)
}
