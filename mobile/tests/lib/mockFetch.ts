interface RouteHandler {
  status: number;
  body?: unknown;
  bodyText?: string;
  capture?: (request: { method: string; headers: Headers; body?: unknown; url: string }) => void;
}

interface MockState {
  routes: Map<string, RouteHandler[]>;
  baseUrl: string;
  installed: boolean;
}

const state: MockState = {
  routes: new Map(),
  baseUrl: 'http://api.test',
  installed: false,
};

function key(method: string, path: string): string {
  return `${method.toUpperCase()} ${state.baseUrl}${path}`;
}

function mockFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const url = typeof input === 'string' ? input : (input as URL).toString();
  const method = (init?.method ?? 'GET').toUpperCase();
  const k = `${method} ${url}`;
  const handlers = state.routes.get(k);
  if (!handlers || handlers.length === 0) {
    return Promise.reject(new Error(`unhandled ${k}`));
  }
  const handler = handlers.length > 1 ? handlers.shift()! : handlers[0]!;
  if (handler.capture) {
    const headers = new Headers(init?.headers as HeadersInit);
    const body = typeof init?.body === 'string' ? JSON.parse(init.body as string) : init?.body;
    handler.capture({ method, headers, body, url });
  }
  const text = handler.bodyText ?? (handler.body !== undefined ? JSON.stringify(handler.body) : '');
  return Promise.resolve(
    new Response(text, {
      status: handler.status,
      headers: { 'content-type': 'application/json' },
    }),
  );
}

export const fetchMock = {
  baseUrl: (): string => state.baseUrl,
  install(baseUrl = 'http://api.test'): void {
    state.baseUrl = baseUrl;
    state.routes.clear();
    if (!state.installed) {
      globalThis.fetch = mockFetch as unknown as typeof fetch;
      state.installed = true;
    }
  },
  reset(): void {
    state.routes.clear();
  },
  on(method: string, path: string, handler: RouteHandler): void {
    const k = key(method, path);
    const list = state.routes.get(k) ?? [];
    list.push(handler);
    state.routes.set(k, list);
  },
};
