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
  useSessionStore.setState({ status: 'authed', student: null });
});

describe('EditProfileScreen', () => {
  it('rejects a too-short display name without submitting', () => {
    const { getByLabelText, getByText } = render(<EditProfileScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('Nome'), 'A');
    fireEvent.press(getByText('Salvar'));

    expect(getByText('Mínimo 2 caracteres.')).toBeTruthy();
    expect(mockBack).not.toHaveBeenCalled();
  });

  it('saves and navigates back on success', async () => {
    fetchMock.on('PATCH', '/v1/students/me', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Maria Nova',
        email: 'maria@preuni.com',
        xp_total: 0,
        streak_count: 0,
        readiness_score: 0,
        onboarding_completed: true,
      },
    });

    const { getByLabelText, getByText } = render(<EditProfileScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('Nome'), 'Maria Nova');
    fireEvent.press(getByText('Salvar'));

    await waitFor(() => expect(mockBack).toHaveBeenCalled());
  });

  it('cancels without saving', () => {
    const { getByText } = render(<EditProfileScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.press(getByText('Cancelar'));
    expect(mockBack).toHaveBeenCalled();
  });
});
