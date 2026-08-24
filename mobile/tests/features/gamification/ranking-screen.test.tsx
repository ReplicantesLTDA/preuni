import { render, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';
import RankingScreen from '../../../app/(tabs)/perfil/ranking';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('RankingScreen', () => {
  it('renders the caller rank, leaderboard, and medals once loaded', async () => {
    fetchMock.on('GET', '/v1/ranking/me', {
      status: 200,
      body: {
        user_id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Aluno',
        weekly_score: 450,
        league_tier: 'silver',
        rank_in_tier: 2,
      },
    });
    fetchMock.on('GET', '/v1/ranking/weekly?tier=silver', {
      status: 200,
      body: [
        {
          user_id: '00000000-0000-4000-a000-000000000011',
          display_name: 'Aluno',
          weekly_score: 450,
          league_tier: 'silver',
        },
        {
          user_id: '00000000-0000-4000-a000-000000000012',
          display_name: 'Outro',
          weekly_score: 900,
          league_tier: 'silver',
        },
      ],
    });
    fetchMock.on('GET', '/v1/medals/me', {
      status: 200,
      body: [{ type: 'streak_7_day', earned_at: '2026-08-20T00:00:00Z' }],
    });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<RankingScreen />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText(/Aluno/)).toBeTruthy());
    expect(getByText(/Outro/)).toBeTruthy();
    await waitFor(() => expect(getByText('Ofensiva de 7 dias')).toBeTruthy());
  });

  it('shows empty states when nothing has loaded yet', async () => {
    fetchMock.on('GET', '/v1/ranking/me', {
      status: 200,
      body: {
        user_id: '00000000-0000-4000-a000-000000000011',
        display_name: 'Aluno',
        weekly_score: 0,
        league_tier: 'bronze',
      },
    });
    fetchMock.on('GET', '/v1/ranking/weekly?tier=bronze', { status: 200, body: [] });
    fetchMock.on('GET', '/v1/medals/me', { status: 200, body: [] });

    const { Wrapper } = buildWrapper();
    const { getByText } = render(<RankingScreen />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText('Nenhuma medalha ainda. Continue sua ofensiva para desbloquear!')).toBeTruthy());
  });
});
