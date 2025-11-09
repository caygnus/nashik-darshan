package types

import (
	"github.com/shopspring/decimal"
)

// Location represents a geographic location with latitude and longitude (WGS84)
type Location struct {
	Latitude  decimal.Decimal `json:"latitude"`
	Longitude decimal.Decimal `json:"longitude"`
}

// NewLocation creates a new Location
func NewLocation(lat, lng decimal.Decimal) *Location {
	return &Location{
		Latitude:  lat,
		Longitude: lng,
	}
}

// Validate validates the Location coordinates
func (l Location) Validate() error {
	return ValidateCoordinates(l.Latitude, l.Longitude)
}

// IsValid returns true if the location has valid coordinates
func (l Location) IsValid() bool {
	return l.Validate() == nil
}

// IsZero returns true if both coordinates are zero
func (l Location) IsZero() bool {
	return l.Latitude.IsZero() && l.Longitude.IsZero()
}
