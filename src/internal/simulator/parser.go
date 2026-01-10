package simulator

import (
	"errors"
	"fmt"
	"math/big"
	"trustflow/src/pkg/types"

	"github.com/ethereum/go-ethereum/common"
)

// ParseIntent converts a generic high-level Intent into a low-level TxCandidate
func ParseIntent(intent types.Intent) (*TxCandidate, error) {
	switch intent.Action {
	case "payment":
		return parsePayment(intent.Params)
	default:
		return nil, fmt.Errorf("unknown action type: %s", intent.Action)
	}
}

func parsePayment(params map[string]string) (*TxCandidate, error) {
	// 1. Validate Recipient
	recipientStr, ok := params["recipient"]
	if !ok || recipientStr == "" {
		return nil, errors.New("missing recipient parameter")
	}
	if !common.IsHexAddress(recipientStr) {
		return nil, errors.New("invalid recipient address format")
	}
	toAddr := common.HexToAddress(recipientStr)

	// 2. Validate Amount
	amountStr, ok := params["amount"]
	if !ok || amountStr == "" {
		return nil, errors.New("missing amount parameter")
	}

	amount := new(big.Int)
	amount, success := amount.SetString(amountStr, 10) // Assume decimal input for simplicity
	if !success {
		return nil, errors.New("invalid amount format (must be decimal integer)")
	}
	if amount.Sign() <= 0 {
		return nil, errors.New("amount must be positive")
	}

	// 3. Construct Candidate
	return &TxCandidate{
		ToAddress: &toAddr,
		Value:     amount,
		Data:      nil, // Native transfer has no data
	}, nil
}
