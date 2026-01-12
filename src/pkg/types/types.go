package types

// Intent represents the high-level user request
type Intent struct {
	ID        string            `json:"id,omitempty"`
	Action    string            `json:"action"` // Deprecated in favor of Steps, but kept for backward compat if needed
	Params    map[string]string `json:"params,omitempty"`
	Steps     []IntentStep      `json:"steps,omitempty"` // For multi-step workflows
	CreatedAt int64             `json:"created_at,omitempty"`
}

// IntentStep represents a single atomic action within a workflow
type IntentStep struct {
	ID     string            `json:"id,omitempty"`
	Action string            `json:"action" binding:"required"`
	Params map[string]string `json:"params" binding:"required"`
}

// IntentResponse is the standard API response for intent submission
type IntentResponse struct {
	Status          string   `json:"status"`
	IntentID        string   `json:"intent_id"`
	Message         string   `json:"message"`
	TxHash          string   `json:"tx_hash,omitempty"`           // For single step
	TxHashes        []string `json:"tx_hashes,omitempty"`         // For multi-step
	FailedStepIndex *int     `json:"failed_step_index,omitempty"` // If failed, which step (0-based)
	Error           string   `json:"error,omitempty"`             // Error details
}

// SimulationResponse provides details about a dry-run execution
type SimulationResponse struct {
	Valid     bool   `json:"valid"`
	GasLimit  uint64 `json:"gas_limit"`
	GasPrice  string `json:"gas_price"`
	TotalCost string `json:"total_cost"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}
