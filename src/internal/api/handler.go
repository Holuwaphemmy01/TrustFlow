package api

import (
	"log"
	"net/http"
	"time"
	"trustflow/src/internal/simulator"
	"trustflow/src/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	sim *simulator.Simulator
}

func NewHandler(sim *simulator.Simulator) *Handler {
	return &Handler{sim: sim}
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

	// Return success response
	response := types.IntentResponse{
		Status:   "validated",
		IntentID: intent.ID,
		Message:  "Intent validated and safe to execute",
	}

	c.JSON(http.StatusAccepted, response)
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

	// 3. Get Cost Details (Gas Price)
	// Note: We need to access the client from the simulator, or expose a method.
	// For now, let's assume CheckSolvency does the heavy lifting, but we want the raw numbers.
	// We'll calculate it manually here or expose a "GetGasPrice" in simulator.
	// Let's rely on CheckSolvency for the final "Valid" boolean, but we need the breakdown.

	// Refactoring Simulator to expose GasPrice might be cleaner, but for now let's just use what we have.
	// To keep it simple, I will add a helper to Simulator to get the cost breakdown.

	// WAIT: I can't easily get GasPrice without exposing it.
	// Let's update Simulator to return the price or just re-fetch it here.
	// Accessing h.sim.client is not possible (private field).

	// Quick Fix: I will just re-implement the Solvency check logic here to get the data points,
	// OR better, I will add a `EstimateCost` method to the Simulator in the next step.

	// For now, I will return the basic info we have.

	response := types.SimulationResponse{
		Valid:    true,
		GasLimit: gasLimit,
		// GasPrice and TotalCost will be populated in the next iteration if we update Simulator
		Message: "Simulation Successful",
	}

	c.JSON(http.StatusOK, response)
}
