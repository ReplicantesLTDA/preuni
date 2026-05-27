# Phase 0 Research — 012-expo-rn-frontend

**Date**: 2026-05-26
**Status**: complete — no NEEDS CLARIFICATION outstanding.

This document records the decisions that resolve every open question before implementation. Spec inputs and four user-confirmed choices (delete KMP, multiplatform iOS+Android+Web, wireframe-driven scope, Expo Router + Zustand + TanStack Query + Zod) frame the work; the items below detail the supporting choices.

---

## 1. Framework & runtime

- **Decision**: React Native 0.76+ on Expo SDK 52+, TypeScript 5 (`strict: true`).
- **Rationale**: Expo SDK 52 is the first to ship the New Architecture (Fabric + TurboModules) on by default and React Native 0.76 web parity via `react-native-web`. Expo Go covers iOS + Android dev; `expo export --platform web` produces the web build. Strongest agent/tooling ecosystem of the multiplatform options (per user motivation).
- **Alternatives considered**:
  - **Bare React Native** — rejected: loses Expo Go fast-feedback loop (FR-016, SC-006).
  - **Flutter** — rejected: weaker LLM-tooling ecosystem (against the explicit motivation for switching).
  - **Continue KMP** — rejected by user prior to spec.

## 2. Navigation

- **Decision**: **Expo Router** (file-based, on top of React Navigation), with route groups for `(auth)`, `(onboarding)`, `(tabs)`.
- **Rationale**: File-based routing maps 1:1 to the wireframe flow, encodes auth gating at the layout level, supports typed routes (`typedRoutes: true`), and is the path Expo officially recommends. Universal links + web routing come free.
- **Alternatives considered**:
  - **React Navigation directly** — rejected: more boilerplate, no file-based discovery, weaker web parity.

## 3. State + data

- **Decision**: **Zustand** for client/session UI state; **TanStack Query v5** for server cache + mutations; **Zod** for runtime schema validation at the API boundary.
- **Rationale**: User-selected. Matches the spec's "agent ships a screen in under a day" goal (SC-003) — minimal ceremony, colocated logic, schema-first contracts. TanStack Query handles refetch/invalidation so the FR-010 "Trilha updates after activity" requirement is one `invalidateQueries` call.
- **Alternatives considered**:
  - **Redux Toolkit + RTK Query** — rejected: heavier, more boilerplate per screen.
  - **React Context + useReducer** — rejected: scaling pain past ~5 features, no cache.
  - **Yup / class-validator** — rejected: Zod has the strongest TS inference story.

## 4. Auth & session

- **Decision**: `expo-secure-store` for `access_token` + `refresh_token`. Single `apiClient` (typed `fetch` wrapper) with an interceptor that, on 401, calls `/v1/auth/refresh` once, retries the original request, and on failure clears tokens + routes to `/login`. Session machine lives in a Zustand store (`status: 'loading' | 'authed' | 'anon'`).
- **Rationale**: Mirrors the current KMP `TokenStore`/`ApiClient` behavior so the backend (and the `{error:{...}}` envelope) needs no changes (FR-018, FR-019, FR-020). SecureStore uses Keychain (iOS) / EncryptedSharedPreferences (Android); on web it falls back to `localStorage` — acceptable for v1 (web is dev-review surface, not production).
- **Alternatives considered**:
  - **AsyncStorage** — rejected: not encrypted, would weaken token storage on mobile.
  - **Cookie-based session via web `fetch` credentials** — rejected: doesn't translate to RN mobile.

## 5. Design tokens (sourced from `wireframe.html`)

- **Decision**: Tokens live in `mobile/src/theme/tokens.ts` and are the *only* source of color/spacing/typography/radius/shadow values. Components consume tokens via a thin `useTheme()` hook. No raw hex/px in component styles (Constitution III).
- **Palette extracted from wireframe**:
  - `paper.0` `#fafaf6` (canvas), `paper.1` `#f3f1ea` (raised surfaces, cards)
  - `ink.0` `#1a1a1a` (primary text), `ink.1` `#2a2a2a` (headings), `ink.2` `#4a4a4a` (body), `ink.muted` `#8a8a8a` (placeholders)
  - `accent` `#ff8c42` (CTAs, streak, mascot accents)
  - `success` `#3a8a3a`, `danger` `#b03030`
- **Typography**: hand-drawn family stack — `Caveat Brush` (display/H1), `Patrick Hand` (H2/H3), `Architects Daughter` (body), `Kalam` (numbers + metrics), `Caveat` (accents). Loaded via `expo-font` at root layout. Fallback to system sans for first paint.
- **Spacing scale**: `0, 4, 8, 12, 16, 20, 24, 32, 40, 56` (px → `dp` on native, `px` on web).
- **Radius**: `sm: 8`, `md: 12`, `lg: 20`, `pill: 999`.
- **Shadow / elevation**: single token `card` (small offset, low blur, 8% black) — matches the wireframe's flat-paper feel.
- **Rationale**: Tokens captured directly from the wireframe satisfy FR-001 / FR-004 and Constitution III. Hand-drawn fonts are part of the design identity — not optional.
- **Alternatives considered**:
  - **Use Material Design 3 defaults** — rejected: would erase the wireframe's identity.

