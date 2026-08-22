import type { ApiClient } from '@/lib/api/client';
import { StreakSchema, type Streak } from '@/types/essay';

export function makeStreakApi(api: ApiClient) {
  return {
    getMe(): Promise<Streak> {
      return api.request({
        method: 'GET',
        path: '/v1/streaks/me',
        schema: StreakSchema,
      });
    },
  };
}

export type StreakApi = ReturnType<typeof makeStreakApi>;
