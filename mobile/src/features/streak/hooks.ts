import { useQuery } from '@tanstack/react-query';
import { useApi } from '@/lib/api/context';
import { qk } from '@/lib/query/keys';
import { makeStreakApi } from './api';

export function useStreakApi() {
  const api = useApi();
  return makeStreakApi(api);
}

export function useStreak() {
  const api = useStreakApi();
  return useQuery({
    queryKey: qk.streakMe(),
    queryFn: api.getMe,
  });
}
