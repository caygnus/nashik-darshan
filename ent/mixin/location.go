package mixin

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/shopspring/decimal"
)

// GeoPoint represents a PostGIS geography Point with longitude and latitude
type GeoPoint struct {
	Lng decimal.Decimal
	Lat decimal.Decimal
}

// Value implements driver.Valuer interface
// Returns PostGIS WKT format: "POINT(lng lat)"
func (p GeoPoint) Value() (driver.Value, error) {
	if p.Lat.IsZero() && p.Lng.IsZero() {
		return nil, nil
	}
	// PostGIS expects POINT(longitude latitude) - note the order!
	return fmt.Sprintf("POINT(%f %f)", p.Lng.InexactFloat64(), p.Lat.InexactFloat64()), nil
}

// Scan implements sql.Scanner interface
// Handles PostGIS output in various formats (WKT, EWKB, etc.)
func (p *GeoPoint) Scan(src interface{}) error {
	if src == nil {
		p.Lng = decimal.Zero
		p.Lat = decimal.Zero
		return nil
	}

	switch v := src.(type) {
	case string:
		// Handle WKT format: "POINT(lng lat)" or "SRID=4326;POINT(lng lat)"
		return p.scanWKT(v)
	case []byte:
		// Handle byte array (could be WKT or EWKB)
		return p.scanWKT(string(v))
	default:
		return fmt.Errorf("unsupported type for GeoPoint: %T", src)
	}
}

// scanWKT parses WKT format string
func (p *GeoPoint) scanWKT(wkt string) error {
	// Remove SRID prefix if present
	if idx := strings.Index(wkt, ";"); idx != -1 {
		wkt = wkt[idx+1:]
	}

	// Parse "POINT(lng lat)" format
	var lng, lat float64
	_, err := fmt.Sscanf(wkt, "POINT(%f %f)", &lng, &lat)
	if err != nil {
		return fmt.Errorf("failed to parse GeoPoint WKT: %w", err)
	}

	p.Lng = decimal.NewFromFloat(lng)
	p.Lat = decimal.NewFromFloat(lat)
	return nil
}

// ToLocation converts GeoPoint to lat, lng decimals
func (p GeoPoint) ToLocation() (decimal.Decimal, decimal.Decimal) {
	return p.Lat, p.Lng
}

// NewGeoPoint creates GeoPoint from latitude and longitude
func NewGeoPoint(lat, lng decimal.Decimal) GeoPoint {
	return GeoPoint{
		Lat: lat,
		Lng: lng,
	}
}

// LocationMixin adds a PostGIS geography(Point, 4326) location field
// with a GiST spatial index for efficient geospatial queries
type LocationMixin struct {
	mixin.Schema
}

// Fields of the LocationMixin
func (LocationMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Other("location", &GeoPoint{}).
			SchemaType(map[string]string{
				dialect.Postgres: "geography(Point,4326)",
			}).
			Comment("PostGIS geography point for geospatial queries"),
	}
}

// Indexes of the LocationMixin
func (LocationMixin) Indexes() []ent.Index {
	return []ent.Index{
		// GiST index for spatial queries (ST_DWithin, ST_Distance, etc.)
		// Note: GiST index will be created via migration script
		// Ent doesn't directly support USING GIST in index annotations,
		// so we'll create it manually in the migration
		index.Fields("location"),
	}
}
