package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/omkar273/nashikdarshan/internal/types"
)

// EventOverride moves a single occurrence to new start/end (patch one instance).
type EventOverride struct {
	ent.Schema
}

func (EventOverride) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			DefaultFunc(func() string {
				return types.GenerateUUIDWithPrefix(types.UUID_PREFIX_EVENT_OVERRIDE)
			}).
			Immutable(),

		field.String("event_id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty(),

		field.Time("original_start_at").
			Comment("Start of the occurrence being moved"),
		field.Time("new_start_at"),
		field.Time("new_end_at"),
	}
}

func (EventOverride) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).
			Ref("overrides").
			Unique().
			Required().
			Field("event_id"),
	}
}

func (EventOverride) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "original_start_at").
			Unique(),
	}
}
