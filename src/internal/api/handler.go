package api

import (
	"log"
	"net/http"
	"time"
	"trustflow/src/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SubmitIntentHandler handles the POST /intent request
func SubmitIntentHandler(c *gin.Context) {
	var intent types.Intent

	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&intent); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Assign ID and Timestamp if missing (Stub logic)
	if intent.ID == "" {
		intent.ID = uuid.New().String()
	}
	if intent.CreatedAt == 0 {
		intent.CreatedAt = time.Now().Unix()
	}

	// Stub: Log the intent (In reality, this would go to the Orchestrator)
	// log.Printf("Received Intent: %+v", intent)

	// Return success response
	response := types.IntentResponse{
		Status:   "pending",
		IntentID: intent.ID,
		Message:  "Intent received and queued for simulation",
	}

	c.JSON(http.StatusAccepted, response)
}
