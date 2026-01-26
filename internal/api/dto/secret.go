package dto

import (
	"time"

	"github.com/omkar273/nashikdarshan/internal/domain/secret"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/omkar273/nashikdarshan/internal/validator"
	"github.com/samber/lo"
)

type CreateAPIKeyRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=255"`
	Type        string   `json:"type" binding:"required,oneof=private_key publishable_key"`
	Permissions []string `json:"permissions,omitempty"`
}

// Validate validates the CreateAPIKeyRequest
func (req *CreateAPIKeyRequest) Validate() error {
	// Validate struct tags
	if err := validator.ValidateRequest(req); err != nil {
		return err
	}

	// Validate type using enum
	secretType := types.SecretType(req.Type)
	if err := secretType.Validate(); err != nil {
		return err
	}

	return nil
}

// ToSecretType converts the string type to types.SecretType
func (req *CreateAPIKeyRequest) ToSecretType() types.SecretType {
	return types.SecretType(req.Type)
}

type UpdateAPIKeyRequest struct {
	Name        *string       `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Permissions []string      `json:"permissions,omitempty"`
	Status      *types.Status `json:"status,omitempty" binding:"omitempty,oneof=published archived"`
}

// Validate validates the UpdateAPIKeyRequest
func (req *UpdateAPIKeyRequest) Validate() error {
	// Validate struct tags
	if err := validator.ValidateRequest(req); err != nil {
		return err
	}

	// Validate status if provided
	if req.Status != nil {
		if *req.Status != types.StatusPublished && *req.Status != types.StatusArchived {
			return ierr.NewError("status must be 'published' or 'archived'").
				WithHint("Please provide a valid status").
				Mark(ierr.ErrValidation)
		}
	}

	return nil
}

// APIKeyResponse represents an API key in responses (never includes raw key)
type APIKeyResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Prefix      string    `json:"prefix"`
	Permissions []string  `json:"permissions"`
	Status      string    `json:"status"`
	ExpiresAt   *string   `json:"expires_at,omitempty"`
	LastUsedAt  *string   `json:"last_used_at,omitempty"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// FromSecret converts a domain Secret to APIKeyResponse
func FromSecret(s *secret.Secret) *APIKeyResponse {
	resp := &APIKeyResponse{
		ID:          s.ID,
		Name:        s.Name,
		Type:        string(s.Type),
		Prefix:      s.Prefix,
		Permissions: s.Permissions,
		Status:       string(s.Status),
		CreatedAt:    s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    s.UpdatedAt.Format(time.RFC3339),
	}

	if s.ExpiresAt != nil {
		expiresAt := s.ExpiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &expiresAt
	}

	if s.LastUsedAt != nil {
		lastUsedAt := s.LastUsedAt.Format(time.RFC3339)
		resp.LastUsedAt = &lastUsedAt
	}

	return resp
}

// CreateAPIKeyResponse includes the raw key (only returned once)
type CreateAPIKeyResponse struct {
	*APIKeyResponse
	Key string `json:"key"` // Raw key, only returned once
}

// ListAPIKeysResponse represents a list of API keys
type ListAPIKeysResponse struct {
	Items []*APIKeyResponse `json:"items"`
	Total int               `json:"total"`
}

// NewListAPIKeysResponse creates a new list response
func NewListAPIKeysResponse(secrets []*secret.Secret) *ListAPIKeysResponse {
	items := lo.Map(secrets, func(s *secret.Secret, _ int) *APIKeyResponse {
		return FromSecret(s)
	})

	return &ListAPIKeysResponse{
		Items: items,
		Total: len(items),
	}
}
