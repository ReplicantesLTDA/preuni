import { ScrollView, Text, View } from 'react-native';
import { Card } from '@/components/Card';
import { EmptyState } from '@/components/EmptyState';
import { ScreenContainer } from '@/components/ScreenContainer';
import { useTheme } from '@/theme';
import { t } from '@/lib/i18n/pt-BR';
import { useMyMedals, useMyRanking, useWeeklyLeaderboard } from '@/features/gamification/hooks';
import type { MedalTypeValue } from '@/types/gamification';

const TIER_LABEL: Record<string, string> = {
  bronze: t.gamification.tierBronze,
  silver: t.gamification.tierSilver,
  gold: t.gamification.tierGold,
  platinum: t.gamification.tierPlatinum,
  diamond: t.gamification.tierDiamond,
};

const MEDAL_LABEL: Record<MedalTypeValue, string> = {
  streak_7_day: t.gamification.medalStreak7Day,
  streak_30_day: t.gamification.medalStreak30Day,
  streak_100_day: t.gamification.medalStreak100Day,
  tier_promotion: t.gamification.medalTierPromotion,
  weekly_top_finish: t.gamification.medalWeeklyTopFinish,
};

export default function RankingScreen() {
  const { color, space, font, size } = useTheme();
  const { data: myRanking } = useMyRanking();
  const { data: leaderboard } = useWeeklyLeaderboard(myRanking?.leagueTier ?? 'bronze');
  const { data: medals } = useMyMedals();

  return (
    <ScreenContainer>
      <ScrollView contentContainerStyle={{ padding: space[5], paddingBottom: space[8] }}>
        <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'], marginBottom: space[4] }}>
          {t.gamification.title}
        </Text>

        {myRanking ? (
          <Card>
            <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.sm }}>
              {t.gamification.myRank}
            </Text>
            <Text style={{ color: color.ink[1], fontFamily: font.display, fontSize: size['2xl'] }}>
              {TIER_LABEL[myRanking.leagueTier] ?? myRanking.leagueTier}
              {myRanking.rankInTier != null ? ` · #${myRanking.rankInTier}` : ''}
            </Text>
            <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.sm }}>
              {myRanking.weeklyScore} pontos esta semana
            </Text>
          </Card>
        ) : null}

        <View style={{ height: space[5] }} />

        <Text style={{ color: color.ink[1], fontFamily: font.heading, fontSize: size.lg, marginBottom: space[2] }}>
          {t.gamification.leaderboard}
        </Text>
        {leaderboard && leaderboard.length > 0 ? (
          <View style={{ gap: space[2] }}>
            {leaderboard.map((entry, i) => (
              <Card key={entry.userId}>
                <Text style={{ color: color.ink[1], fontFamily: font.heading, fontSize: size.md }}>
                  #{i + 1} {entry.displayName}
                </Text>
                <Text style={{ color: color.ink[2], fontFamily: font.body, fontSize: size.sm }}>
                  {entry.weeklyScore} pontos
                </Text>
              </Card>
            ))}
          </View>
        ) : (
          <EmptyState title={t.gamification.leaderboard} body="Ninguém pontuou nesta liga ainda." mascot="thinking" />
        )}

        <View style={{ height: space[5] }} />

        <Text style={{ color: color.ink[1], fontFamily: font.heading, fontSize: size.lg, marginBottom: space[2] }}>
          {t.gamification.medals}
        </Text>
        {medals && medals.length > 0 ? (
          <View style={{ gap: space[2] }}>
            {medals.map((m, i) => (
              <Card key={`${m.type}-${i}`}>
                <Text style={{ color: color.ink[1], fontFamily: font.body, fontSize: size.md }}>
                  {MEDAL_LABEL[m.type] ?? m.type}
                </Text>
              </Card>
            ))}
          </View>
        ) : (
          <EmptyState title={t.gamification.medals} body={t.gamification.noMedalsYet} mascot="reading" />
        )}
      </ScrollView>
    </ScreenContainer>
  );
}
