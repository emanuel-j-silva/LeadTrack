package domain

import (
	"time"
)

// WeekStart calculates the Monday of the week for a given time in the specified timezone.
func WeekStart(t time.Time, loc *time.Location) time.Time {
	tLoc := t.In(loc)
	// Truncate to date
	year, month, day := tLoc.Date()
	date := time.Date(year, month, day, 0, 0, 0, 0, loc)

	weekday := date.Weekday()
	// Weekday: Sunday = 0, Monday = 1, ..., Saturday = 6
	daysToSubtract := int(weekday - time.Monday)
	if weekday == time.Sunday {
		daysToSubtract = 6
	}

	return date.AddDate(0, 0, -daysToSubtract)
}
