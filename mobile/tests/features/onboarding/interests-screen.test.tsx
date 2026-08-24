import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

const mockReplace = jest.fn();
jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

const InterestsScreen = require('../../../app/(onboarding)/interests').default;

beforeAll(() => fetchMock.install());
afterEach(() => {
  fetchMock.reset();
  mockReplace.mockClear();
});

describe('InterestsScreen', () => {
  it('toggles subjects and completes onboarding with the selected ids', async () => {
    let captured: unknown;
    fetchMock.on('PATCH', '/v1/students/me/onboarding', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Maria',
        email: 'maria@preuni.com',
        xp_total: 0,
        streak_count: 0,
        readiness_score: 0,
        onboarding_completed: true,
      },
      capture: ({ body }) => {
        captured = body;
      },
    });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<InterestsScreen />, { wrapper: Wrapper });

    fireEvent.press(getByText('Matemática'));
    fireEvent.press(getByText('Redação'));
    fireEvent.press(getByText('Concluir'));

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/(tabs)/trilha'));
    expect(captured).toMatchObject({ interests: ['matematica', 'redacao'] });
  });

  it('finishing with nothing selected sends an empty interests array', async () => {
    fetchMock.on('PATCH', '/v1/students/me/onboarding', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Maria',
        email: 'maria@preuni.com',
        xp_total: 0,
        streak_count: 0,
        readiness_score: 0,
        onboarding_completed: true,
      },
    });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<InterestsScreen />, { wrapper: Wrapper });

    fireEvent.press(getByText('Concluir'));

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/(tabs)/trilha'));
  });
});
