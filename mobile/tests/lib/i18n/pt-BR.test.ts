import { t } from '@/lib/i18n/pt-BR';

describe('t.trilha.streak', () => {
  it('uses the singular form for exactly 1 day', () => {
    expect(t.trilha.streak(1)).toBe('1 dia de ofensiva');
  });

  it('uses the plural form for any other count', () => {
    expect(t.trilha.streak(0)).toBe('0 dias de ofensiva');
    expect(t.trilha.streak(4)).toBe('4 dias de ofensiva');
  });
});
