# Contract: Shared UI Components

Primitives live under `mobile/src/components/`. All consume design tokens; none accept raw colors/sizes as props (only token keys). Every interactive element ships `accessibilityRole` + `accessibilityLabel`.

## Core primitives

### `<Button>`

```ts
type ButtonProps = {
  label: string;
  onPress: () => void;
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';   // default 'primary'
  size?: 'sm' | 'md' | 'lg';                                 // default 'md'
  loading?: boolean;
  disabled?: boolean;
  leftIcon?: IconName;
  fullWidth?: boolean;
  accessibilityLabel?: string;                               // default = label
};
```

- `primary` uses `tokens.color.accent` on `tokens.color.ink[0]` text.
- `loading` swaps the label for a spinner; the button keeps its width (no layout shift).
- `disabled` reduces opacity and disables `onPress`.

### `<FormField>`

Composable input wrapper: label + input + helper/error + optional right icon.

```ts
type FormFieldProps = {
  label: string;
  value: string;
  onChangeText: (v: string) => void;
  error?: string;            // shown below in `tokens.color.danger`
  helper?: string;           // shown below when no error
  keyboardType?: KeyboardTypeOptions;
  secureTextEntry?: boolean;
  autoComplete?: TextInputProps['autoComplete'];
  testID?: string;
};
```

### `<OtpInput>`

Six-digit OTP entry. Auto-advance, paste-fill, autocomplete one-time-code.

```ts
type OtpInputProps = {
  value: string;
  onChange: (v: string) => void;
  length?: number;        // default 6
  error?: string;
  autoFocus?: boolean;
};
```

### `<TopStatusBar>`

Carries forward the feature 011 contract.

```ts
type TopStatusBarProps = {
  streak: number;
  xp: number;
  avatarUrl?: string | null;
  onAvatarPress?: () => void;     // routes to /(tabs)/perfil
};
```

- Sticky at the top of `(tabs)` layout.
- Hides on `redacao/[promptId]` and `trilha/[trackId]` editor-style routes via layout opt-out.
- Streak + XP always visible when data exists; placeholder values use `tokens.color.ink.muted` (clearly placeholder).

### `<BottomNavigation>`

Carries forward feature 011 contract. Four tabs: Trilha, Redação, Simulado, Perfil. Selected tab is visually distinct (filled icon + accent underline). Every icon has `accessibilityLabel`.

### `<EmptyState>`

```ts
type EmptyStateProps = {
  title: string;
  body?: string;
  mascot?: 'thinking' | 'reading' | 'cheering';   // selects MascotPlaceholder variant
  cta?: { label: string; onPress: () => void };   // optional primary CTA
};
```

### `<MascotPlaceholder>`

Carries forward feature 011 contract. Static SVG variants in `assets/images/mascot/`. Layout does not shift when the mascot is later replaced with artwork.

### `<Card>`

Paper-surface container. `tokens.color.paper[1]`, `tokens.radius.md`, `tokens.shadow.card`, padding `space[4]`.

## Patterns

| Pattern | Component |
|---------|-----------|
| Loading a screen | `<ScreenLoader>` — full-bleed centered skeleton on `paper[0]` |
| Loading inside a card | `<Skeleton>` lines via `react-native-reanimated` |
| Network error | `<ErrorState onRetry>` — same shape as EmptyState, uses `danger` accent |
| Success toast | `<Toast variant="success">` via context provider in root layout |

All loading/error/empty states are routed through these four components — no one-off spinners or ad-hoc error strings (Constitution III).

## Acceptance per primitive

For each primitive:
- [ ] Component test renders default + every variant prop.
- [ ] Snapshot test on iOS + web simulator catches token regressions.
- [ ] Storybook-equivalent dev screen at `/(tabs)/perfil/__dev-components` (gated behind `__DEV__`) lists every primitive — single place to eyeball.
