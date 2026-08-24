import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useSessionStore } from '@/stores/sessionStore';

const mockPush = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

const PerfilHome = require('../../../app/(tabs)/perfil/index').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockPush.mockClear();
  useSessionStore.setState({ status: 'authed', student: null });
});

describe('PerfilHome', () => {
  it('shows a fallback when there is no student loaded yet', () => {
    const { getAllByText } = render(<PerfilHome />, { wrapper: buildWrapper().Wrapper });
    expect(getAllByText('—').length).toBeGreaterThan(0);
  });

  it('shows the student profile summary', () => {
    useSessionStore.setState({
      status: 'authed',
      student: {
        id: '1',
        displayName: 'Maria',
        email: 'maria@preuni.com',
        xpTotal: 120,
        streakCount: 5,
        readinessScore: 42.6,
        onboardingCompleted: true,
      },
    });

    const { getByText } = render(<PerfilHome />, { wrapper: buildWrapper().Wrapper });
    expect(getByText('Maria')).toBeTruthy();
    expect(getByText('maria@preuni.com')).toBeTruthy();
    expect(getByText('120')).toBeTruthy();
    expect(getByText('5')).toBeTruthy();
    expect(getByText('43')).toBeTruthy();
  });

  it('navigates to each menu destination', () => {
    const { getByLabelText } = render(<PerfilHome />, { wrapper: buildWrapper().Wrapper });

    fireEvent.press(getByLabelText('Amigos'));
    expect(mockPush).toHaveBeenCalledWith('/(tabs)/perfil/friends');

    fireEvent.press(getByLabelText('Ranking e medalhas'));
    expect(mockPush).toHaveBeenCalledWith('/(tabs)/perfil/ranking');

    fireEvent.press(getByLabelText('Editar perfil'));
    expect(mockPush).toHaveBeenCalledWith('/(tabs)/perfil/edit');

    fireEvent.press(getByLabelText('Alterar e-mail'));
    expect(mockPush).toHaveBeenCalledWith('/(tabs)/perfil/change-email');

    fireEvent.press(getByLabelText('Alterar senha'));
    expect(mockPush).toHaveBeenCalledWith('/(tabs)/perfil/change-password');

    fireEvent.press(getByLabelText('Exportar meus dados'));
    expect(mockPush).toHaveBeenCalledWith('/(tabs)/perfil/data-export');

    fireEvent.press(getByLabelText('Excluir conta'));
    expect(mockPush).toHaveBeenCalledWith('/(tabs)/perfil/delete-account');
  });

  it('logs out and clears the session on sign-out', async () => {
    fetchMock.on('POST', '/v1/auth/logout', { status: 200, body: {} });
    useSessionStore.setState({
      status: 'authed',
      student: {
        id: '1',
        displayName: 'Maria',
        email: 'maria@preuni.com',
        xpTotal: 0,
        streakCount: 0,
        readinessScore: 0,
        onboardingCompleted: true,
      },
    });

    const { getByText } = render(<PerfilHome />, { wrapper: buildWrapper().Wrapper });
    fireEvent.press(getByText('Sair'));

    await waitFor(() => expect(useSessionStore.getState().status).toBe('anon'));
  });
});
