import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/lib/api/context';
import { qk } from '@/lib/query/keys';
import { makeEssayApi } from './api';

export function useEssayApi() {
  const api = useApi();
  return makeEssayApi(api);
}

export function useSubmitEssay() {
  const api = useEssayApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.submit,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.essayList() });
      void qc.invalidateQueries({ queryKey: qk.streakMe() });
    },
  });
}

/** Polls the submission until it leaves `pending` (submit-then-poll, per
 * Constitution Principle V: correction is async, never blocking). */
export function useEssay(id: string | undefined) {
  const api = useEssayApi();
  return useQuery({
    queryKey: qk.essayDetail(id ?? ''),
    queryFn: () => api.get(id as string),
    enabled: Boolean(id),
    refetchInterval: (query) => (query.state.data?.status === 'pending' ? 3000 : false),
  });
}

export function useEssayList() {
  const api = useEssayApi();
  return useQuery({
    queryKey: qk.essayList(),
    queryFn: api.list,
  });
}
