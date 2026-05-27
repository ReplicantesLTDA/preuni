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

describe('PATCH /v1/students/me — updateMe', () => {
  it('posts snake_case body and returns Student', async () => {
    let captured: unknown = null;
    fetchMock.on('PATCH', '/v1/students/me', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000000',
        display_name: 'NovoNome',
        email: 'a@b.com',
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
    const me = await api.updateMe({ displayName: 'NovoNome' });
    expect(captured).toEqual({ display_name: 'NovoNome' });
    expect(me.displayName).toBe('NovoNome');
  });
});

describe('POST /v1/auth/password/change — changePassword', () => {
  it('posts snake_case current/new', async () => {
    let captured: unknown = null;
    fetchMock.on('POST', '/v1/auth/password/change', {
      status: 200,
      bodyText: '',
      capture: ({ body }) => {
        captured = body;
      },
    });
    const api = build();
    await api.changePassword({ currentPassword: 'Old1', newPassword: 'New12345' });
    expect(captured).toEqual({ current_password: 'Old1', new_password: 'New12345' });
  });

  it('surfaces 422 on weak password', async () => {
    fetchMock.on('POST', '/v1/auth/password/change', {
      status: 422,
      body: { error: { field: 'new_password', message: 'fraca' } },
    });
    const api = build();
    await expect(
      api.changePassword({ currentPassword: 'Old1', newPassword: 'x' }),
    ).rejects.toMatchObject({ kind: 'Validation' });
  });
});

describe('POST /v1/auth/email/change/{request,confirm} — changeEmail', () => {
  it('request posts new_email', async () => {
    let captured: unknown = null;
    fetchMock.on('POST', '/v1/auth/email/change/request', {
      status: 200,
      bodyText: '',
      capture: ({ body }) => {
        captured = body;
      },
    });
    const api = build();
    await api.changeEmailRequest({ newEmail: 'new@b.com' });
    expect(captured).toEqual({ new_email: 'new@b.com' });
  });

  it('confirm posts new_email + otp', async () => {
    let captured: unknown = null;
    fetchMock.on('POST', '/v1/auth/email/change/confirm', {
      status: 200,
      bodyText: '',
      capture: ({ body }) => {
        captured = body;
      },
    });
    const api = build();
    await api.changeEmailConfirm({ newEmail: 'new@b.com', otp: '123456' });
    expect(captured).toEqual({ new_email: 'new@b.com', otp: '123456' });
  });
});

describe('DELETE /v1/students/me — deleteMe', () => {
  it('posts confirmation body', async () => {
    let captured: unknown = null;
    fetchMock.on('DELETE', '/v1/students/me', {
      status: 200,
      bodyText: '',
      capture: ({ body }) => {
        captured = body;
      },
    });
    const api = build();
    await api.deleteMe('DELETE');
    expect(captured).toEqual({ confirmation: 'DELETE' });
  });
});
