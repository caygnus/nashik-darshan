package dto

import (
	"context"
	"time"

	eventdomain "github.com/omkar273/nashikdarshan/internal/domain/event"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/omkar273/nashikdarshan/internal/validator"
	"github.com/shopspring/decimal"
)

// ========== Event DTOs ==========

// CreateEventRequest represents a request to create an event
type CreateEventRequest struct {
	Slug          string                 `json:"slug" binding:"required,min=3,max=100"`
	Type          string                 `json:"type" binding:"required"`
	Title         string                 `json:"title" binding:"required,min=2,max=255"`
	Subtitle      *string                `json:"subtitle,omitempty" binding:"omitempty,max=500"`
	Description   *string                `json:"description,omitempty" binding:"omitempty,max=10000"`
	PlaceID       *string                `json:"place_id,omitempty"`
	StartDate     string                 `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate       *string                `json:"end_date,omitempty"`            // YYYY-MM-DD
	CoverImageURL *string                `json:"cover_image_url,omitempty" binding:"omitempty,url,max=500"`
	Images        []string               `json:"images,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Latitude      *decimal.Decimal       `json:"latitude,omitempty"`
	Longitude     *decimal.Decimal       `json:"longitude,omitempty"`
	LocationName  *string                `json:"location_name,omitempty" binding:"omitempty,max=255"`
	Status        *string                `json:"status,omitempty"` // draft, published, archived
}

// Validate validates the CreateEventRequest
func (req *CreateEventRequest) Validate() error {
	if err := validator.ValidateRequest(req); err != nil {
		return err
	}

	// Validate slug format
	if err := validator.ValidateSlugFormat(req.Slug); err != nil {
		return err
	}

	// Validate event type
	if err := validator.ValidateEventType(req.Type); err != nil {
		return err
	}

	// Parse and validate dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return ierr.NewError("Invalid start_date format, expected YYYY-MM-DD").Mark(ierr.ErrValidation)
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return ierr.NewError("Invalid end_date format, expected YYYY-MM-DD").Mark(ierr.ErrValidation)
		}
		endDate = &parsed
	}

	if err := validator.ValidateEventDates(startDate, endDate); err != nil {
		return err
	}

	// Validate location: must have either place_id or (latitude+longitude+location_name)
	hasPlaceID := req.PlaceID != nil && *req.PlaceID != ""
	hasCoordinates := req.Latitude != nil && req.Longitude != nil

	if !hasPlaceID && !hasCoordinates {
		return ierr.NewError("Event must have either place_id or coordinates (latitude+longitude)").
			Mark(ierr.ErrValidation)
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := []string{"draft", "published", "archived"}
		valid := false
		for _, s := range validStatuses {
			if *req.Status == s {
				valid = true
				break
			}
		}
		if !valid {
			return ierr.NewError("Invalid status. Must be one of: draft, published, archived").
				Mark(ierr.ErrValidation)
		}
	}

	return nil
}

// ToEvent converts CreateEventRequest to domain Event
func (req *CreateEventRequest) ToEvent(ctx context.Context) (*eventdomain.Event, error) {
	baseModel := types.GetDefaultBaseModel(ctx)

	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	var endDate *time.Time
	if req.EndDate != nil {
		parsed, _ := time.Parse("2006-01-02", *req.EndDate)
		endDate = &parsed
	}

	status := types.StatusDraft
	if req.Status != nil {
		status = types.Status(*req.Status)
	}

	return &eventdomain.Event{
		ID:            types.GenerateUUIDWithPrefix(types.UUID_PREFIX_EVENT),
		Slug:          req.Slug,
		Type:          types.EventType(req.Type),
		Title:         req.Title,
		Subtitle:      req.Subtitle,
		Description:   req.Description,
		PlaceID:       req.PlaceID,
		StartDate:     startDate,
		EndDate:       endDate,
		CoverImageURL: req.CoverImageURL,
		Images:        req.Images,
		Tags:          req.Tags,
		Metadata:      req.Metadata,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		LocationName:  req.LocationName,
		Status:        status,
		BaseModel:     baseModel,
	}, nil
}

