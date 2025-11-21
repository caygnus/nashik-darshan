package service

import (
	"sort"
	"time"

	eventdomain "github.com/omkar273/nashikdarshan/internal/domain/event"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/types"
)

// EventExpander handles expansion of recurrence rules into concrete occurrences
type EventExpander struct {
	timezone *time.Location
}

// NewEventExpander creates a new event expander with IST timezone
func NewEventExpander() *EventExpander {
	// Load IST timezone (Asia/Kolkata)
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fallback to UTC+5:30 if timezone database not available
		ist = time.FixedZone("IST", 5*60*60+30*60)
	}

	return &EventExpander{
		timezone: ist,
	}
}

// ExpandOccurrences expands all occurrences for an event within a date range
func (e *EventExpander) ExpandOccurrences(
	event *eventdomain.Event,
	occurrences []*eventdomain.EventOccurrence,
	fromDateStr, toDateStr string,
) ([]*eventdomain.ExpandedOccurrence, error) {
	// Parse date range
	fromDate, err := time.Parse("2006-01-02", fromDateStr)
	if err != nil {
		return nil, ierr.NewError("Invalid from_date format, expected YYYY-MM-DD").Mark(ierr.ErrValidation)
	}

	toDate, err := time.Parse("2006-01-02", toDateStr)
	if err != nil {
		return nil, ierr.NewError("Invalid to_date format, expected YYYY-MM-DD").Mark(ierr.ErrValidation)
	}

	// Ensure dates are within event validity window
	if fromDate.Before(event.StartDate) {
		fromDate = event.StartDate
	}
	if event.EndDate != nil && toDate.After(*event.EndDate) {
		toDate = *event.EndDate
	}

	var expanded []*eventdomain.ExpandedOccurrence

	// Expand each occurrence
	for _, occ := range occurrences {
		instances := e.expandSingleOccurrence(event, occ, fromDate, toDate)
		expanded = append(expanded, instances...)
	}

	// Sort by start time
	sort.Slice(expanded, func(i, j int) bool {
		return expanded[i].StartTime.Before(expanded[j].StartTime)
	})

	// Mark today and upcoming
	now := time.Now().In(e.timezone)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, e.timezone)

	for _, exp := range expanded {
		expDate := time.Date(
			exp.StartTime.Year(),
			exp.StartTime.Month(),
			exp.StartTime.Day(),
			0, 0, 0, 0, e.timezone,
		)

		exp.IsToday = expDate.Equal(today)
		exp.IsUpcoming = expDate.After(today) || exp.IsToday
		if exp.IsUpcoming {
			exp.DaysUntil = int(expDate.Sub(today).Hours() / 24)
		}
	}

	return expanded, nil
}

// expandSingleOccurrence expands a single occurrence based on its recurrence type
func (e *EventExpander) expandSingleOccurrence(
	event *eventdomain.Event,
	occ *eventdomain.EventOccurrence,
	fromDate, toDate time.Time,
) []*eventdomain.ExpandedOccurrence {
	switch occ.RecurrenceType {
	case types.RecurrenceNone:
		return e.expandNone(event, occ, fromDate, toDate)
	case types.RecurrenceDaily:
		return e.expandDaily(event, occ, fromDate, toDate)
	case types.RecurrenceWeekly:
		return e.expandWeekly(event, occ, fromDate, toDate)
	case types.RecurrenceMonthly:
		return e.expandMonthly(event, occ, fromDate, toDate)
	case types.RecurrenceYearly:
		return e.expandYearly(event, occ, fromDate, toDate)
	default:
		return nil
	}
}

// expandNone handles one-time events
func (e *EventExpander) expandNone(
	event *eventdomain.Event,
	occ *eventdomain.EventOccurrence,
	fromDate, toDate time.Time,
) []*eventdomain.ExpandedOccurrence {
	// For NONE type, use event.StartDate as the occurrence date
	occDate := event.StartDate

	// Check if within range
	if occDate.Before(fromDate) || occDate.After(toDate) {
		return nil
	}

	// Check if in exception list
	if e.isException(occ, occDate) {
		return nil
	}

	startTime := e.combineDateAndTime(occDate, occ.StartTime)
	endTime := e.combineDateAndTime(occDate, occ.EndTime)

	return []*eventdomain.ExpandedOccurrence{{
		EventID:      event.ID,
		OccurrenceID: occ.ID,
		Title:        event.Title,
		StartTime:    startTime,
		EndTime:      endTime,
	}}
}

// expandDaily handles daily recurring events
func (e *EventExpander) expandDaily(
	event *eventdomain.Event,
	occ *eventdomain.EventOccurrence,
	fromDate, toDate time.Time,
) []*eventdomain.ExpandedOccurrence {
	var expanded []*eventdomain.ExpandedOccurrence

	current := fromDate
	for !current.After(toDate) {
		if !e.isException(occ, current) {
			startTime := e.combineDateAndTime(current, occ.StartTime)
			endTime := e.combineDateAndTime(current, occ.EndTime)

			expanded = append(expanded, &eventdomain.ExpandedOccurrence{
				EventID:      event.ID,
				OccurrenceID: occ.ID,
				Title:        event.Title,
				StartTime:    startTime,
				EndTime:      endTime,
			})
		}
		current = current.AddDate(0, 0, 1)
	}

	return expanded
}

