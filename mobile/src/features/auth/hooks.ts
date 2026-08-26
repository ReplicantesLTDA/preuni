import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/lib/api/context';
import { tokenStore } from '@/lib/auth/tokenStore';
import { useSessionStore } from '@/stores/sessionStore';
import { qk } from '@/lib/query/keys';
import { makeAuthApi } from './api';
import type { Session } from '@/types/auth';
import type { Student } from '@/types/student';

async function persistSession(
  session: Session,
  fetchMe: () => Promise<Student>,
  student?: Student | null,
) {
  await tokenStore.save({
    accessToken: session.accessToken,
    refreshToken: session.refreshToken,
    expiresAt: session.accessTokenExpiresAt ?? null,
  });
  const stu = student ?? session.student ?? null;
  if (stu) {
    useSessionStore.getState().setAuthed(stu);
    return;
  }
  try {
    const me = await fetchMe();
    useSessionStore.getState().setAuthed(me);
  } catch {
    // bootstrap will pick this up on next mount
  }
}

export function useAuthApi() {
  const api = useApi();
  return makeAuthApi(api);
}

export function useStudentMe() {
  const auth = useAuthApi();
  return useQuery({
    queryKey: qk.studentMe(),
    queryFn: () => auth.getMe(),
    enabled: useSessionStore((s) => s.status === 'authed'),
  });
}

export function useRegister() {
  const auth = useAuthApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: auth.register,
    onSuccess: async (session) => {
      await persistSession(session, () => auth.getMe());
      await qc.invalidateQueries({ queryKey: qk.studentMe() });
    },
  });
}

export function useLogin() {
  const auth = useAuthApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: auth.login,
    onSuccess: async (session) => {
      await persistSession(session, () => auth.getMe());
      await qc.invalidateQueries({ queryKey: qk.studentMe() });
    },
  });
}

export function useGoogleLogin() {
  const auth = useAuthApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (idToken: string) => auth.googleLogin({ idToken }),
    onSuccess: async (session) => {
      await persistSession(session, () => auth.getMe());
      await qc.invalidateQueries({ queryKey: qk.studentMe() });
    },
  });
}

export function useLogout() {
  const auth = useAuthApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => auth.logout().catch(() => undefined),
    onSettled: async () => {
      await tokenStore.clear();
      useSessionStore.getState().setAnon();
      qc.clear();
    },
  });
}

export function useVerifyEmail() {
  const auth = useAuthApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: auth.verifyEmail,
    onSuccess: async () => {
      // post-verify: re-fetch me to update emailVerified / route to onboarding or tabs
      await qc.invalidateQueries({ queryKey: qk.studentMe() });
      try {
        const me = await auth.getMe();
        useSessionStore.getState().setAuthed(me);
      } catch {
        // ignore — bootstrap will recover
      }
    },
  });
}

export function useResendVerificationOtp() {
  const auth = useAuthApi();
  return useMutation({ mutationFn: auth.resendVerificationOtp });
}

export function useOtpLoginRequest() {
  const auth = useAuthApi();
  return useMutation({ mutationFn: auth.requestOtpLogin });
}

export function useOtpLoginVerify() {
  const auth = useAuthApi();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: auth.verifyOtpLogin,
    onSuccess: async (session) => {
      await persistSession(session, () => auth.getMe());
      await qc.invalidateQueries({ queryKey: qk.studentMe() });
    },
  });
}

export function usePasswordResetRequest() {
  const auth = useAuthApi();
  return useMutation({ mutationFn: auth.requestPasswordReset });
}

export function usePasswordResetConfirm() {
  const auth = useAuthApi();
  return useMutation({ mutationFn: auth.confirmPasswordReset });
}
