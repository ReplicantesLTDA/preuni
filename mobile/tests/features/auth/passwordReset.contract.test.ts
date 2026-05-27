import { fetchMock } from '../../lib/mockFetch';
import { buildAuthApi } from '../../lib/buildAuthApi';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('POST /v1/auth/password/reset/request + confirm', () => {
  it('requestPasswordReset accepts email and returns empty', async () => {
    let captured: unknown = null;
    fetchMock.on('POST', '/v1/auth/password/reset/request', {
      status: 200,
      bodyText: '',
      capture: ({ body }) => {
        captured = body;
      },
    });
    const { auth } = buildAuthApi();
    await auth.requestPasswordReset({ email: 'a@b.com' });
    expect(captured).toEqual({ email: 'a@b.com' });
  });

  it('confirmPasswordReset posts new credentials', async () => {
    let captured: unknown = null;
    fetchMock.on('POST', '/v1/auth/password/reset/confirm', {
      status: 200,
      bodyText: '',
      capture: ({ body }) => {
        captured = body;
      },
    });
    const { auth } = buildAuthApi();
    await auth.confirmPasswordReset({
      email: 'a@b.com',
      otp: '123456',
      newPassword: 'NovaSenha1',
    });
    expect(captured).toEqual({
      email: 'a@b.com',
      otp: '123456',
      new_password: 'NovaSenha1',
    });
  });

  it('confirmPasswordReset surfaces 422 weak password', async () => {
    fetchMock.on('POST', '/v1/auth/password/reset/confirm', {
      status: 422,
      body: { error: { field: 'newPassword', message: 'Senha fraca.' } },
    });
    const { auth } = buildAuthApi();
    await expect(
      auth.confirmPasswordReset({ email: 'a@b.com', otp: '123456', newPassword: 'x' }),
    ).rejects.toMatchObject({ kind: 'Validation', field: 'newPassword' });
  });
});
