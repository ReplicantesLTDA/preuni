// Package domain holds the pure, DB-free streak state machine.
package domain

import "time"

// NextStreak computes the resulting streak state for a submission accepted
// "today" (a fixed UTC calendar day — constitution clarification: streak
// day boundary is fixed UTC, not per-user local time), given the user's
// previously recorded last-active day and streak counters.
//
// Rules:
//   - No prior submission (lastActiveDay is nil): streak starts at 1.
//   - A submission already recorded today: no-op (idempotent — a Pro user
//     submitting multiple times in one day doesn't inflate the streak).
//   - Last submission was yesterday: streak continues, +1.
//   - Any larger gap: the streak was broken; it resets to 1.
//
// longestStreak is only ever raised, never lowered.
func NextStreak(
	lastActiveDay *time.Time,
	today time.Time,
	currentStreak int,
	longestStreak int,
) (newCurrent int, newLongest int, changed bool) {
	today = truncateUTCDay(today)

	switch {
	case lastActiveDay == nil:
		newCurrent = 1
	case truncateUTCDay(*lastActiveDay).Equal(today):
		newCurrent = currentStreak
	case truncateUTCDay(*lastActiveDay).Equal(today.AddDate(0, 0, -1)):
		newCurrent = currentStreak + 1
	default:
		newCurrent = 1
	}

	newLongest = longestStreak
	if newCurrent > newLongest {
		newLongest = newCurrent
	}

	changed = newCurrent != currentStreak || newLongest != longestStreak ||
		lastActiveDay == nil || !truncateUTCDay(*lastActiveDay).Equal(today)

	return newCurrent, newLongest, changed
}

func truncateUTCDay(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}
