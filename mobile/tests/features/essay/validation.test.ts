import { SubmitEssaySchema } from '@/features/essay/validation';

describe('SubmitEssaySchema', () => {
  it('accepts a fully filled submission', () => {
    const result = SubmitEssaySchema.safeParse({
      promptThemeTitle: 'Tema',
      promptThemeContext: 'Contexto',
      essayText: 'Texto da redação.',
    });
    expect(result.success).toBe(true);
  });

  it('rejects an empty theme title', () => {
    const result = SubmitEssaySchema.safeParse({
      promptThemeTitle: '',
      promptThemeContext: 'Contexto',
      essayText: 'Texto',
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.path).toEqual(['promptThemeTitle']);
    }
  });

  it('rejects an empty theme context', () => {
    const result = SubmitEssaySchema.safeParse({
      promptThemeTitle: 'Tema',
      promptThemeContext: '',
      essayText: 'Texto',
    });
    expect(result.success).toBe(false);
  });

  it('rejects an empty essay text', () => {
    const result = SubmitEssaySchema.safeParse({
      promptThemeTitle: 'Tema',
      promptThemeContext: 'Contexto',
      essayText: '',
    });
    expect(result.success).toBe(false);
  });
});
