import { render, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import { useSessionStore } from '@/stores/sessionStore';

const TrilhaHome = require('../../../app/(tabs)/trilha/index').default;

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('TrilhaHome', () => {
  it('shows a greeting, streak, and XP once loaded', async () => {
    useSessionStore.setState({ status: 'authed', student: null });
    fetchMock.on('GET', '/v1/students/me', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Maria Silva',
        email: 'maria@preuni.com',
        xp_total: 500,
        streak_count: 4,
        readiness_score: 70,
        onboarding_completed: true,
      },
    });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<TrilhaHome />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText(/Maria/)).toBeTruthy());
    expect(getByText(/4 dias de ofensiva/)).toBeTruthy();
    expect(getByText(/500 XP/)).toBeTruthy();
  });

  it('shows an error state with a retry option on failure', async () => {
    useSessionStore.setState({ status: 'authed', student: null });
    fetchMock.on('GET', '/v1/students/me', { status: 500, bodyText: '{"error":{"code":"INTERNAL_ERROR","message":"boom"}}' });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<TrilhaHome />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText('Sem conexão')).toBeTruthy());
  });
});
