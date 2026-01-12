package executor

import (
	"context"
	"fmt"
	"trustflow/src/internal/chain"
	"trustflow/src/internal/simulator"
)

type Executor struct {
	client *chain.ChainClient
}

func NewExecutor(client *chain.ChainClient) *Executor {
	return &Executor{client: client}
}

// Execute signs and broadcasts the transaction candidate
func (e *Executor) Execute(ctx context.Context, candidate *simulator.TxCandidate, gasLimit uint64) (string, error) {
	// Call the ChainClient's SendTransaction method
	txHash, err := e.client.SendTransaction(
		ctx,
		candidate.ToAddress,
		candidate.Value,
		candidate.Data,
		gasLimit,
	)
	if err != nil {
		return "", fmt.Errorf("execution failed: %w", err)
	}

	return txHash, nil
}
