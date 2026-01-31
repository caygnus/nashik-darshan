package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omkar273/nashikdarshan/internal/logger"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type HealthHandler struct {
	logger *logger.Logger
}

func NewHealthHandler(logger *logger.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

// @Summary Health check
// @Description Health check
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
func (h *HealthHandler) Health(c *gin.Context) {
	request := map[string]string{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	// log the incoming request
	h.logger.Infof("incoming request: %s %s | body: %s", c.Request.Method, c.Request.URL.Path, request)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HealthPost is a health check that requires authentication (X-API-Key or Bearer token).
// Use it to verify the API key or JWT works. Returns status and the authenticated user_id.
//
// @Summary Authenticated health check
// @Description Health check that validates API key or JWT; returns status and user_id
// @Tags Health
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /health [post]
func (h *HealthHandler) HealthPost(c *gin.Context) {
	userID := types.GetUserID(c.Request.Context())
	userEmail := types.GetUserEmail(c.Request.Context())
	resp := gin.H{
		"status":        "ok",
		"authenticated": true,
		"user_id":       userID,
	}
	if userEmail != "" {
		resp["user_email"] = userEmail
	}
	c.JSON(http.StatusOK, resp)
}
