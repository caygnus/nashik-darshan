package service

import (
	"context"

	"github.com/omkar273/nashikdarshan/internal/api/dto"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/types"
)

// ReviewService defines the interface for review business logic
type ReviewService interface {
	// Basic CRUD operations
	CreateReview(ctx context.Context, req *dto.CreateReviewRequest) (*dto.ReviewResponse, error)
	GetReview(ctx context.Context, id string) (*dto.ReviewResponse, error)
	UpdateReview(ctx context.Context, id string, req *dto.UpdateReviewRequest) (*dto.ReviewResponse, error)
	DeleteReview(ctx context.Context, id string) error

	// List and filtering
	ListReviews(ctx context.Context, filter *types.ReviewFilter) (types.ListResponse[*dto.ReviewResponse], error)

	// Rating and statistics
	GetRatingStats(ctx context.Context, req *dto.GetRatingStatsRequest) (*dto.RatingStatsResponse, error)
}

// reviewService implements ReviewService
type reviewService struct {
	ServiceParams
}

// NewReviewService creates a new review service
func NewReviewService(params ServiceParams) ReviewService {
	return &reviewService{
		ServiceParams: params,
	}
}

// CreateReview creates a new review
func (s *reviewService) CreateReview(ctx context.Context, req *dto.CreateReviewRequest) (*dto.ReviewResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	reviewModel := req.ToReview(ctx)

	createdReview, err := s.ReviewRepo.Create(ctx, reviewModel)
	if err != nil {
		return nil, ierr.WithError(err).
			WithMessage("failed to create review").
			Mark(ierr.ErrDatabase)
	}

	return dto.NewReviewResponse(createdReview), nil
}

// GetReview retrieves a review by ID
func (s *reviewService) GetReview(ctx context.Context, id string) (*dto.ReviewResponse, error) {
	reviewModel, err := s.ReviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.NewReviewResponse(reviewModel), nil
}

// UpdateReview updates an existing review
func (s *reviewService) UpdateReview(ctx context.Context, id string, req *dto.UpdateReviewRequest) (*dto.ReviewResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	review, err := s.ReviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Rating != nil {
		review.Rating = *req.Rating
	}
	if req.Title != nil {
		review.Title = req.Title
	}
	if req.Content != nil {
		review.Content = req.Content
	}
	if req.Tags != nil {
		review.Tags = req.Tags
	}
	if req.Images != nil {
		review.Images = req.Images
	}

	updatedReview, err := s.ReviewRepo.Update(ctx, id, review)
	if err != nil {
		return nil, ierr.WithError(err).
			WithMessage("failed to update review").
			Mark(ierr.ErrDatabase)
	}

	return dto.NewReviewResponse(updatedReview), nil
}

// DeleteReview deletes a review
func (s *reviewService) DeleteReview(ctx context.Context, id string) error {
	err := s.ReviewRepo.Delete(ctx, id)
	if err != nil {
		return ierr.WithError(err).
			WithMessage("failed to delete review").
			Mark(ierr.ErrDatabase)
	}

	return nil
}

// ListReviews lists reviews with filtering
func (s *reviewService) ListReviews(ctx context.Context, filter *types.ReviewFilter) (types.ListResponse[*dto.ReviewResponse], error) {
	if err := filter.Validate(); err != nil {
		return types.ListResponse[*dto.ReviewResponse]{}, err
	}

	reviews, err := s.ReviewRepo.List(ctx, filter)
	if err != nil {
		return types.ListResponse[*dto.ReviewResponse]{}, ierr.WithError(err).
			WithMessage("failed to list reviews").
			Mark(ierr.ErrDatabase)
	}

	total, err := s.ReviewRepo.Count(ctx, filter)
	if err != nil {
		return types.ListResponse[*dto.ReviewResponse]{}, ierr.WithError(err).
			WithMessage("failed to count reviews").
			Mark(ierr.ErrDatabase)
	}

	// Convert to DTOs
	reviewResponses := make([]*dto.ReviewResponse, len(reviews))
	for i, r := range reviews {
		reviewResponses[i] = dto.NewReviewResponse(r)
	}

	return types.NewListResponse(reviewResponses, total, filter.GetLimit(), filter.GetOffset()), nil
}

// GetRatingStats gets rating statistics for an entity
func (s *reviewService) GetRatingStats(ctx context.Context, req *dto.GetRatingStatsRequest) (*dto.RatingStatsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	stats, err := s.ReviewRepo.GetRatingStats(ctx, req.EntityType, req.EntityID)
	if err != nil {
		return nil, ierr.WithError(err).
			WithMessage("failed to get rating stats").
			Mark(ierr.ErrDatabase)
	}

	return dto.NewRatingStatsResponse(stats), nil
}
