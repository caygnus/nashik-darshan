package secret

import (
	"time"

	"github.com/omkar273/nashikdarshan/ent"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/samber/lo"
)

// Secret represents a credential in the system
type Secret struct {
	ID          string
	Name        string
	Type        types.SecretType
	Provider    types.SecretProvider
	Value       string
	Prefix      string
	Permissions []string
	ExpiresAt   *time.Time
	LastUsedAt  *time.Time
	Metadata    *types.Metadata
	types.BaseModel
}

// FromEnt converts an ent.Secret to a domain Secret
func FromEnt(e *ent.Secret) *Secret {
	if e == nil {
		return nil
	}

	metadata := types.NewMetadataFromMap(e.Metadata)

	return &Secret{
		ID:          e.ID,
		Name:        e.Name,
		Type:        e.Type,
		Provider:    e.Provider,
		Value:       e.Value,
		Prefix:      e.Prefix,
		Permissions: e.Permissions,
		ExpiresAt:   e.ExpiresAt,
		LastUsedAt:  e.LastUsedAt,
		Metadata:    metadata,
		BaseModel: types.BaseModel{
			Status:    types.Status(e.Status),
			CreatedAt: e.CreatedAt,
			UpdatedAt: e.UpdatedAt,
			CreatedBy: e.CreatedBy,
			UpdatedBy: e.UpdatedBy,
		},
	}
}

// FromEntList converts a list of ent.Secret to domain Secrets
func FromEntList(list []*ent.Secret) []*Secret {
	if list == nil {
		return nil
	}

	return lo.Map(list, func(e *ent.Secret, _ int) *Secret {
		return FromEnt(e)
	})
}
