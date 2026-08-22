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
import { createRefreshFn } from '@/lib/api/refreshInterceptor';
import { tokenStore } from '@/lib/auth/tokenStore';

beforeAll(() => fetchMock.install());
afterEach(async () => {
  fetchMock.reset();
  await tokenStore.clear();
});

describe('createRefreshFn', () => {
  it('returns false when no refresh token is stored', async () => {
    const refresh = createRefreshFn(fetchMock.baseUrl());
    expect(await refresh()).toBe(false);
  });

  it('persists new tokens on 200', async () => {
    await tokenStore.save({ accessToken: 'old-a', refreshToken: 'old-r' });
    fetchMock.on('POST', '/v1/auth/refresh', {
      status: 200,
      body: {
        accessToken: 'new-a',
        refreshToken: 'new-r',
        accessTokenExpiresAt: '2030-01-01T00:00:00.000Z',
      },
    });
    const refresh = createRefreshFn(fetchMock.baseUrl());
    expect(await refresh()).toBe(true);
    expect(await tokenStore.getAccessToken()).toBe('new-a');
    expect(await tokenStore.getRefreshToken()).toBe('new-r');
  });

  it('clears tokens on non-2xx', async () => {
    await tokenStore.save({ accessToken: 'old-a', refreshToken: 'old-r' });
    fetchMock.on('POST', '/v1/auth/refresh', {
      status: 401,
      body: { error: { message: 'refresh expired' } },
    });
    const refresh = createRefreshFn(fetchMock.baseUrl());
    expect(await refresh()).toBe(false);
    expect(await tokenStore.getAccessToken()).toBeNull();
    expect(await tokenStore.getRefreshToken()).toBeNull();
  });

  it('clears tokens on malformed response body', async () => {
    await tokenStore.save({ accessToken: 'old-a', refreshToken: 'old-r' });
    fetchMock.on('POST', '/v1/auth/refresh', {
      status: 200,
      body: { foo: 'bar' },
    });
    const refresh = createRefreshFn(fetchMock.baseUrl());
    expect(await refresh()).toBe(false);
    expect(await tokenStore.getAccessToken()).toBeNull();
  });
});
