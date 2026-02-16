package category

import (
	"github.com/omkar273/nashikdarshan/ent"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/samber/lo"
)

type Category struct {
	ID          string          `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Slug        string          `json:"slug" db:"slug"`
	Description string          `json:"description,omitempty" db:"description"`
	ImageURL    string          `json:"image_url" db:"image_url"`
	Metadata    *types.Metadata `json:"metadata,omitempty" db:"metadata"`
	types.BaseModel
}

func FromEnt(category *ent.Category) *Category {
	metadata := types.NewMetadataFromMap(category.Metadata)

	return &Category{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ImageURL:    category.ImageURL,
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
