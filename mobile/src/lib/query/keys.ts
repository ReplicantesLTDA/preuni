export const qk = {
  studentMe: () => ['student', 'me'] as const,
  trilhaHome: (trackId?: string) =>
    trackId ? (['trilha', 'home', trackId] as const) : (['trilha', 'home'] as const),
  trilhaTrack: (trackId: string) => ['trilha', 'track', trackId] as const,
  redacaoList: () => ['redacao', 'list'] as const,
  redacaoPrompt: (id: string) => ['redacao', 'prompt', id] as const,
  simuladoList: () => ['simulado', 'list'] as const,
  essayList: () => ['essay', 'list'] as const,
  essayDetail: (id: string) => ['essay', 'detail', id] as const,
  streakMe: () => ['streak', 'me'] as const,
};
