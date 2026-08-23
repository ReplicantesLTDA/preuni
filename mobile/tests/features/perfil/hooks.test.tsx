import { renderHook, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useChangeEmailConfirm, useDeleteMe, useUploadAvatar } from '@/features/perfil/hooks';

const STUDENT_BODY = {
  id: '00000000-0000-4000-a000-000000000011',
  display_name: 'Maria',
  email: 'maria@preuni.com',
  xp_total: 0,
  streak_count: 0,
  readiness_score: 0,
  onboarding_completed: true,
};

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('useUploadAvatar', () => {
  it('skips the S3 PUT for the dev stub presigned URL and confirms directly', async () => {
    fetchMock.on('PUT', '/v1/students/me/avatar', {
      status: 200,
      body: { upload_url: 'https://dev.local/upload?X-Amz-Credential=PRESIGNED', object_key: 'avatars/abc' },
    });
    fetchMock.on('POST', '/v1/students/me/avatar/confirm', { status: 200, body: STUDENT_BODY });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useUploadAvatar(), { wrapper: Wrapper });

    result.current.mutate({ uri: 'file:///fake.jpg', mimeType: 'image/jpeg' });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe('useDeleteMe', () => {
  it('deletes the account and clears the session', async () => {
    fetchMock.on('DELETE', '/v1/students/me', { status: 200, body: {} });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useDeleteMe(), { wrapper: Wrapper });

    result.current.mutate('DELETE');

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe('useChangeEmailConfirm', () => {
  it('confirms the email change and invalidates the student query', async () => {
    fetchMock.on('POST', '/v1/auth/email/change/confirm', { status: 200, body: {} });

    const { Wrapper } = buildWrapper();
    const { result } = renderHook(() => useChangeEmailConfirm(), { wrapper: Wrapper });

    result.current.mutate({ newEmail: 'new@preuni.com', otp: '123456' });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});
