package types

import (
	"database/sql/driver"
	"encoding/json"

	ierr "github.com/omkar273/nashikdarshan/internal/errors"
)

// Metadata represents a JSONB field for storing key-value pairs
type Metadata map[string]string

// Scan implements the sql.Scanner interface for Metadata
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(Metadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ierr.NewError("failed to unmarshal JSONB value").
			WithHint("Please provide a valid JSON value").
			Mark(ierr.ErrValidation)
	}

	result := make(Metadata)
	err := json.Unmarshal(bytes, &result)
	*m = result
	return err
}

// Value implements the driver.Valuer interface for Metadata
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(make(Metadata))
	}
	return json.Marshal(m)
}

// FromMap converts a map[string]string to a Metadata pointer.
// Returns nil if the map is empty or nil.
func NewMetadataFromMap(m map[string]string) *Metadata {
	if len(m) == 0 {
		return nil
	}
	metadata := Metadata(m)
	return &metadata
}

// ToMap converts a Metadata pointer to a map[string]string.
// Returns an empty map if the Metadata pointer is nil.
func (m *Metadata) ToMap() map[string]string {
	if m == nil {
		return make(map[string]string)
	}
	return map[string]string(*m)
}
