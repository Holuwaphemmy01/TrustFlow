package simulator

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TxCandidate represents a transaction that has been parsed but not yet signed or broadcast.
// It serves as the intermediate format for simulation.
type TxCandidate struct {
	ToAddress *common.Address // Pointer because it can be nil (contract creation)
	Value     *big.Int        // Amount in Wei
	Data      []byte          // Call data (for smart contracts) or empty (for payments)
}
