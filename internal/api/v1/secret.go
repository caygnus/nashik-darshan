package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omkar273/nashikdarshan/internal/api/dto"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/service"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type SecretHandler struct {
	secretService service.SecretService
}

func NewSecretHandler(secretService service.SecretService) *SecretHandler {
	return &SecretHandler{secretService: secretService}
}

// @Summary Create a new API key
// @Description Create a new API key (private or publishable). The raw key is returned only once.
// @Tags Secrets
// @Accept json
// @Produce json
// @Param request body dto.CreateAPIKeyRequest true "Create API key request"
// @Success 201 {object} dto.CreateAPIKeyResponse
// @Failure 400 {object} ierr.ErrorResponse
// @Failure 409 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /secrets/api-keys [post]
// @Security Authorization
// @Security ApiKeyAuth
func (h *SecretHandler) CreateAPIKey(c *gin.Context) {
	var req dto.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(ierr.WithError(err).
			WithHint("Please check the request payload").
			Mark(ierr.ErrValidation))
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(err)
		return
	}

	userID := types.GetUserID(c.Request.Context())
	if userID == "" {
		c.Error(ierr.NewError("user not authenticated").
			WithHint("Please authenticate to create an API key").
			Mark(ierr.ErrPermissionDenied))
		return
	}

	secretEntity, rawKey, err := h.secretService.CreateAPIKey(
		c.Request.Context(),
		userID,
		req.Name,
		req.ToSecretType(),
		req.Permissions,
	)
	if err != nil {
		c.Error(err)
		return
	}

	response := &dto.CreateAPIKeyResponse{
		APIKeyResponse: dto.FromSecret(secretEntity),
		Key:            rawKey,
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary List API keys
// @Description List all API keys for the current user
// @Tags Secrets
// @Accept json
// @Produce json
// @Success 200 {object} dto.ListAPIKeysResponse
// @Failure 401 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /secrets/api-keys [get]
// @Security Authorization
// @Security ApiKeyAuth
func (h *SecretHandler) ListAPIKeys(c *gin.Context) {
	userID := types.GetUserID(c.Request.Context())
	if userID == "" {
		c.Error(ierr.NewError("user not authenticated").
			WithHint("Please authenticate to list API keys").
			Mark(ierr.ErrPermissionDenied))
		return
	}

	secrets, err := h.secretService.ListAPIKeys(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	response := dto.NewListAPIKeysResponse(secrets)
	c.JSON(http.StatusOK, response)
}

// @Summary Get API key by ID
// @Description Get a specific API key by its ID
// @Tags Secrets
// @Accept json
// @Produce json
// @Param id path string true "API Key ID"
// @Success 200 {object} dto.APIKeyResponse
// @Failure 404 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /secrets/api-keys/{id} [get]
// @Security Authorization
// @Security ApiKeyAuth
func (h *SecretHandler) GetAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(ierr.NewError("API key ID is required").
			WithHint("Please provide a valid API key ID").
			Mark(ierr.ErrValidation))
		return
	}

	secretEntity, err := h.secretService.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	// Verify ownership
	userID := types.GetUserID(c.Request.Context())
	if secretEntity.CreatedBy != userID {
		c.Error(ierr.NewError("access denied").
			WithHint("You can only access your own API keys").
			Mark(ierr.ErrPermissionDenied))
		return
	}

	response := dto.FromSecret(secretEntity)
	c.JSON(http.StatusOK, response)
}

// @Summary Update API key
// @Description Update an API key (name, permissions, or status)
// @Tags Secrets
// @Accept json
// @Produce json
// @Param id path string true "API Key ID"
// @Param request body dto.UpdateAPIKeyRequest true "Update API key request"
// @Success 200 {object} dto.APIKeyResponse
// @Failure 400 {object} ierr.ErrorResponse
// @Failure 404 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /secrets/api-keys/{id} [patch]
// @Security Authorization
// @Security ApiKeyAuth
func (h *SecretHandler) UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(ierr.NewError("API key ID is required").
			WithHint("Please provide a valid API key ID").
			Mark(ierr.ErrValidation))
		return
	}

	var req dto.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(ierr.WithError(err).
			WithHint("Please check the request payload").
			Mark(ierr.ErrValidation))
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(err)
		return
	}

	// Verify ownership
	secretEntity, err := h.secretService.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	userID := types.GetUserID(c.Request.Context())
	if secretEntity.CreatedBy != userID {
		c.Error(ierr.NewError("access denied").
			WithHint("You can only update your own API keys").
			Mark(ierr.ErrPermissionDenied))
		return
	}

	// Determine status
	status := secretEntity.Status
	if req.Status != nil {
		status = *req.Status
	}

	// Determine name
	name := secretEntity.Name
	if req.Name != nil && *req.Name != "" {
		name = *req.Name
	}

	// Determine permissions
	permissions := secretEntity.Permissions
	if req.Permissions != nil {
		permissions = req.Permissions
	}

	err = h.secretService.UpdateAPIKey(c.Request.Context(), id, name, permissions, status)
	if err != nil {
		c.Error(err)
		return
	}

	// Fetch updated secret
	updatedSecret, err := h.secretService.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response := dto.FromSecret(updatedSecret)
	c.JSON(http.StatusOK, response)
}

// @Summary Delete API key
// @Description Delete (archive) an API key
// @Tags Secrets
// @Accept json
// @Produce json
// @Param id path string true "API Key ID"
// @Success 204 "No Content"
// @Failure 404 {object} ierr.ErrorResponse
// @Failure 500 {object} ierr.ErrorResponse
// @Router /secrets/api-keys/{id} [delete]
// @Security Authorization
// @Security ApiKeyAuth
func (h *SecretHandler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(ierr.NewError("API key ID is required").
			WithHint("Please provide a valid API key ID").
			Mark(ierr.ErrValidation))
		return
	}

	// Verify ownership
	secretEntity, err := h.secretService.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	userID := types.GetUserID(c.Request.Context())
	if secretEntity.CreatedBy != userID {
		c.Error(ierr.NewError("access denied").
			WithHint("You can only delete your own API keys").
			Mark(ierr.ErrPermissionDenied))
		return
	}

	err = h.secretService.DeleteAPIKey(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
