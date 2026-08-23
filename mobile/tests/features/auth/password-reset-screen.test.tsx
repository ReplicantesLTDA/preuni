import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockReplace = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

const PasswordResetScreen = require('../../../app/(auth)/password-reset').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockReplace.mockClear();
});

describe('PasswordResetScreen', () => {
  it('requests a code, confirms the new password, and returns to login', async () => {
    fetchMock.on('POST', '/v1/auth/password/reset/request', { status: 200, body: {} });
    fetchMock.on('POST', '/v1/auth/password/reset/confirm', { status: 200, body: {} });

    const { getByLabelText, getByText, findByText } = render(<PasswordResetScreen />, {
      wrapper: buildWrapper().Wrapper,
    });

    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.press(getByText('Enviar código'));

    expect(await findByText('Código enviado para maria@preuni.com.')).toBeTruthy();

    fireEvent.changeText(getByLabelText('Código de verificação'), '123456');
    fireEvent.changeText(getByLabelText('Nova senha'), 'Senha1234');
    fireEvent.press(getByText('Atualizar senha'));

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/(auth)/login'));
  });

  it('shows a field error when the confirm request fails', async () => {
    fetchMock.on('POST', '/v1/auth/password/reset/request', { status: 200, body: {} });
    fetchMock.on('POST', '/v1/auth/password/reset/confirm', {
      status: 422,
      body: { error: { field: 'otp', message: 'O código tem 6 dígitos.' } },
    });

    const { getByLabelText, getByText, findByText } = render(<PasswordResetScreen />, {
      wrapper: buildWrapper().Wrapper,
    });

    fireEvent.changeText(getByLabelText('E-mail'), 'maria@preuni.com');
    fireEvent.press(getByText('Enviar código'));
    await findByText('Código enviado para maria@preuni.com.');

    fireEvent.changeText(getByLabelText('Código de verificação'), '654321');
    fireEvent.changeText(getByLabelText('Nova senha'), 'Senha1234');
    fireEvent.press(getByText('Atualizar senha'));

    expect(await findByText('O código tem 6 dígitos.')).toBeTruthy();
    expect(mockReplace).not.toHaveBeenCalled();
  });
});
