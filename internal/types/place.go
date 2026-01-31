package types

import (
	"strings"

	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

type PlaceType string

const (
	PlaceTypeTemple     PlaceType = "temple"
	PlaceTypeRestaurant PlaceType = "restaurant"
	PlaceTypeMuseum     PlaceType = "museum"
	PlaceTypePark       PlaceType = "park"
	PlaceTypeExperience PlaceType = "experience"
)

func (pt PlaceType) Validate() error {
	allowedPlaceTypes := []string{
		string(PlaceTypeTemple),
		string(PlaceTypeRestaurant),
		string(PlaceTypeMuseum),
		string(PlaceTypePark),
		string(PlaceTypeExperience),
	}
	if !lo.Contains(allowedPlaceTypes, string(pt)) {
		return ierr.NewError("invalid place type").
			WithHintf("valid place types are: %s", strings.Join(allowedPlaceTypes, ", ")).
			WithReportableDetails(map[string]any{"place_type": pt}).
			Mark(ierr.ErrValidation)
	}
	return nil
}

// ValidateCoordinates validates latitude and longitude values
func ValidateCoordinates(latitude, longitude decimal.Decimal) error {
	// Validate latitude range (-90 to 90)
	if latitude.LessThan(decimal.NewFromInt(-90)) || latitude.GreaterThan(decimal.NewFromInt(90)) {
		return ierr.NewError("latitude must be between -90 and 90").
			WithHint("Please provide a valid latitude value").
			Mark(ierr.ErrValidation)
	}

	// Validate longitude range (-180 to 180)
	if longitude.LessThan(decimal.NewFromInt(-180)) || longitude.GreaterThan(decimal.NewFromInt(180)) {
		return ierr.NewError("longitude must be between -180 and 180").
			WithHint("Please provide a valid longitude value").
			Mark(ierr.ErrValidation)
	}

	return nil
}

type PlaceFilter struct {
	*QueryFilter
	*TimeRangeFilter

	// Custom filters
	Slug       []string `json:"slug,omitempty" form:"slug" validate:"omitempty"`
	PlaceTypes []string `json:"place_types,omitempty" form:"place_types" validate:"omitempty"`

	// Geospatial filters (center point + radius in meters)
	Center  *GeoPoint        `json:"center,omitempty" form:"center"`
	RadiusM *decimal.Decimal `json:"radius_m,omitempty" form:"radius_m" validate:"omitempty"` // radius in meters (cap: 15km for v1)

	// Search
	SearchQuery *string `json:"search_query,omitempty" form:"search_query" validate:"omitempty"`

	// Optional quality gate (e.g. min reviews / min rating to appear in feed)
	MinRatingCount *int             `json:"min_rating_count,omitempty" form:"min_rating_count" validate:"omitempty,min=0"`
	MinRatingAvg   *decimal.Decimal `json:"min_rating_avg,omitempty" form:"min_rating_avg" validate:"omitempty"`
}

// ApplyFlatGeospatialParams sets Center and RadiusM from flat query params (latitude, longitude, radius_m or radius_km)
// when Center was not bound from nested form. Call after ShouldBindQuery. Only sets both when radius is present.
func (f *PlaceFilter) ApplyFlatGeospatialParams(latStr, lngStr, radiusMStr, radiusKmStr string) {
	latStr = strings.TrimSpace(latStr)
	lngStr = strings.TrimSpace(lngStr)
	radiusMStr = strings.TrimSpace(radiusMStr)
	radiusKmStr = strings.TrimSpace(radiusKmStr)
	if latStr == "" || lngStr == "" {
		return
	}
	var radiusM decimal.Decimal
	if radiusMStr != "" {
		var err error
		radiusM, err = decimal.NewFromString(radiusMStr)
		if err != nil {
			return
		}
	} else if radiusKmStr != "" {
		radiusKm, err := decimal.NewFromString(radiusKmStr)
		if err != nil {
			return
		}
		radiusM = radiusKm.Mul(decimal.NewFromInt(1000))
	} else {
		return
	}
	lat, errLat := decimal.NewFromString(latStr)
	lng, errLng := decimal.NewFromString(lngStr)
	if errLat != nil || errLng != nil {
		return
	}
	f.Center = &GeoPoint{Lat: lat, Lng: lng}
	f.RadiusM = &radiusM
}

func (f *PlaceFilter) Validate() error {
	if f.QueryFilter != nil {
		if err := f.QueryFilter.Validate(); err != nil {
			return err
		}
	}

	if f.TimeRangeFilter != nil {
		if err := f.TimeRangeFilter.Validate(); err != nil {
			return err
		}
	}

	// Validate place types
	if len(f.PlaceTypes) > 0 {
		for _, pt := range f.PlaceTypes {
			if err := PlaceType(pt).Validate(); err != nil {
				return err
			}
		}
	}

	// Optional quality gate
	if f.MinRatingCount != nil && *f.MinRatingCount < 0 {
		return ierr.NewError("min_rating_count must be >= 0").
			WithHint("Please provide a non-negative min_rating_count").
			Mark(ierr.ErrValidation)
	}
	if f.MinRatingAvg != nil && (f.MinRatingAvg.LessThan(decimal.Zero) || f.MinRatingAvg.GreaterThan(decimal.NewFromInt(5))) {
		return ierr.NewError("min_rating_avg must be between 0 and 5").
			WithHint("Please provide min_rating_avg in range [0, 5]").
			Mark(ierr.ErrValidation)
	}

	// Validate geospatial filters
	if f.Center != nil || f.RadiusM != nil {
		if f.Center == nil || f.RadiusM == nil {
			return ierr.NewError("center and radius_m must both be provided for geospatial search").
				WithHint("Please provide center (latitude/longitude) and radius_m").
				Mark(ierr.ErrValidation)
		}
		if err := f.Center.Validate(); err != nil {
			return err
		}
		// Cap radius at 15km (15000m) for v1
		if f.RadiusM.GreaterThan(decimal.NewFromInt(15000)) {
			return ierr.NewError("radius_m cannot exceed 15000 meters (15km)").
				WithHint("Maximum allowed radius is 15km (15000 meters)").
				Mark(ierr.ErrValidation)
		}
		// Radius must be positive
		if f.RadiusM.LessThanOrEqual(decimal.Zero) {
			return ierr.NewError("radius_m must be greater than 0").
				WithHint("Please provide a positive value for radius_m").
				Mark(ierr.ErrValidation)
		}
	}

	return nil
}

func NewPlaceFilter() *PlaceFilter {
	return &PlaceFilter{
		QueryFilter:     NewDefaultQueryFilter(),
		TimeRangeFilter: &TimeRangeFilter{},
	}
}

func NewNoLimitPlaceFilter() *PlaceFilter {
	return &PlaceFilter{
		QueryFilter:     NewNoLimitQueryFilter(),
		TimeRangeFilter: &TimeRangeFilter{},
	}
}

// GetLimit implements BaseFilter interface
func (f *PlaceFilter) GetLimit() int {
	if f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetLimit()
	}
	return f.QueryFilter.GetLimit()
}

