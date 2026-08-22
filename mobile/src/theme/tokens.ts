export const tokens = {
  color: {
    paper: { 0: '#fafaf6', 1: '#f3f1ea' },
    ink: { 0: '#1a1a1a', 1: '#2a2a2a', 2: '#4a4a4a', muted: '#8a8a8a' },
    accent: '#ff8c42',
    success: '#3a8a3a',
    danger: '#b03030',
    transparent: 'transparent',
  },
  space: [0, 4, 8, 12, 16, 20, 24, 32, 40, 56] as const,
  radius: { sm: 8, md: 12, lg: 20, pill: 999 },
  font: {
    display: 'CaveatBrush',
    heading: 'PatrickHand',
    body: 'ArchitectsDaughter',
    numeric: 'Kalam',
    numericBold: 'KalamBold',
    accent: 'Caveat',
    accentBold: 'CaveatBold',
  },
  size: {
    xs: 12,
    sm: 14,
    md: 16,
    lg: 20,
    xl: 24,
    '2xl': 32,
    '3xl': 40,
  },
  shadow: {
    card: {
      shadowColor: '#000000',
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.08,
      shadowRadius: 6,
      elevation: 2,
    },
  },
} as const;

export type Tokens = typeof tokens;
export type SpaceIndex = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9;
