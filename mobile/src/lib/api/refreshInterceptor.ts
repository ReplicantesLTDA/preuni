import { z } from 'zod';
import { tokenStore } from '@/lib/auth/tokenStore';

const RefreshResponseSchema = z.object({
  accessToken: z.string(),
  refreshToken: z.string(),
  accessTokenExpiresAt: z.string().datetime().optional(),
});

export function createRefreshFn(baseUrl: string) {
  return async function refreshTokens(): Promise<boolean> {
    const refreshToken = await tokenStore.getRefreshToken();
    if (!refreshToken) return false;

    try {
      const resp = await fetch(`${baseUrl.replace(/\/$/, '')}/v1/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify({ refreshToken }),
      });
      if (!resp.ok) {
        await tokenStore.clear();
        return false;
      }
      const json = (await resp.json()) as unknown;
      const parsed = RefreshResponseSchema.safeParse(json);
      if (!parsed.success) {
        await tokenStore.clear();
        return false;
      }
      await tokenStore.save({
        accessToken: parsed.data.accessToken,
        refreshToken: parsed.data.refreshToken,
        expiresAt: parsed.data.accessTokenExpiresAt ?? null,
      });
      return true;
    } catch {
      await tokenStore.clear();
      return false;
    }
  };
}
