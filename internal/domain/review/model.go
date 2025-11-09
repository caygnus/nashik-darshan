package review

import (
	"github.com/omkar273/nashikdarshan/ent"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

type Review struct {
	ID              string                 `json:"id" db:"id"`
	EntityType      types.ReviewEntityType `json:"entity_type" db:"entity_type"`
	EntityID        string                 `json:"entity_id" db:"entity_id"`
	UserID          string                 `json:"user_id" db:"user_id"`
	Rating          decimal.Decimal        `json:"rating" db:"rating"`
	Title           *string                `json:"title,omitempty" db:"title"`
	Content         *string                `json:"content,omitempty" db:"content"`
	Tags            []string               `json:"tags" db:"tags"`
	Images          []string               `json:"images" db:"images"`
	HelpfulCount    int                    `json:"helpful_count" db:"helpful_count"`
	NotHelpfulCount int                    `json:"not_helpful_count" db:"not_helpful_count"`
	IsVerified      bool                   `json:"is_verified" db:"is_verified"`
	IsFeatured      bool                   `json:"is_featured" db:"is_featured"`
	types.BaseModel
}

// FromEnt converts ent.Review to domain Review
func FromEnt(review *ent.Review) *Review {
	r := &Review{
		ID:              review.ID,
		EntityType:      types.ReviewEntityType(review.EntityType),
		EntityID:        review.EntityID,
		UserID:          review.UserID,
		Rating:          review.Rating,
		Title:           lo.ToPtr(review.Title),
		Content:         lo.ToPtr(review.Content),
		Tags:            review.Tags,
		Images:          review.Images,
		HelpfulCount:    review.HelpfulCount,
		NotHelpfulCount: review.NotHelpfulCount,
		IsVerified:      review.IsVerified,
		IsFeatured:      review.IsFeatured,
		BaseModel: types.BaseModel{
			Status:    types.Status(review.Status),
			CreatedAt: review.CreatedAt,
			UpdatedAt: review.UpdatedAt,
			CreatedBy: review.CreatedBy,
			UpdatedBy: review.UpdatedBy,
		},
	}

	return r
}

// FromEntList converts a list of ent.Review to domain Review
func FromEntList(reviews []*ent.Review) []*Review {
	return lo.Map(reviews, func(review *ent.Review, _ int) *Review {
		return FromEnt(review)
	})
}

// GetHelpfulnessRatio returns the ratio of helpful votes to total votes
func (r *Review) GetHelpfulnessRatio() float64 {
	totalVotes := r.HelpfulCount + r.NotHelpfulCount
	if totalVotes == 0 {
		return 0.0
	}
	return float64(r.HelpfulCount) / float64(totalVotes)
}

// GetEngagementScore calculates an engagement score based on helpfulness and rating
func (r *Review) GetEngagementScore() decimal.Decimal {
	// Base score from rating (1-5 scale)
	baseScore := r.Rating

	// Helpfulness boost (up to 2x multiplier)
	helpfulnessRatio := decimal.NewFromFloat(r.GetHelpfulnessRatio())
	helpfulnessBoost := decimal.NewFromFloat(1.0).Add(helpfulnessRatio)

	// Content quality boost
	contentBoost := decimal.NewFromFloat(1.0)
	if r.Content != nil && len(*r.Content) > 50 {
		contentBoost = decimal.NewFromFloat(1.2) // 20% boost for detailed reviews
	}
	if len(r.Images) > 0 {
		contentBoost = contentBoost.Add(decimal.NewFromFloat(0.1)) // 10% boost for images
	}

	// Verification boost
	verificationBoost := decimal.NewFromFloat(1.0)
	if r.IsVerified {
		verificationBoost = decimal.NewFromFloat(1.3) // 30% boost for verified reviews
	}

	return baseScore.
		Mul(helpfulnessBoost).
		Mul(contentBoost).
		Mul(verificationBoost)
}

// HasContent returns true if the review has meaningful content
func (r *Review) HasContent() bool {
	return r.Content != nil && len(*r.Content) > 0
}

// HasImages returns true if the review has attached images
func (r *Review) HasImages() bool {
	return len(r.Images) > 0
}
