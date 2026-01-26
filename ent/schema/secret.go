package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	baseMixin "github.com/omkar273/nashikdarshan/ent/mixin"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type Secret struct {
	ent.Schema
}

func (Secret) Mixin() []ent.Mixin {
	return []ent.Mixin{
		baseMixin.BaseMixin{},
		baseMixin.MetadataMixin{},
	}
}

func (Secret) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			DefaultFunc(func() string {
				return types.GenerateUUIDWithPrefix(types.UUID_PREFIX_SECRET)
			}).
			Immutable(),

		field.String("name").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty(),

		field.String("type").
			SchemaType(map[string]string{
				"postgres": "varchar(50)",
			}).
			GoType(types.SecretType("")).
			Default(string(types.SecretTypePrivateKey)).
			NotEmpty().
			Immutable().
			Comment("Type of secret: private_key or publishable_key"),

		field.String("provider").
			SchemaType(map[string]string{
				"postgres": "varchar(50)",
			}).
			GoType(types.SecretProvider("")).
			Default(string(types.SecretProviderNashikDarshan)).
			NotEmpty().
			Immutable().
			Comment("Provider of the secret"),

		field.String("value").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			NotEmpty().
			Immutable().
			Comment("Hashed API key value"),

		field.String("prefix").
			SchemaType(map[string]string{
				"postgres": "varchar(8)",
			}).
			NotEmpty().
			Immutable().
			Comment("First 8 characters of the key for display"),

		field.Strings("permissions").
			SchemaType(map[string]string{
				"postgres": "text[]",
			}).
			Optional().
			Comment("Array of permissions (read, write, etc.)"),

		field.Time("expires_at").
			Optional().
			Nillable().
			Immutable().
			Comment("Optional expiration date"),

		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("Last time the API key was used"),
	}
}

func (Secret) Indexes() []ent.Index {
	return []ent.Index{
		// Index on value for fast lookup during validation
		index.Fields("value"),
		// Index on type and provider for filtering
		index.Fields("type", "provider"),
		// Index on status for active keys
		index.Fields("status"),
		// Index on created_by for user-scoped queries
		index.Fields("created_by"),
		// Composite index for listing user's keys by type
		index.Fields("created_by", "type", "status"),
	}
}
