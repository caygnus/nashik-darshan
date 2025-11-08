package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	baseMixin "github.com/omkar273/nashikdarshan/ent/mixin"
	"github.com/shopspring/decimal"
)

type Review struct {
	ent.Schema
}

func (Review) Mixin() []ent.Mixin {
	return []ent.Mixin{
		baseMixin.BaseMixin{},
		baseMixin.MetadataMixin{},
	}
}

func (Review) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty().
			Immutable(),

		// Generic entity reference
		field.String("entity_type").
			SchemaType(map[string]string{
				"postgres": "varchar(50)",
			}).
			NotEmpty().
			Comment("Type of entity being reviewed (place, hotel, event, etc.)"),

		field.String("entity_id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty().
			Comment("ID of the entity being reviewed"),

		// User who submitted the review
		field.String("user_id").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			NotEmpty(),

		// Review content
		field.Other("rating", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "decimal(2,1)",
			}).
			Comment("Rating from 1.0 to 5.0"),

		field.String("title").
			SchemaType(map[string]string{
				"postgres": "varchar(255)",
			}).
			Optional().
			Comment("Optional review title"),

		field.Text("content").
			Optional().
			Comment("Review text content"),

		// Review metadata
		field.Strings("tags").
			SchemaType(map[string]string{
				"postgres": "text[]",
			}).
			Optional().
			Comment("Review tags (e.g., 'family-friendly', 'romantic', 'budget')"),

		field.JSON("images", []string{}).
			SchemaType(map[string]string{
				"postgres": "jsonb",
			}).
			Optional().
			Comment("Array of image URLs attached to review"),

		// Engagement metrics
		field.Int("helpful_count").
			SchemaType(map[string]string{
				"postgres": "integer",
			}).
			Default(0).
			NonNegative().
			Comment("Number of users who found this review helpful"),

		field.Int("not_helpful_count").
			SchemaType(map[string]string{
				"postgres": "integer",
			}).
			Default(0).
			NonNegative().
			Comment("Number of users who found this review not helpful"),

		// Verification status
		field.Bool("is_verified").
			Default(false).
			Comment("Whether this review is from a verified visit/purchase"),

		field.Bool("is_featured").
			Default(false).
			Comment("Whether this review is featured/highlighted"),
	}
}

func (Review) Indexes() []ent.Index {
	return []ent.Index{
		// Primary lookup indexes
		index.Fields("entity_type", "entity_id"),
		index.Fields("entity_type", "entity_id", "status"),

		// User reviews
		index.Fields("user_id", "status"),

		// Rating queries
		index.Fields("entity_type", "entity_id", "rating"),

		// Featured reviews
		index.Fields("entity_type", "entity_id", "is_featured", "status"),

		// Helpful reviews
		index.Fields("entity_type", "entity_id", "helpful_count"),

		// Time-based queries
		index.Fields("entity_type", "entity_id", "created_at"),

		// Composite index for feed queries
		index.Fields("entity_type", "entity_id", "rating", "helpful_count", "created_at"),
	}
}
