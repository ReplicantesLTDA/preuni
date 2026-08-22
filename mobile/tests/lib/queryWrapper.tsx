import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { ApiProvider } from '@/lib/api/context';
import { ThemeProvider } from '@/theme';
import { ToastProvider } from '@/components/Toast';
import { fetchMock } from './mockFetch';

export function buildQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: 0, gcTime: 0, staleTime: 0 },
      mutations: { retry: 0 },
    },
  });
}

export function buildWrapper(client = buildQueryClient()) {
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>
        <ThemeProvider>
          <ApiProvider baseUrl={fetchMock.baseUrl()}>
            <ToastProvider>{children}</ToastProvider>
          </ApiProvider>
        </ThemeProvider>
      </QueryClientProvider>
    );
  }
  return { Wrapper, client };
}
