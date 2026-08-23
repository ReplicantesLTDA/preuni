import { render, waitFor, fireEvent } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import FriendsScreen from '../../../app/(tabs)/perfil/friends';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('FriendsScreen', () => {
  it('renders the friend list with streak and grade', async () => {
    fetchMock.on('GET', '/v1/friends', {
      status: 200,
      body: [
        {
          friendship_id: '00000000-0000-4000-a000-000000000010',
          user_id: '00000000-0000-4000-a000-000000000011',
          display_name: 'Amigo',
          current_streak: 5,
          latest_overall_score: 700,
        },
      ],
    });
    const { Wrapper } = buildWrapper();
    const { getByText } = render(<FriendsScreen />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText('Amigo')).toBeTruthy());
    expect(getByText(/700/)).toBeTruthy();
  });

  it('shows the empty state with no friends', async () => {
    fetchMock.on('GET', '/v1/friends', { status: 200, body: [] });
    const { Wrapper } = buildWrapper();
    const { getByText } = render(<FriendsScreen />, { wrapper: Wrapper });

    await waitFor(() =>
      expect(
        getByText('Você ainda não tem amigos. Adicione alguém para acompanhar a ofensiva e as notas.'),
      ).toBeTruthy(),
    );
  });

  it('sends a friend request by user id', async () => {
    fetchMock.on('GET', '/v1/friends', { status: 200, body: [] });
    let captured: unknown;
    fetchMock.on('POST', '/v1/friends/requests', {
      status: 201,
      body: { id: '00000000-0000-4000-a000-000000000020', status: 'pending' },
      capture: ({ body }) => {
        captured = body;
      },
    });
    const { Wrapper } = buildWrapper();
    const { getAllByLabelText } = render(<FriendsScreen />, { wrapper: Wrapper });

    const [input, button] = getAllByLabelText('Adicionar amigo');
    fireEvent.changeText(input!, '00000000-0000-4000-a000-000000000099');
    fireEvent.press(button!);

    await waitFor(() => expect(captured).toMatchObject({ addressee_id: '00000000-0000-4000-a000-000000000099' }));
  });

  it('shows a toast when the friend request fails', async () => {
    fetchMock.on('GET', '/v1/friends', { status: 200, body: [] });
    fetchMock.on('POST', '/v1/friends/requests', {
      status: 500,
      bodyText: '{"error":{"code":"INTERNAL_ERROR","message":"boom"}}',
    });
    const { Wrapper } = buildWrapper();
    const { getAllByLabelText, findByText } = render(<FriendsScreen />, { wrapper: Wrapper });

    const [input, button] = getAllByLabelText('Adicionar amigo');
    fireEvent.changeText(input!, '00000000-0000-4000-a000-000000000099');
    fireEvent.press(button!);

    expect(await findByText('Não foi possível enviar o pedido.')).toBeTruthy();
  });

  it('removes a friend', async () => {
    fetchMock.on('GET', '/v1/friends', {
      status: 200,
      body: [
        {
          friendship_id: '00000000-0000-4000-a000-000000000010',
          user_id: '00000000-0000-4000-a000-000000000011',
          display_name: 'Amigo',
          current_streak: 5,
          latest_overall_score: null,
        },
      ],
    });
    let removedId: unknown;
    fetchMock.on('DELETE', '/v1/friends/00000000-0000-4000-a000-000000000010', {
      status: 204,
      capture: () => {
        removedId = '00000000-0000-4000-a000-000000000010';
      },
    });
    const { Wrapper } = buildWrapper();
    const { getByText, findByText } = render(<FriendsScreen />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText('Amigo')).toBeTruthy());
    expect(await findByText('Ainda sem redação corrigida')).toBeTruthy();

    fireEvent.press(getByText('Remover amigo'));
    await waitFor(() => expect(removedId).toBe('00000000-0000-4000-a000-000000000010'));
  });
});
