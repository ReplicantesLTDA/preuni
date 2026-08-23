import { QueryClient } from '@tanstack/react-query';
import { createQueryClient } from '@/lib/query/client';

describe('createQueryClient', () => {
  it('builds a QueryClient with the configured defaults', () => {
    const client = createQueryClient();
    expect(client).toBeInstanceOf(QueryClient);
    const defaults = client.getDefaultOptions();
    expect(defaults.queries?.retry).toBe(1);
    expect(defaults.queries?.staleTime).toBe(30_000);
    expect(defaults.queries?.refetchOnWindowFocus).toBe(false);
    expect(defaults.mutations?.retry).toBe(0);
  });
});
