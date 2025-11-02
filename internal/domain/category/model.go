package category

import (
	"github.com/omkar273/nashikdarshan/ent"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/samber/lo"
)

// Note: This file requires the ent code to be regenerated after adding the slug field to the schema.
// Run the ent generation command to update ent.Category with the Slug field.

type Category struct {
	ID          string          `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Slug        string          `json:"slug" db:"slug"`
	Description *string         `json:"description,omitempty" db:"description"`
	Metadata    *types.Metadata `json:"metadata,omitempty" db:"metadata"`
	types.BaseModel
}

func FromEnt(category *ent.Category) *Category {
	metadata := types.NewMetadataFromMap(category.Metadata)
	var description *string
	if category.Description != "" {
		description = lo.ToPtr(category.Description)
	}
	return &Category{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: description,
		Metadata:    metadata,
		BaseModel: types.BaseModel{
			Status:    types.Status(category.Status),
			CreatedAt: category.CreatedAt,
			UpdatedAt: category.UpdatedAt,
			CreatedBy: category.CreatedBy,
			UpdatedBy: category.UpdatedBy,
		},
	}
}

func FromEntList(categories []*ent.Category) []*Category {
	return lo.Map(categories, func(category *ent.Category, _ int) *Category {
		return FromEnt(category)
	})
}
