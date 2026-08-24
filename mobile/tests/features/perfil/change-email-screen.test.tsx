import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockBack = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ back: mockBack }),
}));

const ChangeEmailScreen = require('../../../app/(tabs)/perfil/change-email').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockBack.mockClear();
});

describe('ChangeEmailScreen', () => {
  it('walks through request -> confirm', async () => {
    fetchMock.on('POST', '/v1/auth/email/change/request', { status: 200, body: {} });
    fetchMock.on('POST', '/v1/auth/email/change/confirm', { status: 200, body: {} });

    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText, queryByText } = render(<ChangeEmailScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Novo e-mail'), 'novo@preuni.com');
    fireEvent.press(getByText('Enviar código'));

    await waitFor(() => expect(queryByText(/Enviamos um código/)).toBeTruthy());

    fireEvent.press(getByText('Confirmar'));
    // OTP still empty -> schema rejects, no navigation yet
    expect(mockBack).not.toHaveBeenCalled();
  });

  it('shows a toast when the request step fails', async () => {
    fetchMock.on('POST', '/v1/auth/email/change/request', {
      status: 409,
      body: { error: { message: 'Esse e-mail já está em uso.' } },
    });

    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText, findByText } = render(<ChangeEmailScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Novo e-mail'), 'novo@preuni.com');
    fireEvent.press(getByText('Enviar código'));

    expect(await findByText('Esse e-mail já está em uso.')).toBeTruthy();
  });

  it('confirms the change and navigates back', async () => {
    fetchMock.on('POST', '/v1/auth/email/change/request', { status: 200, body: {} });
    fetchMock.on('POST', '/v1/auth/email/change/confirm', { status: 200, body: {} });

    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText, queryByText } = render(<ChangeEmailScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Novo e-mail'), 'novo@preuni.com');
    fireEvent.press(getByText('Enviar código'));
    await waitFor(() => expect(queryByText(/Enviamos um código/)).toBeTruthy());

    fireEvent.changeText(getByLabelText('Código de verificação'), '123456');
    fireEvent.press(getByText('Confirmar'));

    await waitFor(() => expect(mockBack).toHaveBeenCalled());
  });

  it('shows a field error when the confirm step fails', async () => {
    fetchMock.on('POST', '/v1/auth/email/change/request', { status: 200, body: {} });
    fetchMock.on('POST', '/v1/auth/email/change/confirm', {
      status: 422,
      body: { error: { field: 'otp', message: 'Código inválido.' } },
    });

    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText, findByText, queryByText } = render(<ChangeEmailScreen />, {
      wrapper: Wrapper,
    });

    fireEvent.changeText(getByLabelText('Novo e-mail'), 'novo@preuni.com');
    fireEvent.press(getByText('Enviar código'));
    await waitFor(() => expect(queryByText(/Enviamos um código/)).toBeTruthy());

    fireEvent.changeText(getByLabelText('Código de verificação'), '000000');
    fireEvent.press(getByText('Confirmar'));

    expect(await findByText('Código inválido.')).toBeTruthy();
    expect(mockBack).not.toHaveBeenCalled();
  });

  it('rejects an invalid new email', () => {
    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText } = render(<ChangeEmailScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Novo e-mail'), 'not-an-email');
    fireEvent.press(getByText('Enviar código'));

    expect(getByText('Informe um e-mail válido.')).toBeTruthy();
  });
});
