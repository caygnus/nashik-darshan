package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/omkar273/nashikdarshan/internal/domain/secret"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/samber/lo"
)

type SecretService interface {
	// API Key operations
	CreateAPIKey(ctx context.Context, userID, name string, keyType types.SecretType, permissions []string) (*secret.Secret, string, error)
	ValidateAPIKey(ctx context.Context, apiKey string) (userID string, permissions []string, valid bool)
	ListAPIKeys(ctx context.Context, userID string) ([]*secret.Secret, error)
	GetAPIKey(ctx context.Context, id string) (*secret.Secret, error)
	UpdateAPIKey(ctx context.Context, id, name string, permissions []string, status types.Status) error
	DeleteAPIKey(ctx context.Context, id string) error
}

type secretService struct {
	ServiceParams
}

// NewSecretService creates a new secret service
func NewSecretService(params ServiceParams) SecretService {
	return &secretService{
		ServiceParams: params,
	}
}

// CreateAPIKey creates a new API key
func (s *secretService) CreateAPIKey(ctx context.Context, userID, name string, keyType types.SecretType, permissions []string) (*secret.Secret, string, error) {
	// Validate input
	if name == "" {
		return nil, "", ierr.NewError("name is required").
			WithHint("Please provide a name for the API key").
			Mark(ierr.ErrValidation)
	}

	// Validate key type
	if err := keyType.Validate(); err != nil {
		return nil, "", err
	}

	// Check if publishable key already exists for this user
	if keyType == types.SecretTypePublishableKey {
		existingKeys, err := s.SecretRepo.GetByType(ctx, userID, types.SecretTypePublishableKey)
		if err != nil {
			return nil, "", ierr.WithError(err).
				WithHint("Failed to check existing keys").
				Mark(ierr.ErrDatabase)
		}

		if len(existingKeys) > 0 {
			return nil, "", ierr.NewError("user already has a publishable key").
				WithHint("Each user can only have one publishable key. Please delete the existing one first.").
				Mark(ierr.ErrInvalidOperation)
		}
	}

	// Generate a new API key
	rawKey, err := s.generateAPIKey()
	if err != nil {
		return nil, "", ierr.WithError(err).
			WithHint("Failed to generate API key").
			Mark(ierr.ErrSystem)
	}

	// Hash the key for storage
	hashedKey := s.EncryptionService.Hash(rawKey)
	prefix := rawKey[:8] // Store first 8 chars as prefix

	// Create the secret entity
	now := time.Now().UTC()
	newSecret := &secret.Secret{
		ID:          types.GenerateUUIDWithPrefix(types.UUID_PREFIX_SECRET),
		Name:        name,
		Type:        keyType,
		Provider:    types.SecretProviderNashikDarshan,
		Value:       hashedKey,
		Prefix:      prefix,
		Permissions: permissions,
		Metadata:    types.NewMetadataFromMap(map[string]string{}),
		BaseModel: types.BaseModel{
			Status:    types.StatusPublished,
			CreatedBy: userID,
			UpdatedBy: userID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Save to repository
	if err := s.SecretRepo.Create(ctx, newSecret); err != nil {
		return nil, "", ierr.WithError(err).
			WithHint("Failed to create API key").
			Mark(ierr.ErrDatabase)
	}

	// Return the created secret and the raw key (which won't be stored)
	return newSecret, rawKey, nil
}

// ValidateAPIKey validates an API key and returns user ID and permissions if valid
// TODO: Implement caching layer (Redis or in-memory) to avoid database hits on every request
func (s *secretService) ValidateAPIKey(ctx context.Context, apiKey string) (string, []string, bool) {
	if apiKey == "" {
		return "", nil, false
	}

	// Hash the API key
	hashedKey := s.EncryptionService.Hash(apiKey)

	// Query database
	secretEntity, err := s.SecretRepo.GetByValue(ctx, hashedKey)
	if err != nil || secretEntity == nil {
		return "", nil, false
	}

	// Check if key is active
	if secretEntity.Status != types.StatusPublished {
		return "", nil, false
	}

	// Check expiration if set
	if secretEntity.ExpiresAt != nil && secretEntity.ExpiresAt.Before(time.Now().UTC()) {
		return "", nil, false
	}

	// Update last used timestamp asynchronously
	go func() {
		secretEntity.LastUsedAt = lo.ToPtr(time.Now().UTC())
		s.SecretRepo.Update(context.Background(), secretEntity)
	}()

	return secretEntity.CreatedBy, secretEntity.Permissions, true
}

// ListAPIKeys lists all API keys for a user
func (s *secretService) ListAPIKeys(ctx context.Context, userID string) ([]*secret.Secret, error) {
	secrets, err := s.SecretRepo.List(ctx, userID)
	if err != nil {
		return nil, ierr.WithError(err).
			WithHint("Failed to list API keys").
			Mark(ierr.ErrDatabase)
	}

	return secrets, nil
}

// GetAPIKey retrieves a specific API key by ID
func (s *secretService) GetAPIKey(ctx context.Context, id string) (*secret.Secret, error) {
	secretEntity, err := s.SecretRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return secretEntity, nil
}

// UpdateAPIKey updates an API key
func (s *secretService) UpdateAPIKey(ctx context.Context, id, name string, permissions []string, status types.Status) error {
	// Get existing secret
	secretEntity, err := s.SecretRepo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Update fields
	if name != "" {
		secretEntity.Name = name
	}
	if permissions != nil {
		secretEntity.Permissions = permissions
	}
	secretEntity.Status = status
	secretEntity.UpdatedBy = types.GetUserID(ctx)
	secretEntity.UpdatedAt = time.Now().UTC()

	// Save updates
	if err := s.SecretRepo.Update(ctx, secretEntity); err != nil {
		return ierr.WithError(err).
			WithHint("Failed to update API key").
			Mark(ierr.ErrDatabase)
	}

	return nil
}

// DeleteAPIKey deletes (archives) an API key
func (s *secretService) DeleteAPIKey(ctx context.Context, id string) error {
	if err := s.SecretRepo.Delete(ctx, id); err != nil {
		return ierr.WithError(err).
			WithHint("Failed to delete API key").
			Mark(ierr.ErrDatabase)
	}

	return nil
}

// generateAPIKey generates a new random API key
func (s *secretService) generateAPIKey() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	return hex.EncodeToString(bytes), nil
}
