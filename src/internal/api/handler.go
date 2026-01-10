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
