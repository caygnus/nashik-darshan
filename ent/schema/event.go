package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	baseMixin "github.com/omkar273/nashikdarshan/ent/mixin"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type Event struct {
	ent.Schema
}

func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		baseMixin.BaseMixin{},
		baseMixin.MetadataMixin{},
	}
}

func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			DefaultFunc(func() string {
				return types.GenerateUUIDWithPrefix(types.UUID_PREFIX_EVENT)
			}).
			Immutable(),

		field.String("place_id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			Optional().
			Nillable().
			Comment("FK to place; NULL = city-level event"),

		field.String("title").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			NotEmpty(),

		field.String("subtitle").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			Optional(),

		field.Text("description").
			Optional(),

		field.Time("start_at").
			Comment("First occurrence start"),
		field.Time("end_at").
			Comment("First occurrence end; duration reused for expanded occurrences"),

		field.String("rrule").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			Optional().
			Nillable(),
		field.Time("rrule_until").
			Optional().
			Nillable(),
		field.Int("rrule_count").
			Optional().
			Nillable(),

		field.String("type").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			GoType(types.EventType("")).
			Comment("aarti, festival, cultural, city_event, other"),

		field.String("cover_image_url").
			SchemaType(map[string]string{
				"postgres": "text",
			}).
			Optional(),
	}
}

func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("place", Place.Type).
			Ref("events").
			Unique().
			Field("place_id"),
		edge.To("overrides", EventOverride.Type),
		edge.To("exceptions", EventException.Type),
	}
}

func (Event) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("place_id", "status"),
		index.Fields("type", "status"),
		index.Fields("start_at", "end_at"),
		index.Fields("status", "start_at"),
	}
}
