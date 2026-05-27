import type { z } from 'zod';
import { parseApiErrorEnvelope } from '@/types/api';
import { toAppError, networkError, type AppError } from './errors';
import { toCamelCase, toSnakeCase } from './caseConvert';

export type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';

export interface ApiRequestInput<T> {
  method: HttpMethod;
  path: string;
  body?: unknown;
  schema: z.ZodType<T>;
  auth?: boolean;
  signal?: AbortSignal;
  headers?: Record<string, string>;
}

export interface ApiClient {
  request<T>(input: ApiRequestInput<T>): Promise<T>;
}

export interface CreateApiClientOptions {
  baseUrl: string;
  getAccessToken: () => Promise<string | null>;
  refreshTokens: () => Promise<boolean>;
  onUnauthorized: () => void;
  logger?: (entry: { method: HttpMethod; path: string; status: number; durationMs: number }) => void;
}

export function createApiClient(opts: CreateApiClientOptions): ApiClient {
  let refreshing: Promise<boolean> | null = null;

  function joinUrl(path: string): string {
    if (path.startsWith('http://') || path.startsWith('https://')) return path;
    const base = opts.baseUrl.replace(/\/$/, '');
    const tail = path.startsWith('/') ? path : `/${path}`;
    return `${base}${tail}`;
  }

  async function execute<T>(input: ApiRequestInput<T>, retried: boolean): Promise<T> {
    const started = Date.now();
    const url = joinUrl(input.path);
    const headers: Record<string, string> = {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...(input.headers ?? {}),
    };
    if (input.auth !== false) {
      const token = await opts.getAccessToken();
      if (token) headers.Authorization = `Bearer ${token}`;
    }

    let response: Response;
    try {
      response = await fetch(url, {
        method: input.method,
        headers,
        body: input.body !== undefined ? JSON.stringify(toSnakeCase(input.body)) : undefined,
        signal: input.signal,
      });
    } catch (err) {
      const reason = err instanceof Error ? err.message : undefined;
      throw networkError(reason);
    }
    const durationMs = Date.now() - started;
    opts.logger?.({ method: input.method, path: input.path, status: response.status, durationMs });

    if (response.status === 401 && input.auth !== false && !retried) {
      const refreshed = await singleFlightRefresh();
      if (refreshed) return execute(input, true);
      opts.onUnauthorized();
      throw toAppError(401, null);
    }

    const text = await response.text();
    if (response.ok) {
      const raw = text.length > 0 ? (JSON.parse(text) as unknown) : {};
      const json = toCamelCase(raw);
      const parsed = input.schema.safeParse(json);
      if (!parsed.success) {
        throw {
          kind: 'Unknown',
          message: 'Resposta do servidor em formato inesperado.',
        } satisfies AppError;
      }
      return parsed.data;
    }

    const envelope = parseApiErrorEnvelope(text);
    throw toAppError(response.status, envelope);
  }

  async function singleFlightRefresh(): Promise<boolean> {
    if (!refreshing) {
      refreshing = opts
        .refreshTokens()
        .catch(() => false)
        .finally(() => {
          refreshing = null;
        });
    }
    return refreshing;
  }

  return {
    request: (input) => execute(input, false),
  };
}
