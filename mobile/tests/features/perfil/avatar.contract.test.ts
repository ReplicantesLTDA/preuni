import { fetchMock } from '../../lib/mockFetch';
import { createApiClient } from '@/lib/api/client';
import { makePerfilApi } from '@/features/perfil/api';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

function build() {
  const client = createApiClient({
    baseUrl: fetchMock.baseUrl(),
    getAccessToken: jest.fn(async () => 'tok'),
    refreshTokens: jest.fn(async () => false),
    onUnauthorized: jest.fn(),
  });
  return makePerfilApi(client);
}

describe('PUT /v1/students/me/avatar — getAvatarUploadUrl', () => {
  it('returns uploadUrl + objectKey (snake_case round-trip)', async () => {
    fetchMock.on('PUT', '/v1/students/me/avatar', {
      status: 200,
      body: {
        upload_url: 'https://bucket.s3.us-east-1.amazonaws.com/avatars/u/123.webp?X-Amz=...',
        object_key: 'avatars/u/123.webp',
      },
    });
    const api = build();
    const url = await api.getAvatarUploadUrl();
    expect(url.uploadUrl).toMatch(/^https:\/\//);
    expect(url.objectKey).toBe('avatars/u/123.webp');
  });
});

describe('POST /v1/students/me/avatar/confirm — confirmAvatar', () => {
  it('posts snake_case object_key and returns updated Student', async () => {
    let captured: unknown = null;
    fetchMock.on('POST', '/v1/students/me/avatar/confirm', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000000',
        display_name: 'X',
        email: 'x@y.com',
        avatar_url: 'https://bucket.s3.us-east-1.amazonaws.com/avatars/u/123.webp',
        xp_total: 0,
        streak_count: 0,
        readiness_score: 0,
        onboarding_completed: true,
      },
      capture: ({ body }) => {
        captured = body;
      },
    });
    const api = build();
    const me = await api.confirmAvatar('avatars/u/123.webp');
    expect(captured).toEqual({ object_key: 'avatars/u/123.webp' });
    expect(me.avatarUrl).toMatch(/avatars\/u\/123\.webp$/);
  });
});
