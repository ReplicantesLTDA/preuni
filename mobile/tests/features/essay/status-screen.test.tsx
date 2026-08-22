import { render, waitFor } from '@testing-library/react-native';
import { fetchMock } from '../../lib/mockFetch';
import { buildWrapper } from '../../lib/queryWrapper';

jest.mock('expo-router', () => ({
  useLocalSearchParams: () => ({ id: '00000000-0000-4000-a000-000000000001' }),
}));

const EssayStatusScreen = require('../../../app/(tabs)/redacao/[id]').default;

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

describe('EssayStatusScreen', () => {
  it('shows a pending indicator while grading', async () => {
    fetchMock.on('GET', '/v1/essays/00000000-0000-4000-a000-000000000001', {
      status: 200,
      body: { id: '00000000-0000-4000-a000-000000000001', status: 'pending', submitted_at: '2026-08-22T12:00:00Z' },
    });
    const { Wrapper } = buildWrapper();
    const { getByText } = render(<EssayStatusScreen />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText('Corrigindo sua redação…')).toBeTruthy());
  });

  it('shows the failed message', async () => {
    fetchMock.on('GET', '/v1/essays/00000000-0000-4000-a000-000000000001', {
      status: 200,
      body: { id: '00000000-0000-4000-a000-000000000001', status: 'failed', submitted_at: '2026-08-22T12:00:00Z' },
    });
    const { Wrapper } = buildWrapper();
    const { getByText } = render(<EssayStatusScreen />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText('Não foi possível corrigir esta redação.')).toBeTruthy());
  });

  it('shows the full graded breakdown', async () => {
    fetchMock.on('GET', '/v1/essays/00000000-0000-4000-a000-000000000001', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000001',
        status: 'graded',
        submitted_at: '2026-08-22T12:00:00Z',
        grade: {
          overall_score: 880,
          graded_at: '2026-08-22T12:05:00Z',
          competencies: [
            { competency: 1, score: 160, justification_pt_br: 'Bom domínio.', excerpt: 'trecho um' },
            { competency: 2, score: 180, justification_pt_br: 'Compreende o tema.', excerpt: 'trecho dois' },
            { competency: 3, score: 180, justification_pt_br: 'Boa argumentação.', excerpt: 'trecho três' },
            { competency: 4, score: 180, justification_pt_br: 'Boa coesão.', excerpt: 'trecho quatro' },
            { competency: 5, score: 180, justification_pt_br: 'Proposta clara.', excerpt: 'trecho cinco' },
          ],
        },
      },
    });
    const { Wrapper } = buildWrapper();
    const { getByText } = render(<EssayStatusScreen />, { wrapper: Wrapper });

    await waitFor(() => expect(getByText(/880/)).toBeTruthy());
    expect(getByText('Bom domínio.')).toBeTruthy();
    expect(getByText(/trecho cinco/)).toBeTruthy();
  });
});
