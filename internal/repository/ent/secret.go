package ent

import (
	"context"
	"time"

	"github.com/omkar273/nashikdarshan/ent"
	"github.com/omkar273/nashikdarshan/ent/secret"
	domainSecret "github.com/omkar273/nashikdarshan/internal/domain/secret"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/logger"
	"github.com/omkar273/nashikdarshan/internal/postgres"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type secretRepository struct {
	client    postgres.IClient
	log       logger.Logger
	queryOpts SecretQueryOptions
}

func NewSecretRepository(client postgres.IClient, logger *logger.Logger) domainSecret.Repository {
	return &secretRepository{
		client:    client,
		log:       *logger,
		queryOpts: SecretQueryOptions{},
	}
}

func (r *secretRepository) Create(ctx context.Context, s *domainSecret.Secret) error {
	client := r.client.Querier(ctx)

	r.log.Debugw("creating secret",
		"secret_id", s.ID,
		"name", s.Name,
		"type", s.Type,
		"provider", s.Provider,
	)

	create := client.Secret.Create().
		SetID(s.ID).
		SetName(s.Name).
		SetType(s.Type).
		SetProvider(s.Provider).
		SetValue(s.Value).
		SetPrefix(s.Prefix).
		SetPermissions(s.Permissions).
		SetStatus(string(s.Status)).
		SetCreatedAt(s.CreatedAt).
		SetUpdatedAt(s.UpdatedAt).
		SetCreatedBy(s.CreatedBy).
		SetUpdatedBy(s.UpdatedBy)

	if s.ExpiresAt != nil {
		create = create.SetExpiresAt(*s.ExpiresAt)
	}

	if s.LastUsedAt != nil {
		create = create.SetLastUsedAt(*s.LastUsedAt)
	}

	// Set metadata
	metadataMap := make(map[string]string)
	if s.Metadata != nil && len(s.Metadata.ToMap()) > 0 {
		metadataMap = s.Metadata.ToMap()
	}
	create = create.SetMetadata(metadataMap)

	_, err := create.Save(ctx)

	if err != nil {
		if ent.IsConstraintError(err) {
			return ierr.WithError(err).
				WithHint("Secret with this ID already exists").
				WithReportableDetails(map[string]any{
					"secret_id": s.ID,
					"type":      s.Type,
					"provider":  s.Provider,
				}).
				Mark(ierr.ErrAlreadyExists)
		}
		return ierr.WithError(err).
			WithHint("Failed to create secret").
			WithReportableDetails(map[string]any{
				"secret_id": s.ID,
				"name":      s.Name,
			}).
			Mark(ierr.ErrDatabase)
	}

	return nil
}

func (r *secretRepository) Get(ctx context.Context, id string) (*domainSecret.Secret, error) {
	client := r.client.Querier(ctx)

	r.log.Debugw("getting secret", "secret_id", id)

	entSecret, err := client.Secret.Query().
		Where(secret.ID(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ierr.WithError(err).
				WithHintf("Secret with ID %s was not found", id).
				WithReportableDetails(map[string]any{
					"secret_id": id,
				}).
				Mark(ierr.ErrNotFound)
		}
		return nil, ierr.WithError(err).
			WithHint("Failed to get secret").
			WithReportableDetails(map[string]any{
				"secret_id": id,
			}).
			Mark(ierr.ErrDatabase)
	}

	return domainSecret.FromEnt(entSecret), nil
}

func (r *secretRepository) GetByValue(ctx context.Context, hashedValue string) (*domainSecret.Secret, error) {
	client := r.client.Querier(ctx)

	r.log.Debugw("getting secret by value")

	entSecret, err := client.Secret.Query().
		Where(
			secret.Value(hashedValue),
			secret.Status(string(types.StatusPublished)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ierr.WithError(err).
				WithHint("Secret with this value was not found").
				Mark(ierr.ErrNotFound)
		}
		return nil, ierr.WithError(err).
			WithHint("Failed to get secret by value").
			Mark(ierr.ErrDatabase)
	}

	return domainSecret.FromEnt(entSecret), nil
}

func (r *secretRepository) GetByType(ctx context.Context, userID string, secretType types.SecretType) ([]*domainSecret.Secret, error) {
	client := r.client.Querier(ctx)

	r.log.Debugw("getting secrets by type",
		"user_id", userID,
		"type", secretType,
	)

	entSecrets, err := client.Secret.Query().
		Where(
			secret.CreatedBy(userID),
			secret.Type(secretType),
			secret.StatusNotIn(string(types.StatusArchived), string(types.StatusDeleted)),
		).
		Order(ent.Desc(secret.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, ierr.WithError(err).
			WithHint("Failed to get secrets by type").
			WithReportableDetails(map[string]any{
				"user_id": userID,
				"type":    secretType,
			}).
			Mark(ierr.ErrDatabase)
	}

	return domainSecret.FromEntList(entSecrets), nil
}

func (r *secretRepository) List(ctx context.Context, userID string) ([]*domainSecret.Secret, error) {
	client := r.client.Querier(ctx)

	r.log.Debugw("listing secrets", "user_id", userID)

	entSecrets, err := client.Secret.Query().
		Where(
			secret.CreatedBy(userID),
			secret.StatusNotIn(string(types.StatusArchived), string(types.StatusDeleted)),
		).
		Order(ent.Desc(secret.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, ierr.WithError(err).
			WithHint("Failed to list secrets").
			WithReportableDetails(map[string]any{
				"user_id": userID,
			}).
			Mark(ierr.ErrDatabase)
	}

	return domainSecret.FromEntList(entSecrets), nil
}

func (r *secretRepository) Update(ctx context.Context, s *domainSecret.Secret) error {
	client := r.client.Querier(ctx)

	r.log.Debugw("updating secret",
		"secret_id", s.ID,
		"name", s.Name,
	)

	update := client.Secret.UpdateOneID(s.ID).
		SetName(s.Name).
		SetPermissions(s.Permissions).
		SetStatus(string(s.Status)).
		SetUpdatedAt(time.Now().UTC()).
		SetUpdatedBy(types.GetUserID(ctx))

	// expires_at is immutable, cannot be updated
	// last_used_at is updated separately in ValidateAPIKey
	if s.LastUsedAt != nil {
		update = update.SetLastUsedAt(*s.LastUsedAt)
	}

	if s.Metadata != nil {
		update = update.SetMetadata(s.Metadata.ToMap())
	}

	_, err := update.Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return ierr.WithError(err).
				WithHintf("Secret with ID %s was not found", s.ID).
				WithReportableDetails(map[string]any{
					"secret_id": s.ID,
				}).
				Mark(ierr.ErrNotFound)
		}
		return ierr.WithError(err).
			WithHint("Failed to update secret").
			WithReportableDetails(map[string]any{
				"secret_id": s.ID,
			}).
			Mark(ierr.ErrDatabase)
	}

	return nil
}

func (r *secretRepository) Delete(ctx context.Context, id string) error {
	client := r.client.Querier(ctx)

	r.log.Debugw("deleting secret", "secret_id", id)

	_, err := client.Secret.UpdateOneID(id).
		SetStatus(string(types.StatusArchived)).
		SetUpdatedAt(time.Now().UTC()).
		SetUpdatedBy(types.GetUserID(ctx)).
		Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return ierr.WithError(err).
				WithHintf("Secret with ID %s was not found", id).
				WithReportableDetails(map[string]any{
					"secret_id": id,
				}).
				Mark(ierr.ErrNotFound)
		}
		return ierr.WithError(err).
			WithHint("Failed to delete secret").
			WithReportableDetails(map[string]any{
				"secret_id": id,
			}).
			Mark(ierr.ErrDatabase)
	}

	return nil
}

