package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	baseMixin "github.com/omkar273/nashikdarshan/ent/mixin"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type User struct {
	ent.Schema
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		baseMixin.BaseMixin{},
		baseMixin.MetadataMixin{},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			DefaultFunc(func() string {
				return types.GenerateUUIDWithPrefix(types.UUID_PREFIX_USER)
			}).
			Immutable(),
		field.String("name").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty(),
		field.String("email").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty(),
		field.String("phone").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			Nillable(),

		field.String("role").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			Default(string(types.UserRoleUser)).
			NotEmpty(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		// Email must be unique
		index.Fields("email").
			Unique(),
		// Phone should be unique only when not NULL
		// Note: Ent doesn't support partial indexes directly, so we remove the unique constraint on phone
		// and handle uniqueness at the application level for non-null phones
		index.Fields("phone"),
	}
}
