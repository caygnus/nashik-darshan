package ent

import (
	"entgo.io/ent/dialect/sql"
	"github.com/shopspring/decimal"
)

// WithinRadius creates a PostGIS ST_DWithin predicate for finding locations within a radius
// locationColumn: the database column name (typically "location")
// lng: longitude of the center point
// lat: latitude of the center point
// radiusM: radius in meters
// Returns a SQL predicate that can be used in Ent queries
func WithinRadius(locationColumn string, lng, lat decimal.Decimal, radiusM decimal.Decimal) func(*sql.Selector) {
	return func(s *sql.Selector) {
		// ST_DWithin(location, ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography, radiusM)
		// Note: locationColumn must be the actual column name in the table
		s.Where(sql.ExprP(
			"ST_DWithin("+locationColumn+", ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
			lng.InexactFloat64(),
			lat.InexactFloat64(),
			radiusM.InexactFloat64(),
		))
	}
}

// OrderByDistance creates a PostGIS ST_Distance ordering expression
// locationColumn: the database column name (typically "location")
// lng: longitude of the center point
// lat: latitude of the center point
// ascending: true for nearest first, false for farthest first
// Returns a SQL order expression that can be used in Ent queries
func OrderByDistance(locationColumn string, lng, lat decimal.Decimal, ascending bool) func(*sql.Selector) {
	return func(s *sql.Selector) {
		order := "ASC"
		if !ascending {
			order = "DESC"
		}
		// ST_Distance(location, ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography) ASC/DESC
		s.OrderExpr(sql.ExprP(
			"ST_Distance("+locationColumn+", ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography) "+order,
			lng.InexactFloat64(),
			lat.InexactFloat64(),
		))
	}
}
