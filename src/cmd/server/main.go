package main

import (
	"log"
	"trustflow/src/internal/api"
	"trustflow/src/internal/chain"
	"trustflow/src/internal/config"
	"trustflow/src/internal/executor"
	"trustflow/src/internal/orchestrator"
	"trustflow/src/internal/simulator"
	"trustflow/src/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Chain Client
	client, err := chain.NewChainClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to chain: %v", err)
	}
	defer client.Close()
	log.Printf("✅ Connected to Chain ID: %s", client.GetAddress().Hex())

	// 3. Initialize Simulator
	sim := simulator.NewSimulator(client)

	// 4. Initialize Executor
	exec := executor.NewExecutor(client)

	// 5. Initialize Storage
	store, err := storage.NewStorage("trustflow.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Println("✅ Connected to SQLite Storage")

	// 6. Initialize Orchestrator
	orch := orchestrator.NewOrchestrator(sim, exec, store)

	// 7. Initialize API Handler
	handler := api.NewHandler(orch, sim)

	// Initialize Gin router
	router := gin.Default()

	// Define Routes
	router.POST("/intent", handler.SubmitIntent)
	router.POST("/simulate", handler.SimulateIntent)
	router.GET("/status/:id", handler.GetStatus)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Start Server
	log.Println("Starting TrustFlow Orchestrator on :8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
