package category

import (
	"github.com/omkar273/nashikdarshan/ent"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/samber/lo"
)

type Category struct {
	ID               string          `json:"id" db:"id"`
	Name             string          `json:"name" db:"name"`
	Subtitle         string          `json:"subtitle,omitempty" db:"subtitle"`
	ShortDescription string          `json:"short_description,omitempty" db:"short_description"`
	Slug             string          `json:"slug" db:"slug"`
	Description      string          `json:"description,omitempty" db:"description"`
	ImageURL         string          `json:"image_url" db:"image_url"`
	Icon             string          `json:"icon" db:"icon"`
	Tags             []string        `json:"tags,omitempty" db:"tags"`
	Metadata         *types.Metadata `json:"metadata,omitempty" db:"metadata"`
	types.BaseModel
}

func FromEnt(category *ent.Category) *Category {
	metadata := types.NewMetadataFromMap(category.Metadata)

	return &Category{
		ID:               category.ID,
		Name:             category.Name,
		Subtitle:         category.Subtitle,
		ShortDescription: category.ShortDescription,
		Slug:             category.Slug,
		Description:      category.Description,
		ImageURL:         category.ImageURL,
		Icon:             category.Icon,
		Tags:             category.Tags,
		Metadata:         metadata,
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
