import { toAppError, isAppError, networkError } from '@/lib/api/errors';

describe('toAppError', () => {
  it('maps 401 → Unauthorized with envelope message', () => {
    const err = toAppError(401, { error: { message: 'token inválido' } });
    expect(err).toEqual({ kind: 'Unauthorized', message: 'token inválido' });
  });

  it('maps 403 → Forbidden with default message when envelope missing', () => {
    const err = toAppError(403, null);
    expect(err.kind).toBe('Forbidden');
    expect(err.message).toMatch(/[Aa]cesso/);
  });

  it('maps 409 → Conflict', () => {
    const err = toAppError(409, { error: { code: 'EMAIL_TAKEN', message: 'já existe' } });
    expect(err).toEqual({ kind: 'Conflict', message: 'já existe' });
  });

  it('maps 422 → Validation with field passthrough', () => {
    const err = toAppError(422, { error: { field: 'email', message: 'inválido' } });
    expect(err).toEqual({ kind: 'Validation', field: 'email', message: 'inválido' });
  });

  it('maps 429 → RateLimited', () => {
    const err = toAppError(429, null);
    expect(err.kind).toBe('RateLimited');
  });

  it('falls back to Unknown for other statuses', () => {
    const err = toAppError(500, null);
    expect(err.kind).toBe('Unknown');
    expect(err.message).toMatch(/[Aa]lgo/);
  });
});

describe('networkError + isAppError', () => {
  it('builds Network kind', () => {
    expect(networkError('econn')).toEqual({ kind: 'Network', message: 'econn' });
  });

  it('detects AppError shapes', () => {
    expect(isAppError({ kind: 'Conflict', message: 'x' })).toBe(true);
    expect(isAppError({ message: 'x' })).toBe(false);
    expect(isAppError(null)).toBe(false);
  });
});
