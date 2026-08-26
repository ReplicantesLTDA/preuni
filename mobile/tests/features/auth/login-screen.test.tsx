import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockReplace = jest.fn();
const mockPush = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace, push: mockPush }),
}));

const mockUseGoogleIdTokenRequest = jest.fn();
jest.mock('@/features/auth/useGoogleIdTokenRequest', () => ({
  useGoogleIdTokenRequest: () => mockUseGoogleIdTokenRequest(),
}));

const LoginScreen = require('../../../app/(auth)/login').default;

beforeAll(() => fetchMock.install());
beforeEach(() => {
  mockUseGoogleIdTokenRequest.mockReturnValue({
    request: null,
    response: null,
    promptAsync: jest.fn(),
    configured: false,
  });
});
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

  it('shows a field error for a server-side validation failure', async () => {
    fetchMock.on('POST', '/v1/auth/login', {
      status: 422,
      body: { error: { field: 'email', message: 'Esse e-mail parece inválido.' } },
    });

    const { getByLabelText, findByText, getByRole } = render(<LoginScreen />, {
      wrapper: buildWrapper().Wrapper,
    });

    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.changeText(getByLabelText('Senha'), 'Senha1234');
    fireEvent.press(getByRole('button', { name: 'Entrar' }));

    expect(await findByText('Esse e-mail parece inválido.')).toBeTruthy();
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

  it('hides the Google button when Google Sign-In is not configured', () => {
    const { queryByText } = render(<LoginScreen />, { wrapper: buildWrapper().Wrapper });
    expect(queryByText('Continuar com Google')).toBeNull();
  });

  it('prompts Google sign-in when the button is pressed', () => {
    const promptAsync = jest.fn();
    mockUseGoogleIdTokenRequest.mockReturnValue({
      request: {},
      response: null,
      promptAsync,
      configured: true,
    });

    const { getByText } = render(<LoginScreen />, { wrapper: buildWrapper().Wrapper });
    fireEvent.press(getByText('Continuar com Google'));

    expect(promptAsync).toHaveBeenCalled();
  });

  it('logs in and navigates to the trilha tab on a successful Google response', async () => {
    fetchMock.on('POST', '/v1/auth/google', {
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
    mockUseGoogleIdTokenRequest.mockReturnValue({
      request: {},
      response: { type: 'success', params: { id_token: 'fake-id-token' } },
      promptAsync: jest.fn(),
      configured: true,
    });

    render(<LoginScreen />, { wrapper: buildWrapper().Wrapper });

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/(tabs)/trilha'));
  });

  it('shows a toast when the Google sign-in request fails', async () => {
    fetchMock.on('POST', '/v1/auth/google', {
      status: 401,
      body: { error: { code: 'invalid_credentials', message: 'invalid Google ID token' } },
    });
    mockUseGoogleIdTokenRequest.mockReturnValue({
      request: {},
      response: { type: 'success', params: { id_token: 'fake-id-token' } },
      promptAsync: jest.fn(),
      configured: true,
    });

    const { findByText } = render(<LoginScreen />, { wrapper: buildWrapper().Wrapper });

    expect(await findByText('Não foi possível entrar com o Google.')).toBeTruthy();
    expect(mockReplace).not.toHaveBeenCalled();
  });

  it('ignores a Google response with no id_token', () => {
    mockUseGoogleIdTokenRequest.mockReturnValue({
      request: {},
      response: { type: 'success', params: {} },
      promptAsync: jest.fn(),
      configured: true,
    });

    render(<LoginScreen />, { wrapper: buildWrapper().Wrapper });

    expect(mockReplace).not.toHaveBeenCalled();
  });
});
