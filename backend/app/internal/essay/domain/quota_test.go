package domain

import "testing"

func TestQuotaExceeded_FreeTierFirstSubmissionOfDay(t *testing.T) {
	if QuotaExceeded("free", false) {
		t.Fatal("free tier's first submission of the day must be allowed")
	}
}

func TestQuotaExceeded_FreeTierSecondSubmissionOfDay(t *testing.T) {
	if !QuotaExceeded("free", true) {
		t.Fatal("free tier's second submission of the day must be blocked")
	}
}

func TestQuotaExceeded_ProTierAlwaysAllowed(t *testing.T) {
	if QuotaExceeded(ProTier, true) {
		t.Fatal("Pro tier must never be blocked by the daily quota")
	}
	if QuotaExceeded(ProTier, false) {
		t.Fatal("Pro tier must never be blocked by the daily quota")
	}
}
