import { renderHook, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useGoogleLogin, useLogin, useLogout, useStudentMe, useVerifyEmail } from '@/features/auth/hooks';
import { useSessionStore } from '@/stores/sessionStore';

const STUDENT_BODY = {
  id: '00000000-0000-4000-a000-000000000011',
  display_name: 'Maria',
  email: 'maria@preuni.com',
  xp_total: 0,
  streak_count: 0,
  readiness_score: 0,
  onboarding_completed: true,
};

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  useSessionStore.setState({ status: 'loading', student: null });
});

describe('persistSession (via useLogin)', () => {
  it('uses the student embedded in the session response without an extra getMe call', async () => {
    fetchMock.on('POST', '/v1/auth/login', {
      status: 200,
      body: { access_token: 'a', refresh_token: 'b', student: STUDENT_BODY },
    });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useLogin(), { wrapper: Wrapper });

    result.current.mutate({ email: 'maria@preuni.com', password: 'Senha1234' });

    await waitFor(() => expect(useSessionStore.getState().status).toBe('authed'));
    expect(useSessionStore.getState().student?.displayName).toBe('Maria');
  });
});

describe('persistSession (via useGoogleLogin)', () => {
  it('uses the student embedded in the session response without an extra getMe call', async () => {
    fetchMock.on('POST', '/v1/auth/google', {
      status: 200,
      body: { access_token: 'a', refresh_token: 'b', student: STUDENT_BODY },
    });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useGoogleLogin(), { wrapper: Wrapper });

    result.current.mutate('fake-google-id-token');

    await waitFor(() => expect(useSessionStore.getState().status).toBe('authed'));
    expect(useSessionStore.getState().student?.displayName).toBe('Maria');
  });
});

describe('useLogout', () => {
  it('clears the token store and session on success', async () => {
    fetchMock.on('POST', '/v1/auth/logout', { status: 200, body: {} });
    useSessionStore.setState({ status: 'authed', student: null });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useLogout(), { wrapper: Wrapper });

    result.current.mutate();

    await waitFor(() => expect(useSessionStore.getState().status).toBe('anon'));
  });

  it('still clears the session even if the logout request fails', async () => {
    fetchMock.on('POST', '/v1/auth/logout', {
      status: 500,
      bodyText: '{"error":{"code":"INTERNAL_ERROR","message":"boom"}}',
    });
    useSessionStore.setState({ status: 'authed', student: null });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useLogout(), { wrapper: Wrapper });

    result.current.mutate();

    await waitFor(() => expect(useSessionStore.getState().status).toBe('anon'));
  });
});

describe('useStudentMe', () => {
  it('is disabled while the session is not authed', () => {
    useSessionStore.setState({ status: 'anon', student: null });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useStudentMe(), { wrapper: Wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches the student once authed', async () => {
    useSessionStore.setState({ status: 'authed', student: null });
    fetchMock.on('GET', '/v1/students/me', { status: 200, body: STUDENT_BODY });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useStudentMe(), { wrapper: Wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.displayName).toBe('Maria');
  });
});

describe('useVerifyEmail', () => {
  it('re-fetches and authenticates the session on success', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify', { status: 200, body: {} });
    fetchMock.on('GET', '/v1/students/me', { status: 200, body: STUDENT_BODY });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useVerifyEmail(), { wrapper: Wrapper });

    result.current.mutate({ email: 'maria@preuni.com', otp: '123456' });

    await waitFor(() => expect(useSessionStore.getState().status).toBe('authed'));
  });

  it('does not throw when the post-verify getMe call fails', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify', { status: 200, body: {} });
    fetchMock.on('GET', '/v1/students/me', { status: 500, bodyText: '{"error":{"code":"INTERNAL_ERROR","message":"boom"}}' });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useVerifyEmail(), { wrapper: Wrapper });

    result.current.mutate({ email: 'maria@preuni.com', otp: '123456' });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(useSessionStore.getState().status).not.toBe('authed');
  });
});
