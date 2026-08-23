import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockReplace = jest.fn();
const mockPush = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace, push: mockPush }),
}));

const LoginScreen = require('../../../app/(auth)/login').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockReplace.mockClear();
  mockPush.mockClear();
});

describe('LoginScreen', () => {
  it('shows a validation error for an invalid email and does not submit', () => {
    const { getByLabelText, getByText, getByRole } = render(<LoginScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('E-mail'), 'not-an-email');
    fireEvent.press(getByRole('button', { name: 'Entrar' }));

    expect(getByText('Informe um e-mail válido.')).toBeTruthy();
  });

  it('logs in and navigates to the trilha tab on success', async () => {
    fetchMock.on('POST', '/v1/auth/login', {
      status: 200,
      body: { accessToken: 'a', refreshToken: 'b' },
    });
    fetchMock.on('GET', '/v1/students/me', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Maria',
        email: 'maria@preuni.com',
        xp_total: 0,
        streak_count: 0,
        readiness_score: 0,
        onboarding_completed: true,
      },
    });

    const { getByLabelText, getByRole } = render(<LoginScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.changeText(getByLabelText('Senha'), 'Senha1234');
    fireEvent.press(getByRole('button', { name: 'Entrar' }));

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/(tabs)/trilha'));
  });

  it('shows a toast on invalid credentials', async () => {
    fetchMock.on('POST', '/v1/auth/login', {
      status: 401,
      body: { error: { code: 'invalid_credentials', message: 'E-mail ou senha incorretos.' } },
    });

    const { getByLabelText, findByText, getByRole } = render(<LoginScreen />, {
      wrapper: buildWrapper().Wrapper,
    });

    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.changeText(getByLabelText('Senha'), 'wrongpass');
    fireEvent.press(getByRole('button', { name: 'Entrar' }));

    expect(await findByText('E-mail ou senha incorretos.')).toBeTruthy();
    expect(mockReplace).not.toHaveBeenCalled();
  });

  it('navigates to password reset and register screens', () => {
    const { getByText } = render(<LoginScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.press(getByText('Esqueci minha senha'));
    expect(mockPush).toHaveBeenCalledWith('/(auth)/password-reset');

    fireEvent.press(getByText('Entrar com código'));
    expect(mockPush).toHaveBeenCalledWith('/(auth)/otp-login');

    fireEvent.press(getByText('Criar conta'));
    expect(mockPush).toHaveBeenCalledWith('/(auth)/register');
  });
});
