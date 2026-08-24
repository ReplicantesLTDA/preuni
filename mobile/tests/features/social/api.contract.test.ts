import { fetchMock } from '../../lib/mockFetch';
import { createApiClient } from '@/lib/api/client';
import { makeSocialApi } from '@/features/social/api';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

function build() {
  const client = createApiClient({
    baseUrl: fetchMock.baseUrl(),
    getAccessToken: jest.fn(async () => 'tok'),
    refreshTokens: jest.fn(async () => false),
    onUnauthorized: jest.fn(),
  });
  return makeSocialApi(client);
}

describe('Social API', () => {
  it('listFriends(): parses a friend with streak + latest grade', async () => {
    fetchMock.on('GET', '/v1/friends', {
      status: 200,
      body: [
        {
          friendship_id: '00000000-0000-4000-a000-000000000010',
          user_id: '00000000-0000-4000-a000-000000000011',
          display_name: 'Amigo',
          current_streak: 3,
          latest_overall_score: 720,
          latest_graded_at: '2026-08-22T12:00:00Z',
        },
      ],
    });
    const api = build();
    const friends = await api.listFriends();
    expect(friends).toHaveLength(1);
    expect(friends[0]?.friendshipId).toBe('00000000-0000-4000-a000-000000000010');
    expect(friends[0]?.displayName).toBe('Amigo');
    expect(friends[0]?.latestOverallScore).toBe(720);
  });

  it('listFriends(): a friend with no grade yet omits latest_overall_score', async () => {
    fetchMock.on('GET', '/v1/friends', {
      status: 200,
      body: [
        {
          friendship_id: '00000000-0000-4000-a000-000000000010',
          user_id: '00000000-0000-4000-a000-000000000011',
          display_name: 'Amigo',
          current_streak: 0,
        },
      ],
    });
    const api = build();
    const friends = await api.listFriends();
    expect(friends[0]?.latestOverallScore).toBeUndefined();
  });

  it('sendRequest(): posts addresseeId as snake_case and parses the pending request', async () => {
    let captured: unknown;
    fetchMock.on('POST', '/v1/friends/requests', {
      status: 201,
      body: { id: '00000000-0000-4000-a000-000000000020', status: 'pending' },
      capture: ({ body }) => {
        captured = body;
      },
    });
    const api = build();
    const req = await api.sendRequest('00000000-0000-4000-a000-000000000011');
    expect(req.status).toBe('pending');
    expect(captured).toMatchObject({ addressee_id: '00000000-0000-4000-a000-000000000011' });
  });

  it('acceptRequest(): parses the accepted status', async () => {
    fetchMock.on('POST', '/v1/friends/requests/00000000-0000-4000-a000-000000000020/accept', {
      status: 200,
      body: { id: '00000000-0000-4000-a000-000000000020', status: 'accepted' },
    });
    const api = build();
    const req = await api.acceptRequest('00000000-0000-4000-a000-000000000020');
    expect(req.status).toBe('accepted');
  });

  it('removeFriend(): sends DELETE and tolerates an empty 204-style body', async () => {
    fetchMock.on('DELETE', '/v1/friends/00000000-0000-4000-a000-000000000010', {
      status: 200,
      body: {},
    });
    const api = build();
    await expect(api.removeFriend('00000000-0000-4000-a000-000000000010')).resolves.toEqual({});
  });
});
