package secret

import (
	"context"

	"github.com/omkar273/nashikdarshan/internal/types"
)

// Repository defines the interface for secret storage
type Repository interface {
	// Core CRUD operations
	Create(ctx context.Context, secret *Secret) error
	Get(ctx context.Context, id string) (*Secret, error)
	Update(ctx context.Context, secret *Secret) error
	Delete(ctx context.Context, id string) error

	// Specialized operations
	GetByValue(ctx context.Context, hashedValue string) (*Secret, error)
	GetByType(ctx context.Context, userID string, secretType types.SecretType) ([]*Secret, error)
	List(ctx context.Context, userID string) ([]*Secret, error)
}
