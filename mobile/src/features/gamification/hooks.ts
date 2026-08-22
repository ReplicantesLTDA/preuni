import { useQuery } from '@tanstack/react-query';
import { useApi } from '@/lib/api/context';
import { qk } from '@/lib/query/keys';
import { makeGamificationApi } from './api';

export function useGamificationApi() {
  const api = useApi();
  return makeGamificationApi(api);
}

export function useWeeklyLeaderboard(tier: string) {
  const api = useGamificationApi();
  return useQuery({
    queryKey: qk.rankingWeekly(tier),
    queryFn: () => api.weeklyLeaderboard(tier),
  });
}

export function useMyRanking() {
  const api = useGamificationApi();
  return useQuery({
    queryKey: qk.rankingMe(),
    queryFn: api.myRanking,
  });
}

export function useMyMedals() {
  const api = useGamificationApi();
  return useQuery({
    queryKey: qk.medalsMe(),
    queryFn: api.myMedals,
  });
}