// UpdateEventRequest represents a request to update an event
type UpdateEventRequest struct {
	Type          *string                `json:"type,omitempty"`
	Title         *string                `json:"title,omitempty" binding:"omitempty,min=2,max=255"`
	Subtitle      *string                `json:"subtitle,omitempty" binding:"omitempty,max=500"`
	Description   *string                `json:"description,omitempty" binding:"omitempty,max=10000"`
	PlaceID       *string                `json:"place_id,omitempty"`
	StartDate     *string                `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate       *string                `json:"end_date,omitempty"`   // YYYY-MM-DD
	CoverImageURL *string                `json:"cover_image_url,omitempty" binding:"omitempty,url,max=500"`
	Images        []string               `json:"images,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Latitude      *decimal.Decimal       `json:"latitude,omitempty"`
	Longitude     *decimal.Decimal       `json:"longitude,omitempty"`
	LocationName  *string                `json:"location_name,omitempty" binding:"omitempty,max=255"`
	Status        *string                `json:"status,omitempty"`
}

// Validate validates the UpdateEventRequest
func (req *UpdateEventRequest) Validate() error {
	if err := validator.ValidateRequest(req); err != nil {
		return err
	}

	// Validate event type if provided
	if req.Type != nil {
		if err := validator.ValidateEventType(*req.Type); err != nil {
			return err
		}
	}

	// Parse and validate dates if provided
	var startDate *time.Time
	if req.StartDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return ierr.NewError("Invalid start_date format, expected YYYY-MM-DD").Mark(ierr.ErrValidation)
		}
		startDate = &parsed
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return ierr.NewError("Invalid end_date format, expected YYYY-MM-DD").Mark(ierr.ErrValidation)
		}
		endDate = &parsed
	}

	// If both dates are provided, validate them
	if startDate != nil && endDate != nil {
		if err := validator.ValidateEventDates(*startDate, endDate); err != nil {
			return err
		}
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := []string{"draft", "published", "archived"}
		valid := false
		for _, s := range validStatuses {
			if *req.Status == s {
				valid = true
				break
			}
		}
		if !valid {
			return ierr.NewError("Invalid status. Must be one of: draft, published, archived").
				Mark(ierr.ErrValidation)
		}
	}

	return nil
}

// ApplyToEvent applies the update request to an existing event
func (req *UpdateEventRequest) ApplyToEvent(ctx context.Context, event *eventdomain.Event) error {
	if req.Type != nil {
		event.Type = types.EventType(*req.Type)
	}
	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Subtitle != nil {
		event.Subtitle = req.Subtitle
	}
	if req.Description != nil {
		event.Description = req.Description
	}
	if req.PlaceID != nil {
		event.PlaceID = req.PlaceID
	}
	if req.StartDate != nil {
		parsed, _ := time.Parse("2006-01-02", *req.StartDate)
		event.StartDate = parsed
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			event.EndDate = nil
		} else {
			parsed, _ := time.Parse("2006-01-02", *req.EndDate)
			event.EndDate = &parsed
		}
	}
	if req.CoverImageURL != nil {
		event.CoverImageURL = req.CoverImageURL
	}
	if req.Images != nil {
		event.Images = req.Images
	}
	if req.Tags != nil {
		event.Tags = req.Tags
	}
	if req.Metadata != nil {
		event.Metadata = req.Metadata
	}
	if req.Latitude != nil {
		event.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		event.Longitude = req.Longitude
	}
	if req.LocationName != nil {
		event.LocationName = req.LocationName
	}
	if req.Status != nil {
		event.Status = types.Status(*req.Status)
	}

	// Update audit fields
	event.UpdatedBy = types.GetUserID(ctx)
	event.UpdatedAt = time.Now().UTC()

	return nil
}

// EventResponse represents an event in the response
type EventResponse struct {
	*eventdomain.Event
}

// ListEventsResponse represents a paginated list of events
type ListEventsResponse = types.ListResponse[*EventResponse]

// NewEventResponse creates an EventResponse from domain Event
func NewEventResponse(e *eventdomain.Event) *EventResponse {
	return &EventResponse{
		Event: e,
	}
}

// NewListEventsResponse creates a paginated list response
func NewListEventsResponse(events []*eventdomain.Event, total, limit, offset int) *ListEventsResponse {
	responses := make([]*EventResponse, len(events))
	for i, e := range events {
		responses[i] = NewEventResponse(e)
	}

	response := types.NewListResponse(responses, total, limit, offset)
	return &response
}

// ========== Occurrence DTOs ==========

