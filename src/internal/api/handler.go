package api

import (
	"log"
	"math/big"
	"net/http"
	"time"
	"trustflow/src/internal/executor"
	"trustflow/src/internal/simulator"
	"trustflow/src/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	sim  *simulator.Simulator
	exec *executor.Executor
}

func NewHandler(sim *simulator.Simulator, exec *executor.Executor) *Handler {
	return &Handler{
		sim:  sim,
		exec: exec,
	}
}

// SubmitIntent handles the POST /intent request
func (h *Handler) SubmitIntent(c *gin.Context) {
	var intent types.Intent

	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&intent); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Assign ID and Timestamp if missing
	if intent.ID == "" {
		intent.ID = uuid.New().String()
	}
	if intent.CreatedAt == 0 {
		intent.CreatedAt = time.Now().Unix()
	}

	// 1. Parse Intent
	candidate, err := simulator.ParseIntent(intent)
	if err != nil {
		log.Printf("Intent parsing failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Intent: " + err.Error()})
		return
	}

	// 2. Simulate (Safety Check)
	gasLimit, err := h.sim.Simulate(c.Request.Context(), candidate)
	if err != nil {
		log.Printf("Simulation failed: %v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Simulation Failed: " + err.Error()})
		return
	}

	// 3. Solvency Check
	if err := h.sim.CheckSolvency(c.Request.Context(), gasLimit, candidate.Value); err != nil {
		log.Printf("Solvency check failed: %v", err)
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient Funds: " + err.Error()})
		return
	}

	// 4. Execute Transaction
	txHash, err := h.exec.Execute(c.Request.Context(), candidate, gasLimit)
	if err != nil {
		log.Printf("Execution failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Execution Failed: " + err.Error()})
		return
	}

	// Return success response
	response := types.IntentResponse{
		Status:   "success",
		IntentID: intent.ID,
		Message:  "Transaction executed successfully",
		TxHash:   txHash,
	}

	c.JSON(http.StatusOK, response)
}

// SimulateIntent handles the POST /simulate request
func (h *Handler) SimulateIntent(c *gin.Context) {
	var intent types.Intent

	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&intent); err != nil {
		c.JSON(http.StatusBadRequest, types.SimulationResponse{
			Valid: false,
			Error: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// 1. Parse Intent
	candidate, err := simulator.ParseIntent(intent)
	if err != nil {
		c.JSON(http.StatusOK, types.SimulationResponse{
			Valid: false,
			Error: "Parsing Failed: " + err.Error(),
		})
		return
	}

	// 2. Simulate (Get Gas Limit)
	gasLimit, err := h.sim.Simulate(c.Request.Context(), candidate)
	if err != nil {
		c.JSON(http.StatusOK, types.SimulationResponse{
			Valid: false,
			Error: "Simulation Reverted: " + err.Error(),
		})
		return
	}

	// 3. Get Cost Details
	gasPrice, err := h.sim.GetGasPrice(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, types.SimulationResponse{
			Valid: false,
			Error: "Failed to fetch gas price: " + err.Error(),
		})
		return
	}

	// Calculate Total Cost (Gas * Price)
	totalCost := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), gasPrice)

	response := types.SimulationResponse{
		Valid:     true,
		GasLimit:  gasLimit,
		GasPrice:  gasPrice.String(),
		TotalCost: totalCost.String(),
		Message:   "Simulation Successful",
	}

	c.JSON(http.StatusOK, response)
}
