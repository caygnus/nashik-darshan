package ent

import (
	"context"
	"time"

	"github.com/omkar273/nashikdarshan/ent"
	"github.com/omkar273/nashikdarshan/ent/category"
	domainCategory "github.com/omkar273/nashikdarshan/internal/domain/category"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/logger"
	"github.com/omkar273/nashikdarshan/internal/postgres"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type CategoryRepository struct {
	client    postgres.IClient
	log       logger.Logger
	queryOpts CategoryQueryOptions
}

func NewCategoryRepository(client postgres.IClient, logger *logger.Logger) domainCategory.Repository {
	return &CategoryRepository{
		client:    client,
		log:       *logger,
		queryOpts: CategoryQueryOptions{},
	}
}

func (r *CategoryRepository) Create(ctx context.Context, categoryData *domainCategory.Category) error {
	client := r.client.Querier(ctx)

	r.log.Debugw("creating category",
		"category_id", categoryData.ID,
		"name", categoryData.Name,
		"slug", categoryData.Slug,
	)

	create := client.Category.Create().
		SetID(categoryData.ID).
		SetName(categoryData.Name).
		SetSlug(categoryData.Slug).
		SetStatus(string(categoryData.Status)).
		SetCreatedAt(categoryData.CreatedAt).
		SetUpdatedAt(categoryData.UpdatedAt).
		SetCreatedBy(categoryData.CreatedBy).
		SetUpdatedBy(categoryData.UpdatedBy)

	if categoryData.Description != nil {
		create = create.SetDescription(*categoryData.Description)
	}

	if categoryData.Metadata != nil && len(categoryData.Metadata.ToMap()) > 0 {
		create = create.SetMetadata(categoryData.Metadata.ToMap())
	}

	_, err := create.Save(ctx)

	if err != nil {
		if ent.IsConstraintError(err) {
			return ierr.WithError(err).
				WithHint("Category with this slug already exists").
				WithReportableDetails(map[string]any{
					"category_id": categoryData.ID,
					"slug":        categoryData.Slug,
				}).
				Mark(ierr.ErrAlreadyExists)
		}
		return ierr.WithError(err).
			WithHint("Failed to create category").
			WithReportableDetails(map[string]any{
				"category_id": categoryData.ID,
				"name":        categoryData.Name,
			}).
			Mark(ierr.ErrDatabase)
	}

	return nil
}

func (r *CategoryRepository) Get(ctx context.Context, id string) (*domainCategory.Category, error) {
	client := r.client.Querier(ctx)

	r.log.Debugw("getting category", "category_id", id)

	entCategory, err := client.Category.Query().
		Where(category.ID(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ierr.WithError(err).
				WithHintf("Category with ID %s was not found", id).
				WithReportableDetails(map[string]any{
					"category_id": id,
				}).
				Mark(ierr.ErrNotFound)
		}
		return nil, ierr.WithError(err).
			WithHint("Failed to get category").
			WithReportableDetails(map[string]any{
				"category_id": id,
			}).
			Mark(ierr.ErrDatabase)
	}

	return domainCategory.FromEnt(entCategory), nil
}

func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*domainCategory.Category, error) {
	client := r.client.Querier(ctx)

	r.log.Debugw("getting category by slug", "slug", slug)

	entCategory, err := client.Category.Query().
		Where(category.Slug(slug)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ierr.WithError(err).
				WithHintf("Category with slug %s was not found", slug).
				WithReportableDetails(map[string]any{
					"slug": slug,
				}).
				Mark(ierr.ErrNotFound)
		}
		return nil, ierr.WithError(err).
			WithHint("Failed to get category by slug").
			WithReportableDetails(map[string]any{
				"slug": slug,
			}).
			Mark(ierr.ErrDatabase)
	}

	return domainCategory.FromEnt(entCategory), nil
}

func (r *CategoryRepository) List(ctx context.Context, filter *types.CategoryFilter) ([]*domainCategory.Category, error) {
	client := r.client.Querier(ctx)

	r.log.Debugw("listing categories",
		"limit", filter.GetLimit(),
		"offset", filter.GetOffset(),
	)

	query := client.Category.Query()
	query = r.queryOpts.ApplyBaseFilters(ctx, query, filter)
	query = r.queryOpts.ApplyEntityQueryOptions(ctx, filter, query)

	categories, err := query.All(ctx)
	if err != nil {
		return nil, ierr.WithError(err).
			WithHint("Failed to list categories").
			Mark(ierr.ErrDatabase)
	}

	return domainCategory.FromEntList(categories), nil
}

func (r *CategoryRepository) ListAll(ctx context.Context, filter *types.CategoryFilter) ([]*domainCategory.Category, error) {
	if filter == nil {
		filter = types.NewNoLimitCategoryFilter()
	}

	if filter.QueryFilter == nil {
		filter.QueryFilter = types.NewNoLimitQueryFilter()
	}

	categories, err := r.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) Count(ctx context.Context, filter *types.CategoryFilter) (int, error) {
	client := r.client.Querier(ctx)

	r.log.Debugw("counting categories")

	query := client.Category.Query()
	query = r.queryOpts.ApplyBaseFilters(ctx, query, filter)
	query = r.queryOpts.ApplyEntityQueryOptions(ctx, filter, query)

	count, err := query.Count(ctx)
	if err != nil {
		return 0, ierr.WithError(err).
			WithHint("Failed to count categories").
			Mark(ierr.ErrDatabase)
	}

	return count, nil
}

