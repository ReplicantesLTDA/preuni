# Contract: Design Tokens

**Source of truth**: `wireframe.html` (repo root). Tokens extracted into `mobile/src/theme/tokens.ts`. Components MUST NOT use raw hex/px values (Constitution III).

## Module shape

```ts
// mobile/src/theme/tokens.ts
export const tokens = {
  color: {
    paper: { 0: '#fafaf6', 1: '#f3f1ea' },
    ink:   { 0: '#1a1a1a', 1: '#2a2a2a', 2: '#4a4a4a', muted: '#8a8a8a' },
    accent:  '#ff8c42',
    success: '#3a8a3a',
    danger:  '#b03030',
  },
  space: [0, 4, 8, 12, 16, 20, 24, 32, 40, 56] as const,
  radius: { sm: 8, md: 12, lg: 20, pill: 999 },
  font: {
    display:  'CaveatBrush',
    heading:  'PatrickHand',
    body:     'ArchitectsDaughter',
    numeric:  'Kalam',
    accent:   'Caveat',
  },
  size: {
    xs: 12, sm: 14, md: 16, lg: 20, xl: 24, '2xl': 32, '3xl': 40,
  },
  shadow: {
    card: { offset: { width: 0, height: 2 }, opacity: 0.08, radius: 6, elevation: 2 },
  },
} as const;

export type Tokens = typeof tokens;
```

## Required font assets

Loaded once in the root layout via `expo-font.useFonts`:

| Family | File | Source |
|--------|------|--------|
| `CaveatBrush` | `assets/fonts/CaveatBrush-Regular.ttf` | Google Fonts |
| `PatrickHand` | `assets/fonts/PatrickHand-Regular.ttf` | Google Fonts |
| `ArchitectsDaughter` | `assets/fonts/ArchitectsDaughter-Regular.ttf` | Google Fonts |
| `Kalam` | `assets/fonts/Kalam-Regular.ttf`, `Kalam-Bold.ttf` | Google Fonts |
| `Caveat` | `assets/fonts/Caveat-Regular.ttf`, `Caveat-Bold.ttf` | Google Fonts |

## Usage contract

```ts
import { useTheme } from '@/theme';

const { color, space, font, radius } = useTheme();
const styles = StyleSheet.create({
  card: {
    backgroundColor: color.paper[1],
    padding: space[4],          // 16
    borderRadius: radius.md,
    fontFamily: font.body,
  },
});
```

## Acceptance

- [ ] All raw color/spacing/typography values in `mobile/src/` come from `tokens.ts`.
- [ ] ESLint rule (or repo lint) forbids `#[0-9a-f]{3,6}` in component styles.
- [ ] All five hand-drawn fonts load before the first authenticated screen renders (splash held until fonts ready).