// GetOffset implements BaseFilter interface
func (f *PlaceFilter) GetOffset() int {
	if f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetOffset()
	}
	return f.QueryFilter.GetOffset()
}

// GetStatus implements BaseFilter interface
func (f *PlaceFilter) GetStatus() string {
	if f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetStatus()
	}
	return f.QueryFilter.GetStatus()
}

// GetSort implements BaseFilter interface
func (f *PlaceFilter) GetSort() string {
	if f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetSort()
	}
	return f.QueryFilter.GetSort()
}

// GetOrder implements BaseFilter interface
func (f *PlaceFilter) GetOrder() string {
	if f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetOrder()
	}
	return f.QueryFilter.GetOrder()
}

// GetExpand implements BaseFilter interface
func (f *PlaceFilter) GetExpand() Expand {
	if f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetExpand()
	}
	return f.QueryFilter.GetExpand()
}

func (f *PlaceFilter) IsUnlimited() bool {
	if f.QueryFilter == nil {
		return NewDefaultQueryFilter().IsUnlimited()
	}
	return f.QueryFilter.IsUnlimited()
}

// FeedSectionType represents the type of feed section
type FeedSectionType string

const (
	SectionTypeLatest   FeedSectionType = "latest"
	SectionTypeTrending FeedSectionType = "trending"
	SectionTypePopular  FeedSectionType = "popular"
	SectionTypeNearby   FeedSectionType = "nearby"
	SectionTypeDiscover FeedSectionType = "discover"
)

// FeedSectionTypes contains all valid feed section types
var FeedSectionTypes = []string{
	string(SectionTypeLatest),
	string(SectionTypeTrending),
	string(SectionTypePopular),
	string(SectionTypeNearby),
	string(SectionTypeDiscover),
}

func (f FeedSectionType) Validate() error {
	if !lo.Contains(FeedSectionTypes, string(f)) {
		return ierr.NewError("invalid section type").
			WithHintf("valid types are: latest, trending, popular, nearby, discover").
			WithReportableDetails(map[string]any{"section_type": f}).
			Mark(ierr.ErrValidation)
	}
	return nil
}
