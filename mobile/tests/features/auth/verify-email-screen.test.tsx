import { render, fireEvent, waitFor, act } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockReplace = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace }),
  useLocalSearchParams: () => ({ email: 'maria@preuni.com' }),
}));

const VerifyEmailScreen = require('../../../app/(auth)/verify-email').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockReplace.mockClear();
});

describe('VerifyEmailScreen', () => {
  it('verifies the otp and navigates home', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify', { status: 200, body: {} });

    const { getByLabelText, getByText } = render(<VerifyEmailScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('Código de verificação'), '123456');
    fireEvent.press(getByText('Enviar'));

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/'));
  });

  it('shows a field error for an invalid otp', async () => {
    fetchMock.on('POST', '/v1/auth/email/verify', {
      status: 422,
      body: { error: { code: 'otp_invalid', field: 'otp', message: 'Código inválido.' } },
    });

    const { getByLabelText, getByText, findByText } = render(<VerifyEmailScreen />, {
      wrapper: buildWrapper().Wrapper,
    });

    fireEvent.changeText(getByLabelText('Código de verificação'), '000000');
    fireEvent.press(getByText('Enviar'));

    expect(await findByText('Código inválido.')).toBeTruthy();
    expect(mockReplace).not.toHaveBeenCalled();
  });

  it('starts with the resend button in cooldown, disabled', () => {
    const { getByText } = render(<VerifyEmailScreen />, { wrapper: buildWrapper().Wrapper });

    expect(getByText('Aguarde 60s para reenviar')).toBeTruthy();
  });

  it('rejects a code shorter than 6 digits without submitting', () => {
    const { getByLabelText, getByText } = render(<VerifyEmailScreen />, { wrapper: buildWrapper().Wrapper });

    fireEvent.changeText(getByLabelText('Código de verificação'), '123');
    fireEvent.press(getByText('Enviar'));

    expect(getByText('O código tem 6 dígitos.')).toBeTruthy();
    expect(mockReplace).not.toHaveBeenCalled();
  });

  it('resends the code once the cooldown elapses', async () => {
    jest.useFakeTimers();
    try {
      fetchMock.on('POST', '/v1/auth/email/verify-resend', { status: 200, body: {} });

      const { getByText } = render(<VerifyEmailScreen />, { wrapper: buildWrapper().Wrapper });

      act(() => {
        jest.advanceTimersByTime(60_000);
      });

      expect(getByText('Reenviar código')).toBeTruthy();
      fireEvent.press(getByText('Reenviar código'));
      await act(async () => {
        await Promise.resolve();
      });

      expect(getByText('Aguarde 60s para reenviar')).toBeTruthy();
    } finally {
      jest.useRealTimers();
    }
  });

  it('shows a toast when resend fails', async () => {
    jest.useFakeTimers();
    try {
      fetchMock.on('POST', '/v1/auth/email/verify-resend', {
        status: 429,
        body: { error: { message: 'Muitas tentativas.' } },
      });

      const { getByText } = render(<VerifyEmailScreen />, { wrapper: buildWrapper().Wrapper });

      act(() => {
        jest.advanceTimersByTime(60_000);
      });

      fireEvent.press(getByText('Reenviar código'));
      await act(async () => {
        await Promise.resolve();
        await Promise.resolve();
      });

      expect(getByText('Muitas tentativas.')).toBeTruthy();
    } finally {
      jest.useRealTimers();
    }
  });
});
