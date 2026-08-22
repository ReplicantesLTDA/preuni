# Implementation Plan: Migrate frontend to React Native + Expo (012)

**Branch**: `012-expo-rn-frontend` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/012-expo-rn-frontend/spec.md`

## Summary

Replace the Kotlin Multiplatform frontend (`mobile/`) with a fresh **React Native + TypeScript** app on **Expo**, targeting iOS, Android and Web from one codebase. The new app re-implements the existing product following `wireframe.html` as the visual contract (warm-paper / hand-drawn notebook aesthetic) and adapts any KMP-only flows into that language. The backend monolith is untouched. Navigation: Expo Router. Client state: Zustand. Server cache + mutations: TanStack Query. Schema validation: Zod. Old `mobile/` tree is removed only after P1 + P2 stories reach parity.

## Technical Context

**Language/Version**: TypeScript 5.x, React Native 0.76+, Expo SDK 52+, React 18.3+.
**Primary Dependencies**: `expo`, `expo-router`, `react-native-web`, `@tanstack/react-query`, `zustand`, `zod`, `expo-secure-store`, `expo-image-picker`, `expo-image`, `expo-font`, `expo-localization`, `@expo/vector-icons`, `react-native-reanimated`, `react-native-gesture-handler`, `react-native-safe-area-context`, `react-native-svg`.
**Storage**: `expo-secure-store` (session tokens, `active_track_id`, `welcome_seen`). No client-side relational DB in v1; TanStack Query handles server cache.
**Testing**: Jest + `@testing-library/react-native` (unit/component), `msw` (mock API at network boundary), Detox (smoke E2E on Android emulator, post-MVP). Lint: ESLint + `@typescript-eslint`. Format: Prettier. Type-check: `tsc --noEmit` in CI.
**Target Platform**: iOS 15+, Android 8+ (API 26+), modern evergreen browsers (Chrome/Safari/Firefox/Edge latest 2 versions). Built via Expo Go for dev, EAS Build for store binaries (post-MVP), `expo export --platform web` for web.
**Project Type**: Mobile + Web client (single codebase) consuming an existing HTTP API. New top-level package: `mobile/` at repo root.
**Performance Goals**: LCP ≤ 2.5s on throttled 4G (web build); INP ≤ 200ms; Trilha cold open → first interactive < 2s on mid-tier Android; hot reload < 10s.
**Constraints**: No backend changes for parity behavior; consume existing `{ error: { code, field, message } }` envelope; pt-BR copy only; design tokens from `wireframe.html`, no raw hex/px in components.
**Scale/Scope**: ~25 screens at parity (auth ×6, onboarding ×3, trilha ×4, redação ×4, simulado ×3, perfil ×5). One internal-testing cohort (~50 users) for first ship.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | PASS | Strict TS (`strict: true`), ESLint, no dead code policy, single-purpose modules; no premature abstraction — features compose primitives, abstractions extracted on 3rd repetition. |
| II. Testing Standards | PASS | Jest + RTL for components, msw for API contracts, Zod schemas double as runtime contract tests; ≥80% line coverage on new code; test names describe behavior. Test-first for new screens. |
| III. UX Consistency | PASS | Design tokens (`mobile/src/theme/`) sourced from wireframe; no raw values in styles; reusable loading/empty/error patterns; pt-BR copy module; mobile-first; back-button predictable via Expo Router. Accessibility: every interactive element gets `accessibilityLabel`; tested with screen reader on iOS + Android. |
| IV. Performance | PASS | Route-based code splitting via Expo Router; bundle audit per route < 150 kB gzip on web; images via `expo-image` (WebP/AVIF); explicit width/height; TanStack Query dedups + caches → avoids N+1 from naive re-renders. |

No violations. No `Complexity Tracking` rows needed.

**Backend monolith principle**: Untouched. New client only consumes existing `/v1/*` routes through NGINX.

## Project Structure

### Documentation (this feature)

```text
specs/012-expo-rn-frontend/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── design-tokens.md
│   ├── navigation.md
│   ├── api-client.md
│   └── components.md
├── checklists/
│   └── requirements.md
└── tasks.md          # created by /speckit-tasks
```

### Source Code (repository root)

```text
mobile/                         # NEW — React Native + Expo app
├── app/                           # Expo Router routes (file-based)
│   ├── (auth)/
│   │   ├── login.tsx
│   │   ├── register.tsx
│   │   ├── verify-email.tsx
│   │   ├── password-reset.tsx
│   │   └── _layout.tsx
│   ├── (onboarding)/
│   │   ├── welcome.tsx
│   │   ├── interests.tsx
│   │   └── _layout.tsx
│   ├── (tabs)/
│   │   ├── trilha/
│   │   ├── redacao/
│   │   ├── simulado/
│   │   ├── perfil/
│   │   └── _layout.tsx            # bottom-nav + TopStatusBar host
│   ├── +not-found.tsx
│   └── _layout.tsx                # root: providers, theme, fonts, auth gate
├── src/
│   ├── theme/                     # design tokens (colors/spacing/typography/radius/shadow)
│   ├── components/                # shared UI primitives (Button, Card, EmptyState, MascotPlaceholder, TopStatusBar, BottomNav, FormField, OtpInput, ...)
│   ├── features/                  # vertical feature slices
│   │   ├── auth/                  # api, hooks, schemas, components
│   │   ├── onboarding/
│   │   ├── trilha/
│   │   ├── redacao/
│   │   ├── simulado/
│   │   └── perfil/
│   ├── lib/
│   │   ├── api/                   # ky/fetch client, refresh interceptor, error envelope
│   │   ├── auth/                  # tokenStore (expo-secure-store), session machine
│   │   ├── query/                 # TanStack Query client config
│   │   ├── analytics/             # noop in v1
│   │   └── i18n/                  # pt-BR copy
│   ├── stores/                    # zustand stores (session, ui flags)
│   └── types/                     # zod schemas + inferred TS types
├── assets/
│   ├── fonts/                     # Architects Daughter, Caveat, Caveat Brush, Patrick Hand, Kalam
│   └── images/                    # mascot placeholders, illustrations
├── tests/
│   ├── components/                # RTL component tests
│   ├── features/                  # feature-level hook + flow tests (msw)
│   └── setup.ts
├── app.json                       # Expo config
├── babel.config.js
├── metro.config.js
├── tsconfig.json
├── package.json
└── README.md

mobile/                            # EXISTING KMP tree — frozen, removed after P1+P2 parity (FR-017)
backend/                           # unchanged
infra/                             # unchanged
specs/012-expo-rn-frontend/        # this feature
```

**Structure Decision**: New top-level `mobile/` directory. Coexists with `mobile/` (KMP) until the new app reaches P1 + P2 parity; then `mobile/` is removed in a follow-up commit (FR-017). File-based routing via Expo Router groups (`(auth)`, `(onboarding)`, `(tabs)`) maps directly to the wireframe's flow sections. Feature-sliced source layout (`src/features/<domain>/`) keeps each domain's API, schemas, hooks, and UI co-located — agent-friendly and easy to delete.

## Complexity Tracking

> No constitution violations. Section intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
