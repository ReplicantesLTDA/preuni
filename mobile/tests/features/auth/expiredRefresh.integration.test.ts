jest.mock('expo-secure-store', () => {
  const store = new Map<string, string>();
  return {
    __store: store,
    getItemAsync: jest.fn(async (k: string) => store.get(k) ?? null),
    setItemAsync: jest.fn(async (k: string, v: string) => {
      store.set(k, v);
    }),
    deleteItemAsync: jest.fn(async (k: string) => {
      store.delete(k);
    }),
  };
});

import { fetchMock } from '../../lib/mockFetch';
import { createApiClient } from '@/lib/api/client';
import { createRefreshFn } from '@/lib/api/refreshInterceptor';
import { tokenStore } from '@/lib/auth/tokenStore';
import { useSessionStore } from '@/stores/sessionStore';
import { bootstrapSession } from '@/lib/auth/bootstrap';

function buildClient(onUnauthorized: () => void) {
  return createApiClient({
    baseUrl: fetchMock.baseUrl(),
    getAccessToken: () => tokenStore.getAccessToken(),
    refreshTokens: createRefreshFn(fetchMock.baseUrl()),
    onUnauthorized,
  });
}

beforeAll(() => fetchMock.install());
beforeEach(() => {
  useSessionStore.setState({ status: 'loading', student: null });
});
afterEach(async () => {
  fetchMock.reset();
  await tokenStore.clear();
});

describe('expired refresh token → anon (FR-020)', () => {
  it('bootstrap with no tokens lands in anon', async () => {
    const api = buildClient(jest.fn());
    await bootstrapSession(api);
    expect(useSessionStore.getState().status).toBe('anon');
  });

  it('bootstrap with stored tokens whose access + refresh both fail → anon, tokens cleared', async () => {
    await tokenStore.save({ accessToken: 'access', refreshToken: 'refresh' });
    // GET /v1/students/me → 401 (access dead), POST /v1/auth/refresh → 401 (refresh dead)
    fetchMock.on('GET', '/v1/students/me', {
      status: 401,
      body: { error: { message: 'expired' } },
    });
    fetchMock.on('POST', '/v1/auth/refresh', {
      status: 401,
      body: { error: { message: 'refresh expired' } },
    });

    const onUnauthorized = jest.fn(() => {
      // mirrors ApiProvider's wiring
      void tokenStore.clear();
      useSessionStore.getState().setAnon();
    });
    const api = buildClient(onUnauthorized);
    await bootstrapSession(api);

    expect(useSessionStore.getState().status).toBe('anon');
    expect(await tokenStore.getAccessToken()).toBeNull();
    expect(await tokenStore.getRefreshToken()).toBeNull();
    expect(onUnauthorized).toHaveBeenCalled();
  });

  it('bootstrap with valid access transitions to authed and loads student', async () => {
    await tokenStore.save({ accessToken: 'good', refreshToken: 'good-r' });
    fetchMock.on('GET', '/v1/students/me', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000000',
        email: 'admin@preuni.com',
        displayName: 'Admin',
        username: 'admin',
        avatarUrl: null,
        xpTotal: 0,
        streakCount: 0,
        readinessScore: 100,
        onboardingCompleted: true,
      },
    });
    const api = buildClient(jest.fn());
    await bootstrapSession(api);

    const s = useSessionStore.getState();
    expect(s.status).toBe('authed');
    expect(s.student?.email).toBe('admin@preuni.com');
    expect(s.student?.onboardingCompleted).toBe(true);
  });
});
