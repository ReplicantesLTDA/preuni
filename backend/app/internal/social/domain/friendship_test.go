package domain

import "testing"

func TestCanSendRequest_NoExistingRelationship(t *testing.T) {
	if !CanSendRequest(nil) {
		t.Fatal("a first-ever request between two users must be allowed")
	}
}

func TestCanSendRequest_AlreadyPendingBlocked(t *testing.T) {
	pending := StatusPending
	if CanSendRequest(&pending) {
		t.Fatal("a second request while one is already pending must be blocked")
	}
}

func TestCanSendRequest_AlreadyAcceptedBlocked(t *testing.T) {
	accepted := StatusAccepted
	if CanSendRequest(&accepted) {
		t.Fatal("a request between users who are already friends must be blocked")
	}
}

func TestCanSendRequest_PreviouslyRemovedAllowed(t *testing.T) {
	removed := StatusRemoved
	if !CanSendRequest(&removed) {
		t.Fatal("spec Edge Cases: a request after a prior removal must be treated as new and allowed")
	}
}

func TestIsVisible_OnlyAcceptedGrantsVisibility(t *testing.T) {
	if IsVisible(StatusPending) {
		t.Fatal("a pending request must not grant visibility")
	}
	if IsVisible(StatusRemoved) {
		t.Fatal("a removed friendship must not grant visibility")
	}
	if !IsVisible(StatusAccepted) {
		t.Fatal("an accepted friendship must grant visibility")
	}
}
