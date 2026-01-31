package types

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// GeoPoint represents a PostGIS geography Point (WGS84) and is the single source of truth for location.
// JSON and form use "latitude" and "longitude" for API compatibility.
type GeoPoint struct {
	Lat decimal.Decimal `json:"latitude" form:"latitude"`
	Lng decimal.Decimal `json:"longitude" form:"longitude"`
}

// Value implements driver.Valuer. Returns PostGIS WKT format: "POINT(lng lat)".
func (p GeoPoint) Value() (driver.Value, error) {
	if p.Lat.IsZero() && p.Lng.IsZero() {
		return nil, nil
	}
	return fmt.Sprintf("POINT(%f %f)", p.Lng.InexactFloat64(), p.Lat.InexactFloat64()), nil
}

// Scan implements sql.Scanner. Handles PostGIS output (WKT, etc.).
func (p *GeoPoint) Scan(src interface{}) error {
	if src == nil {
		p.Lng = decimal.Zero
		p.Lat = decimal.Zero
		return nil
	}
	switch v := src.(type) {
	case string:
		return p.scanWKT(v)
	case []byte:
		return p.scanWKT(string(v))
	default:
		return fmt.Errorf("unsupported type for GeoPoint: %T", src)
	}
}

func (p *GeoPoint) scanWKT(wkt string) error {
	if idx := strings.Index(wkt, ";"); idx != -1 {
		wkt = wkt[idx+1:]
	}
	var lng, lat float64
	_, err := fmt.Sscanf(wkt, "POINT(%f %f)", &lng, &lat)
	if err != nil {
		return fmt.Errorf("failed to parse GeoPoint WKT: %w", err)
	}
	p.Lng = decimal.NewFromFloat(lng)
	p.Lat = decimal.NewFromFloat(lat)
	return nil
}

// Validate validates the GeoPoint coordinates.
func (p GeoPoint) Validate() error {
	return ValidateCoordinates(p.Lat, p.Lng)
}

// IsZero returns true if both coordinates are zero.
func (p GeoPoint) IsZero() bool {
	return p.Lat.IsZero() && p.Lng.IsZero()
}

// NewGeoPoint creates a GeoPoint from latitude and longitude.
func NewGeoPoint(lat, lng decimal.Decimal) GeoPoint {
	return GeoPoint{Lat: lat, Lng: lng}
}
