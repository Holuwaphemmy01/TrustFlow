package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"
	"trustflow/src/internal/executor"
	"trustflow/src/internal/simulator"
	"trustflow/src/internal/storage"
	"trustflow/src/pkg/types"
)

type Orchestrator struct {
	sim   *simulator.Simulator
	exec  *executor.Executor
	store *storage.Storage
}

func NewOrchestrator(sim *simulator.Simulator, exec *executor.Executor, store *storage.Storage) *Orchestrator {
	return &Orchestrator{
		sim:   sim,
		exec:  exec,
		store: store,
	}
}

// GetIntentStatus retrieves the current state of an intent
func (o *Orchestrator) GetIntentStatus(id string) (*types.IntentState, error) {
	return o.store.GetIntent(id)
}

// ListIntents retrieves the most recent intents
func (o *Orchestrator) ListIntents(limit int) ([]types.IntentState, error) {
	return o.store.GetRecentIntents(limit)
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

	// Save Intent to DB
	if err := o.store.SaveIntent(intent); err != nil {
		log.Printf("Failed to save intent: %v", err)
	}

	// 2. Execution Loop
	for i, step := range steps {
		log.Printf("🔄 Processing Step %d/%d: %s", i+1, len(steps), step.Action)

		// Save Step to DB
		if err := o.store.SaveStep(intent.ID, i, step.Action); err != nil {
			log.Printf("Failed to save step: %v", err)
		}

		// Helper to return partial failure
		returnFailure := func(err error) (*types.IntentResponse, error) {
			o.store.UpdateIntentStatus(intent.ID, "failed", err.Error())
			o.store.UpdateStepStatus(intent.ID, i, "failed", "", err.Error())

			failedIdx := i
			return &types.IntentResponse{
				Status:          "failed",
				IntentID:        intent.ID,
				Message:         fmt.Sprintf("Execution halted at step %d: %v", i+1, err),
				TxHashes:        txHashes,
				FailedStepIndex: &failedIdx,
				Error:           err.Error(),
			}, nil // We return nil error because we want to return the structured response
		}

		// A. Parse Step
		tempIntent := types.Intent{Action: step.Action, Params: step.Params}
		candidate, err := simulator.ParseIntent(tempIntent)
		if err != nil {
			return returnFailure(fmt.Errorf("parse failed: %w", err))
		}

		// B. Simulate (Safety Check)
		gasLimit, err := o.sim.Simulate(ctx, candidate)
		if err != nil {
			return returnFailure(fmt.Errorf("simulation failed: %w", err))
		}

		// C. Execute
		txHash, err := o.exec.Execute(ctx, candidate, gasLimit)
		if err != nil {
			return returnFailure(fmt.Errorf("execution failed: %w", err))
		}

		log.Printf("✅ Step %d Executed. Hash: %s", i+1, txHash)
		txHashes = append(txHashes, txHash)
		o.store.UpdateStepStatus(intent.ID, i, "success", txHash, "")

		// D. Wait for Confirmation (if there are more steps)
		if i < len(steps)-1 {
			log.Printf("⏳ Waiting for confirmation of %s...", txHash)
			time.Sleep(5 * time.Second)
		}
	}

	o.store.UpdateIntentStatus(intent.ID, "success", "All steps executed successfully")

	return &types.IntentResponse{
		Status:   "success",
		IntentID: intent.ID,
		Message:  fmt.Sprintf("Successfully executed %d steps", len(steps)),
		TxHashes: txHashes,
		TxHash:   txHashes[len(txHashes)-1], // Last hash for backward compatibility
	}, nil
}
