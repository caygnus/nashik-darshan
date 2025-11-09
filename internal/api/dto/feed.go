package dto

import (
	"github.com/omkar273/nashikdarshan/internal/domain/place"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/shopspring/decimal"
)

// FeedSectionRequest represents a single section request in the feed
type FeedSectionRequest struct {
	Type types.FeedSectionType `json:"type" binding:"required" validate:"required"`

	// Section-specific filters (override global filters if provided)
	*types.QueryFilter     `json:",inline,omitempty"`
	*types.TimeRangeFilter `json:",inline,omitempty"`

	// Geospatial fields (for nearby section)
	Latitude  *decimal.Decimal `json:"latitude,omitempty" validate:"omitempty"`
	Longitude *decimal.Decimal `json:"longitude,omitempty" validate:"omitempty"`
	RadiusKm  *decimal.Decimal `json:"radius_km,omitempty" validate:"omitempty,min=0.1,max=50" default:"5"`
}

// FeedRequest represents the main feed request
type FeedRequest struct {
	Sections []FeedSectionRequest `json:"sections" binding:"required,min=1,max=10" validate:"required,min=1,max=10,dive"`

	// Global filters (applied to all sections unless overridden)
	*types.QueryFilter     `json:",inline,omitempty"`
	*types.TimeRangeFilter `json:",inline,omitempty"`
}

// FeedSectionResponse represents a single section response in the feed
type FeedSectionResponse struct {
	Type       types.FeedSectionType    `json:"type"`
	Items      []*PlaceResponse         `json:"items"`
	Pagination types.PaginationResponse `json:"pagination"`
}

// FeedResponse represents the main feed response
type FeedResponse struct {
	Sections []FeedSectionResponse `json:"sections"`
}

// Validate validates the FeedRequest
func (req *FeedRequest) Validate() error {
	// Validate global filters
	if req.QueryFilter != nil {
		if err := req.QueryFilter.Validate(); err != nil {
			return err
		}
	}
	if req.TimeRangeFilter != nil {
		if err := req.TimeRangeFilter.Validate(); err != nil {
			return err
		}
	}

	// Validate each section
	for _, section := range req.Sections {
		if err := section.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates a single FeedSectionRequest
func (req *FeedSectionRequest) Validate() error {
	// Validate section type
	if err := req.Type.Validate(); err != nil {
		return err
	}

	// Validate section-specific filters
	if req.QueryFilter != nil {
		if err := req.QueryFilter.Validate(); err != nil {
			return err
		}
	}
	if req.TimeRangeFilter != nil {
		if err := req.TimeRangeFilter.Validate(); err != nil {
			return err
		}
	}

	// Validate geospatial fields for nearby section
	if req.Type == types.SectionTypeNearby {
		if req.Latitude == nil || req.Longitude == nil {
			return ierr.NewError("latitude and longitude are required for nearby section").
				WithHint("Please provide both latitude and longitude for nearby section").
				Mark(ierr.ErrValidation)
		}

		// Create location and validate coordinates
		location := types.NewLocation(*req.Latitude, *req.Longitude)
		if err := location.Validate(); err != nil {
			return err
		}

		// Validate radius if provided
		if req.RadiusKm != nil {
			if req.RadiusKm.LessThanOrEqual(decimal.Zero) {
				return ierr.NewError("radius_km must be greater than 0").
					WithHint("Please provide a positive radius value").
					Mark(ierr.ErrValidation)
			}
			if req.RadiusKm.GreaterThan(decimal.NewFromInt(50)) {
				return ierr.NewError("radius_km cannot exceed 50 kilometers").
					WithHint("Please provide a radius value within 50km limit").
					Mark(ierr.ErrValidation)
			}
		}
	}

	return nil
}

// ToPlaceFilter converts a FeedSectionRequest to a PlaceFilter, merging with global filters
func (req *FeedSectionRequest) ToPlaceFilter(globalFilter *FeedRequest) *types.PlaceFilter {
	filter := types.NewPlaceFilter()

	// Start with global filters if provided
	if globalFilter != nil {
		if globalFilter.QueryFilter != nil {
			filter.QueryFilter = globalFilter.QueryFilter
		}
		if globalFilter.TimeRangeFilter != nil {
			filter.TimeRangeFilter = globalFilter.TimeRangeFilter
		}
	}

	// Override with section-specific filters if provided
	if req.QueryFilter != nil {
		if filter.QueryFilter == nil {
			filter.QueryFilter = req.QueryFilter
		} else {
			// Merge section filters into global filters
			filter.QueryFilter.Merge(*req.QueryFilter)
		}
	}
	if req.TimeRangeFilter != nil {
		filter.TimeRangeFilter = req.TimeRangeFilter
	}

	// Ensure we have default query filter if none provided
	if filter.QueryFilter == nil {
		filter.QueryFilter = types.NewDefaultQueryFilter()
	}

	// Configure default sorting based on section type (can be overridden by filters)
	if filter.QueryFilter.Sort == nil || filter.QueryFilter.Order == nil {
		switch req.Type {
		case types.SectionTypeLatest:
			if filter.QueryFilter.Sort == nil {
				sort := "created_at"
				filter.QueryFilter.Sort = &sort
			}
			if filter.QueryFilter.Order == nil {
				order := "desc"
				filter.QueryFilter.Order = &order
			}

		case types.SectionTypeTrending, types.SectionTypePopular:
			if filter.QueryFilter.Sort == nil {
				sort := "popularity_score"
				filter.QueryFilter.Sort = &sort
			}
			if filter.QueryFilter.Order == nil {
				order := "desc"
				filter.QueryFilter.Order = &order
			}

		case types.SectionTypeNearby:
			// For nearby, we still want popularity sorting within the geographic area
			if filter.QueryFilter.Sort == nil {
				sort := "popularity_score"
				filter.QueryFilter.Sort = &sort
			}
			if filter.QueryFilter.Order == nil {
				order := "desc"
				filter.QueryFilter.Order = &order
			}
		}
	}

	// Handle geospatial fields for nearby section
	if req.Type == types.SectionTypeNearby && req.Latitude != nil && req.Longitude != nil {
		filter.Latitude = req.Latitude
		filter.Longitude = req.Longitude

		// Set default radius if not provided (5km)
		if req.RadiusKm != nil {
			radiusM := req.RadiusKm.Mul(decimal.NewFromInt(1000)) // Convert km to meters
			filter.RadiusM = &radiusM
		} else {
			defaultRadiusM := decimal.NewFromInt(5000) // 5km default
			filter.RadiusM = &defaultRadiusM
		}
	}

	return filter
}

// NewFeedResponse creates a new FeedResponse
func NewFeedResponse(sections []FeedSectionResponse) *FeedResponse {
	return &FeedResponse{
		Sections: sections,
	}
}

// NewFeedSectionResponse creates a new FeedSectionResponse
func NewFeedSectionResponse(sectionType types.FeedSectionType, places []*PlaceResponse, total, limit, offset int) FeedSectionResponse {
	return FeedSectionResponse{
		Type:       sectionType,
		Items:      places,
		Pagination: types.NewPaginationResponse(total, limit, offset),
	}
}

// NewFeedSectionResponseFromDomain creates a FeedSectionResponse from domain places
func NewFeedSectionResponseFromDomain(sectionType types.FeedSectionType, places []*place.Place, total, limit, offset int) FeedSectionResponse {
	// Convert domain places to DTOs
	placeResponses := make([]*PlaceResponse, len(places))
	for i, p := range places {
		placeResponses[i] = NewPlaceResponse(p)
	}

	return NewFeedSectionResponse(sectionType, placeResponses, total, limit, offset)
}