// CreateOccurrenceRequest represents a request to create an occurrence
type CreateOccurrenceRequest struct {
	RecurrenceType string                 `json:"recurrence_type" binding:"required"` // NONE, DAILY, WEEKLY, MONTHLY, YEARLY
	StartTime      string                 `json:"start_time" binding:"required"`      // HH:MM format
	EndTime        string                 `json:"end_time" binding:"required"`        // HH:MM format
	DayOfWeek      *int                   `json:"day_of_week,omitempty"`              // 0-6 for WEEKLY
	DayOfMonth     *int                   `json:"day_of_month,omitempty"`             // 1-31 for MONTHLY/YEARLY
	MonthOfYear    *int                   `json:"month_of_year,omitempty"`            // 1-12 for YEARLY
	ExceptionDates []string               `json:"exception_dates,omitempty"`          // ["2025-12-25", ...]
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Status         *string                `json:"status,omitempty"` // active, paused, archived
}

// Validate validates the CreateOccurrenceRequest
func (req *CreateOccurrenceRequest) Validate() error {
	if err := validator.ValidateRequest(req); err != nil {
		return err
	}

	// Validate recurrence type
	if err := validator.ValidateRecurrenceType(req.RecurrenceType); err != nil {
		return err
	}

	// Parse and validate times
	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return ierr.NewError("Invalid start_time format, expected HH:MM (24-hour format)").Mark(ierr.ErrValidation)
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return ierr.NewError("Invalid end_time format, expected HH:MM (24-hour format)").Mark(ierr.ErrValidation)
	}

	if err := validator.ValidateOccurrenceTimes(startTime, endTime); err != nil {
		return err
	}

	// Validate recurrence rules
	recurrenceType := types.RecurrenceType(req.RecurrenceType)
	if err := validator.ValidateRecurrenceRules(recurrenceType, req.DayOfWeek, req.DayOfMonth, req.MonthOfYear); err != nil {
		return err
	}

	// Validate exception dates
	if len(req.ExceptionDates) > 0 {
		if err := validator.ValidateExceptionDates(req.ExceptionDates); err != nil {
			return err
		}
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := []string{"active", "paused", "archived"}
		valid := false
		for _, s := range validStatuses {
			if *req.Status == s {
				valid = true
				break
			}
		}
		if !valid {
			return ierr.NewError("Invalid status. Must be one of: active, paused, archived").
				Mark(ierr.ErrValidation)
		}
	}

	return nil
}

