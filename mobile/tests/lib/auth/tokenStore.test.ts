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

import { Platform } from 'react-native';
import { tokenStore } from '@/lib/auth/tokenStore';

describe('tokenStore (native / SecureStore)', () => {
  afterEach(async () => {
    await tokenStore.clear();
  });

  it('returns null from load() when nothing is stored', async () => {
    expect(await tokenStore.load()).toBeNull();
  });

  it('saves and loads a full session', async () => {
    await tokenStore.save({ accessToken: 'a1', refreshToken: 'r1', expiresAt: '2026-01-01T00:00:00Z' });
    const loaded = await tokenStore.load();
    expect(loaded).toEqual({ accessToken: 'a1', refreshToken: 'r1', expiresAt: '2026-01-01T00:00:00Z' });
    expect(await tokenStore.getAccessToken()).toBe('a1');
    expect(await tokenStore.getRefreshToken()).toBe('r1');
  });

  it('removes the expiry key when expiresAt is omitted', async () => {
    await tokenStore.save({ accessToken: 'a2', refreshToken: 'r2', expiresAt: '2026-01-01T00:00:00Z' });
    await tokenStore.save({ accessToken: 'a3', refreshToken: 'r3' });
    const loaded = await tokenStore.load();
    expect(loaded?.expiresAt).toBeNull();
  });

  it('clear() removes everything', async () => {
    await tokenStore.save({ accessToken: 'a4', refreshToken: 'r4' });
    await tokenStore.clear();
    expect(await tokenStore.load()).toBeNull();
  });
});

describe('tokenStore (web / localStorage)', () => {
  const originalOS = Platform.OS;
  let store: Map<string, string>;

  beforeEach(() => {
    Platform.OS = 'web';
    store = new Map();
    // @ts-expect-error -- RN test env has no DOM localStorage; provide a minimal fake
    global.localStorage = {
      getItem: (k: string) => store.get(k) ?? null,
      setItem: (k: string, v: string) => store.set(k, v),
      removeItem: (k: string) => store.delete(k),
    };
  });

  afterEach(() => {
    Platform.OS = originalOS;
    // @ts-expect-error -- cleanup
    delete global.localStorage;
  });

  it('saves and loads via localStorage on web', async () => {
    await tokenStore.save({ accessToken: 'wa', refreshToken: 'wr', expiresAt: '2026-01-01T00:00:00Z' });
    const loaded = await tokenStore.load();
    expect(loaded).toEqual({ accessToken: 'wa', refreshToken: 'wr', expiresAt: '2026-01-01T00:00:00Z' });
  });

  it('clear() removes web-stored tokens', async () => {
    await tokenStore.save({ accessToken: 'wa2', refreshToken: 'wr2' });
    await tokenStore.clear();
    expect(await tokenStore.load()).toBeNull();
  });
});
