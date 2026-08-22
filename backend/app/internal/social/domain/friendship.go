// Package domain holds the social domain's pure, DB-free decision logic.
package domain

// Status mirrors social.friendship_status.
type Status string

const (
	StatusPending  Status = "pending"
	StatusAccepted Status = "accepted"
	StatusRemoved  Status = "removed"
)

// CanSendRequest reports whether a new friend request may be created,
// given the most recent existing friendship row between the same two
// users (nil if none exists). Per spec Edge Cases: a request to a user who
// previously removed the requester is treated as a brand new request —
// only an already-pending or already-accepted relationship blocks it.
func CanSendRequest(existing *Status) bool {
	if existing == nil {
		return true
	}
	return *existing == StatusRemoved
}

// IsVisible reports whether a friendship status grants visibility of the
// other user's streak and latest grade (constitution Principle III context:
// only accepted friends see this — spec FR-004).
func IsVisible(status Status) bool {
	return status == StatusAccepted
}
