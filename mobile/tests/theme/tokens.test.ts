import { tokens } from '@/theme/tokens';

describe('design tokens', () => {
  it('exposes the wireframe palette exactly', () => {
    expect(tokens.color.paper[0]).toBe('#fafaf6');
    expect(tokens.color.paper[1]).toBe('#f3f1ea');
    expect(tokens.color.ink[0]).toBe('#1a1a1a');
    expect(tokens.color.ink[1]).toBe('#2a2a2a');
    expect(tokens.color.ink[2]).toBe('#4a4a4a');
    expect(tokens.color.ink.muted).toBe('#8a8a8a');
    expect(tokens.color.accent).toBe('#ff8c42');
    expect(tokens.color.success).toBe('#3a8a3a');
    expect(tokens.color.danger).toBe('#b03030');
  });

  it('uses the documented space scale', () => {
    expect([...tokens.space]).toEqual([0, 4, 8, 12, 16, 20, 24, 32, 40, 56]);
  });

  it('uses radius tokens sm/md/lg/pill', () => {
    expect(tokens.radius).toEqual({ sm: 8, md: 12, lg: 20, pill: 999 });
  });

  it('exports all hand-drawn font family names', () => {
    expect(tokens.font).toEqual({
      display: 'CaveatBrush',
      heading: 'PatrickHand',
      body: 'ArchitectsDaughter',
      numeric: 'Kalam',
      numericBold: 'KalamBold',
      accent: 'Caveat',
      accentBold: 'CaveatBold',
    });
  });
});
