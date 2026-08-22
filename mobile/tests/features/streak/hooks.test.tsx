import { renderHook, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useStreak } from '@/features/streak/hooks';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('useStreak', () => {
  it('fetches and exposes the current streak', async () => {
    fetchMock.on('GET', '/v1/streaks/me', {
      status: 200,
      body: { current_streak: 4, longest_streak: 9, last_active_day: '2026-08-22' },
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useStreak(), { wrapper: Wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.currentStreak).toBe(4);
  });
});
