import { renderHook, waitFor, act } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useSubmitEssay, useEssay, useEssayList } from '@/features/essay/hooks';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('useSubmitEssay', () => {
  it('submits and returns the pending submission', async () => {
    fetchMock.on('POST', '/v1/essays', {
      status: 202,
      body: { id: '00000000-0000-4000-a000-000000000001', status: 'pending', submitted_at: '2026-08-22T12:00:00Z' },
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useSubmitEssay(), { wrapper: Wrapper });

    act(() => {
      result.current.mutate({
        promptThemeTitle: 'Tema',
        promptThemeContext: 'Contexto',
        essayText: 'Texto',
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.status).toBe('pending');
  });
});

describe('useEssay', () => {
  it('fetches a graded submission by id', async () => {
    fetchMock.on('GET', '/v1/essays/00000000-0000-4000-a000-000000000001', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000001',
        status: 'graded',
        submitted_at: '2026-08-22T12:00:00Z',
        grade: { overall_score: 800, competencies: [], graded_at: '2026-08-22T12:05:00Z' },
      },
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useEssay('00000000-0000-4000-a000-000000000001'), { wrapper: Wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.status).toBe('graded');
  });

  it('does not fetch when id is undefined', () => {
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useEssay(undefined), { wrapper: Wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useEssayList', () => {
  it('fetches the caller submission list', async () => {
    fetchMock.on('GET', '/v1/essays', {
      status: 200,
      body: [{ id: '00000000-0000-4000-a000-000000000001', status: 'graded', submitted_at: '2026-08-22T12:00:00Z' }],
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useEssayList(), { wrapper: Wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
  });
});
