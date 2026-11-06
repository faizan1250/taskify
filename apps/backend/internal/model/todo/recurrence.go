package todo

import (
	"time"
)

// Recurrence is the repeat rule of a todo. RecurrenceNone means the todo
// never repeats; any other value makes the todo a template whose next
// occurrence is tracked in NextRunAt and materialized by the
// recurring-todos cron job.
type Recurrence string

const (
	RecurrenceNone    Recurrence = "none"
	RecurrenceDaily   Recurrence = "daily"
	RecurrenceWeekly  Recurrence = "weekly"
	RecurrenceMonthly Recurrence = "monthly"
)

// IsRecurring reports whether the rule produces future occurrences.
func (r Recurrence) IsRecurring() bool {
	switch r {
	case RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly:
		return true
	default:
		return false
	}
}

// NextOccurrence advances from by one step of r. It returns false for
// RecurrenceNone and any unknown value so callers never schedule on a
// rule they do not understand. Monthly steps clamp to the last day of the
// target month (Jan 31 -> Feb 28) instead of overflowing into the month
// after, so a monthly template never drifts off month-end.
func NextOccurrence(from time.Time, r Recurrence) (time.Time, bool) {
	switch r {
	case RecurrenceDaily:
		return from.Add(24 * time.Hour), true
	case RecurrenceWeekly:
		return from.Add(7 * 24 * time.Hour), true
	case RecurrenceMonthly:
		return addCalendarMonth(from), true
	default:
		return time.Time{}, false
	}
}

// addCalendarMonth returns the same wall-clock time one calendar month
// later, clamped to the last day of the target month when the source day
// does not exist there.
func addCalendarMonth(from time.Time) time.Time {
	y, m, d := from.Date()
	target := m + 1
	targetYear := y
	if target > 12 {
		target = 1
		targetYear++
	}
	last := lastDayOfMonth(targetYear, target)
	if d > last {
		d = last
	}
	return time.Date(targetYear, target, d, from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), from.Location())
}

// lastDayOfMonth reports the number of days in the given month.
func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// ResolveNextRun computes the initial next_run_at for a todo with rule r.
// It returns nil for non-recurring todos (nothing to schedule). When a due
// date is set the first occurrence fires at the due date itself — even when
// that date is past, so the cron picks it up on its next pass and the
// template then advances. Without a due date the first occurrence is one
// full step from now.
func ResolveNextRun(r Recurrence, dueDate *time.Time, now time.Time) *time.Time {
	if !r.IsRecurring() {
		return nil
	}
	if dueDate != nil {
		next := *dueDate
		return &next
	}
	next, _ := NextOccurrence(now, r)
	return &next
}
