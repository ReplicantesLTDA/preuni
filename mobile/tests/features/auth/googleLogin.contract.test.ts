import { fetchMock } from '../../lib/mockFetch';
import { buildAuthApi } from '../../lib/buildAuthApi';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('POST /v1/auth/google', () => {
  it('returns a Session on 200', async () => {
    fetchMock.on('POST', '/v1/auth/google', {
      status: 200,
      body: { accessToken: 'a', refreshToken: 'b' },
    });
    const { auth } = buildAuthApi();
    const session = await auth.googleLogin({ idToken: 'fake-google-id-token' });
    expect(session).toEqual({ accessToken: 'a', refreshToken: 'b' });
  });

  it('sends id_token in the request body', async () => {
    let capturedBody: unknown;
    fetchMock.on('POST', '/v1/auth/google', {
      status: 200,
      body: { accessToken: 'a', refreshToken: 'b' },
      capture: (req) => {
        capturedBody = req.body;
      },
    });
    const { auth } = buildAuthApi();
    await auth.googleLogin({ idToken: 'fake-google-id-token' });
    expect(capturedBody).toEqual({ id_token: 'fake-google-id-token' });
  });

  it('maps 401 invalid token to AppError.Unauthorized', async () => {
    fetchMock.on('POST', '/v1/auth/google', {
      status: 401,
      body: { error: { code: 'invalid_credentials', message: 'invalid Google ID token' } },
    });
    const { auth } = buildAuthApi({ refreshTokens: jest.fn(async () => true) });
    await expect(
      auth.googleLogin({ idToken: 'bad-token' }),
    ).rejects.toMatchObject({ kind: 'Unauthorized' });
  });
});
