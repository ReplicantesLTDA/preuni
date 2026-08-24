import { render, waitFor, fireEvent } from '@testing-library/react-native';
import { RefreshControl } from 'react-native';
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

  it('shows a toast when tapping the next-activity CTA or a subject tile', async () => {
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
    const { getByText, getByLabelText, findByText } = render(<TrilhaHome />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText(/Maria/)).toBeTruthy());

    fireEvent.press(getByLabelText('Em breve'));
    expect(await findByText('Trilha de atividades em construção.')).toBeTruthy();

    fireEvent.press(getByLabelText('Matemática'));
    expect(await findByText('Matemática: conteúdo em breve.')).toBeTruthy();
  });

  it('refetches student data when pulled to refresh', async () => {
    useSessionStore.setState({ status: 'authed', student: null });
    let callCount = 0;
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
      capture: () => {
        callCount += 1;
      },
    });

    const { Wrapper } = buildWrapper();
    const { getByText, UNSAFE_getByType } = render(<TrilhaHome />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText(/Maria/)).toBeTruthy());
    expect(callCount).toBe(1);

    const refreshControl = UNSAFE_getByType(RefreshControl);
    fireEvent(refreshControl, 'refresh');

    await waitFor(() => expect(callCount).toBeGreaterThan(1));
  });

  it.each([
    [8, 'Bom dia, Maria!'],
    [15, 'Boa tarde, Maria!'],
    [21, 'Boa noite, Maria!'],
  ])('greets appropriately for hour %i', async (hour, expected) => {
    const spy = jest.spyOn(Date.prototype, 'getHours').mockReturnValue(hour);
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

    await waitFor(() => expect(getByText(expected)).toBeTruthy());
    spy.mockRestore();
  });

  it('shows an error state with a retry option on failure', async () => {
    useSessionStore.setState({ status: 'authed', student: null });
    fetchMock.on('GET', '/v1/students/me', { status: 500, bodyText: '{"error":{"code":"INTERNAL_ERROR","message":"boom"}}' });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<TrilhaHome />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText('Sem conexão')).toBeTruthy());
  });
});
