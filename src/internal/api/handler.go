package api

import (
	"log"
	"math/big"
	"net/http"
	"time"
	"trustflow/src/internal/orchestrator"
	"trustflow/src/internal/simulator"
	"trustflow/src/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	orch *orchestrator.Orchestrator
	sim  *simulator.Simulator // Kept for SimulateIntent (dry-run)
}

func NewHandler(orch *orchestrator.Orchestrator, sim *simulator.Simulator) *Handler {
	return &Handler{
		orch: orch,
		sim:  sim,
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

	// Delegate to Orchestrator
	response, err := h.orch.ProcessIntent(c.Request.Context(), intent)
	if err != nil {
		log.Printf("Orchestration failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
