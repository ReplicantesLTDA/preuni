import { z } from 'zod';

export const RankingEntrySchema = z.object({
  userId: z.string().uuid(),
  displayName: z.string(),
  weeklyScore: z.number().int().nonnegative(),
  leagueTier: z.enum(['bronze', 'silver', 'gold', 'platinum', 'diamond']),
  rankInTier: z.number().int().positive().nullable().optional(),
});
export type RankingEntry = z.infer<typeof RankingEntrySchema>;

export const MedalTypeSchema = z.enum([
  'streak_7_day',
  'streak_30_day',
  'streak_100_day',
  'tier_promotion',
  'weekly_top_finish',
]);
export type MedalTypeValue = z.infer<typeof MedalTypeSchema>;

export const MedalSchema = z.object({
  type: MedalTypeSchema,
  earnedAt: z.string(),
});
export type Medal = z.infer<typeof MedalSchema>;
