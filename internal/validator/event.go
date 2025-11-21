package validator

import (
	"time"

	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/types"
)

// ValidateEventType validates if the event type is valid
func ValidateEventType(eventType string) error {
	validTypes := []types.EventType{
		types.EventTypeAarti,
		types.EventTypeFestival,
		types.EventTypeCultural,
		types.EventTypeWorkshop,
		types.EventTypeSpecialDarshan,
		types.EventTypeOther,
	}

	for _, valid := range validTypes {
		if types.EventType(eventType) == valid {
			return nil
		}
	}

	return ierr.NewError("Invalid event type. Must be one of: AARTI, FESTIVAL, CULTURAL, WORKSHOP, SPECIAL_DARSHAN, OTHER").
		Mark(ierr.ErrValidation)
}

// ValidateRecurrenceType validates if the recurrence type is valid
func ValidateRecurrenceType(recurrenceType string) error {
	validTypes := []types.RecurrenceType{
		types.RecurrenceNone,
		types.RecurrenceDaily,
		types.RecurrenceWeekly,
		types.RecurrenceMonthly,
		types.RecurrenceYearly,
	}

	for _, valid := range validTypes {
		if types.RecurrenceType(recurrenceType) == valid {
			return nil
		}
	}

	return ierr.NewError("Invalid recurrence type. Must be one of: NONE, DAILY, WEEKLY, MONTHLY, YEARLY").
		Mark(ierr.ErrValidation)
}

// ValidateEventDates validates event date logic
func ValidateEventDates(startDate time.Time, endDate *time.Time) error {
	// Start date must not be too far in the past
	now := time.Now().UTC()
	oneYearAgo := now.AddDate(-1, 0, 0)

	if startDate.Before(oneYearAgo) {
		return ierr.NewError("Start date cannot be more than 1 year in the past").
			Mark(ierr.ErrValidation)
	}

	// End date must be after start date if provided
	if endDate != nil && !endDate.After(startDate) {
		return ierr.NewError("End date must be after start date").
			Mark(ierr.ErrValidation)
	}

	// Event duration should not exceed 10 years
	if endDate != nil {
		maxEndDate := startDate.AddDate(10, 0, 0)
		if endDate.After(maxEndDate) {
			return ierr.NewError("Event duration cannot exceed 10 years").
				Mark(ierr.ErrValidation)
		}
	}

	return nil
}

// ValidateOccurrenceTimes validates occurrence time logic
func ValidateOccurrenceTimes(startTime, endTime time.Time) error {
	// Extract only time components for comparison
	startHour, startMin := startTime.Hour(), startTime.Minute()
	endHour, endMin := endTime.Hour(), endTime.Minute()

	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	if endMinutes <= startMinutes {
		return ierr.NewError("End time must be after start time").
			Mark(ierr.ErrValidation)
	}

	// Duration should be reasonable (max 12 hours)
	duration := endMinutes - startMinutes
	if duration > 12*60 {
		return ierr.NewError("Occurrence duration cannot exceed 12 hours").
			Mark(ierr.ErrValidation)
	}

	return nil
}

// ValidateDayOfWeek validates day of week (0-6, Sunday=0)
func ValidateDayOfWeek(day int) error {
	if day < 0 || day > 6 {
		return ierr.NewError("Day of week must be between 0 (Sunday) and 6 (Saturday)").
			Mark(ierr.ErrValidation)
	}
	return nil
}

// ValidateDayOfMonth validates day of month (1-31)
func ValidateDayOfMonth(day int) error {
	if day < 1 || day > 31 {
		return ierr.NewError("Day of month must be between 1 and 31").
			Mark(ierr.ErrValidation)
	}
	return nil
}

// ValidateMonthOfYear validates month of year (1-12)
func ValidateMonthOfYear(month int) error {
	if month < 1 || month > 12 {
		return ierr.NewError("Month of year must be between 1 (January) and 12 (December)").
			Mark(ierr.ErrValidation)
	}
	return nil
}

// ValidateRecurrenceRules validates recurrence-specific rules
func ValidateRecurrenceRules(recurrenceType types.RecurrenceType, dayOfWeek, dayOfMonth, monthOfYear *int) error {
	switch recurrenceType {
	case types.RecurrenceNone:
		// No day restrictions for one-time events
		return nil

	case types.RecurrenceDaily:
		// Daily events don't need day specifications
		return nil

	case types.RecurrenceWeekly:
		// Weekly events MUST have day_of_week
		if dayOfWeek == nil {
			return ierr.NewError("Weekly recurrence requires day_of_week (0-6)").
				Mark(ierr.ErrValidation)
		}
		if err := ValidateDayOfWeek(*dayOfWeek); err != nil {
			return err
		}
		return nil

	case types.RecurrenceMonthly:
		// Monthly events MUST have day_of_month
		if dayOfMonth == nil {
			return ierr.NewError("Monthly recurrence requires day_of_month (1-31)").
				Mark(ierr.ErrValidation)
		}
		if err := ValidateDayOfMonth(*dayOfMonth); err != nil {
			return err
		}
		return nil

	case types.RecurrenceYearly:
		// Yearly events MUST have both day_of_month and month_of_year
		if dayOfMonth == nil {
			return ierr.NewError("Yearly recurrence requires day_of_month (1-31)").
				Mark(ierr.ErrValidation)
		}
		if monthOfYear == nil {
			return ierr.NewError("Yearly recurrence requires month_of_year (1-12)").
				Mark(ierr.ErrValidation)
		}
		if err := ValidateDayOfMonth(*dayOfMonth); err != nil {
			return err
		}
		if err := ValidateMonthOfYear(*monthOfYear); err != nil {
			return err
		}
		return nil

	default:
		return ierr.NewError("Invalid recurrence type").
			Mark(ierr.ErrValidation)
	}
}

// ValidateExceptionDates validates exception date format
func ValidateExceptionDates(dates []string) error {
	for _, dateStr := range dates {
		_, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return ierr.NewError("Exception dates must be in YYYY-MM-DD format").
				WithReportableDetails(map[string]any{
					"invalid_date": dateStr,
				}).
				Mark(ierr.ErrValidation)
		}
	}
	return nil
}
