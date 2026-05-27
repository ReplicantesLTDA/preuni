import { z } from 'zod';
import { createApiClient } from '@/lib/api/client';
import { isAppError } from '@/lib/api/errors';

const PingSchema = z.object({ ok: z.boolean() });
const baseUrl = 'http://api.test';

interface RouteHandler {
  status: number;
  body?: unknown;
  bodyText?: string;
  capture?: (request: { method: string; headers: Headers; body?: unknown }) => void;
}

interface MockState {
  routes: Map<string, RouteHandler[]>;
}

const state: MockState = { routes: new Map() };

function reset() {
  state.routes.clear();
}

function on(method: string, path: string, handler: RouteHandler) {
  const key = `${method.toUpperCase()} ${baseUrl}${path}`;
  const list = state.routes.get(key) ?? [];
  list.push(handler);
  state.routes.set(key, list);
}

function mockFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const url = typeof input === 'string' ? input : (input as URL).toString();
  const method = (init?.method ?? 'GET').toUpperCase();
  const key = `${method} ${url}`;
  const handlers = state.routes.get(key);
  if (!handlers || handlers.length === 0) {
    return Promise.reject(new Error(`unhandled ${key}`));
  }
  const handler = handlers.length > 1 ? handlers.shift()! : handlers[0]!;
  if (handler.capture) {
    const headers = new Headers(init?.headers as HeadersInit);
    const body = typeof init?.body === 'string' ? JSON.parse(init.body as string) : init?.body;
    handler.capture({ method, headers, body });
  }
  const text = handler.bodyText ?? (handler.body !== undefined ? JSON.stringify(handler.body) : '');
  return Promise.resolve(
    new Response(text, {
      status: handler.status,
      headers: { 'content-type': 'application/json' },
    }),
  );
}

beforeAll(() => {
  globalThis.fetch = mockFetch as unknown as typeof fetch;
});
afterEach(() => reset());

function buildClient(overrides: Partial<Parameters<typeof createApiClient>[0]> = {}) {
  const resolved = {
    onUnauthorized: jest.fn(),
    refreshTokens: jest.fn(async () => false),
    getAccessToken: jest.fn(async () => 'tok'),
    ...overrides,
  };
  const client = createApiClient({ baseUrl, ...resolved });
  return { client, ...resolved };
}

describe('apiClient.request — success', () => {
  it('returns parsed body on 200', async () => {
    on('GET', '/v1/ping', { status: 200, body: { ok: true } });
    const { client } = buildClient();
    const out = await client.request({ method: 'GET', path: '/v1/ping', schema: PingSchema });
    expect(out).toEqual({ ok: true });
  });

  it('attaches Bearer when auth is on', async () => {
    let seen: string | null = null;
    on('GET', '/v1/ping', {
      status: 200,
      body: { ok: true },
      capture: ({ headers }) => {
        seen = headers.get('authorization');
      },
    });
    const { client } = buildClient();
    await client.request({ method: 'GET', path: '/v1/ping', schema: PingSchema });
    expect(seen).toBe('Bearer tok');
  });
});

describe('apiClient.request — refresh flow', () => {
  it('on 401 refresh succeeds → retries and returns 200', async () => {
    on('GET', '/v1/me', { status: 401, body: { error: { message: 'expired' } } });
    on('GET', '/v1/me', { status: 200, body: { ok: true } });
    const { client, onUnauthorized, refreshTokens } = buildClient({
      refreshTokens: jest.fn(async () => true),
    });
    const out = await client.request({ method: 'GET', path: '/v1/me', schema: PingSchema });
    expect(out).toEqual({ ok: true });
    expect(refreshTokens).toHaveBeenCalledTimes(1);
    expect(onUnauthorized).not.toHaveBeenCalled();
  });

  it('on 401 refresh fails → onUnauthorized called and Unauthorized thrown', async () => {
    on('GET', '/v1/me', { status: 401, body: { error: { message: 'expired' } } });
    const { client, onUnauthorized } = buildClient({ refreshTokens: jest.fn(async () => false) });
    await expect(
      client.request({ method: 'GET', path: '/v1/me', schema: PingSchema }),
    ).rejects.toMatchObject({ kind: 'Unauthorized' });
    expect(onUnauthorized).toHaveBeenCalledTimes(1);
  });
});

describe('apiClient.request — error mapping', () => {
  it('422 → Validation with field + message', async () => {
    on('POST', '/v1/things', {
      status: 422,
      body: { error: { field: 'email', message: 'inválido' } },
    });
    const { client } = buildClient();
    try {
      await client.request({
        method: 'POST',
        path: '/v1/things',
        body: { x: 1 },
        schema: PingSchema,
      });
      throw new Error('should have thrown');
    } catch (err) {
      expect(isAppError(err)).toBe(true);
      expect(err).toEqual({ kind: 'Validation', field: 'email', message: 'inválido' });
    }
  });

  it('500 with unparseable body → Unknown fallback', async () => {
    on('GET', '/v1/x', { status: 500, bodyText: 'boom' });
    const { client } = buildClient();
    await expect(
      client.request({ method: 'GET', path: '/v1/x', schema: PingSchema }),
    ).rejects.toMatchObject({ kind: 'Unknown' });
  });
});