// ToOccurrence converts CreateOccurrenceRequest to domain EventOccurrence
func (req *CreateOccurrenceRequest) ToOccurrence(ctx context.Context, eventID string) (*eventdomain.EventOccurrence, error) {
	baseModel := types.GetDefaultBaseModel(ctx)

	// Parse times (we use a reference date, only time matters)
	refDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	startTime, _ := time.Parse("15:04", req.StartTime)
	startTime = time.Date(refDate.Year(), refDate.Month(), refDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)

	endTime, _ := time.Parse("15:04", req.EndTime)
	endTime = time.Date(refDate.Year(), refDate.Month(), refDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	// Calculate duration
	duration := int(endTime.Sub(startTime).Minutes())

	status := types.OccurrenceActive
	if req.Status != nil {
		status = types.OccurrenceStatus(*req.Status)
	}

	return &eventdomain.EventOccurrence{
		ID:              types.GenerateUUIDWithPrefix(types.UUID_PREFIX_OCCURRENCE),
		EventID:         eventID,
		RecurrenceType:  types.RecurrenceType(req.RecurrenceType),
		StartTime:       startTime,
		EndTime:         endTime,
		DurationMinutes: &duration,
		DayOfWeek:       req.DayOfWeek,
		DayOfMonth:      req.DayOfMonth,
		MonthOfYear:     req.MonthOfYear,
		ExceptionDates:  req.ExceptionDates,
		Metadata:        req.Metadata,
		Status:          status,
		BaseModel:       baseModel,
	}, nil
}

// UpdateOccurrenceRequest represents a request to update an occurrence
type UpdateOccurrenceRequest struct {
	RecurrenceType *string                `json:"recurrence_type,omitempty"`
	StartTime      *string                `json:"start_time,omitempty"` // HH:MM format
	EndTime        *string                `json:"end_time,omitempty"`   // HH:MM format
	DayOfWeek      *int                   `json:"day_of_week,omitempty"`
	DayOfMonth     *int                   `json:"day_of_month,omitempty"`
	MonthOfYear    *int                   `json:"month_of_year,omitempty"`
	ExceptionDates []string               `json:"exception_dates,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Status         *string                `json:"status,omitempty"`
}

// Validate validates the UpdateOccurrenceRequest
func (req *UpdateOccurrenceRequest) Validate() error {
	if err := validator.ValidateRequest(req); err != nil {
		return err
	}

	// Validate recurrence type if provided
	if req.RecurrenceType != nil {
		if err := validator.ValidateRecurrenceType(*req.RecurrenceType); err != nil {
			return err
		}
	}

	// Parse and validate times if provided
	var startTime, endTime *time.Time
	if req.StartTime != nil {
		parsed, err := time.Parse("15:04", *req.StartTime)
		if err != nil {
			return ierr.NewError("Invalid start_time format, expected HH:MM (24-hour format)").Mark(ierr.ErrValidation)
		}
		startTime = &parsed
	}

	if req.EndTime != nil {
		parsed, err := time.Parse("15:04", *req.EndTime)
		if err != nil {
			return ierr.NewError("Invalid end_time format, expected HH:MM (24-hour format)").Mark(ierr.ErrValidation)
		}
		endTime = &parsed
	}

	// If both times are provided, validate them
	if startTime != nil && endTime != nil {
		if err := validator.ValidateOccurrenceTimes(*startTime, *endTime); err != nil {
			return err
		}
	}

	// Validate exception dates
	if len(req.ExceptionDates) > 0 {
		if err := validator.ValidateExceptionDates(req.ExceptionDates); err != nil {
			return err
		}
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := []string{"active", "paused", "archived"}
		valid := false
		for _, s := range validStatuses {
			if *req.Status == s {
				valid = true
				break
			}
		}
		if !valid {
			return ierr.NewError("Invalid status. Must be one of: active, paused, archived").
				Mark(ierr.ErrValidation)
		}
	}

	return nil
}

// ApplyToOccurrence applies the update request to an existing occurrence
func (req *UpdateOccurrenceRequest) ApplyToOccurrence(ctx context.Context, occ *eventdomain.EventOccurrence) error {
	refDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	if req.RecurrenceType != nil {
		occ.RecurrenceType = types.RecurrenceType(*req.RecurrenceType)
	}
	if req.StartTime != nil {
		parsed, _ := time.Parse("15:04", *req.StartTime)
		occ.StartTime = time.Date(refDate.Year(), refDate.Month(), refDate.Day(),
			parsed.Hour(), parsed.Minute(), 0, 0, time.UTC)
	}
	if req.EndTime != nil {
		parsed, _ := time.Parse("15:04", *req.EndTime)
		occ.EndTime = time.Date(refDate.Year(), refDate.Month(), refDate.Day(),
			parsed.Hour(), parsed.Minute(), 0, 0, time.UTC)
	}

	// Recalculate duration
	duration := int(occ.EndTime.Sub(occ.StartTime).Minutes())
	occ.DurationMinutes = &duration

	if req.DayOfWeek != nil {
		occ.DayOfWeek = req.DayOfWeek
	}
	if req.DayOfMonth != nil {
		occ.DayOfMonth = req.DayOfMonth
	}
	if req.MonthOfYear != nil {
		occ.MonthOfYear = req.MonthOfYear
	}
	if req.ExceptionDates != nil {
		occ.ExceptionDates = req.ExceptionDates
	}
	if req.Metadata != nil {
		occ.Metadata = req.Metadata
	}
	if req.Status != nil {
		occ.Status = types.OccurrenceStatus(*req.Status)
	}

	// Update audit fields
	occ.UpdatedBy = types.GetUserID(ctx)
	occ.UpdatedAt = time.Now().UTC()

	return nil
}

// OccurrenceResponse represents an occurrence in the response
type OccurrenceResponse struct {
	*eventdomain.EventOccurrence
}

// NewOccurrenceResponse creates an OccurrenceResponse from domain EventOccurrence
func NewOccurrenceResponse(occ *eventdomain.EventOccurrence) *OccurrenceResponse {
	return &OccurrenceResponse{
		EventOccurrence: occ,
	}
}

// ExpandedOccurrenceResponse represents a concrete expanded occurrence instance
type ExpandedOccurrenceResponse struct {
	*eventdomain.ExpandedOccurrence
}

// NewExpandedOccurrenceResponse creates an ExpandedOccurrenceResponse from domain ExpandedOccurrence
func NewExpandedOccurrenceResponse(expanded *eventdomain.ExpandedOccurrence) *ExpandedOccurrenceResponse {
	return &ExpandedOccurrenceResponse{
		ExpandedOccurrence: expanded,
	}
}
