package domain

import (
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestNextStreak_FirstEverSubmission(t *testing.T) {
	today := day(2026, 8, 22)
	current, longest, changed := NextStreak(nil, today, 0, 0)
	if current != 1 || longest != 1 || !changed {
		t.Fatalf("got (%d, %d, %v), want (1, 1, true)", current, longest, changed)
	}
}

func TestNextStreak_ConsecutiveDayIncrements(t *testing.T) {
	yesterday := day(2026, 8, 21)
	today := day(2026, 8, 22)
	current, longest, changed := NextStreak(&yesterday, today, 5, 5)
	if current != 6 || longest != 6 || !changed {
		t.Fatalf("got (%d, %d, %v), want (6, 6, true)", current, longest, changed)
	}
}

func TestNextStreak_SameDayIsIdempotent(t *testing.T) {
	today := day(2026, 8, 22)
	current, longest, changed := NextStreak(&today, today, 5, 7)
	if current != 5 || longest != 7 || changed {
		t.Fatalf("got (%d, %d, %v), want (5, 7, false) — a second same-day submission must not change the streak", current, longest, changed)
	}
}

func TestNextStreak_GapOfTwoDaysResets(t *testing.T) {
	twoDaysAgo := day(2026, 8, 20)
	today := day(2026, 8, 22)
	current, longest, changed := NextStreak(&twoDaysAgo, today, 10, 10)
	if current != 1 || longest != 10 || !changed {
		t.Fatalf("got (%d, %d, %v), want (1, 10, true) — longest_streak must never decrease", current, longest, changed)
	}
}

func TestNextStreak_NewRecordRaisesLongest(t *testing.T) {
	yesterday := day(2026, 8, 21)
	today := day(2026, 8, 22)
	current, longest, changed := NextStreak(&yesterday, today, 9, 9)
	if current != 10 || longest != 10 || !changed {
		t.Fatalf("got (%d, %d, %v), want (10, 10, true)", current, longest, changed)
	}
}

func TestNextStreak_UTCBoundaryNotLocalTime(t *testing.T) {
	// 23:59 UTC on the 21st and 00:01 UTC on the 22nd are different UTC
	// calendar days even though only two minutes apart (spec Edge Cases).
	lastActive := time.Date(2026, 8, 21, 23, 59, 0, 0, time.UTC)
	today := time.Date(2026, 8, 22, 0, 1, 0, 0, time.UTC)
	current, _, changed := NextStreak(&lastActive, today, 3, 3)
	if current != 4 || !changed {
		t.Fatalf("got (%d, %v), want (4, true) — UTC day boundary crossed", current, changed)
	}
}
