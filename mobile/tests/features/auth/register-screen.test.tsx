import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockReplace = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

const RegisterScreen = require('../../../app/(auth)/register').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockReplace.mockClear();
});

describe('RegisterScreen', () => {
  it('rejects a weak password without submitting', () => {
    const { getByLabelText, getByText } = render(<RegisterScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('Como podemos te chamar?'), 'Maria');
    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.changeText(getByLabelText('Senha'), 'ab1');
    fireEvent.press(getByText('Enviar'));

    expect(getByText('A senha precisa ter pelo menos 8 caracteres.')).toBeTruthy();
  });

  it('registers and navigates to verify-email with the entered address', async () => {
    fetchMock.on('POST', '/v1/auth/register', {
      status: 200,
      body: { accessToken: 'a', refreshToken: 'b' },
    });

    const { getByLabelText, getByText } = render(<RegisterScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('Como podemos te chamar?'), 'Maria');
    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.changeText(getByLabelText('Senha'), 'Senha1234');
    fireEvent.press(getByText('Enviar'));

    await waitFor(() =>
      expect(mockReplace).toHaveBeenCalledWith({
        pathname: '/(auth)/verify-email',
        params: { email: 'maria@preuni.com' },
      }),
    );
  });

  it('shows a toast when the email is already taken', async () => {
    fetchMock.on('POST', '/v1/auth/register', {
      status: 409,
      body: { error: { code: 'email_already_taken', message: 'Esse e-mail já está em uso.' } },
    });

    const { getByLabelText, getByText, findByText } = render(<RegisterScreen />, {
      wrapper: buildWrapper().Wrapper,
    });

    fireEvent.changeText(getByLabelText('Como podemos te chamar?'), 'Maria');
    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.changeText(getByLabelText('Senha'), 'Senha1234');
    fireEvent.press(getByText('Enviar'));

    expect(await findByText('Esse e-mail já está em uso.')).toBeTruthy();
    expect(mockReplace).not.toHaveBeenCalled();
  });
});