// expandWeekly handles weekly recurring events
func (e *EventExpander) expandWeekly(
	event *eventdomain.Event,
	occ *eventdomain.EventOccurrence,
	fromDate, toDate time.Time,
) []*eventdomain.ExpandedOccurrence {
	if occ.DayOfWeek == nil {
		return nil
	}

	targetWeekday := time.Weekday(*occ.DayOfWeek)
	var expanded []*eventdomain.ExpandedOccurrence

	// Find first occurrence of target weekday on or after fromDate
	current := fromDate
	for current.Weekday() != targetWeekday {
		current = current.AddDate(0, 0, 1)
		if current.After(toDate) {
			return nil
		}
	}

	// Generate all occurrences
	for !current.After(toDate) {
		if !e.isException(occ, current) {
			startTime := e.combineDateAndTime(current, occ.StartTime)
			endTime := e.combineDateAndTime(current, occ.EndTime)

			expanded = append(expanded, &eventdomain.ExpandedOccurrence{
				EventID:      event.ID,
				OccurrenceID: occ.ID,
				Title:        event.Title,
				StartTime:    startTime,
				EndTime:      endTime,
			})
		}
		current = current.AddDate(0, 0, 7) // Next week
	}

	return expanded
}

// expandMonthly handles monthly recurring events
func (e *EventExpander) expandMonthly(
	event *eventdomain.Event,
	occ *eventdomain.EventOccurrence,
	fromDate, toDate time.Time,
) []*eventdomain.ExpandedOccurrence {
	if occ.DayOfMonth == nil {
		return nil
	}

	targetDay := *occ.DayOfMonth
	var expanded []*eventdomain.ExpandedOccurrence

	// Start from the first day of fromDate's month
	current := time.Date(fromDate.Year(), fromDate.Month(), 1, 0, 0, 0, 0, e.timezone)

	for !current.After(toDate) {
		// Check if this month has the target day
		lastDayOfMonth := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, e.timezone).Day()

		if targetDay <= lastDayOfMonth {
			occDate := time.Date(current.Year(), current.Month(), targetDay, 0, 0, 0, 0, e.timezone)

			// Check if within range
			if !occDate.Before(fromDate) && !occDate.After(toDate) && !e.isException(occ, occDate) {
				startTime := e.combineDateAndTime(occDate, occ.StartTime)
				endTime := e.combineDateAndTime(occDate, occ.EndTime)

				expanded = append(expanded, &eventdomain.ExpandedOccurrence{
					EventID:      event.ID,
					OccurrenceID: occ.ID,
					Title:        event.Title,
					StartTime:    startTime,
					EndTime:      endTime,
				})
			}
		}

		// Move to next month
		current = current.AddDate(0, 1, 0)
	}

	return expanded
}

// expandYearly handles yearly recurring events
func (e *EventExpander) expandYearly(
	event *eventdomain.Event,
	occ *eventdomain.EventOccurrence,
	fromDate, toDate time.Time,
) []*eventdomain.ExpandedOccurrence {
	if occ.DayOfMonth == nil || occ.MonthOfYear == nil {
		return nil
	}

	targetDay := *occ.DayOfMonth
	targetMonth := time.Month(*occ.MonthOfYear)
	var expanded []*eventdomain.ExpandedOccurrence

	// Start from fromDate's year
	for year := fromDate.Year(); year <= toDate.Year(); year++ {
		// Check if this year has valid date (handle Feb 29)
		lastDayOfMonth := time.Date(year, targetMonth+1, 0, 0, 0, 0, 0, e.timezone).Day()

		if targetDay <= lastDayOfMonth {
			occDate := time.Date(year, targetMonth, targetDay, 0, 0, 0, 0, e.timezone)

			// Check if within range
			if !occDate.Before(fromDate) && !occDate.After(toDate) && !e.isException(occ, occDate) {
				startTime := e.combineDateAndTime(occDate, occ.StartTime)
				endTime := e.combineDateAndTime(occDate, occ.EndTime)

				expanded = append(expanded, &eventdomain.ExpandedOccurrence{
					EventID:      event.ID,
					OccurrenceID: occ.ID,
					Title:        event.Title,
					StartTime:    startTime,
					EndTime:      endTime,
				})
			}
		}
	}

	return expanded
}

// combineDateAndTime combines a date with a time-of-day
func (e *EventExpander) combineDateAndTime(date, timeOfDay time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		timeOfDay.Hour(), timeOfDay.Minute(), timeOfDay.Second(),
		0, e.timezone,
	)
}

// isException checks if a date is in the exception list
func (e *EventExpander) isException(occ *eventdomain.EventOccurrence, date time.Time) bool {
	dateStr := date.Format("2006-01-02")
	for _, exDate := range occ.ExceptionDates {
		if exDate == dateStr {
			return true
		}
	}
	return false
}
