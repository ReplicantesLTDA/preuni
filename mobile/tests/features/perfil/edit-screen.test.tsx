import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useSessionStore } from '@/stores/sessionStore';

const mockBack = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ back: mockBack }),
}));

const EditProfileScreen = require('../../../app/(tabs)/perfil/edit').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockBack.mockClear();
});

describe('EditProfileScreen', () => {
  it('saves a new display name', async () => {
    useSessionStore.getState().setAuthed({
      id: '00000000-0000-4000-a000-000000000011',
      email: 'aluno@preuni.com',
      displayName: 'Aluno',
      username: null,
      avatarUrl: null,
      xpTotal: 0,
      streakCount: 0,
      readinessScore: 0,
      onboardingCompleted: true,
    });

    let captured: unknown;
    fetchMock.on('PATCH', '/v1/students/me', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000011',
        email: 'aluno@preuni.com',
        display_name: 'Novo Nome',
        xp_total: 0,
        streak_count: 0,
        readiness_score: 0,
        onboarding_completed: true,
      },
      capture: ({ body }) => {
        captured = body;
      },
    });

    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText } = render(<EditProfileScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Nome'), 'Novo Nome');
    fireEvent.press(getByText('Salvar'));

    await waitFor(() => expect(mockBack).toHaveBeenCalled());
    expect(captured).toMatchObject({ display_name: 'Novo Nome' });
  });
});
