import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockReplace = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

const OtpLoginScreen = require('../../../app/(auth)/otp-login').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockReplace.mockClear();
});

describe('OtpLoginScreen', () => {
  it('rejects an invalid email and does not request a code', () => {
    const { getByLabelText, getByText } = render(<OtpLoginScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('E-mail'), 'not-an-email');
    fireEvent.press(getByText('Enviar código'));

    expect(getByText('Informe um e-mail válido.')).toBeTruthy();
  });

  it('requests a code then verifies it and navigates home', async () => {
    fetchMock.on('POST', '/v1/auth/otp/request', { status: 200, body: {} });
    fetchMock.on('POST', '/v1/auth/otp/verify', {
      status: 200,
      body: { accessToken: 'a', refreshToken: 'b' },
    });

    const { getByLabelText, getByText, findByText } = render(<OtpLoginScreen />, {
      wrapper: buildWrapper().Wrapper,
    });

    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.press(getByText('Enviar código'));

    expect(await findByText('Enviamos um código para maria@preuni.com.')).toBeTruthy();

    fireEvent.changeText(getByLabelText('Código de verificação'), '123456');
    fireEvent.press(getByText('Entrar'));

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/'));
  });

  it('shows a field error for an invalid otp code', async () => {
    fetchMock.on('POST', '/v1/auth/otp/request', { status: 200, body: {} });
    fetchMock.on('POST', '/v1/auth/otp/verify', {
      status: 422,
      body: { error: { code: 'otp_invalid', field: 'otp', message: 'Código inválido.' } },
    });

    const { getByLabelText, getByText, findByText } = render(<OtpLoginScreen />, {
      wrapper: buildWrapper().Wrapper,
    });

    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.press(getByText('Enviar código'));
    await findByText('Enviamos um código para maria@preuni.com.');

    fireEvent.changeText(getByLabelText('Código de verificação'), '000000');
    fireEvent.press(getByText('Entrar'));

    expect(await findByText('Código inválido.')).toBeTruthy();
    expect(mockReplace).not.toHaveBeenCalled();
  });
});
