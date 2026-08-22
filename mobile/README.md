# preuni — mobile

React Native + TypeScript + Expo client for preuni.com.br. Replaces the legacy KMP app under `mobile/`.

See [`specs/012-expo-rn-frontend/quickstart.md`](../specs/012-expo-rn-frontend/quickstart.md) for the canonical setup + smoke tests.

## Prereqs

- Node 20 LTS
- pnpm 9+ (`corepack enable && corepack prepare pnpm@9 --activate` or `npm i -g pnpm`)
- Expo Go on a real iOS/Android device, **or** Xcode + Android Studio for simulators

## Install + run

```bash
pnpm install
cp .env.example .env.local
# edit EXPO_PUBLIC_API_BASE_URL to your LAN IP if testing on a real phone
pnpm start                # QR code; scan with Expo Go
pnpm ios | pnpm android | pnpm web
```

## Test

```bash
pnpm typecheck
pnpm lint
pnpm test --coverage
```

## Layout

```
app/        Expo Router routes (file-based)
src/
  theme/        design tokens + provider
  components/   shared UI primitives
  features/     vertical slices (auth, onboarding, trilha, redacao, simulado, perfil)
  lib/          api/, auth/, query/, i18n/
  stores/       Zustand stores
  types/        Zod schemas + inferred TS types
assets/     fonts + images
tests/      Jest + RTL + msw
```

## Constitution gates

- TS strict + ESLint forbid raw hex/px in component styles.
- All loading/error/empty states route through shared components in `src/components/`.
- Coverage floor: 80% lines on `src/`.
- Web bundle budget: 150 kB gzip per route chunk.
