import { fetchMock } from '../../lib/mockFetch';
import { createApiClient } from '@/lib/api/client';
import { makeStreakApi } from '@/features/streak/api';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

function build() {
  const client = createApiClient({
    baseUrl: fetchMock.baseUrl(),
    getAccessToken: jest.fn(async () => 'tok'),
    refreshTokens: jest.fn(async () => false),
    onUnauthorized: jest.fn(),
  });
  return makeStreakApi(client);
}

describe('Streak API', () => {
  it('getMe(): parses current/longest streak and last active day', async () => {
    fetchMock.on('GET', '/v1/streaks/me', {
      status: 200,
      body: { current_streak: 5, longest_streak: 12, last_active_day: '2026-08-22' },
    });
    const api = build();
    const streak = await api.getMe();
    expect(streak.currentStreak).toBe(5);
    expect(streak.longestStreak).toBe(12);
    expect(streak.lastActiveDay).toBe('2026-08-22');
  });

  it('getMe(): accepts a null last_active_day (brand-new user)', async () => {
    fetchMock.on('GET', '/v1/streaks/me', {
      status: 200,
      body: { current_streak: 0, longest_streak: 0, last_active_day: null },
    });
    const api = build();
    const streak = await api.getMe();
    expect(streak.lastActiveDay).toBeNull();
  });
});
