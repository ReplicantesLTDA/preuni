import { fetchMock } from '../../lib/mockFetch';
import { createApiClient } from '@/lib/api/client';
import { makeGamificationApi } from '@/features/gamification/api';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

function build() {
  const client = createApiClient({
    baseUrl: fetchMock.baseUrl(),
    getAccessToken: jest.fn(async () => 'tok'),
    refreshTokens: jest.fn(async () => false),
    onUnauthorized: jest.fn(),
  });
  return makeGamificationApi(client);
}

describe('Gamification API', () => {
  it('weeklyLeaderboard(): fetches the tier-scoped leaderboard', async () => {
    fetchMock.on('GET', '/v1/ranking/weekly?tier=bronze', {
      status: 200,
      body: [
        {
          user_id: '00000000-0000-4000-a000-000000000011',
          display_name: 'Aluno',
          weekly_score: 900,
          league_tier: 'bronze',
          rank_in_tier: 1,
        },
      ],
    });
    const api = build();
    const board = await api.weeklyLeaderboard('bronze');
    expect(board).toHaveLength(1);
    expect(board[0]?.leagueTier).toBe('bronze');
    expect(board[0]?.rankInTier).toBe(1);
  });

  it('myRanking(): parses the caller rank with no rank_in_tier yet (week in progress)', async () => {
    fetchMock.on('GET', '/v1/ranking/me', {
      status: 200,
      body: {
        user_id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Aluno',
        weekly_score: 400,
        league_tier: 'silver',
      },
    });
    const api = build();
    const mine = await api.myRanking();
    expect(mine.leagueTier).toBe('silver');
    expect(mine.rankInTier).toBeUndefined();
  });

  it('myMedals(): parses earned medals', async () => {
    fetchMock.on('GET', '/v1/medals/me', {
      status: 200,
      body: [
        { type: 'streak_7_day', earned_at: '2026-08-20T00:00:00Z' },
        { type: 'tier_promotion', earned_at: '2026-08-22T00:00:00Z' },
      ],
    });
    const api = build();
    const medals = await api.myMedals();
    expect(medals).toHaveLength(2);
    expect(medals[0]?.type).toBe('streak_7_day');
  });

  it('myMedals(): rejects an unknown medal type (schema drift guard)', async () => {
    fetchMock.on('GET', '/v1/medals/me', {
      status: 200,
      body: [{ type: 'not_a_real_medal', earned_at: '2026-08-22T00:00:00Z' }],
    });
    const api = build();
    await expect(api.myMedals()).rejects.toMatchObject({ kind: 'Unknown' });
  });
});
