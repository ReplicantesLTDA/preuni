import { render, waitFor, fireEvent } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockPush = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

const RedacaoHome = require('../../../app/(tabs)/redacao/index').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockPush.mockClear();
});

describe('RedacaoHome', () => {
  it('shows the current streak and essay history once loaded', async () => {
    fetchMock.on('GET', '/v1/streaks/me', {
      status: 200,
      body: { current_streak: 3, longest_streak: 5, last_active_day: '2026-08-22' },
    });
    fetchMock.on('GET', '/v1/essays', {
      status: 200,
      body: [{ id: '00000000-0000-4000-a000-000000000001', status: 'graded', submitted_at: '2026-08-22T12:00:00Z' }],
    });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<RedacaoHome />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText(/3 dias de ofensiva/)).toBeTruthy());
    expect(getByText(/Ver correção/)).toBeTruthy();
  });

  it('shows the empty state with no submissions', async () => {
    fetchMock.on('GET', '/v1/streaks/me', {
      status: 200,
      body: { current_streak: 0, longest_streak: 0, last_active_day: null },
    });
    fetchMock.on('GET', '/v1/essays', { status: 200, body: [] });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<RedacaoHome />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText('Nenhuma redação ainda')).toBeTruthy());
  });

  it('navigates to the write screen', async () => {
    fetchMock.on('GET', '/v1/streaks/me', {
      status: 200,
      body: { current_streak: 0, longest_streak: 0, last_active_day: null },
    });
    fetchMock.on('GET', '/v1/essays', { status: 200, body: [] });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<RedacaoHome />, { wrapper: Wrapper });

    fireEvent.press(getByText('Escrever agora'));
    expect(mockPush).toHaveBeenCalledWith('/redacao/write');
  });
});
