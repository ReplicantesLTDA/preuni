import { fetchMock } from '../../lib/mockFetch';
import { buildAuthApi } from '../../lib/buildAuthApi';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('POST /v1/auth/email/verify + verify-resend', () => {
  it('verifyEmail returns void on 204-style empty 200', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify', { status: 200, bodyText: '' });
    const { auth } = buildAuthApi();
    await expect(auth.verifyEmail({ email: 'a@b.com', otp: '123456' })).resolves.toEqual({});
  });

  it('verifyEmail surfaces 422 invalid OTP', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify', {
      status: 422,
      body: { error: { field: 'otp', message: 'Código inválido.' } },
    });
    const { auth } = buildAuthApi();
    await expect(
      auth.verifyEmail({ email: 'a@b.com', otp: '000000' }),
    ).rejects.toMatchObject({ kind: 'Validation', field: 'otp' });
  });

  it('resendVerificationOtp returns void on success', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify-resend', { status: 200, bodyText: '' });
    const { auth } = buildAuthApi();
    await expect(auth.resendVerificationOtp({ email: 'a@b.com' })).resolves.toEqual({});
  });

  it('resendVerificationOtp surfaces 409 already-verified', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify-resend', {
      status: 409,
      body: { error: { code: 'email_already_verified', message: 'Seu e-mail já foi verificado.' } },
    });
    const { auth } = buildAuthApi();
    await expect(
      auth.resendVerificationOtp({ email: 'a@b.com' }),
    ).rejects.toMatchObject({ kind: 'Conflict' });
  });

  it('resendVerificationOtp surfaces 429 rate-limit', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify-resend', {
      status: 429,
      body: { error: { message: 'Aguarde antes de reenviar.' } },
    });
    const { auth } = buildAuthApi();
    await expect(
      auth.resendVerificationOtp({ email: 'a@b.com' }),
    ).rejects.toMatchObject({ kind: 'RateLimited' });
  });
});
