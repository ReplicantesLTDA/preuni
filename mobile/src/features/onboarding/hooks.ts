import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/lib/api/context';
import { useSessionStore } from '@/stores/sessionStore';
import { qk } from '@/lib/query/keys';
import { makeOnboardingApi } from './api';

export function useCompleteOnboarding() {
  const api = useApi();
  const onboarding = makeOnboardingApi(api);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: onboarding.complete,
    onSuccess: (student) => {
      useSessionStore.getState().setAuthed(student);
      void qc.invalidateQueries({ queryKey: qk.studentMe() });
    },
  });
}
