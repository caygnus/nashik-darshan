package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/omkar273/nashikdarshan/internal/types"
)

// LocationMixin adds a PostGIS geography(Point, 4326) location field
// with a GiST spatial index for efficient geospatial queries
type LocationMixin struct {
	mixin.Schema
}

// Fields of the LocationMixin
func (LocationMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Other("location", &types.GeoPoint{}).
			SchemaType(map[string]string{
				dialect.Postgres: "geography(Point,4326)",
			}).
			Comment("PostGIS geography point for geospatial queries"),
	}
}

// Indexes of the LocationMixin
func (LocationMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("location"),
	}
}
