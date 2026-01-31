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
		lngF, latF, radiusF := lng.InexactFloat64(), lat.InexactFloat64(), radiusM.InexactFloat64()
		col := s.C(locationColumn)
		s.Where(sql.P(func(b *sql.Builder) {
			b.WriteString("ST_DWithin(").WriteString(col).WriteString(", ST_SetSRID(ST_MakePoint(")
			b.Arg(lngF).WriteString(", ").Arg(latF).WriteString("), 4326)::geography, ")
			b.Arg(radiusF).WriteString(")")
		}))
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
		lngF, latF := lng.InexactFloat64(), lat.InexactFloat64()
		col := s.C(locationColumn)
		s.OrderExpr(sql.ExprFunc(func(b *sql.Builder) {
			b.WriteString("ST_Distance(").WriteString(col).WriteString(", ST_SetSRID(ST_MakePoint(")
			b.Arg(lngF).WriteString(", ").Arg(latF).WriteString("), 4326)::geography) ").WriteString(order)
		}))
	}
}
