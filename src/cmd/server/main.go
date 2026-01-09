package main

import (
	"log"
	"trustflow/src/internal/api"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Gin router
	router := gin.Default()

	// Define Routes
	router.POST("/intent", api.SubmitIntentHandler)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Start Server
	log.Println("Starting TrustFlow Orchestrator on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
