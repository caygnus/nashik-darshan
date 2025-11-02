package category

import (
	"context"

	"github.com/omkar273/nashikdarshan/internal/types"
)

// Repository defines the interface for category persistence operations
type Repository interface {
	// Core operations
	Create(ctx context.Context, category *Category) error
	Get(ctx context.Context, id string) (*Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, category *Category) error

	// List operations
	List(ctx context.Context, filter *types.CategoryFilter) ([]*Category, error)
	ListAll(ctx context.Context, filter *types.CategoryFilter) ([]*Category, error)
	Count(ctx context.Context, filter *types.CategoryFilter) (int, error)
}
