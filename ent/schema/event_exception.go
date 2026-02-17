package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/omkar273/nashikdarshan/internal/types"
)

// EventException cancels a single occurrence.
type EventException struct {
	ent.Schema
}

func (EventException) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			DefaultFunc(func() string {
				return types.GenerateUUIDWithPrefix(types.UUID_PREFIX_EVENT_EXCEPTION)
			}).
			Immutable(),

		field.String("event_id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty(),

		field.Time("occurrence_start_at").
			Comment("Start of the occurrence being cancelled"),
	}
}

func (EventException) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).
			Ref("exceptions").
			Unique().
			Required().
			Field("event_id"),
	}
}

func (EventException) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "occurrence_start_at").
			Unique(),
	}
}
