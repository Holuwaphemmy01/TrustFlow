package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"trustflow/src/internal/api"
	"trustflow/src/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSubmitIntentHandler(t *testing.T) {
	// Setup Gin router
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/intent", api.SubmitIntentHandler)

	t.Run("Valid Intent", func(t *testing.T) {
		intent := types.Intent{
			AgentID: "agent-007",
			Action:  "payment",
			Params: map[string]string{
				"recipient": "0x123",
				"amount":    "100",
			},
		}
		body, _ := json.Marshal(intent)

		req, _ := http.NewRequest("POST", "/intent", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response types.IntentResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		assert.Equal(t, "pending", response.Status)
		assert.NotEmpty(t, response.IntentID)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/intent", bytes.NewBuffer([]byte("{invalid-json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
