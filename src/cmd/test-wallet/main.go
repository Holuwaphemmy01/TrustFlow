package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"trustflow/src/internal/chain"
	"trustflow/src/internal/config"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v\nDid you create a .env file?", err)
	}

	// 2. Initialize Chain Client
	client, err := chain.NewChainClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create chain client: %v", err)
	}
	defer client.Close()

	// 3. Get Address
	address := client.GetAddress()
	fmt.Printf("✅ Connected!\n")
	fmt.Printf("📍 Wallet Address: %s\n", address.Hex())

	// 4. Get Balance
	balance, err := client.GetBalance(context.Background())
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	// Convert Wei to Eth/TCRO (Display only)
	ethValue := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	fmt.Printf("💰 Balance: %s Wei (%f TCRO)\n", balance.String(), ethValue)
}
