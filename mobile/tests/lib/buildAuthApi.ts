import { createApiClient } from '@/lib/api/client';
import { makeAuthApi } from '@/features/auth/api';
import { fetchMock } from './mockFetch';

export function buildAuthApi(opts?: {
  getAccessToken?: () => Promise<string | null>;
  refreshTokens?: () => Promise<boolean>;
  onUnauthorized?: () => void;
}) {
  const onUnauthorized = opts?.onUnauthorized ?? jest.fn();
  const refreshTokens = opts?.refreshTokens ?? jest.fn(async () => false);
  const getAccessToken = opts?.getAccessToken ?? jest.fn(async () => null);
  const client = createApiClient({
    baseUrl: fetchMock.baseUrl(),
    getAccessToken,
    refreshTokens,
    onUnauthorized,
  });
  return { auth: makeAuthApi(client), client, onUnauthorized, refreshTokens, getAccessToken };
}
