package types

// Intent represents the high-level goal of the agent
type Intent struct {
	ID          string            `json:"id"`
	AgentID     string            `json:"agent_id"`
	Action      string            `json:"action"` // e.g., "payment", "swap"
	Params      map[string]string `json:"params"` // e.g., {"recipient": "0x...", "amount": "100", "token": "USDC"}
	Conditions  map[string]string `json:"conditions,omitempty"` // e.g., {"max_gas": "500000"}
	CreatedAt   int64             `json:"created_at"`
}

// IntentResponse is the standard response structure
type IntentResponse struct {
	Status  string `json:"status"`
	IntentID string `json:"intent_id"`
	Message string `json:"message"`
}
