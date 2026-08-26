import { renderHook } from '@testing-library/react-native';
import { useGoogleIdTokenRequest } from '@/features/auth/useGoogleIdTokenRequest';

const ENV_KEYS = [
  'EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID',
  'EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID',
  'EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID',
] as const;

const originalEnv: Record<string, string | undefined> = {};
beforeAll(() => {
  for (const key of ENV_KEYS) originalEnv[key] = process.env[key];
});
afterEach(() => {
  for (const key of ENV_KEYS) delete process.env[key];
});
afterAll(() => {
  for (const key of ENV_KEYS) {
    if (originalEnv[key] !== undefined) process.env[key] = originalEnv[key];
  }
});

describe('useGoogleIdTokenRequest', () => {
  it('reports not configured when no client ID env vars are set', () => {
    const { result } = renderHook(() => useGoogleIdTokenRequest());
    expect(result.current.configured).toBe(false);
  });

  it('reports configured when only the web client ID is set', () => {
    process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID = 'web-client-id';
    const { result } = renderHook(() => useGoogleIdTokenRequest());
    expect(result.current.configured).toBe(true);
  });

  it('reports configured when only the iOS client ID is set', () => {
    process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID = 'ios-client-id';
    const { result } = renderHook(() => useGoogleIdTokenRequest());
    expect(result.current.configured).toBe(true);
  });

  it('reports configured when only the Android client ID is set', () => {
    process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID = 'android-client-id';
    const { result } = renderHook(() => useGoogleIdTokenRequest());
    expect(result.current.configured).toBe(true);
  });

  it('returns the underlying request/response/promptAsync from expo-auth-session', () => {
    const { result } = renderHook(() => useGoogleIdTokenRequest());
    expect(result.current.request).toBeNull();
    expect(result.current.response).toBeNull();
    expect(typeof result.current.promptAsync).toBe('function');
  });
});
