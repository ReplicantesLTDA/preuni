import { fetchMock } from '../../lib/mockFetch';
import { buildAuthApi } from '../../lib/buildAuthApi';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('POST /v1/auth/login', () => {
  it('returns a Session on 200', async () => {
    fetchMock.on('POST', '/v1/auth/login', {
      status: 200,
      body: { accessToken: 'a', refreshToken: 'b' },
    });
    const { auth } = buildAuthApi();
    const session = await auth.login({ email: 'a@b.com', password: 'Senha1234' });
    expect(session).toEqual({ accessToken: 'a', refreshToken: 'b' });
  });

  it('maps 401 invalid credentials to AppError.Unauthorized', async () => {
    fetchMock.on('POST', '/v1/auth/login', {
      status: 401,
      body: { error: { code: 'invalid_credentials', message: 'E-mail ou senha incorretos.' } },
    });
    // refresh never fires on a non-auth endpoint
    const { auth } = buildAuthApi({ refreshTokens: jest.fn(async () => true) });
    await expect(
      auth.login({ email: 'a@b.com', password: 'wrong' }),
    ).rejects.toMatchObject({ kind: 'Unauthorized' });
  });

  it('maps 422 validation to AppError.Validation with field', async () => {
    fetchMock.on('POST', '/v1/auth/login', {
      status: 422,
      body: { error: { field: 'email', message: 'inválido' } },
    });
    const { auth } = buildAuthApi();
    await expect(
      auth.login({ email: 'not-an-email', password: 'x' }),
    ).rejects.toMatchObject({ kind: 'Validation', field: 'email' });
  });
});
