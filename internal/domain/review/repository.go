package review

import (
	"context"

	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/shopspring/decimal"
)

type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, review *Review) (*Review, error)
	GetByID(ctx context.Context, id string) (*Review, error)
	Update(ctx context.Context, id string, review *Review) (*Review, error)
	Delete(ctx context.Context, id string) error

	// List and filtering
	List(ctx context.Context, filter *types.ReviewFilter) ([]*Review, error)
	Count(ctx context.Context, filter *types.ReviewFilter) (int, error)

	// Rating aggregation operations
	GetAverageRating(ctx context.Context, entityType types.ReviewEntityType, entityID string) (decimal.Decimal, error)
	GetRatingDistribution(ctx context.Context, entityType types.ReviewEntityType, entityID string) (map[int]int, error)
	GetRatingStats(ctx context.Context, entityType types.ReviewEntityType, entityID string) (*RatingStats, error)

	// Moderation operations
	SetFeatured(ctx context.Context, reviewID string, featured bool) error
	SetVerified(ctx context.Context, reviewID string, verified bool) error

	// Analytics operations (require aggregation, can't be replaced by filters)
	GetAverageRatingByTimeRange(ctx context.Context, entityType types.ReviewEntityType, entityID string, filter *types.TimeRangeFilter) (decimal.Decimal, error)
}

// RatingStats represents aggregated rating statistics for an entity
type RatingStats struct {
	EntityType         types.ReviewEntityType `json:"entity_type"`
	EntityID           string                 `json:"entity_id"`
	AverageRating      decimal.Decimal        `json:"average_rating"`
	TotalReviews       int                    `json:"total_reviews"`
	RatingDistribution map[int]int            `json:"rating_distribution"` // rating -> count
	FiveStarCount      int                    `json:"five_star_count"`
	FourStarCount      int                    `json:"four_star_count"`
	ThreeStarCount     int                    `json:"three_star_count"`
	TwoStarCount       int                    `json:"two_star_count"`
	OneStarCount       int                    `json:"one_star_count"`
	VerifiedReviews    int                    `json:"verified_reviews"`
	ReviewsWithImages  int                    `json:"reviews_with_images"`
	ReviewsWithContent int                    `json:"reviews_with_content"`
}
