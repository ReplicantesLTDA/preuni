import { renderHook, waitFor, act } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useFriends, useSendFriendRequest, useAcceptFriendRequest, useRemoveFriend } from '@/features/social/hooks';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('useFriends', () => {
  it('fetches the friend list', async () => {
    fetchMock.on('GET', '/v1/friends', {
      status: 200,
      body: [
        {
          friendship_id: '00000000-0000-4000-a000-000000000010',
          user_id: '00000000-0000-4000-a000-000000000011',
          display_name: 'Amigo',
          current_streak: 2,
        },
      ],
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useFriends(), { wrapper: Wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
  });
});

describe('useSendFriendRequest', () => {
  it('sends a request', async () => {
    fetchMock.on('POST', '/v1/friends/requests', {
      status: 201,
      body: { id: '00000000-0000-4000-a000-000000000020', status: 'pending' },
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useSendFriendRequest(), { wrapper: Wrapper });

    act(() => {
      result.current.mutate('00000000-0000-4000-a000-000000000011');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe('useAcceptFriendRequest', () => {
  it('accepts a request', async () => {
    fetchMock.on('POST', '/v1/friends/requests/00000000-0000-4000-a000-000000000020/accept', {
      status: 200,
      body: { id: '00000000-0000-4000-a000-000000000020', status: 'accepted' },
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useAcceptFriendRequest(), { wrapper: Wrapper });

    act(() => {
      result.current.mutate('00000000-0000-4000-a000-000000000020');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe('useRemoveFriend', () => {
  it('removes a friend and invalidates the list', async () => {
    fetchMock.on('DELETE', '/v1/friends/00000000-0000-4000-a000-000000000010', {
      status: 200,
      body: {},
    });
    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useRemoveFriend(), { wrapper: Wrapper });

    act(() => {
      result.current.mutate('00000000-0000-4000-a000-000000000010');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});
