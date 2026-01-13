package main

import (
	"log"
	"net/http"
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

	openAPISpec := `{
	  "openapi": "3.0.0",
	  "info": {
	    "title": "TrustFlow API",
	    "version": "1.0.0",
	    "description": "Safety-first orchestration for AI blockchain intents"
	  },
	  "servers": [
	    { "url": "/" }
	  ],
	  "paths": {
	    "/health": {
	      "get": {
	        "summary": "Health check",
	        "description": "Service liveness check",
	        "responses": {
	          "200": {
	            "description": "OK",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/HealthResponse" },
	                "example": { "status": "ok" }
	              }
	            }
	          }
	        }
	      }
	    },
	    "/intent": {
	      "post": {
	        "summary": "Submit intent",
	        "description": "Submit a single action or multi-step workflow for execution",
	        "requestBody": {
	          "required": true,
	          "content": {
	            "application/json": {
	              "schema": { "$ref": "#/components/schemas/Intent" },
	              "examples": {
	                "single_action": {
	                  "summary": "Single payment",
	                  "value": {
	                    "action": "payment",
	                    "params": {
	                      "recipient": "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
	                      "amount": "100000000000000000"
	                    }
	                  }
	                },
	                "multi_step": {
	                  "summary": "Multi-step workflow",
	                  "value": {
	                    "steps": [
	                      {
	                        "action": "payment",
	                        "params": {
	                          "recipient": "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
	                          "amount": "100000000000000000"
	                        }
	                      },
	                      {
	                        "action": "payment",
	                        "params": {
	                          "recipient": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
	                          "amount": "200000000000000000"
	                        }
	                      }
	                    ]
	                  }
	                }
	              }
	            }
	          }
	        },
	        "responses": {
	          "200": {
	            "description": "Intent processed",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/IntentResponse" },
	                "example": {
	                  "status": "success",
	                  "intent_id": "a55470d4-784f-485b-b36f-ce70e540da3b",
	                  "message": "Successfully executed 2 steps",
	                  "tx_hash": "0xabc...",
	                  "tx_hashes": ["0xabc...", "0xdef..."]
	                }
	              }
	            }
	          },
	          "422": {
	            "description": "Intent failed",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/IntentResponse" },
	                "example": {
	                  "status": "failed",
	                  "intent_id": "a55470d4-784f-485b-b36f-ce70e540da3b",
	                  "message": "Execution halted at step 1: insufficient funds",
	                  "failed_step_index": 0,
	                  "error": "insufficient funds: have 0 wei, want 21000000000000 wei"
	                }
	              }
	            }
	          },
	          "400": {
	            "description": "Invalid JSON",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
	                "example": { "error": "json: cannot unmarshal string into Go struct field Intent.steps of type []types.IntentStep" }
	              }
	            }
	          },
	          "500": {
	            "description": "Internal server error",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
	                "example": { "error": "orchestration failed: execution failed" }
	              }
	            }
	          }
	        }
	      }
	    },
	    "/simulate": {
	      "post": {
	        "summary": "Dry-run simulation",
	        "description": "Pre-flight checks including gas estimate and price",
	        "requestBody": {
	          "required": true,
	          "content": {
	            "application/json": {
	              "schema": { "$ref": "#/components/schemas/Intent" },
	              "examples": {
	                "payment": {
	                  "value": {
	                    "action": "payment",
	                    "params": {
	                      "recipient": "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
	                      "amount": "100000000000000000"
	                    }
	                  }
	                }
	              }
	            }
	          }
	        },
	        "responses": {
	          "200": {
	            "description": "Simulation result",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/SimulationResponse" },
	                "examples": {
	                  "success": {
	                    "value": {
	                      "valid": true,
	                      "gas_limit": 21000,
	                      "gas_price": "1000000000",
	                      "total_cost": "21000000000000",
	                      "message": "Simulation Successful"
	                    }
	                  },
	                  "revert": {
	                    "value": {
	                      "valid": false,
	                      "error": "Simulation Reverted: execution reverted"
	                    }
	                  }
	                }
	              }
	            }
	          },
	          "400": {
	            "description": "Invalid JSON",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/SimulationResponse" },
	                "example": { "valid": false, "error": "Invalid JSON: ..." }
	              }
	            }
	          }
	        }
	      }
	    },
	    "/status/{id}": {
	      "get": {
	        "summary": "Get intent status",
	        "description": "Retrieve the current state of a submitted intent including step results",
	        "parameters": [
	          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
	        ],
	        "responses": {
	          "200": {
	            "description": "Current intent state",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/IntentState" },
	                "example": {
	                  "intent_id": "a55470d4-784f-485b-b36f-ce70e540da3b",
	                  "status": "failed",
	                  "created_at": 1700000000,
	                  "message": "Execution halted at step 1: insufficient funds",
	                  "raw_intent": "{\"action\":\"payment\",\"params\":{\"recipient\":\"0x...\",\"amount\":\"100\"}}",
	                  "steps": [
	                    { "step_index": 0, "action": "payment", "status": "failed", "error": "insufficient funds" }
	                  ]
	                }
	              }
	            }
	          },
	          "404": { "description": "Not found" },
	          "500": {
	            "description": "Internal server error",
	            "content": {
	              "application/json": {
	                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
	                "example": { "error": "Internal server error" }
	              }
	            }
	          }
	        }
	      }
	    },
	    "/intents": {
	      "get": {
	        "summary": "List recent intents",
	        "description": "List latest intents (default limit 50)",
	        "responses": {
	          "200": {
	            "description": "List of intents",
	            "content": {
	              "application/json": {
	                "schema": {
	                  "type": "array",
	                  "items": { "$ref": "#/components/schemas/IntentState" }
	                },
	                "example": [
	                  { "intent_id": "592e2677-6321-4ace-a2e9-ad17f641dd83", "status": "success", "created_at": 1700000000, "steps": [] },
	                  { "intent_id": "a55470d4-784f-485b-b36f-ce70e540da3b", "status": "failed", "created_at": 1700000001, "steps": [ { "step_index": 0, "action": "payment", "status": "failed", "error": "insufficient funds" } ] }
	                ]
	              }
	            }
	          }
	        }
	      }
	    }
	  },
	  "components": {
	    "schemas": {
	      "Intent": {
	        "type": "object",
	        "properties": {
	          "id": { "type": "string" },
	          "action": { "type": "string" },
	          "params": { "type": "object", "additionalProperties": { "type": "string" } },
	          "steps": {
	            "type": "array",
	            "items": { "$ref": "#/components/schemas/IntentStep" }
	          },
	          "created_at": { "type": "integer", "format": "int64" }
	        }
	      },
	      "IntentStep": {
	        "type": "object",
	        "properties": {
	          "id": { "type": "string" },
	          "action": { "type": "string" },
	          "params": { "type": "object", "additionalProperties": { "type": "string" } }
	        },
	        "required": ["action", "params"]
	      },
	      "IntentResponse": {
	        "type": "object",
	        "properties": {
	          "status": { "type": "string" },
	          "intent_id": { "type": "string" },
	          "message": { "type": "string" },
	          "tx_hash": { "type": "string" },
	          "tx_hashes": { "type": "array", "items": { "type": "string" } },
	          "failed_step_index": { "type": "integer" },
	          "error": { "type": "string" }
	        }
	      },
	      "StepState": {
	        "type": "object",
	        "properties": {
	          "step_index": { "type": "integer" },
	          "action": { "type": "string" },
	          "status": { "type": "string" },
	          "tx_hash": { "type": "string" },
	          "error": { "type": "string" }
	        }
	      },
	      "IntentState": {
	        "type": "object",
	        "properties": {
	          "intent_id": { "type": "string" },
	          "status": { "type": "string" },
	          "created_at": { "type": "integer", "format": "int64" },
	          "message": { "type": "string" },
	          "raw_intent": { "type": "string" },
	          "steps": { "type": "array", "items": { "$ref": "#/components/schemas/StepState" } }
	        }
	      },
	      "SimulationResponse": {
	        "type": "object",
	        "properties": {
	          "valid": { "type": "boolean" },
	          "gas_limit": { "type": "integer", "format": "int64" },
	          "gas_price": { "type": "string" },
	          "total_cost": { "type": "string" },
	          "message": { "type": "string" },
	          "error": { "type": "string" }
	        }
	      },
	      "ErrorResponse": {
	        "type": "object",
	        "properties": { "error": { "type": "string" } }
	      },
	      "HealthResponse": {
	        "type": "object",
	        "properties": { "status": { "type": "string" } }
	      }
	    }
	  }
	}`

	swaggerHTML := `<!DOCTYPE html><html><head><meta charset="utf-8"/><title>TrustFlow API Docs</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>window.ui=SwaggerUIBundle({url:'/openapi.json',dom_id:'#swagger-ui'});</script></body></html>`

	// Define Routes
	router.POST("/intent", handler.SubmitIntent)
	router.POST("/simulate", handler.SimulateIntent)
	router.GET("/status/:id", handler.GetStatus)
	router.GET("/intents", handler.ListIntents)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	router.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte(openAPISpec))
	})
	router.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, swaggerHTML)
	})

	// Start Server
	log.Println("Starting TrustFlow Orchestrator on :8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
