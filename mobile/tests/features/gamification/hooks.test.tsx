import { renderHook, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useWeeklyLeaderboard, useMyRanking, useMyMedals } from '@/features/gamification/hooks';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('useWeeklyLeaderboard', () => {
  it('fetches the tier leaderboard', async () => {
    fetchMock.on('GET', '/v1/ranking/weekly?tier=gold', {
      status: 200,
      body: [
        {
          user_id: '00000000-0000-4000-a000-000000000011',
          display_name: 'Aluno',
          weekly_score: 500,
          league_tier: 'gold',
        },
      ],
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useWeeklyLeaderboard('gold'), { wrapper: Wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
  });
});

describe('useMyRanking', () => {
  it('fetches the caller ranking', async () => {
    fetchMock.on('GET', '/v1/ranking/me', {
      status: 200,
      body: {
        user_id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Aluno',
        weekly_score: 300,
        league_tier: 'bronze',
      },
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useMyRanking(), { wrapper: Wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.leagueTier).toBe('bronze');
  });
});

describe('useMyMedals', () => {
  it('fetches earned medals', async () => {
    fetchMock.on('GET', '/v1/medals/me', {
      status: 200,
      body: [{ type: 'streak_7_day', earned_at: '2026-08-20T00:00:00Z' }],
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useMyMedals(), { wrapper: Wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
  });
});
