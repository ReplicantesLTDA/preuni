import type { AppError } from '@/lib/api/errors';

const PT_BR: Record<string, string> = {
  invalid_credentials: 'E-mail ou senha incorretos.',
  email_already_taken: 'Esse e-mail já está em uso.',
  email_already_verified: 'Esse e-mail já foi verificado.',
  otp_expired: 'O código expirou. Solicite um novo.',
  otp_invalid: 'Código inválido.',
  password_weak: 'Escolha uma senha mais forte.',
};

export function authErrorMessage(err: AppError | unknown): string {
  if (typeof err === 'object' && err !== null && 'kind' in err) {
    const e = err as AppError;
    if (e.kind === 'Validation') return e.message;
    if (e.kind === 'Conflict') return e.message;
    if (e.kind === 'RateLimited') return e.message;
    return e.message;
  }
  return 'Algo deu errado. Tente novamente.';
}

export function authErrorByCode(code: string | undefined): string | undefined {
  if (!code) return undefined;
  return PT_BR[code];
}
