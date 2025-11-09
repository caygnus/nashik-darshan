package service

import (
	"context"
	"time"

	"github.com/omkar273/nashikdarshan/internal/api/dto"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type FeedService interface {
	// Core feed operations
	GetFeed(ctx context.Context, req *dto.FeedRequest) (*dto.FeedResponse, error)

	// Engagement tracking
	IncrementViewCount(ctx context.Context, placeID string) error

	// Background operations
	UpdatePopularityScores(ctx context.Context) error
}

type feedService struct {
	ServiceParams
}

// NewFeedService creates a new feed service
func NewFeedService(params ServiceParams) FeedService {
	return &feedService{
		ServiceParams: params,
	}
}

// GetFeed retrieves feed data for multiple sections
func (s *feedService) GetFeed(ctx context.Context, req *dto.FeedRequest) (*dto.FeedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sections := make([]dto.FeedSectionResponse, 0, len(req.Sections))

	// Process each section request
	for _, sectionReq := range req.Sections {
		sectionResponse, err := s.processSectionRequest(ctx, sectionReq, req)
		if err != nil {
			// Log error but continue with other sections (graceful degradation)
			s.Logger.Errorw("failed to process section",
				"section_type", sectionReq.Type,
				"error", err)
			continue
		}

		sections = append(sections, sectionResponse)
	}

	return dto.NewFeedResponse(sections), nil
}

// processSectionRequest handles a single section request
func (s *feedService) processSectionRequest(ctx context.Context, sectionReq dto.FeedSectionRequest, globalReq *dto.FeedRequest) (dto.FeedSectionResponse, error) {
	// Create base filter from request
	filter := sectionReq.ToPlaceFilter(globalReq)

	// Apply section-specific filter modifications
	switch sectionReq.Type {
	case types.SectionTypeLatest:
		// Latest uses default sorting (by created_at desc)
	case types.SectionTypeTrending:
		// Add trending-specific time filter (48 hours lookback for last viewed)
		cutoffTime := time.Now().UTC().Add(-48 * time.Hour)
		filter.LastViewedAfter = &cutoffTime
	case types.SectionTypePopular:
		// Popular uses default sorting (by popularity_score desc)
	case types.SectionTypeNearby:
		// Nearby uses geospatial filtering (already handled in ToPlaceFilter)
	default:
		return dto.FeedSectionResponse{}, ierr.NewError("unsupported section type").
			WithHint("Please use a valid section type").
			Mark(ierr.ErrValidation)
	}

	// Execute the common List + Count pattern
	places, err := s.PlaceRepo.List(ctx, filter)
	if err != nil {
		return dto.FeedSectionResponse{}, err
	}

	total, err := s.PlaceRepo.Count(ctx, filter)
	if err != nil {
		return dto.FeedSectionResponse{}, err
	}

	return dto.NewFeedSectionResponseFromDomain(
		sectionReq.Type,
		places,
		total,
		filter.GetLimit(),
		filter.GetOffset(),
	), nil
}

// IncrementViewCount increments the view count for a place
func (s *feedService) IncrementViewCount(ctx context.Context, placeID string) error {
	// Verify place exists
	_, err := s.PlaceRepo.Get(ctx, placeID)
	if err != nil {
		return err
	}

	// Increment view count
	return s.PlaceRepo.IncrementViewCount(ctx, placeID)
}

// UpdatePopularityScores recalculates popularity scores for all places
func (s *feedService) UpdatePopularityScores(ctx context.Context) error {
	s.Logger.Infow("starting popularity score update")

	// Get all places (no limit)
	filter := types.NewNoLimitPlaceFilter()
	places, err := s.PlaceRepo.ListAll(ctx, filter)
	if err != nil {
		return err
	}

	// Calculate and update popularity score for each place
	for _, place := range places {
		score := place.CalculatePopularityScore()
		err = s.PlaceRepo.UpdatePopularityScore(ctx, place.ID, score)
		if err != nil {
			s.Logger.Errorw("failed to update popularity score", "place_id", place.ID, "error", err)
			return err
		}
	}

	s.Logger.Infow("completed popularity score update", "places_updated", len(places))
	return nil
}
