package types

// Intent represents the high-level user request
type Intent struct {
	ID        string            `json:"id,omitempty"`
	Action    string            `json:"action" binding:"required"`
	Params    map[string]string `json:"params" binding:"required"`
	CreatedAt int64             `json:"created_at,omitempty"`
}

// IntentResponse is the standard API response for intent submission
type IntentResponse struct {
	Status   string `json:"status"`
	IntentID string `json:"intent_id"`
	Message  string `json:"message"`
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
