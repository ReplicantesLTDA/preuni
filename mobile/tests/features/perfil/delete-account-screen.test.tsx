import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockReplace = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace, back: jest.fn() }),
}));

const DeleteAccountScreen = require('../../../app/(tabs)/perfil/delete-account').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockReplace.mockClear();
});

describe('DeleteAccountScreen', () => {
  it('blocks deletion until the confirmation word is typed', () => {
    const { Wrapper } = buildWrapper();
    const { getByText, getByLabelText } = render(<DeleteAccountScreen />, { wrapper: Wrapper });

    fireEvent.press(getByText('Excluir minha conta'));
    expect(getByText('Digite EXCLUIR para confirmar.')).toBeTruthy();

    fireEvent.changeText(getByLabelText('Digite "EXCLUIR" para confirmar'), 'nope');
    fireEvent.press(getByText('Excluir minha conta'));
    expect(getByText('Digite EXCLUIR para confirmar.')).toBeTruthy();
  });

  it('deletes the account and redirects to welcome', async () => {
    fetchMock.on('DELETE', '/v1/students/me', { status: 200, body: {} });
    const { Wrapper } = buildWrapper();
    const { getByText, getByLabelText } = render(<DeleteAccountScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Digite "EXCLUIR" para confirmar'), 'EXCLUIR');
    fireEvent.press(getByText('Excluir minha conta'));

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/(auth)/welcome'));
  });

  it('shows a toast when deletion fails', async () => {
    fetchMock.on('DELETE', '/v1/students/me', {
      status: 500,
      bodyText: '{"error":{"code":"INTERNAL_ERROR","message":"boom"}}',
    });
    const { Wrapper } = buildWrapper();
    const { getByText, getByLabelText, findByText } = render(<DeleteAccountScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Digite "EXCLUIR" para confirmar'), 'EXCLUIR');
    fireEvent.press(getByText('Excluir minha conta'));

    expect(await findByText('boom')).toBeTruthy();
    expect(mockReplace).not.toHaveBeenCalled();
  });
});
