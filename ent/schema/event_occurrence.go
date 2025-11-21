package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EventOccurrence holds the schema definition for the EventOccurrence entity.
type EventOccurrence struct {
	ent.Schema
}

// Fields of the EventOccurrence.
func (EventOccurrence) Fields() []ent.Field {
	return []ent.Field{
		// Identity
		field.String("id").
			Immutable().
			NotEmpty().
			Comment("Unique occurrence identifier with prefix occ_"),

		field.String("event_id").
			NotEmpty().
			Comment("FK to parent event"),

		// Recurrence Pattern
		field.Enum("recurrence_type").
			Values("NONE", "DAILY", "WEEKLY", "MONTHLY", "YEARLY").
			Default("NONE").
			Comment("How this occurrence repeats"),

		// Time Configuration
		field.Time("start_time").
			Comment("Time of day (only time component used)"),

		field.Time("end_time").
			Comment("End time of day (only time component used)"),

		field.Int("duration_minutes").
			Optional().
			Nillable().
			Comment("Auto-calculated duration"),

		// Day-specific fields (for recurrence logic)
		field.Int("day_of_week").
			Optional().
			Nillable().
			Min(0).
			Max(6).
			Comment("0=Sunday, 6=Saturday - for WEEKLY"),

		field.Int("day_of_month").
			Optional().
			Nillable().
			Min(1).
			Max(31).
			Comment("1-31 - for MONTHLY/YEARLY"),

		field.Int("month_of_year").
			Optional().
			Nillable().
			Min(1).
			Max(12).
			Comment("1-12 - for YEARLY only"),

		// Exception Dates (skip specific dates)
		field.JSON("exception_dates", []string{}).
			Optional().
			Comment("ISO dates to skip: ['2025-12-25', '2025-01-26']"),

		// Metadata
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("Occurrence-specific data"),

		// Lifecycle
		field.Enum("status").
			Values("active", "paused", "archived", "deleted").
			Default("active").
			Comment("Occurrence status"),

		// Audit
		field.String("created_by").
			NotEmpty().
			Comment("User ID who created"),

		field.String("updated_by").
			NotEmpty().
			Comment("User ID who last updated"),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Creation timestamp"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp"),
	}
}

// Edges of the EventOccurrence.
func (EventOccurrence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).
			Ref("occurrences").
			Field("event_id").
			Unique().
			Required(),
	}
}

// Indexes of the EventOccurrence.
func (EventOccurrence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "status"),
		index.Fields("recurrence_type", "status"),
		index.Fields("day_of_week"),
		index.Fields("day_of_month"),
		index.Fields("created_at"),
	}
}
