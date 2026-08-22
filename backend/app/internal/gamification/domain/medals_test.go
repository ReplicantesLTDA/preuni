package domain

import (
	"reflect"
	"testing"
)

func TestStreakMedalsEarned_CrossingSingleMilestone(t *testing.T) {
	got := StreakMedalsEarned(6, 7)
	want := []MedalType{MedalStreak7Day}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestStreakMedalsEarned_NoMilestoneCrossed(t *testing.T) {
	got := StreakMedalsEarned(3, 4)
	if len(got) != 0 {
		t.Fatalf("expected no medals, got %v", got)
	}
}

func TestStreakMedalsEarned_JumpingMultipleMilestonesAwardsAll(t *testing.T) {
	got := StreakMedalsEarned(5, 30)
	want := []MedalType{MedalStreak7Day, MedalStreak30Day}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestStreakMedalsEarned_HundredDayMilestone(t *testing.T) {
	got := StreakMedalsEarned(99, 100)
	want := []MedalType{MedalStreak100Day}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTierPromotionMedalEarned(t *testing.T) {
	if !TierPromotionMedalEarned(Promote) {
		t.Fatal("a promotion must earn the tier_promotion medal")
	}
	if TierPromotionMedalEarned(Stay) || TierPromotionMedalEarned(Demote) {
		t.Fatal("staying or demoting must not earn the tier_promotion medal")
	}
}

func TestWeeklyTopFinishMedalEarned(t *testing.T) {
	if !WeeklyTopFinishMedalEarned(1) {
		t.Fatal("rank 1 must earn the weekly_top_finish medal")
	}
	if WeeklyTopFinishMedalEarned(2) {
		t.Fatal("rank 2 must not earn the weekly_top_finish medal")
	}
}
