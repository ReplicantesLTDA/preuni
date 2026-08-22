import { fetchMock } from '../../lib/mockFetch';
import { buildAuthApi } from '../../lib/buildAuthApi';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('POST /v1/auth/register', () => {
  it('returns a Session shape and posts credentials JSON', async () => {
    let captured: { body?: unknown; url?: string; method?: string } = {};
    fetchMock.on('POST', '/v1/auth/register', {
      status: 200,
      body: {
        accessToken: 'a.b.c',
        refreshToken: 'rrr',
        accessTokenExpiresAt: '2030-01-01T00:00:00.000Z',
      },
      capture: ({ body, url, method }) => {
        captured = { body, url, method };
      },
    });

    const { auth } = buildAuthApi();
    const session = await auth.register({
      email: 'a@b.com',
      password: 'Senha1234',
      displayName: 'Aluno',
    });

    expect(session.accessToken).toBe('a.b.c');
    expect(session.refreshToken).toBe('rrr');
    expect(captured.method).toBe('POST');
    expect(captured.url).toMatch(/\/v1\/auth\/register$/);
    expect(captured.body).toEqual({
      email: 'a@b.com',
      password: 'Senha1234',
      display_name: 'Aluno',
    });
  });

  it('maps 409 EMAIL_TAKEN to AppError.Conflict', async () => {
    fetchMock.on('POST', '/v1/auth/register', {
      status: 409,
      body: { error: { code: 'EMAIL_TAKEN', message: 'Esse e-mail já está em uso.' } },
    });
    const { auth } = buildAuthApi();
    await expect(
      auth.register({ email: 'a@b.com', password: 'Senha1234', displayName: 'X' }),
    ).rejects.toMatchObject({ kind: 'Conflict', message: 'Esse e-mail já está em uso.' });
  });

  it('does not attach Authorization header (public endpoint)', async () => {
    let auth: string | null = '__unset__';
    fetchMock.on('POST', '/v1/auth/register', {
      status: 200,
      body: { accessToken: 'a', refreshToken: 'b' },
      capture: ({ headers }) => {
        auth = headers.get('authorization');
      },
    });
    const { auth: api } = buildAuthApi({ getAccessToken: async () => 'should-not-be-used' });
    await api.register({ email: 'a@b.com', password: 'Senha1234', displayName: 'X' });
    expect(auth).toBeNull();
  });
});
