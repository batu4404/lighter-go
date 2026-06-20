// Example: Changing the public key associated with an API key slot
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func main() {
	privateKey := examples.GetPrivateKey()
	if privateKey == "" {
		log.Fatal("LIGHTER_PRIVATE_KEY environment variable not set")
	}

	// Ethereum private key for L1 signature (required for ChangePubKey)
	ethPrivateKey := examples.GetEthPrivateKey()
	if ethPrivateKey == "" {
		log.Fatal("LIGHTER_ETH_PRIVATE_KEY environment variable not set")
	}

	apiKeyIndex := uint8(0)
	accountIndex := int64(1) // Your account index

	// Create signer client (uses LIGHTER_NETWORK env var, defaults to mainnet)
	signerClient, err := examples.CreateSignerClient(privateKey, apiKeyIndex, accountIndex)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	network := examples.GetNetwork()
	fmt.Printf("Connected to %s (chain ID: %d)\n\n", network.String(), network.ChainID())

	// Generate a new API key to register
	_, newPublicKey, err := client.GenerateAPIKey()
	if err != nil {
		log.Fatalf("Failed to generate new API key: %v", err)
	}

	// Decode public key from hex to bytes
	pubKeyBytes, err := hexutil.Decode(newPublicKey)
	if err != nil {
		log.Fatalf("Failed to decode public key: %v", err)
	}

	if len(pubKeyBytes) != 40 {
		log.Fatalf("Invalid public key length: expected 40 bytes, got %d", len(pubKeyBytes))
	}

	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	// Create change pub key request
	req := &types.ChangePubKeyReq{
		PubKey: pubKey,
	}

	// ChangePubKey handles L1 signing internally - just pass the eth private key
	resp, err := signerClient.ChangePubKey(context.Background(), ethPrivateKey, req, nil)
	if err != nil {
		log.Fatalf("Failed to submit change pub key: %v", err)
	}

	fmt.Println("Change Pub Key Submitted!")
	fmt.Printf("  TX Hash: %s\n", resp.TxHash)
	fmt.Printf("  New Public Key: %s\n", newPublicKey)
}
