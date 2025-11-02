package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	baseMixin "github.com/omkar273/nashikdarshan/ent/mixin"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type PlaceImage struct {
	ent.Schema
}

func (PlaceImage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		baseMixin.BaseMixin{},
		baseMixin.MetadataMixin{},
	}
}

func (PlaceImage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			DefaultFunc(func() string {
				return types.GenerateUUIDWithPrefix(types.UUID_PREFIX_PLACE_IMAGE)
			}).
			Immutable(),
		field.String("place_id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty(),
		field.String("url").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			NotEmpty(),
		field.String("alt").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			Optional(),
		field.Int("pos").
			SchemaType(map[string]string{
				"postgres": "int",
			}).
			Default(0),
	}
}

func (PlaceImage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("place", Place.Type).
			Ref("images").
			Field("place_id").
			Unique().
			Required(),
	}
}

func (PlaceImage) Indexes() []ent.Index {
	return []ent.Index{
		// Index on place_id for faster lookups
		index.Fields("place_id"),
		// Index on place_id and pos for ordered gallery queries
		index.Fields("place_id", "pos"),
	}
}
