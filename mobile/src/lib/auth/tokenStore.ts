import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

const ACCESS_KEY = 'preuni.access_token';
const REFRESH_KEY = 'preuni.refresh_token';
const EXPIRES_KEY = 'preuni.access_token_expires_at';

type Stored = { accessToken: string; refreshToken: string; expiresAt: string | null };

const webStore = {
  getItem: (k: string) =>
    typeof localStorage !== 'undefined' ? localStorage.getItem(k) : null,
  setItem: (k: string, v: string) => {
    if (typeof localStorage !== 'undefined') localStorage.setItem(k, v);
  },
  removeItem: (k: string) => {
    if (typeof localStorage !== 'undefined') localStorage.removeItem(k);
  },
};

async function get(key: string): Promise<string | null> {
  if (Platform.OS === 'web') return webStore.getItem(key);
  return SecureStore.getItemAsync(key);
}

async function set(key: string, value: string): Promise<void> {
  if (Platform.OS === 'web') {
    webStore.setItem(key, value);
    return;
  }
  await SecureStore.setItemAsync(key, value);
}

async function remove(key: string): Promise<void> {
  if (Platform.OS === 'web') {
    webStore.removeItem(key);
    return;
  }
  await SecureStore.deleteItemAsync(key);
}

export const tokenStore = {
  async load(): Promise<Stored | null> {
    const [accessToken, refreshToken, expiresAt] = await Promise.all([
      get(ACCESS_KEY),
      get(REFRESH_KEY),
      get(EXPIRES_KEY),
    ]);
    if (!accessToken || !refreshToken) return null;
    return { accessToken, refreshToken, expiresAt };
  },

  async save(input: { accessToken: string; refreshToken: string; expiresAt?: string | null }) {
    await Promise.all([
      set(ACCESS_KEY, input.accessToken),
      set(REFRESH_KEY, input.refreshToken),
      input.expiresAt ? set(EXPIRES_KEY, input.expiresAt) : remove(EXPIRES_KEY),
    ]);
  },

  async clear() {
    await Promise.all([remove(ACCESS_KEY), remove(REFRESH_KEY), remove(EXPIRES_KEY)]);
  },

  async getAccessToken(): Promise<string | null> {
    return get(ACCESS_KEY);
  },

  async getRefreshToken(): Promise<string | null> {
    return get(REFRESH_KEY);
  },
};

export type TokenStore = typeof tokenStore;
