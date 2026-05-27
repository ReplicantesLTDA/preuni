import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/lib/api/context';
import { useSessionStore } from '@/stores/sessionStore';
import { qk } from '@/lib/query/keys';
import { makeTrilhaApi } from './api';
import type { Student } from '@/types/student';

export function useTrilhaHome() {
  const api = useApi();
  const trilha = makeTrilhaApi(api);
  const cached = useSessionStore((s) => s.student);
  const status = useSessionStore((s) => s.status);

  return useQuery<Student>({
    queryKey: qk.studentMe(),
    queryFn: async () => {
      const me = await trilha.getMe();
      useSessionStore.getState().setAuthed(me);
      return me;
    },
    enabled: status === 'authed',
    initialData: cached ?? undefined,
    staleTime: 30_000,
  });
}

export function useRefreshTrilha() {
  const qc = useQueryClient();
  return () => qc.invalidateQueries({ queryKey: qk.studentMe() });
}
