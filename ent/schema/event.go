package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/shopspring/decimal"
)

// Event holds the schema definition for the Event entity.
type Event struct {
	ent.Schema
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		// Identity
		field.String("id").
			Immutable().
			NotEmpty().
			Comment("Unique event identifier with prefix evt_"),

		field.String("slug").
			Unique().
			NotEmpty().
			Immutable().
			Comment("URL-friendly unique identifier"),

		// Core Info
		field.Enum("type").
			Values("AARTI", "FESTIVAL", "CULTURAL", "WORKSHOP", "SPECIAL_DARSHAN", "OTHER").
			Comment("Event category for filtering and display"),

		field.String("title").
			NotEmpty().
			MaxLen(255).
			Comment("Event name"),

		field.String("subtitle").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("Brief tagline"),

		field.Text("description").
			Optional().
			Nillable().
			Comment("Detailed description with markdown support"),

		// Association
		field.String("place_id").
			Optional().
			Nillable().
			Comment("FK to place - NULL means citywide event"),

		// Validity Window
		field.Time("start_date").
			Comment("Event becomes active from this date"),

		field.Time("end_date").
			Optional().
			Nillable().
			Comment("Event expires after this date (NULL = ongoing)"),

		// Media
		field.String("cover_image_url").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("Event banner/poster image"),

		field.JSON("images", []string{}).
			Optional().
			Comment("Additional event images"),

		// Metadata
		field.JSON("tags", []string{}).
			Optional().
			Comment("Searchable tags: morning, evening, spiritual, etc"),

		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("Flexible data: stream_url, booking_link, contact, fee, etc"),

		// Location (for citywide events without place_id)
		field.Other("latitude", &decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "decimal(10,8)",
			}).
			Optional().
			Nillable().
			Comment("Latitude for standalone events"),

		field.Other("longitude", &decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "decimal(10,8)",
			}).
			Optional().
			Nillable().
			Comment("Longitude for standalone events"),

		field.String("location_name").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Text location for citywide events"),

		// Stats (cached for performance)
		field.Int("view_count").
			Default(0).
			NonNegative().
			Comment("Total views"),

		field.Int("interested_count").
			Default(0).
			NonNegative().
			Comment("Users who marked interested"),

		// Lifecycle
		field.Enum("status").
			Values("draft", "published", "archived", "deleted").
			Default("draft").
			Comment("Event visibility status"),

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

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		// One event has many occurrence slots
		edge.To("occurrences", EventOccurrence.Type),
	}
}

// Indexes of the Event.
func (Event) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("place_id", "status"),
		index.Fields("type", "status"),
		index.Fields("start_date", "end_date"),
		index.Fields("status", "start_date"),
		index.Fields("created_at"),
	}
}
