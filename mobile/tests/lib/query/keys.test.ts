import { qk } from '@/lib/query/keys';

describe('qk (query keys)', () => {
  it('trilhaHome varies by optional trackId', () => {
    expect(qk.trilhaHome()).toEqual(['trilha', 'home']);
    expect(qk.trilhaHome('track-1')).toEqual(['trilha', 'home', 'track-1']);
  });

  it('builds stable keys for every other query', () => {
    expect(qk.studentMe()).toEqual(['student', 'me']);
    expect(qk.trilhaTrack('t1')).toEqual(['trilha', 'track', 't1']);
    expect(qk.redacaoList()).toEqual(['redacao', 'list']);
    expect(qk.redacaoPrompt('p1')).toEqual(['redacao', 'prompt', 'p1']);
    expect(qk.simuladoList()).toEqual(['simulado', 'list']);
    expect(qk.essayList()).toEqual(['essay', 'list']);
    expect(qk.essayDetail('e1')).toEqual(['essay', 'detail', 'e1']);
    expect(qk.streakMe()).toEqual(['streak', 'me']);
    expect(qk.friendsList()).toEqual(['friends', 'list']);
    expect(qk.rankingWeekly('bronze')).toEqual(['ranking', 'weekly', 'bronze']);
    expect(qk.rankingMe()).toEqual(['ranking', 'me']);
    expect(qk.medalsMe()).toEqual(['medals', 'me']);
  });
});
