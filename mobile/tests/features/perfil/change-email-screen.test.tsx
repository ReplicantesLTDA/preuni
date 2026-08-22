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

  it('rejects an invalid new email', () => {
    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText } = render(<ChangeEmailScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Novo e-mail'), 'not-an-email');
    fireEvent.press(getByText('Enviar código'));

    expect(getByText('Informe um e-mail válido.')).toBeTruthy();
  });
});