## 6. Icons & illustrations

- **Decision**: `@expo/vector-icons` (Feather + Material Community subset) for tab and inline icons. Mascot uses a static asset via `expo-image` with an SVG placeholder fallback (`react-native-svg`), matching the existing `MascotPlaceholder` contract from feature 011.
- **Rationale**: Single icon library keeps bundle small (tree-shaken) and consistent. Mascot is a deliberate placeholder until artwork ships.

## 7. Testing strategy

- **Decision**: Three tiers — (a) **Component tests**: Jest + `@testing-library/react-native` for primitives + screens; (b) **Feature tests**: full flow tests with `msw` mocking the API at fetch boundary (no mocking of our own modules); (c) **Smoke E2E**: Detox on Android emulator for auth + Trilha happy path, post-MVP.
- **Rationale**: Mirrors Constitution II: unit + integration + real-dependency at boundary. Zod schemas double as contract tests — a backend response that fails parse fails the test.
- **Coverage**: ≥ 80% lines on `mobile/src/` (Constitution II floor).
- **Alternatives considered**:
  - **Vitest** — rejected: weaker RN support than Jest.
  - **Mock fetch globally** — rejected: hides URL/contract bugs that `msw` catches.

## 8. CI

- **Decision**: GitHub Actions job `mobile-ci` running, in order: `pnpm install --frozen-lockfile` → `pnpm tsc --noEmit` → `pnpm lint` → `pnpm test --coverage` → `pnpm expo export --platform web` (smoke build). Coverage threshold enforced. Lighthouse CI on the exported web bundle for LCP/INP/CLS (Constitution IV).
- **Rationale**: Quality Gates require green lint/type/tests, coverage non-regression, and Lighthouse no-regression on LCP/INP/CLS. Web export catches Metro-vs-web divergence early.
- **Package manager**: **pnpm** — faster installs, strict hoisting catches phantom deps; aligns with repo's preference for explicit dependencies.

## 9. Coexistence + cutover with `mobile/`

- **Decision**: Both `mobile/` (KMP, frozen) and `mobile/` (active) live in the repo until the new app reaches P1 + P2 parity (FR-017). During coexistence, **only** `mobile/` is shipped to testers. CI for `mobile/` is reduced to "build only" (no new features). Cutover commit deletes `mobile/`, `mobile/shared`, `mobile/androidApp`, `mobile/iosApp`, `mobile/webApp`, and updates `CLAUDE.md` + Makefile.
- **Rationale**: Lets the team validate parity without a hard switch; keeps `git log` clean (one deletion commit instead of churn).

## 10. Backend interface inventory (no changes)

Endpoints the client must consume (already exist after spec 010 + the auth fixes on 011):

| Endpoint | Notes |
|----------|-------|
| `POST /v1/auth/register` | email + password + displayName → session |
| `POST /v1/auth/login` | session |
| `POST /v1/auth/refresh` | refresh-token grant |
| `POST /v1/auth/logout` | invalidates refresh token |
| `POST /v1/auth/email/verify` | OTP verify |
| `POST /v1/auth/email/verify-resend` | resend OTP (added in 011) |
| `POST /v1/auth/otp/request` | OTP-only login flow |
| `POST /v1/auth/otp/verify` | |
| `POST /v1/auth/password/reset/request` | |
| `POST /v1/auth/password/reset/confirm` | |
| `GET /v1/students/me` | profile |
| `PATCH /v1/students/me` | display name, etc. |
| `DELETE /v1/students/me` | account deletion |
| `PUT /v1/students/me/avatar` | avatar upload (changed from GET upload-url in 011) |
| `POST /v1/students/me/avatar/confirm` | confirm uploaded avatar |
| `PATCH /v1/students/me/onboarding` | onboarding completion |
| `GET /v1/students/me/data-export` | LGPD data export |
| ...content/learning/redacao/simulado endpoints | as exposed |

All responses use the `{ error: { code, field, message } }` envelope (FR-019).

## 11. Out of scope (recorded for clarity)

- Native modules requiring a custom dev client (use Expo Go only in v1).
- Push notifications (FCM/APNs) — deferred.
- Offline write queue — deferred; reads can be stale (TanStack Query default).
- Deep links beyond OAuth-style verify callbacks — deferred.
- App Store / Play Store submission — deferred to a follow-up effort.

---

**Outcome**: All Phase 0 unknowns resolved. Proceed to Phase 1.
