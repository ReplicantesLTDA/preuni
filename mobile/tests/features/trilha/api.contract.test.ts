import { fetchMock } from '../../lib/mockFetch';
import { createApiClient } from '@/lib/api/client';
import { makeTrilhaApi } from '@/features/trilha/api';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

function build() {
  const client = createApiClient({
    baseUrl: fetchMock.baseUrl(),
    getAccessToken: jest.fn(async () => 'tok'),
    refreshTokens: jest.fn(async () => false),
    onUnauthorized: jest.fn(),
  });
  return makeTrilhaApi(client);
}

describe('Trilha — GET /v1/students/me', () => {
  it('returns Student-shaped response (snake_case from backend → camelCase via case-converter)', async () => {
    fetchMock.on('GET', '/v1/students/me', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000000',
        display_name: 'Aluno',
        email: 'aluno@preuni.com',
        username: null,
        avatar_url: null,
        xp_total: 1250,
        streak_count: 7,
        readiness_score: 62.5,
        onboarding_completed: true,
        email_verified: true,
      },
    });
    const api = build();
    const me = await api.getMe();
    expect(me.displayName).toBe('Aluno');
    expect(me.xpTotal).toBe(1250);
    expect(me.streakCount).toBe(7);
    expect(me.readinessScore).toBe(62.5);
    expect(me.onboardingCompleted).toBe(true);
  });

  it('attaches Bearer token', async () => {
    let auth: string | null = null;
    fetchMock.on('GET', '/v1/students/me', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000000',
        display_name: 'X',
        email: 'x@y.com',
        xp_total: 0,
        streak_count: 0,
        readiness_score: 0,
        onboarding_completed: false,
      },
      capture: ({ headers }) => {
        auth = headers.get('authorization');
      },
    });
    const api = build();
    await api.getMe();
    expect(auth).toBe('Bearer tok');
  });

  it('rejects malformed response with Unknown', async () => {
    fetchMock.on('GET', '/v1/students/me', {
      status: 200,
      body: { whatever: 'nope' },
    });
    const api = build();
    await expect(api.getMe()).rejects.toMatchObject({ kind: 'Unknown' });
  });
});
