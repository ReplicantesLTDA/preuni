import { fetchMock } from '../../lib/mockFetch';
import { createApiClient } from '@/lib/api/client';
import { makeEssayApi } from '@/features/essay/api';

beforeAll(() => fetchMock.install());
afterEach(() => fetchMock.reset());

function build() {
  const client = createApiClient({
    baseUrl: fetchMock.baseUrl(),
    getAccessToken: jest.fn(async () => 'tok'),
    refreshTokens: jest.fn(async () => false),
    onUnauthorized: jest.fn(),
  });
  return makeEssayApi(client);
}

const GRADE_BODY = {
  overall_score: 880,
  competencies: [
    { competency: 1, score: 160, justification_pt_br: 'Bom.', excerpt: 'trecho' },
    { competency: 2, score: 180, justification_pt_br: 'Bom.', excerpt: 'trecho' },
    { competency: 3, score: 180, justification_pt_br: 'Bom.', excerpt: 'trecho' },
    { competency: 4, score: 180, justification_pt_br: 'Bom.', excerpt: 'trecho' },
    { competency: 5, score: 180, justification_pt_br: 'Bom.', excerpt: 'trecho' },
  ],
  graded_at: '2026-08-22T12:00:00Z',
};

describe('Essay API', () => {
  it('submit(): posts snake_case body and parses the 202-shaped pending response', async () => {
    let captured: unknown;
    fetchMock.on('POST', '/v1/essays', {
      status: 202,
      body: { id: '00000000-0000-4000-a000-000000000001', status: 'pending', submitted_at: '2026-08-22T12:00:00Z' },
      capture: ({ body }) => {
        captured = body;
      },
    });
    const api = build();
    const result = await api.submit({
      promptThemeTitle: 'Tema',
      promptThemeContext: 'Contexto',
      essayText: 'Texto',
    });
    expect(result.status).toBe('pending');
    expect(captured).toMatchObject({
      prompt_theme_title: 'Tema',
      prompt_theme_context: 'Contexto',
      essay_text: 'Texto',
    });
  });

  it('get(): parses a graded submission with all 5 competencies', async () => {
    fetchMock.on('GET', '/v1/essays/00000000-0000-4000-a000-000000000001', {
      status: 200,
      body: {
        id: '00000000-0000-4000-a000-000000000001',
        status: 'graded',
        submitted_at: '2026-08-22T12:00:00Z',
        grade: GRADE_BODY,
      },
    });
    const api = build();
    const result = await api.get('00000000-0000-4000-a000-000000000001');
    expect(result.status).toBe('graded');
    expect(result.grade?.overallScore).toBe(880);
    expect(result.grade?.competencies).toHaveLength(5);
    expect(result.grade?.competencies[0]?.justificationPtBr).toBe('Bom.');
  });

  it('get(): parses a pending submission with no grade field', async () => {
    fetchMock.on('GET', '/v1/essays/00000000-0000-4000-a000-000000000002', {
      status: 200,
      body: { id: '00000000-0000-4000-a000-000000000002', status: 'pending', submitted_at: '2026-08-22T12:00:00Z' },
    });
    const api = build();
    const result = await api.get('00000000-0000-4000-a000-000000000002');
    expect(result.status).toBe('pending');
    expect(result.grade).toBeUndefined();
  });

  it('list(): parses an array of essay summaries', async () => {
    fetchMock.on('GET', '/v1/essays', {
      status: 200,
      body: [
        { id: '00000000-0000-4000-a000-000000000001', status: 'graded', submitted_at: '2026-08-22T12:00:00Z' },
        { id: '00000000-0000-4000-a000-000000000002', status: 'pending', submitted_at: '2026-08-22T13:00:00Z' },
      ],
    });
    const api = build();
    const result = await api.list();
    expect(result).toHaveLength(2);
    expect(result[0]?.status).toBe('graded');
  });

  it('list(): rejects a malformed response', async () => {
    fetchMock.on('GET', '/v1/essays', { status: 200, body: { not: 'an array' } });
    const api = build();
    await expect(api.list()).rejects.toMatchObject({ kind: 'Unknown' });
  });
});