type SecretQuery = *ent.SecretQuery

type SecretQueryOptions struct{}

func (o SecretQueryOptions) ApplyTypeFilter(query SecretQuery, secretType types.SecretType) SecretQuery {
	return query.Where(secret.Type(secretType))
}

func (o SecretQueryOptions) ApplyProviderFilter(query SecretQuery, provider types.SecretProvider) SecretQuery {
	return query.Where(secret.Provider(provider))
}

func (o SecretQueryOptions) ApplyStatusFilter(query SecretQuery, status string) SecretQuery {
	if status != "" {
		return query.Where(secret.Status(status))
	}
	return query
}

func (o SecretQueryOptions) ApplySortFilter(query SecretQuery, field string, order string) SecretQuery {
	fieldName := o.GetFieldName(field)
	if fieldName == "" {
		return query
	}

	if order == "desc" {
		return query.Order(ent.Desc(fieldName))
	}
	return query.Order(ent.Asc(fieldName))
}

func (o SecretQueryOptions) GetFieldName(field string) string {
	switch field {
	case "id":
		return secret.FieldID
	case "name":
		return secret.FieldName
	case "type":
		return secret.FieldType
	case "provider":
		return secret.FieldProvider
	case "prefix":
		return secret.FieldPrefix
	case "expires_at":
		return secret.FieldExpiresAt
	case "last_used_at":
		return secret.FieldLastUsedAt
	case "created_at":
		return secret.FieldCreatedAt
	case "updated_at":
		return secret.FieldUpdatedAt
	default:
		return ""
	}
}
