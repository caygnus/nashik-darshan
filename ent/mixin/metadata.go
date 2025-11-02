package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// MetadataMixin implements the ent.Mixin for sharing metadata fields with package schemas.
type MetadataMixin struct {
	mixin.Schema
}

// Fields of the MetadataMixin.
func (MetadataMixin) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("metadata", map[string]string{}).
			SchemaType(map[string]string{
				"postgres": "jsonb",
			}).
			Default(map[string]string{}).
			Optional(),
	}
}

// Hooks of the MetadataMixin.
func (MetadataMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		// Add hooks for updating updated_at and updated_by
	}
}
