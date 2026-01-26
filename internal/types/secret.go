package types

import (
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/samber/lo"
)

// SecretType represents the type of secret (API key type)
type SecretType string

// Secret types
const (
	SecretTypePrivateKey     SecretType = "private_key"
	SecretTypePublishableKey SecretType = "publishable_key"
)

// Validate validates the SecretType
func (t SecretType) Validate() error {
	allowedSecretTypes := []SecretType{SecretTypePrivateKey, SecretTypePublishableKey}
	if !lo.Contains(allowedSecretTypes, t) {
		return ierr.NewError("invalid secret type").
			WithHint("Invalid secret type").
			Mark(ierr.ErrValidation)
	}
	return nil
}

// String returns the string representation of SecretType
func (t SecretType) String() string {
	return string(t)
}

// SecretProvider represents the provider of the secret
type SecretProvider string

// Provider types
const (
	SecretProviderNashikDarshan SecretProvider = "nashikdarshan"
)

// Validate validates the SecretProvider
func (p SecretProvider) Validate() error {
	allowedSecretProviders := []SecretProvider{
		SecretProviderNashikDarshan,
	}
	if !lo.Contains(allowedSecretProviders, p) {
		return ierr.NewError("invalid secret provider").
			WithHint("Invalid secret provider").
			Mark(ierr.ErrValidation)
	}
	return nil
}

// String returns the string representation of SecretProvider
func (p SecretProvider) String() string {
	return string(p)
}

// SecretFilter defines the filter criteria for secrets
type SecretFilter struct {
	*QueryFilter
	*TimeRangeFilter

	Type     *SecretType     `json:"type,omitempty" form:"type"`
	Provider *SecretProvider `json:"provider,omitempty" form:"provider"`
	Prefix   *string         `json:"prefix,omitempty" form:"prefix"`
}

// NewSecretFilter creates a new SecretFilter with default values
func NewSecretFilter() *SecretFilter {
	return &SecretFilter{
		QueryFilter: NewDefaultQueryFilter(),
	}
}

// NewNoLimitSecretFilter creates a new SecretFilter with no pagination limits
func NewNoLimitSecretFilter() *SecretFilter {
	return &SecretFilter{
		QueryFilter: NewNoLimitQueryFilter(),
	}
}

// Validate validates the SecretFilter
func (f *SecretFilter) Validate() error {
	if f == nil {
		return nil
	}

	if f.QueryFilter != nil {
		if err := f.QueryFilter.Validate(); err != nil {
			return err
		}
	}

	if f.TimeRangeFilter != nil {
		if err := f.TimeRangeFilter.Validate(); err != nil {
			return err
		}
	}

	if !f.GetExpand().IsEmpty() {
		if err := f.GetExpand().Validate(SecretExpandConfig); err != nil {
			return err
		}
	}

	if f.Type != nil {
		if err := f.Type.Validate(); err != nil {
			return err
		}
	}

	if f.Provider != nil {
		if err := f.Provider.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// GetLimit implements BaseFilter interface
func (f *SecretFilter) GetLimit() int {
	if f == nil || f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetLimit()
	}
	return f.QueryFilter.GetLimit()
}

// GetOffset implements BaseFilter interface
func (f *SecretFilter) GetOffset() int {
	if f == nil || f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetOffset()
	}
	return f.QueryFilter.GetOffset()
}

// GetSort implements BaseFilter interface
func (f *SecretFilter) GetSort() string {
	if f == nil || f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetSort()
	}
	return f.QueryFilter.GetSort()
}

// GetStatus implements BaseFilter interface
func (f *SecretFilter) GetStatus() string {
	if f == nil || f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetStatus()
	}
	return f.QueryFilter.GetStatus()
}

// GetOrder implements BaseFilter interface
func (f *SecretFilter) GetOrder() string {
	if f == nil || f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetOrder()
	}
	return f.QueryFilter.GetOrder()
}

// GetExpand implements BaseFilter interface
func (f *SecretFilter) GetExpand() Expand {
	if f == nil || f.QueryFilter == nil {
		return NewDefaultQueryFilter().GetExpand()
	}
	return f.QueryFilter.GetExpand()
}

// IsUnlimited implements BaseFilter interface
func (f *SecretFilter) IsUnlimited() bool {
	if f == nil || f.QueryFilter == nil {
		return NewDefaultQueryFilter().IsUnlimited()
	}
	return f.QueryFilter.IsUnlimited()
}
