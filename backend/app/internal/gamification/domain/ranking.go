// Package domain holds the gamification domain's pure, DB-free decision
// logic: weekly ranking order, tier promotion/demotion, and medal triggers.
package domain

import (
	"sort"
	"time"
)

// ScoreEntry is one user's weekly score input to ranking.
type ScoreEntry struct {
	UserID           string
	Score            int
	EarliestGradedAt time.Time
}

// RankEntries orders entries by score descending. Ties are broken by
// earliest graded_at (spec Assumptions: "earlier submission timestamp
// wins the higher slot"). The result is deterministic for identical input
// regardless of input order (spec SC-004).
func RankEntries(entries []ScoreEntry) []ScoreEntry {
	sorted := make([]ScoreEntry, len(entries))
	copy(sorted, entries)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Score != sorted[j].Score {
			return sorted[i].Score > sorted[j].Score
		}
		return sorted[i].EarliestGradedAt.Before(sorted[j].EarliestGradedAt)
	})
	return sorted
}

// Tier is a league tier, ordered lowest to highest.
type Tier string

const (
	TierBronze   Tier = "bronze"
	TierSilver   Tier = "silver"
	TierGold     Tier = "gold"
	TierPlatinum Tier = "platinum"
	TierDiamond  Tier = "diamond"
)

var tierOrder = []Tier{TierBronze, TierSilver, TierGold, TierPlatinum, TierDiamond}

// Movement is a tier's outcome at week close.
type Movement string

const (
	Promote Movement = "promote"
	Demote  Movement = "demote"
	Stay    Movement = "stay"
)

// minTierSizeForMovement: tiers smaller than this never promote/demote —
// avoids rank thrash in a tiny group where "top 20%" is meaningless
// (documented assumption, spec Assumptions: tie-break/promotion mechanics
// left to planning).
const minTierSizeForMovement = 5

// TierMovement decides whether rankInTier (1-indexed) within a tier of
// tierSize members promotes, demotes, or stays: top/bottom 20% move,
// rounded up to at least one member, only once the tier has enough
// members for that to be meaningful.
func TierMovement(rankInTier, tierSize int) Movement {
	if tierSize < minTierSizeForMovement {
		return Stay
	}
	band := tierSize / 5
	if band < 1 {
		band = 1
	}
	if rankInTier <= band {
		return Promote
	}
	if rankInTier > tierSize-band {
		return Demote
	}
	return Stay
}

// NextTier applies a movement, clamped at the top (Diamond) and bottom
// (Bronze) — you can't promote past Diamond or demote below Bronze.
func NextTier(current Tier, movement Movement) Tier {
	idx := tierIndex(current)
	switch movement {
	case Promote:
		if idx < len(tierOrder)-1 {
			return tierOrder[idx+1]
		}
	case Demote:
		if idx > 0 {
			return tierOrder[idx-1]
		}
	}
	return current
}

func tierIndex(t Tier) int {
	for i, tt := range tierOrder {
		if tt == t {
			return i
		}
	}
	return 0
}
