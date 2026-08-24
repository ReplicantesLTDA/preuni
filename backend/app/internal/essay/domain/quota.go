// Package domain holds the essay domain's pure, DB-free decision logic.
package domain

// ProTier is the subscription_tier value that grants unlimited daily
// submissions (constitution: Free tier one graded essay submission per
// day; Pro tier multiple submissions per day).
const ProTier = "pro"

// QuotaExceeded reports whether a new submission should be rejected, given
// the caller's subscription tier and whether they've already submitted an
// essay on the current UTC day (constitution: streak day boundary and
// quota day boundary are both fixed UTC).
func QuotaExceeded(subscriptionTier string, alreadySubmittedToday bool) bool {
	if subscriptionTier == ProTier {
		return false
	}
	return alreadySubmittedToday
}
