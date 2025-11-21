package service

import (
	"context"

	"github.com/omkar273/nashikdarshan/internal/api/dto"
	eventdomain "github.com/omkar273/nashikdarshan/internal/domain/event"
	"github.com/omkar273/nashikdarshan/internal/types"
)

type EventService interface {
	// Event operations
	Create(ctx context.Context, req *dto.CreateEventRequest) (*dto.EventResponse, error)
	Get(ctx context.Context, id string) (*dto.EventResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dto.EventResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateEventRequest) (*dto.EventResponse, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *types.EventFilter) (*dto.ListEventsResponse, error)

	// Occurrence operations
	CreateOccurrence(ctx context.Context, eventID string, req *dto.CreateOccurrenceRequest) (*dto.OccurrenceResponse, error)
	GetOccurrence(ctx context.Context, id string) (*dto.OccurrenceResponse, error)
	UpdateOccurrence(ctx context.Context, id string, req *dto.UpdateOccurrenceRequest) (*dto.OccurrenceResponse, error)
	DeleteOccurrence(ctx context.Context, id string) error
	ListOccurrences(ctx context.Context, eventID string) ([]*dto.OccurrenceResponse, error)

	// Expanded occurrences (computed from recurrence rules)
	GetExpandedOccurrences(ctx context.Context, eventID string, fromDate, toDate string) ([]*eventdomain.ExpandedOccurrence, error)

	// Stats
	IncrementView(ctx context.Context, id string) error
	IncrementInterested(ctx context.Context, id string) error
}

type eventService struct {
	eventRepo eventdomain.Repository
	expander  *EventExpander
	ServiceParams
}

// NewEventService creates a new event service
// Note: Expects ServiceParams to include EventRepo
func NewEventService(params ServiceParams, eventRepo eventdomain.Repository) EventService {
	return &eventService{
		ServiceParams: params,
		eventRepo:     eventRepo,
		expander:      NewEventExpander(),
	}
}

// Create creates a new event
func (s *eventService) Create(ctx context.Context, req *dto.CreateEventRequest) (*dto.EventResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	event, err := req.ToEvent(ctx)
	if err != nil {
		return nil, err
	}

	err = s.eventRepo.Create(ctx, event)
	if err != nil {
		return nil, err
	}

	return dto.NewEventResponse(event), nil
}

// Get retrieves an event by ID
func (s *eventService) Get(ctx context.Context, id string) (*dto.EventResponse, error) {
	event, err := s.eventRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.NewEventResponse(event), nil
}

// GetBySlug retrieves an event by slug
func (s *eventService) GetBySlug(ctx context.Context, slug string) (*dto.EventResponse, error) {
	event, err := s.eventRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return dto.NewEventResponse(event), nil
}

// Update updates an existing event
func (s *eventService) Update(ctx context.Context, id string, req *dto.UpdateEventRequest) (*dto.EventResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	event, err := s.eventRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = req.ApplyToEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	return dto.NewEventResponse(event), nil
}

// Delete soft deletes an event
func (s *eventService) Delete(ctx context.Context, id string) error {
	_, err := s.eventRepo.Get(ctx, id)
	if err != nil {
		return err
	}

	return s.eventRepo.Delete(ctx, id)
}

// List retrieves a paginated list of events
func (s *eventService) List(ctx context.Context, filter *types.EventFilter) (*dto.ListEventsResponse, error) {
	if filter == nil {
		filter = types.NewEventFilter()
	}

	events, err := s.eventRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	total, err := s.eventRepo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	limit := filter.GetLimit()
	offset := filter.GetOffset()
	response := dto.NewListEventsResponse(events, total, limit, offset)

	return response, nil
}

// CreateOccurrence creates a new occurrence for an event
func (s *eventService) CreateOccurrence(ctx context.Context, eventID string, req *dto.CreateOccurrenceRequest) (*dto.OccurrenceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify event exists
	_, err := s.eventRepo.Get(ctx, eventID)
	if err != nil {
		return nil, err
	}

	occurrence, err := req.ToOccurrence(ctx, eventID)
	if err != nil {
		return nil, err
	}

	err = s.eventRepo.CreateOccurrence(ctx, occurrence)
	if err != nil {
		return nil, err
	}

	return dto.NewOccurrenceResponse(occurrence), nil
}

// GetOccurrence retrieves an occurrence by ID
func (s *eventService) GetOccurrence(ctx context.Context, id string) (*dto.OccurrenceResponse, error) {
	occurrence, err := s.eventRepo.GetOccurrence(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.NewOccurrenceResponse(occurrence), nil
}

// UpdateOccurrence updates an existing occurrence
func (s *eventService) UpdateOccurrence(ctx context.Context, id string, req *dto.UpdateOccurrenceRequest) (*dto.OccurrenceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	occurrence, err := s.eventRepo.GetOccurrence(ctx, id)
	if err != nil {
		return nil, err
	}

	err = req.ApplyToOccurrence(ctx, occurrence)
	if err != nil {
		return nil, err
	}

	err = s.eventRepo.UpdateOccurrence(ctx, occurrence)
	if err != nil {
		return nil, err
	}

	return dto.NewOccurrenceResponse(occurrence), nil
}

// DeleteOccurrence soft deletes an occurrence
func (s *eventService) DeleteOccurrence(ctx context.Context, id string) error {
	_, err := s.eventRepo.GetOccurrence(ctx, id)
	if err != nil {
		return err
	}

	return s.eventRepo.DeleteOccurrence(ctx, id)
}

// ListOccurrences lists all occurrences for an event
func (s *eventService) ListOccurrences(ctx context.Context, eventID string) ([]*dto.OccurrenceResponse, error) {
	occurrences, err := s.eventRepo.ListOccurrencesByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.OccurrenceResponse, len(occurrences))
	for i, occ := range occurrences {
		responses[i] = dto.NewOccurrenceResponse(occ)
	}

	return responses, nil
}

// GetExpandedOccurrences expands recurrence rules into concrete instances
func (s *eventService) GetExpandedOccurrences(ctx context.Context, eventID string, fromDate, toDate string) ([]*eventdomain.ExpandedOccurrence, error) {
	event, err := s.eventRepo.Get(ctx, eventID)
	if err != nil {
		return nil, err
	}

	occurrences, err := s.eventRepo.ListOccurrencesByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Use expander to compute concrete instances
	expanded, err := s.expander.ExpandOccurrences(event, occurrences, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	return expanded, nil
}

// IncrementView increments the view count
func (s *eventService) IncrementView(ctx context.Context, id string) error {
	return s.eventRepo.IncrementViewCount(ctx, id)
}

// IncrementInterested increments the interested count
func (s *eventService) IncrementInterested(ctx context.Context, id string) error {
	return s.eventRepo.IncrementInterestedCount(ctx, id)
}
