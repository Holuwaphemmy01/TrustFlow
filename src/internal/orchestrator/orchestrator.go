package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"
	"trustflow/src/internal/executor"
	"trustflow/src/internal/simulator"
	"trustflow/src/pkg/types"
)

type Orchestrator struct {
	sim  *simulator.Simulator
	exec *executor.Executor
}

func NewOrchestrator(sim *simulator.Simulator, exec *executor.Executor) *Orchestrator {
	return &Orchestrator{
		sim:  sim,
		exec: exec,
	}
}

// ProcessIntent handles both single and multi-step intents
func (o *Orchestrator) ProcessIntent(ctx context.Context, intent types.Intent) (*types.IntentResponse, error) {
	// 1. Normalize: Convert single action to a 1-step workflow
	steps := intent.Steps
	if len(steps) == 0 && intent.Action != "" {
		steps = []types.IntentStep{
			{Action: intent.Action, Params: intent.Params},
		}
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("no actions found in intent")
	}

	var txHashes []string

	// 2. Execution Loop
	for i, step := range steps {
		log.Printf("🔄 Processing Step %d/%d: %s", i+1, len(steps), step.Action)

		// A. Parse Step
		// We reuse the existing Parser, but we might need to map IntentStep back to Intent struct for compatibility
		// or update Parser to accept IntentStep.
		// For now, let's just construct a temporary Intent struct.
		tempIntent := types.Intent{Action: step.Action, Params: step.Params}
		candidate, err := simulator.ParseIntent(tempIntent)
		if err != nil {
			return nil, fmt.Errorf("step %d parse failed: %w", i+1, err)
		}

		// B. Simulate (Safety Check)
		gasLimit, err := o.sim.Simulate(ctx, candidate)
		if err != nil {
			return nil, fmt.Errorf("step %d simulation failed: %w", i+1, err)
		}

		// C. Execute
		txHash, err := o.exec.Execute(ctx, candidate, gasLimit)
		if err != nil {
			return nil, fmt.Errorf("step %d execution failed: %w", i+1, err)
		}

		log.Printf("✅ Step %d Executed. Hash: %s", i+1, txHash)
		txHashes = append(txHashes, txHash)

		// D. Wait for Confirmation (if there are more steps)
		if i < len(steps)-1 {
			log.Printf("⏳ Waiting for confirmation of %s...", txHash)
			// TODO: Implement actual polling. For now, we sleep to simulate block time.
			// In a real zkEVM, finality might be fast, but we should verify.
			time.Sleep(5 * time.Second) 
		}
	}

	return &types.IntentResponse{
		Status:   "success",
		IntentID: intent.ID,
		Message:  fmt.Sprintf("Successfully executed %d steps", len(steps)),
		TxHashes: txHashes,
		TxHash:   txHashes[len(txHashes)-1], // Last hash for backward compatibility
	}, nil
}
