package simulator

import (
	"context"
	"fmt"
	"math/big"
	"trustflow/src/internal/chain"

	"github.com/ethereum/go-ethereum"
)

type Simulator struct {
	client *chain.ChainClient
}

func NewSimulator(client *chain.ChainClient) *Simulator {
	return &Simulator{client: client}
}

// Simulate runs a transaction candidate against the chain to check for validity and estimate gas
func (s *Simulator) Simulate(ctx context.Context, candidate *TxCandidate) (uint64, error) {
	from := s.client.GetAddress()

	callMsg := ethereum.CallMsg{
		From:  from,
		To:    candidate.ToAddress,
		Value: candidate.Value,
		Data:  candidate.Data,
	}

	// If EstimateGas succeeds, it means the transaction didn't revert.
	gasLimit, err := s.client.EstimateGas(ctx, callMsg)
	if err != nil {
		// We wrap the error to give context (e.g. "execution reverted")
		return 0, fmt.Errorf("simulation failed (transaction would revert): %w", err)
	}

	return gasLimit, nil
}

// CheckSolvency ensures the wallet has enough funds for Value + GasCost
func (s *Simulator) CheckSolvency(ctx context.Context, gasLimit uint64, value *big.Int) error {
	// 1. Get Gas Price
	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch gas price: %w", err)
	}

	// 2. Calculate Total Cost: Value + (GasLimit * GasPrice)
	cost := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), gasPrice)
	totalReq := new(big.Int).Add(value, cost)

	// 3. Get Balance
	balance, err := s.client.GetBalance(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch balance: %w", err)
	}

	// 4. Compare
	if balance.Cmp(totalReq) < 0 {
		return fmt.Errorf("insufficient funds: have %s wei, want %s wei (Gas Cost: %s, Value: %s)", 
			balance.String(), totalReq.String(), cost.String(), value.String())
	}

	return nil
}
