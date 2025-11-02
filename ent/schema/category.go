package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	baseMixin "github.com/omkar273/nashikdarshan/ent/mixin"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type Category struct {
	ent.Schema
}

func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		baseMixin.BaseMixin{},
		baseMixin.MetadataMixin{},
	}
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			DefaultFunc(func() string {
				return types.GenerateUUIDWithPrefix(types.UUID_PREFIX_CATEGORY)
			}).
			Immutable(),
		field.String("name").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty(),
		field.String("description").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			Optional(),
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{}
}
