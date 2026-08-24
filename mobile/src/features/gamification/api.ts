import { z } from 'zod';
import type { ApiClient } from '@/lib/api/client';
import { RankingEntrySchema, MedalSchema, type RankingEntry, type Medal } from '@/types/gamification';

const RankingListSchema = z.array(RankingEntrySchema);
const MedalListSchema = z.array(MedalSchema);

export function makeGamificationApi(api: ApiClient) {
  return {
    weeklyLeaderboard(tier: string): Promise<RankingEntry[]> {
      return api.request({
        method: 'GET',
        path: `/v1/ranking/weekly?tier=${encodeURIComponent(tier)}`,
        schema: RankingListSchema,
      });
    },

    myRanking(): Promise<RankingEntry> {
      return api.request({ method: 'GET', path: '/v1/ranking/me', schema: RankingEntrySchema });
    },

    myMedals(): Promise<Medal[]> {
      return api.request({ method: 'GET', path: '/v1/medals/me', schema: MedalListSchema });
    },
  };
}

export type GamificationApi = ReturnType<typeof makeGamificationApi>;
