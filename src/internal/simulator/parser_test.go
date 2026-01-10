package simulator_test

import (
	"math/big"
	"testing"
	"trustflow/src/internal/simulator"
	"trustflow/src/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestParseIntent(t *testing.T) {
	t.Run("Valid Payment", func(t *testing.T) {
		intent := types.Intent{
			Action: "payment",
			Params: map[string]string{
				"recipient": "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
				"amount":    "1000000000000000000", // 1 ETH in Wei
			},
		}

		candidate, err := simulator.ParseIntent(intent)
		assert.NoError(t, err)
		assert.NotNil(t, candidate)
		assert.Equal(t, "0x71C7656EC7ab88b098defB751B7401B5f6d8976F", candidate.ToAddress.Hex())
		assert.Equal(t, big.NewInt(1000000000000000000), candidate.Value)
		assert.Nil(t, candidate.Data)
	})

	t.Run("Invalid Action", func(t *testing.T) {
		intent := types.Intent{
			Action: "unknown_action",
			Params: map[string]string{},
		}
		candidate, err := simulator.ParseIntent(intent)
		assert.Error(t, err)
		assert.Nil(t, candidate)
		assert.Contains(t, err.Error(), "unknown action type")
	})

	t.Run("Payment Missing Recipient", func(t *testing.T) {
		intent := types.Intent{
			Action: "payment",
			Params: map[string]string{
				"amount": "100",
			},
		}
		candidate, err := simulator.ParseIntent(intent)
		assert.Error(t, err)
		assert.Nil(t, candidate)
		assert.Contains(t, err.Error(), "missing recipient")
	})

	t.Run("Payment Invalid Amount", func(t *testing.T) {
		intent := types.Intent{
			Action: "payment",
			Params: map[string]string{
				"recipient": "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
				"amount":    "not_a_number",
			},
		}
		candidate, err := simulator.ParseIntent(intent)
		assert.Error(t, err)
		assert.Nil(t, candidate)
		assert.Contains(t, err.Error(), "invalid amount")
	})
}
