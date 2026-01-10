package main

import (
	"context"
	"fmt"
	"log"

	"trustflow/src/internal/chain"
	"trustflow/src/internal/config"
	"trustflow/src/internal/simulator"
	"trustflow/src/pkg/types"
)

func main() {
	// 1. Setup
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	client, err := chain.NewChainClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	sim := simulator.NewSimulator(client)

	// 2. Define a Test Intent (Send 0.0001 TCRO to self)
	intent := types.Intent{
		Action: "payment",
		Params: map[string]string{
			"recipient": client.GetAddress().Hex(), // Send to self
			"amount":    "100000000000000",         // 0.0001 TCRO in Wei
		},
	}

	fmt.Println("🔍 1. Parsing Intent...")
	candidate, err := simulator.ParseIntent(intent)
	if err != nil {
		log.Fatalf("❌ Parse Failed: %v", err)
	}
	fmt.Printf("✅ Parsed: To=%s, Value=%s\n", candidate.ToAddress.Hex(), candidate.Value)

	fmt.Println("🔄 2. Simulating (EstimateGas)...")
	gasLimit, err := sim.Simulate(context.Background(), candidate)
	if err != nil {
		log.Fatalf("❌ Simulation Failed: %v", err)
	}
	fmt.Printf("✅ Simulation Passed! Gas Estimate: %d\n", gasLimit)

	fmt.Println("💰 3. Checking Solvency...")
	err = sim.CheckSolvency(context.Background(), gasLimit, candidate.Value)
	if err != nil {
		log.Fatalf("❌ Solvency Check Failed: %v", err)
	}
	fmt.Println("✅ Solvency Check Passed! Agent can afford this transaction.")
}
