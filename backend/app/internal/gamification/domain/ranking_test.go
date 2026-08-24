package domain

import (
	"testing"
	"time"
)

func at(h int) time.Time {
	return time.Date(2026, 8, 22, h, 0, 0, 0, time.UTC)
}

func TestTierIndex_UnknownTierDefaultsToZero(t *testing.T) {
	// tierOrder only contains the 5 known tiers; an unrecognized value
	// (e.g. from stale/corrupt data) must fall back to index 0 rather than
	// panicking -- previously untested.
	if got := tierIndex(Tier("not-a-real-tier")); got != 0 {
		t.Fatalf("expected 0 for an unknown tier, got %d", got)
	}
}

func TestRankEntries_OrdersByScoreDescending(t *testing.T) {
	entries := []ScoreEntry{
		{UserID: "a", Score: 600, EarliestGradedAt: at(1)},
		{UserID: "b", Score: 900, EarliestGradedAt: at(2)},
		{UserID: "c", Score: 750, EarliestGradedAt: at(3)},
	}
	ranked := RankEntries(entries)
	if ranked[0].UserID != "b" || ranked[1].UserID != "c" || ranked[2].UserID != "a" {
		t.Fatalf("expected order [b,c,a], got %v", userIDs(ranked))
	}
}

func TestRankEntries_TieBrokenByEarliestGradedAt(t *testing.T) {
	entries := []ScoreEntry{
		{UserID: "later", Score: 800, EarliestGradedAt: at(10)},
		{UserID: "earlier", Score: 800, EarliestGradedAt: at(2)},
	}
	ranked := RankEntries(entries)
	if ranked[0].UserID != "earlier" {
		t.Fatalf("expected the earlier submission to win the tie, got %v", userIDs(ranked))
	}
}

func TestRankEntries_DeterministicRegardlessOfInputOrder(t *testing.T) {
	a := []ScoreEntry{
		{UserID: "x", Score: 500, EarliestGradedAt: at(1)},
		{UserID: "y", Score: 700, EarliestGradedAt: at(1)},
	}
	b := []ScoreEntry{a[1], a[0]}

	ra, rb := RankEntries(a), RankEntries(b)
	if userIDs(ra)[0] != userIDs(rb)[0] {
		t.Fatalf("ranking must not depend on input order: got %v vs %v", userIDs(ra), userIDs(rb))
	}
}

func userIDs(entries []ScoreEntry) []string {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.UserID
	}
	return ids
}

func TestTierMovement_TopBandPromotes(t *testing.T) {
	if TierMovement(1, 10) != Promote {
		t.Fatal("rank 1 of 10 (top 20%) must promote")
	}
	if TierMovement(2, 10) != Promote {
		t.Fatal("rank 2 of 10 (top 20%) must promote")
	}
}

func TestTierMovement_BottomBandDemotes(t *testing.T) {
	if TierMovement(9, 10) != Demote {
		t.Fatal("rank 9 of 10 (bottom 20%) must demote")
	}
	if TierMovement(10, 10) != Demote {
		t.Fatal("rank 10 of 10 (bottom 20%) must demote")
	}
}

func TestTierMovement_MiddleStays(t *testing.T) {
	if TierMovement(5, 10) != Stay {
		t.Fatal("a middling rank must stay")
	}
}

func TestTierMovement_TinyTierNeverMoves(t *testing.T) {
	if TierMovement(1, 3) != Stay {
		t.Fatal("a tier with fewer than 5 members must never promote/demote (avoids rank thrash)")
	}
}

func TestNextTier_PromoteAdvancesOneTier(t *testing.T) {
	if NextTier(TierBronze, Promote) != TierSilver {
		t.Fatal("bronze must promote to silver")
	}
}

func TestNextTier_ClampedAtDiamond(t *testing.T) {
	if NextTier(TierDiamond, Promote) != TierDiamond {
		t.Fatal("diamond must not promote past diamond")
	}
}

func TestNextTier_ClampedAtBronze(t *testing.T) {
	if NextTier(TierBronze, Demote) != TierBronze {
		t.Fatal("bronze must not demote below bronze")
	}
}

func TestNextTier_StayIsNoOp(t *testing.T) {
	if NextTier(TierGold, Stay) != TierGold {
		t.Fatal("stay must not change tier")
	}
}
