package executor_test

import (
	"context"
	"math/big"
	"os"
	"testing"
	"trustflow/src/internal/chain"
	"trustflow/src/internal/config"
	"trustflow/src/internal/executor"
	"trustflow/src/internal/simulator"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_Initialization(t *testing.T) {
	// We need a dummy config to initialize chain client (even if it fails to connect, we just want to test struct creation)
	// However, NewExecutor requires a valid *chain.ChainClient.
	// Let's just test that NewExecutor returns a non-nil object given a nil client (simple struct check)
	exec := executor.NewExecutor(nil)
	assert.NotNil(t, exec)
}

// TestExecutor_Live_Transaction runs a real transaction on the testnet.
// It is skipped unless TEST_LIVE environment variable is set to "1".
func TestExecutor_Live_Transaction(t *testing.T) {
	if os.Getenv("TEST_LIVE") != "1" {
		t.Skip("Skipping live transaction test. Set TEST_LIVE=1 to run.")
	}

	// 1. Setup
	// Manually set the config path or load from root since tests run in package dir
	// os.Chdir("../../../") // Go up to root (src/internal/executor -> src/internal -> src -> root)

	// Debug: Print CWD
	cwd, _ := os.Getwd()
	t.Logf("Current Working Directory: %s", cwd)

	// Try to locate .env by walking up
	// If we are in src/internal/executor, we need to go up 3 levels
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		os.Chdir("../../../")
	}

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	client, err := chain.NewChainClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	exec := executor.NewExecutor(client)

	// 2. Prepare Candidate (Send 0.00001 TCRO to self)
	// Using a very small amount to minimize cost
	toAddr := client.GetAddress()
	amount := big.NewInt(10000000000000) // 0.00001 TCRO

	candidate := &simulator.TxCandidate{
		ToAddress: &toAddr,
		Value:     amount,
		Data:      nil,
	}

	// 3. Execute
	// Use a hardcoded gas limit for safety in test, or use Simulator to estimate.
	// Here we just use a safe buffer.
	gasLimit := uint64(500000)

	txHash, err := exec.Execute(context.Background(), candidate, gasLimit)
	require.NoError(t, err)

	t.Logf("✅ Transaction Executed! Hash: %s", txHash)
	assert.NotEmpty(t, txHash)
	assert.True(t, len(txHash) > 10) // Basic length check
}
