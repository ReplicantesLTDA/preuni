package domain

// MedalType mirrors gamification.medal_type.
type MedalType string

const (
	MedalStreak7Day      MedalType = "streak_7_day"
	MedalStreak30Day     MedalType = "streak_30_day"
	MedalStreak100Day    MedalType = "streak_100_day"
	MedalTierPromotion   MedalType = "tier_promotion"
	MedalWeeklyTopFinish MedalType = "weekly_top_finish"
)

var streakMilestones = []struct {
	days  int
	medal MedalType
}{
	{7, MedalStreak7Day},
	{30, MedalStreak30Day},
	{100, MedalStreak100Day},
}

// StreakMedalsEarned returns the milestone medals newly crossed by a
// streak advancing from oldStreak to newStreak (oldStreak < newStreak).
// Awards every milestone crossed in one jump, not just the highest, so a
// user who somehow skips a milestone still receives all of them.
func StreakMedalsEarned(oldStreak, newStreak int) []MedalType {
	var earned []MedalType
	for _, m := range streakMilestones {
		if oldStreak < m.days && newStreak >= m.days {
			earned = append(earned, m.medal)
		}
	}
	return earned
}

// TierPromotionMedalEarned reports whether a tier movement earns the
// tier_promotion medal (spec: "ranking milestone (e.g., promoted to a new
// tier)").
func TierPromotionMedalEarned(movement Movement) bool {
	return movement == Promote
}

// WeeklyTopFinishMedalEarned reports whether finishing rank 1 in a tier at
// week close earns the weekly_top_finish medal.
func WeeklyTopFinishMedalEarned(rankInTier int) bool {
	return rankInTier == 1
}
