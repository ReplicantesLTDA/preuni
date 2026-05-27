import type { ApiErrorEnvelope } from '@/types/api';

export type AppError =
  | { kind: 'Unauthorized'; message: string }
  | { kind: 'Forbidden'; message: string }
  | { kind: 'Conflict'; message: string }
  | { kind: 'Validation'; field: string; message: string }
  | { kind: 'NotFound'; message: string }
  | { kind: 'RateLimited'; message: string }
  | { kind: 'Network'; message: string }
  | { kind: 'Unknown'; message: string };

const DEFAULT_MESSAGES = {
  Unauthorized: 'Sessão expirada. Entre novamente.',
  Forbidden: 'Acesso negado.',
  Conflict: 'Conflito ao processar a solicitação.',
  Validation: 'Verifique os dados informados.',
  NotFound: 'Não encontrado.',
  RateLimited: 'Muitas tentativas. Aguarde alguns instantes.',
  Network: 'Sem conexão. Verifique sua internet.',
  Unknown: 'Algo deu errado. Tente novamente.',
} as const;

export function toAppError(
  status: number,
  envelope: ApiErrorEnvelope | null,
): AppError {
  const message = envelope?.error.message;
  const field = envelope?.error.field;

  switch (status) {
    case 401:
      return { kind: 'Unauthorized', message: message ?? DEFAULT_MESSAGES.Unauthorized };
    case 403:
      return { kind: 'Forbidden', message: message ?? DEFAULT_MESSAGES.Forbidden };
    case 404:
      return { kind: 'NotFound', message: message ?? DEFAULT_MESSAGES.NotFound };
    case 409:
      return { kind: 'Conflict', message: message ?? DEFAULT_MESSAGES.Conflict };
    case 422:
      return {
        kind: 'Validation',
        field: field ?? 'unknown',
        message: message ?? DEFAULT_MESSAGES.Validation,
      };
    case 429:
      return { kind: 'RateLimited', message: message ?? DEFAULT_MESSAGES.RateLimited };
    default:
      return { kind: 'Unknown', message: message ?? DEFAULT_MESSAGES.Unknown };
  }
}

export function networkError(reason?: string): AppError {
  return { kind: 'Network', message: reason ?? DEFAULT_MESSAGES.Network };
}

export function isAppError(value: unknown): value is AppError {
  return (
    typeof value === 'object' &&
    value !== null &&
    'kind' in value &&
    typeof (value as { kind: unknown }).kind === 'string'
  );
}
