import { authErrorByCode, authErrorMessage } from '@/features/auth/errorMessages';

describe('authErrorMessage', () => {
  it('returns the message for a Validation error', () => {
    expect(authErrorMessage({ kind: 'Validation', field: 'email', message: 'Informe um e-mail válido.' })).toBe(
      'Informe um e-mail válido.',
    );
  });

  it('returns the message for a Conflict error', () => {
    expect(authErrorMessage({ kind: 'Conflict', message: 'Esse e-mail já está em uso.' })).toBe(
      'Esse e-mail já está em uso.',
    );
  });

  it('returns the message for a RateLimited error', () => {
    expect(authErrorMessage({ kind: 'RateLimited', message: 'Muitas tentativas.' })).toBe('Muitas tentativas.');
  });

  it('returns the message for any other AppError kind', () => {
    expect(authErrorMessage({ kind: 'Unauthorized', message: 'Sessão expirada.' })).toBe('Sessão expirada.');
  });

  it('falls back to a generic message for a non-AppError value', () => {
    expect(authErrorMessage(new Error('boom'))).toBe('Algo deu errado. Tente novamente.');
    expect(authErrorMessage('a plain string')).toBe('Algo deu errado. Tente novamente.');
    expect(authErrorMessage(null)).toBe('Algo deu errado. Tente novamente.');
  });
});

describe('authErrorByCode', () => {
  it('returns undefined for an undefined code', () => {
    expect(authErrorByCode(undefined)).toBeUndefined();
  });

  it('returns undefined for an unmapped code', () => {
    expect(authErrorByCode('something_unmapped')).toBeUndefined();
  });

  it('returns the pt-BR message for a known code', () => {
    expect(authErrorByCode('invalid_credentials')).toBe('E-mail ou senha incorretos.');
    expect(authErrorByCode('email_already_taken')).toBe('Esse e-mail já está em uso.');
  });
});