func (r *CategoryRepository) Update(ctx context.Context, categoryData *domainCategory.Category) error {
	client := r.client.Querier(ctx)

	r.log.Debugw("updating category",
		"category_id", categoryData.ID,
		"name", categoryData.Name,
	)

	update := client.Category.UpdateOneID(categoryData.ID).
		SetName(categoryData.Name).
		SetSlug(categoryData.Slug).
		SetStatus(string(categoryData.Status)).
		SetUpdatedAt(time.Now().UTC()).
		SetUpdatedBy(types.GetUserID(ctx))

	if categoryData.Description != nil {
		update = update.SetDescription(*categoryData.Description)
	} else {
		update = update.ClearDescription()
	}

	if categoryData.Metadata != nil {
		update = update.SetMetadata(categoryData.Metadata.ToMap())
	}

	_, err := update.Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return ierr.WithError(err).
				WithHintf("Category with ID %s was not found", categoryData.ID).
				WithReportableDetails(map[string]any{
					"category_id": categoryData.ID,
				}).
				Mark(ierr.ErrNotFound)
		}
		if ent.IsConstraintError(err) {
			return ierr.WithError(err).
				WithHint("Category with this slug already exists").
				WithReportableDetails(map[string]any{
					"category_id": categoryData.ID,
					"slug":        categoryData.Slug,
				}).
				Mark(ierr.ErrAlreadyExists)
		}
		return ierr.WithError(err).
			WithHint("Failed to update category").
			WithReportableDetails(map[string]any{
				"category_id": categoryData.ID,
			}).
			Mark(ierr.ErrDatabase)
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, categoryData *domainCategory.Category) error {
	client := r.client.Querier(ctx)

	r.log.Debugw("deleting category",
		"category_id", categoryData.ID,
	)

	_, err := client.Category.UpdateOneID(categoryData.ID).
		SetStatus(string(types.StatusDeleted)).
		SetUpdatedAt(time.Now().UTC()).
		SetUpdatedBy(types.GetUserID(ctx)).
		Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return ierr.WithError(err).
				WithHintf("Category with ID %s was not found", categoryData.ID).
				WithReportableDetails(map[string]any{
					"category_id": categoryData.ID,
				}).
				Mark(ierr.ErrNotFound)
		}
		return ierr.WithError(err).
			WithHint("Failed to delete category").
			WithReportableDetails(map[string]any{
				"category_id": categoryData.ID,
			}).
			Mark(ierr.ErrDatabase)
	}

	return nil
}

// CategoryQuery type alias for better readability
type CategoryQuery = *ent.CategoryQuery

// CategoryQueryOptions implements query options for category queries
type CategoryQueryOptions struct {
	QueryOptionsHelper
}

// Ensure CategoryQueryOptions implements EntityQueryOptions interface
var _ EntityQueryOptions[CategoryQuery, *types.CategoryFilter] = (*CategoryQueryOptions)(nil)

func (o CategoryQueryOptions) ApplyStatusFilter(query CategoryQuery, status string) CategoryQuery {
	if status == "" {
		return query.Where(category.StatusNotIn(string(types.StatusDeleted)))
	}
	return query.Where(category.Status(status))
}

func (o CategoryQueryOptions) ApplySortFilter(query CategoryQuery, field string, order string) CategoryQuery {
	field, order = o.ValidateSort(field, order)
	fieldName := o.GetFieldName(field)
	if order == types.OrderDesc {
		return query.Order(ent.Desc(fieldName))
	}
	return query.Order(ent.Asc(fieldName))
}

func (o CategoryQueryOptions) ApplyPaginationFilter(query CategoryQuery, limit int, offset int) CategoryQuery {
	limit, offset = o.ValidatePagination(limit, offset)
	return query.Offset(offset).Limit(limit)
}

func (o CategoryQueryOptions) GetFieldName(field string) string {
	switch field {
	case "created_at":
		return category.FieldCreatedAt
	case "updated_at":
		return category.FieldUpdatedAt
	case "name":
		return category.FieldName
	case "slug":
		return category.FieldSlug
	default:
		return field
	}
}

func (o CategoryQueryOptions) ApplyBaseFilters(
	_ context.Context,
	query CategoryQuery,
	filter *types.CategoryFilter,
) CategoryQuery {
	if filter == nil {
		return query.Where(category.StatusNotIn(string(types.StatusDeleted)))
	}

	// Apply status filter
	query = o.ApplyStatusFilter(query, filter.GetStatus())

	// Apply pagination
	if !filter.IsUnlimited() {
		query = o.ApplyPaginationFilter(query, filter.GetLimit(), filter.GetOffset())
	}

	// Apply sorting
	query = o.ApplySortFilter(query, filter.GetSort(), filter.GetOrder())

	return query
}

func (o CategoryQueryOptions) ApplyEntityQueryOptions(
	_ context.Context,
	f *types.CategoryFilter,
	query CategoryQuery,
) CategoryQuery {
	if f == nil {
		return query
	}

	// Apply slug filter if specified
	if len(f.Slug) > 0 {
		query = query.Where(category.SlugIn(f.Slug...))
	}

	// Apply name filter if specified
	if len(f.Name) > 0 {
		query = query.Where(category.NameIn(f.Name...))
	}

	// Apply time range filters if specified
	if f.TimeRangeFilter != nil {
		if f.StartTime != nil {
			query = query.Where(category.CreatedAtGTE(*f.StartTime))
		}
		if f.EndTime != nil {
			query = query.Where(category.CreatedAtLTE(*f.EndTime))
		}
	}

	return query
}
