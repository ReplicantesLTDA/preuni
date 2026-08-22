import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { createApiClient, type ApiClient } from './client';
import { createRefreshFn } from './refreshInterceptor';
import { tokenStore } from '@/lib/auth/tokenStore';
import { useSessionStore } from '@/stores/sessionStore';

const ApiContext = createContext<ApiClient | null>(null);

export function ApiProvider({
  baseUrl,
  children,
}: {
  baseUrl: string;
  children: ReactNode;
}) {
  const client = useMemo(() => {
    const refresh = createRefreshFn(baseUrl);
    return createApiClient({
      baseUrl,
      getAccessToken: () => tokenStore.getAccessToken(),
      refreshTokens: refresh,
      onUnauthorized: () => {
        void tokenStore.clear();
        useSessionStore.getState().setAnon();
      },
      logger:
        process.env.NODE_ENV === 'development'
          ? (entry) => {
              console.log(`[api] ${entry.method} ${entry.path} → ${entry.status} (${entry.durationMs}ms)`);
            }
          : undefined,
    });
  }, [baseUrl]);

  return <ApiContext.Provider value={client}>{children}</ApiContext.Provider>;
}

export function useApi(): ApiClient {
  const ctx = useContext(ApiContext);
  if (!ctx) throw new Error('useApi must be used within <ApiProvider>');
  return ctx;
}
