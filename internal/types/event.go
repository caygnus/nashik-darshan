package types

// EventType represents the category of event
type EventType string

const (
	EventTypeAarti          EventType = "AARTI"
	EventTypeFestival       EventType = "FESTIVAL"
	EventTypeCultural       EventType = "CULTURAL"
	EventTypeWorkshop       EventType = "WORKSHOP"
	EventTypeSpecialDarshan EventType = "SPECIAL_DARSHAN"
	EventTypeOther          EventType = "OTHER"
)

// RecurrenceType represents how an event occurrence repeats
type RecurrenceType string

const (
	RecurrenceNone    RecurrenceType = "NONE"    // One-time event
	RecurrenceDaily   RecurrenceType = "DAILY"   // Every day
	RecurrenceWeekly  RecurrenceType = "WEEKLY"  // Specific day each week
	RecurrenceMonthly RecurrenceType = "MONTHLY" // Specific date each month
	RecurrenceYearly  RecurrenceType = "YEARLY"  // Specific date each year
)

// OccurrenceStatus represents the lifecycle status of an occurrence
type OccurrenceStatus string

const (
	OccurrenceActive   OccurrenceStatus = "active"
	OccurrencePaused   OccurrenceStatus = "paused"
	OccurrenceArchived OccurrenceStatus = "archived"
	OccurrenceDeleted  OccurrenceStatus = "deleted"
)

// EventFilter for querying events
type EventFilter struct {
	*QueryFilter
	Type     *EventType `form:"type"`
	PlaceID  *string    `form:"place_id"`
	FromDate *string    `form:"from_date"` // ISO date YYYY-MM-DD
	ToDate   *string    `form:"to_date"`   // ISO date YYYY-MM-DD
	Tags     []string   `form:"tags"`
}

// NewEventFilter creates a new EventFilter with defaults
func NewEventFilter() *EventFilter {
	return &EventFilter{
		QueryFilter: NewDefaultQueryFilter(),
	}
}

// Validate validates the event filter
func (f *EventFilter) Validate() error {
	if f.QueryFilter != nil {
		return f.QueryFilter.Validate()
	}
	return nil
}

// UUID_PREFIX_EVENT is the prefix for event IDs
const UUID_PREFIX_EVENT = "evt"

// UUID_PREFIX_OCCURRENCE is the prefix for occurrence IDs
const UUID_PREFIX_OCCURRENCE = "occ"
