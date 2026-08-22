import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockBack = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ back: mockBack }),
}));

const ChangePasswordScreen = require('../../../app/(tabs)/perfil/change-password').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockBack.mockClear();
});

describe('ChangePasswordScreen', () => {
  it('shows a validation error for a too-short password', () => {
    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText } = render(<ChangePasswordScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Senha atual'), 'oldpass1');
    fireEvent.changeText(getByLabelText('Nova senha'), 'ab1');
    fireEvent.changeText(getByLabelText('Confirmar nova senha'), 'ab1');
    fireEvent.press(getByText('Atualizar senha'));

    expect(getByText('A senha precisa ter pelo menos 8 caracteres.')).toBeTruthy();
  });

  it('changes the password and navigates back', async () => {
    let captured: unknown;
    fetchMock.on('POST', '/v1/auth/password/change', {
      status: 200,
      body: {},
      capture: ({ body }) => {
        captured = body;
      },
    });
    const { Wrapper } = buildWrapper();
    const { getByLabelText, getByText } = render(<ChangePasswordScreen />, { wrapper: Wrapper });

    fireEvent.changeText(getByLabelText('Senha atual'), 'oldpass1');
    fireEvent.changeText(getByLabelText('Nova senha'), 'newpass1');
    fireEvent.changeText(getByLabelText('Confirmar nova senha'), 'newpass1');
    fireEvent.press(getByText('Atualizar senha'));

    await waitFor(() => expect(mockBack).toHaveBeenCalled());
    expect(captured).toMatchObject({ current_password: 'oldpass1', new_password: 'newpass1' });
  });
});
