package simulator_test

import (
	"math/big"
	"testing"
	"trustflow/src/internal/simulator"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestTxCandidate_Initialization(t *testing.T) {
	// Test Case 1: Standard Payment (Value Transfer)
	t.Run("Standard Payment", func(t *testing.T) {
		toAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		value := big.NewInt(1000000000000000000) // 1 ETH

		candidate := simulator.TxCandidate{
			ToAddress: &toAddr,
			Value:     value,
			Data:      nil,
		}

		assert.NotNil(t, candidate.ToAddress)
		assert.Equal(t, "0x1234567890123456789012345678901234567890", candidate.ToAddress.Hex())
		assert.Equal(t, value, candidate.Value)
		assert.Nil(t, candidate.Data)
	})

	// Test Case 2: Smart Contract Interaction (Data field present)
	t.Run("Contract Interaction", func(t *testing.T) {
		toAddr := common.HexToAddress("0xContractAddress")
		data := []byte{0xde, 0xad, 0xbe, 0xef}

		candidate := simulator.TxCandidate{
			ToAddress: &toAddr,
			Value:     big.NewInt(0),
			Data:      data,
		}

		assert.Equal(t, data, candidate.Data)
		assert.Equal(t, big.NewInt(0), candidate.Value)
	})
}
