---
description: "Task list for 012-expo-rn-frontend"
---

# Tasks: Migrate frontend to React Native + Expo (012)

**Input**: Design documents in `/Users/dwbessa/projects/preuni/specs/012-expo-rn-frontend/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/*, quickstart.md

**Tests**: Test tasks are INCLUDED. Constitution II (NON-NEGOTIABLE) mandates test-first for new features; this whole feature is greenfield code.

**Organization**: Tasks grouped by user story (US1–US5). Phases 1 (Setup) and 2 (Foundational) block all stories; Phases 3–7 are stories in priority order; Phase 8 is polish + cutover.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallelizable — different files, no dependencies on incomplete tasks
- **[Story]**: `[US1]`–`[US5]`; absent on Setup/Foundational/Polish
- Paths absolute or anchored at `mobile/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Bootstrap the `mobile/` Expo + TypeScript project.

- [X] T001 Create `mobile/` directory at repo root with `pnpm create expo-app@latest . --template blank-typescript` (Expo SDK 52+); commit the generated skeleton.
- [X] T002 Update `mobile/package.json` to add deps: `expo-router`, `react-native-web`, `@tanstack/react-query`, `zustand`, `zod`, `expo-secure-store`, `expo-image`, `expo-image-picker`, `expo-font`, `expo-localization`, `@expo/vector-icons`, `react-native-reanimated`, `react-native-gesture-handler`, `react-native-safe-area-context`, `react-native-svg`. Dev deps: `@testing-library/react-native`, `@testing-library/jest-native`, `jest-expo`, `msw`, `@types/jest`, `eslint`, `eslint-config-expo`, `@typescript-eslint/{parser,eslint-plugin}`, `prettier`.
- [X] T003 [P] Configure `mobile/tsconfig.json` with `"strict": true`, `"noUncheckedIndexedAccess": true`, path alias `"@/*": ["./src/*"]`, and `"jsx": "react-jsx"`.
- [X] T004 [P] Configure `mobile/app.json` for Expo Router (`"plugins": ["expo-router"]`), `"scheme": "preuni"`, `"experiments": { "typedRoutes": true }`, iOS/Android bundle ids `com.preuni.app`, web `output: "static"`.
- [X] T005 [P] Add `mobile/babel.config.js` with `babel-preset-expo` + `expo-router/babel` + `react-native-reanimated/plugin` (last position required).
- [X] T006 [P] Configure `mobile/metro.config.js` extending `@expo/metro-config` with SVG transformer (`react-native-svg-transformer`).
- [X] T007 [P] Add `mobile/.eslintrc.cjs` extending `eslint-config-expo` + TypeScript; custom rule forbidding raw `#[0-9a-fA-F]{3,6}` in `src/**` (Constitution III). Add `mobile/.prettierrc`.
- [X] T008 [P] Add `mobile/jest.config.js` (preset `jest-expo`), test setup at `mobile/tests/setup.ts` (jest-native matchers, msw server lifecycle).
- [X] T009 [P] Create `mobile/.env.example` with `EXPO_PUBLIC_API_BASE_URL=http://localhost:8080`; add `.env.local` to `.gitignore`.
- [X] T010 [P] Add scripts to `mobile/package.json`: `"start"`, `"ios"`, `"android"`, `"web"`, `"test"`, `"test:watch"`, `"lint"`, `"typecheck"`, `"build:web": "expo export --platform web"`.
- [X] T011 [P] Update root `Makefile` with `mobile-install`, `mobile-start`, `mobile-test`, `mobile-web` targets that `cd mobile && pnpm …`.
- [X] T012 [P] Add GitHub Actions workflow `.github/workflows/mobile-ci.yml`: install (pnpm), `pnpm typecheck`, `pnpm lint`, `pnpm test --coverage` (≥80% gate), `pnpm build:web`.
- [X] T013 [P] Download hand-drawn fonts (Caveat Brush, Patrick Hand, Architects Daughter, Kalam Regular+Bold, Caveat Regular+Bold) into `mobile/assets/fonts/`.
- [X] T014 [P] Add `mobile/README.md` mirroring `specs/012-expo-rn-frontend/quickstart.md` (install, run, test, web build).

**Checkpoint**: `pnpm start` boots a blank Expo app on iOS/Android/Web.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Design tokens, theme provider, API client, auth store, routing shell, Query/Provider wiring. **No user story work begins until this completes.**

### Design system

- [X] T015 [P] Implement `mobile/src/theme/tokens.ts` per `contracts/design-tokens.md` (palette, space, radius, font, size, shadow). Export `tokens` + `Tokens` type.
- [X] T016 [P] Implement `mobile/src/theme/ThemeProvider.tsx` exposing `useTheme()` returning tokens. Wrap children with React Context.
- [X] T017 [P] Implement `mobile/src/theme/useFonts.ts` calling `useFonts({ CaveatBrush, PatrickHand, ArchitectsDaughter, Kalam, KalamBold, Caveat, CaveatBold })` from `expo-font`; export `fontsReady` boolean.
- [X] T018 [P] Theme unit test `mobile/tests/theme/tokens.test.ts`: assert palette keys + space scale + font names exactly match `contracts/design-tokens.md`.

### Zod schemas + types

- [X] T019 [P] `mobile/src/types/api.ts` — `ApiErrorEnvelopeSchema` + inferred `ApiErrorEnvelope`; helper `parseApiError(text: string): ApiErrorEnvelope | null`.
- [X] T020 [P] `mobile/src/types/auth.ts` — `SessionSchema`, `RegisterRequestSchema`, `LoginRequestSchema`, `VerifyEmailRequestSchema`, `ResendVerifyRequestSchema`, `PasswordResetRequest/ConfirmSchema`.
- [X] T021 [P] `mobile/src/types/student.ts` — `StudentSchema` per data-model.md.

### API client + error model

- [X] T022 `mobile/src/lib/api/errors.ts` — `AppError` discriminated union (Unauthorized, Forbidden, Conflict, Validation, Unknown) + `toAppError(response, parsedBody)` mapping. Unit tests `tests/lib/errors.test.ts` covering each branch.
- [X] T023 `mobile/src/lib/api/client.ts` — `createApiClient({ baseUrl, getAccessToken, onUnauthorized })` returning `request<T>({ method, path, body, schema, auth, signal })`. JSON serialization, header injection, schema validation on 2xx, error mapping on non-2xx via T022.
- [X] T024 `mobile/src/lib/api/refreshInterceptor.ts` — single-flight refresh: on 401, call `/v1/auth/refresh`, persist new tokens, retry once; on failure clear tokens + emit `onUnauthorized`. Wired through `createApiClient`.
- [X] T025 [P] Test `mobile/tests/lib/api/client.test.ts` (msw): success path validates schema; 401 → refresh OK → retry returns 200; 401 → refresh fails → AppError.Unauthorized + onUnauthorized called; 422 → AppError.Validation with field/message.
- [X] T026 [P] Test `mobile/tests/lib/api/envelope.test.ts`: unparseable body falls back to AppError.Unknown with default pt-BR message.

### Token store + session machine

- [X] T027 `mobile/src/lib/auth/tokenStore.ts` — `expo-secure-store` wrapper with `get`, `set`, `clear` for `access_token`, `refresh_token`, `access_token_expires_at`. Web fallback to `localStorage`.
- [X] T028 `mobile/src/stores/sessionStore.ts` — Zustand store `{ status: 'loading' | 'authed' | 'anon', student: Student | null, setSession, clearSession }`.
- [X] T029 `mobile/src/lib/auth/bootstrap.ts` — on app start: hydrate tokens from store; if present and unexpired → `status: 'authed'` and fetch `/v1/students/me`; if missing/expired → attempt refresh once; else → `'anon'`. Unit-tested with msw.

### Query client + providers

- [X] T030 [P] `mobile/src/lib/query/client.ts` — `QueryClient` with `{ retry: 1, staleTime: 30_000 }` defaults; mutations `retry: 0`.
- [X] T031 [P] `mobile/src/lib/query/keys.ts` — namespaced query-key helpers: `qk.studentMe()`, `qk.trilhaHome(trackId)`, `qk.redacao*`, `qk.simulado*`.
- [X] T032 `mobile/src/lib/api/context.tsx` — React context providing the `ApiClient` instance to hooks (uses tokenStore + sessionStore + queryClient).

### Routing shell

- [X] T033 `mobile/app/_layout.tsx` — root layout: `SafeAreaProvider`, `GestureHandlerRootView`, `QueryClientProvider`, `ThemeProvider`, `ApiContextProvider`. Calls `useFonts` + bootstrap; renders `<Splash />` until both ready.
- [X] T034 `mobile/app/index.tsx` — route dispatcher: reads `useSessionStore`, redirects to `/(auth)/welcome`, `/(onboarding)/welcome`, or `/(tabs)/trilha`.
- [X] T035 [P] `mobile/app/+not-found.tsx` — styled 404 using EmptyState + a back-to-home CTA.
- [X] T036 `mobile/app/(auth)/_layout.tsx` — public layout; redirects to `/(tabs)/trilha` when `status === 'authed'`.
- [X] T037 `mobile/app/(onboarding)/_layout.tsx` — gated layout per `contracts/navigation.md`.
- [X] T038 `mobile/app/(tabs)/_layout.tsx` — tab bar host using `<Tabs>` from `expo-router`; mounts `TopStatusBar` and `BottomNavigation`. Gated to `authed + onboardingCompleted`.

### Shared UI primitives (used by every story)

- [X] T039 [P] `mobile/src/components/Button.tsx` per components contract + test `tests/components/Button.test.tsx` covering all variants/sizes + loading + disabled.
- [X] T040 [P] `mobile/src/components/Card.tsx` + test.
- [X] T041 [P] `mobile/src/components/FormField.tsx` + test (label, error, helper).
- [X] T042 [P] `mobile/src/components/OtpInput.tsx` + test (auto-advance, paste).
- [X] T043 [P] `mobile/src/components/EmptyState.tsx` + test.
- [X] T044 [P] `mobile/src/components/MascotPlaceholder.tsx` (SVG variants: thinking, reading, cheering) + test.
- [X] T045 [P] `mobile/src/components/TopStatusBar.tsx` per components contract + test.
- [X] T046 [P] `mobile/src/components/BottomNavigation.tsx` (delegates to `expo-router` `<Tabs.Screen>` styling) + test.
- [X] T047 [P] `mobile/src/components/Skeleton.tsx`, `ScreenLoader.tsx`, `ErrorState.tsx`, `Toast.tsx` + tests.

### i18n / copy

- [X] T048 [P] `mobile/src/lib/i18n/pt-BR.ts` — copy keys for shared chrome (errors, retry, cancel, save). Helper `t(key)` returning typed string.

**Checkpoint**: Blank tabs render at `/(tabs)/trilha` for a manually injected `authed` state; CI green (lint + typecheck + tests).

---

## Phase 3: User Story 2 — Auth + Onboarding (Priority: P1) 🎯 MVP-1

**Order rationale**: US2 is P1 and is a hard dependency for US1 in practice (a user must be signed-in to see Trilha). Built first so demos start from a clean install.

**Goal**: Register → OTP → verify (with resend) → onboarding → land on `(tabs)/trilha` already authed. Login + logout + password reset functional.

**Independent Test**: Clean install → register `qa+rn@preuni.com` → enter OTP (from local mail dev sender) → finish onboarding → arrive on Trilha home.

### Tests for US2

- [X] T049 [P] [US2] Contract test `mobile/tests/features/auth/register.contract.test.ts` (msw): POST `/v1/auth/register` returns session schema; client persists tokens; session store transitions to `authed`.
- [X] T050 [P] [US2] Contract test `mobile/tests/features/auth/login.contract.test.ts`.
- [X] T051 [P] [US2] Contract test `mobile/tests/features/auth/refresh.contract.test.ts` covering 401 → refresh OK → retry.
- [X] T052 [P] [US2] Contract test `mobile/tests/features/auth/verifyEmail.contract.test.ts` — verify + verify-resend both 204 paths; conflict (already verified) → AppError.Conflict.
- [X] T053 [P] [US2] Contract test `mobile/tests/features/auth/passwordReset.contract.test.ts` — request + confirm.
- [X] T054 [P] [US2] Integration test `mobile/tests/features/auth/flow.integration.test.tsx`: register → verify → onboarding → Trilha; assert no extra login step (FR-007).
- [X] T055 [P] [US2] Integration test `mobile/tests/features/auth/expiredRefresh.integration.test.tsx`: open app with expired refresh → routed to `/(auth)/login` (FR-020).

### Auth feature module

- [X] T056 [P] [US2] `mobile/src/features/auth/api.ts` — `register`, `login`, `logout`, `verifyEmail`, `resendVerificationOtp`, `requestOtpLogin`, `verifyOtpLogin`, `requestPasswordReset`, `confirmPasswordReset`. Each typed via Zod schemas from `src/types/auth.ts`.
- [X] T057 [P] [US2] `mobile/src/features/auth/hooks.ts` — TanStack `useMutation` hooks for each, invalidating `qk.studentMe()` on success.
- [X] T058 [P] [US2] `mobile/src/features/auth/validation.ts` — Zod form schemas (email, password ≥8 + letter + digit, otp `/^\d{6}$/`, displayName 2–64).

### Auth screens

- [X] T059 [US2] `mobile/app/(auth)/welcome.tsx` — MascotPlaceholder + "Entrar" + "Criar conta" CTAs.
- [X] T060 [US2] `mobile/app/(auth)/register.tsx` — FormField × 3, submit calls `useRegister`; on success → `/(auth)/verify-email?email=…`.
- [X] T061 [US2] `mobile/app/(auth)/login.tsx` — FormField × 2, submit + "Esqueci minha senha" link + "Entrar com código" link.
- [X] T062 [US2] `mobile/app/(auth)/verify-email.tsx` — OtpInput + resend button (calls `useResendVerificationOtp`, 60s cooldown) + auto-login on success (FR-007).
- [X] T063 [US2] `mobile/app/(auth)/otp-login.tsx` — email request → OtpInput → verify.
- [X] T064 [US2] `mobile/app/(auth)/password-reset.tsx` — request + confirm screens behind a stack inside the route.

### Onboarding

- [X] T065 [P] [US2] `mobile/src/features/onboarding/api.ts` — `PATCH /v1/students/me/onboarding`.
- [X] T066 [P] [US2] `mobile/src/features/onboarding/hooks.ts` — `useCompleteOnboarding` mutation; invalidates `qk.studentMe()`.
- [X] T067 [US2] `mobile/app/(onboarding)/welcome.tsx` — mascot + intro copy.
- [X] T068 [US2] `mobile/app/(onboarding)/profile.tsx` — confirm displayName + optional avatar pick (via expo-image-picker; upload via FR-019/avatar PUT).
- [X] T069 [US2] `mobile/app/(onboarding)/interests.tsx` — multi-select chips for subject interests; final submit triggers `useCompleteOnboarding` → `/(tabs)/trilha`.

### Error + edge

- [X] T070 [US2] Render AppError.Validation as inline field error via FormField; AppError.Conflict / Unknown via Toast or ErrorState. Tests asserting pt-BR strings.
- [X] T071 [US2] Resend rate-limit feedback: on 429 surface "Aguarde antes de reenviar"; cooldown timer driven by local state, restarts on success.

**Checkpoint**: US2 end-to-end demonstrable on a real phone via Expo Go.

---

## Phase 4: User Story 1 — Trilha Core Learning Loop (Priority: P1) 🎯 MVP

**Goal**: Signed-in student sees Trilha home (streak, XP, readiness, next activity) styled per wireframe; completing an activity updates state without manual refresh.

**Independent Test**: Sign in as seeded admin → land on Trilha → see streak/XP/next-step → tap next → finish → Trilha updates.

### Tests for US1

- [X] T072 [P] [US1] Contract test `mobile/tests/features/trilha/home.contract.test.ts` (msw): `GET /v1/students/me` + Trilha endpoints return schemas; UI renders expected fields.
- [ ] T073 [P] [US1] Integration test `mobile/tests/features/trilha/completion.integration.test.tsx`: complete activity → `qk.studentMe()` + `qk.trilhaHome()` invalidated → re-render shows updated values (FR-010).
- [X] T074 [P] [US1] Offline-state test `mobile/tests/features/trilha/offline.integration.test.tsx`: simulate fetch failure → Trilha shows ErrorState (not blank). Edge case from spec.

### Trilha feature

- [X] T075 [P] [US1] `mobile/src/features/trilha/api.ts` — `getMe`, `getTrilhaHome`, `getTrack`, `completeActivity` (use existing backend routes; if a route is not yet exposed, gate the call behind a typed feature flag in `src/lib/api/featureFlags.ts`).
- [X] T076 [P] [US1] `mobile/src/features/trilha/hooks.ts` — `useTrilhaHome`, `useCompleteActivity` (`onSuccess` invalidates `qk.studentMe()` + `qk.trilhaHome()`).
- [X] T077 [P] [US1] `mobile/src/features/trilha/components/TrackPath.tsx` — adapts KMP feature 006 subject-track-path into the wireframe's notebook style using `react-native-svg`. Unit-tested.
- [X] T078 [P] [US1] `mobile/src/features/trilha/components/NextActivityCard.tsx` — styled Card with mascot + CTA.

### Trilha screens

- [X] T079 [US1] `mobile/app/(tabs)/trilha/index.tsx` — TopStatusBar (already in tab layout) + Trilha home: streak summary card, NextActivityCard, TrackPath, EmptyState fallback when no track.
- [ ] T080 [US1] `mobile/app/(tabs)/trilha/[trackId].tsx` — track detail with activity list + completion CTA → calls `useCompleteActivity`.
- [X] T081 [US1] Loading + error states via Skeleton + ErrorState per Constitution III.
- [X] T082 [US1] Telemetry-free analytics noop in `mobile/src/lib/analytics/` (no PII, no network).

**Checkpoint**: MVP complete. US1 + US2 demonstrable end-to-end on iOS, Android, and web (Expo dev).

---

## Phase 5: User Story 3 — Profile Management (Priority: P2)

**Goal**: Edit display name, change avatar, change email (re-verify), change password, sign out.

**Independent Test**: Sign in → open Perfil → change each field → confirm propagation + sign-out routes to `/(auth)/welcome`.

### Tests for US3

- [X] T083 [P] [US3] Contract test `mobile/tests/features/perfil/me.contract.test.ts` — GET/PATCH `/v1/students/me`.
- [X] T084 [P] [US3] Contract test `mobile/tests/features/perfil/avatar.contract.test.ts` — PUT `/v1/students/me/avatar` + confirm.
- [X] T085 [P] [US3] Contract test `mobile/tests/features/perfil/changeEmail.contract.test.ts` — change email → re-verify flow.
- [X] T086 [P] [US3] Contract test `mobile/tests/features/perfil/changePassword.contract.test.ts`.
- [X] T087 [P] [US3] Integration test `mobile/tests/features/perfil/signOut.integration.test.tsx` — sign out → session `anon` → reopen routes to `/(auth)/welcome` (acceptance scenario 4).
- [X] T088 [P] [US3] Integration test `mobile/tests/features/perfil/avatarRollback.integration.test.tsx` — avatar upload fails mid-way → previous avatar preserved (edge case).

### Perfil feature

- [X] T089 [P] [US3] `mobile/src/features/perfil/api.ts` — `updateMe`, `uploadAvatar` (multipart PUT), `confirmAvatar`, `changeEmail`, `changePassword`, `deleteMe`, `requestDataExport`.
- [X] T090 [P] [US3] `mobile/src/features/perfil/hooks.ts` — mutations + invalidations.

### Perfil screens

- [X] T091 [US3] `mobile/app/(tabs)/perfil/index.tsx` — avatar + display name + XP/streak/readiness summary + menu links.
- [X] T092 [US3] `mobile/app/(tabs)/perfil/edit.tsx` — display name edit + avatar picker (expo-image-picker) → upload via PUT, then confirm.
- [X] T093 [US3] `mobile/app/(tabs)/perfil/change-email.tsx` — new email entry → triggers verification on new address; old email keeps working until verified.
- [X] T094 [US3] `mobile/app/(tabs)/perfil/change-password.tsx` — current + new + confirm.
- [X] T095 [US3] `mobile/app/(tabs)/perfil/data-export.tsx` — request + show download link when ready.
- [X] T096 [US3] Sign-out action in Perfil home — calls `logout` mutation, clears tokens, session → `anon`, router back to `/(auth)/welcome`.
- [X] T097 [US3] Account deletion gated behind confirm modal calling `DELETE /v1/students/me`.

**Checkpoint**: US3 functional; no regressions vs KMP profile flow (SC-007).

---

## Phase 6: User Story 4 — Redação + Simulado Tabs (Priority: P2)

**Goal**: Redação and Simulado tab homes render with wireframe styling; surface existing backend data or styled empty states.

**Independent Test**: From Trilha → tap Redação → see styled home (history or empty state) → back → tap Simulado → same.

### Redação

- [ ] T098 [P] [US4] Contract test `mobile/tests/features/redacao/list.contract.test.ts`.
- [ ] T099 [P] [US4] `mobile/src/features/redacao/api.ts` + `hooks.ts` (list prompts, submission detail, submit essay).
- [X] T100 [US4] `mobile/app/(tabs)/redacao/index.tsx` — prompt list / history + EmptyState if none.
- [ ] T101 [US4] `mobile/app/(tabs)/redacao/[promptId].tsx` — prompt detail + writing surface (long-form TextInput inside a Card).

### Simulado

- [ ] T102 [P] [US4] Contract test `mobile/tests/features/simulado/list.contract.test.ts`.
- [ ] T103 [P] [US4] `mobile/src/features/simulado/api.ts` + `hooks.ts`.
- [X] T104 [US4] `mobile/app/(tabs)/simulado/index.tsx` — attempt history + start-new CTA + EmptyState fallback.

**Checkpoint**: US4 tab homes shipped with styled content/empty states.

---

## Phase 7: User Story 5 — Multiplatform (Priority: P3)

**Goal**: One codebase produces working iOS, Android, and Web builds; P1 user stories pass on each.

**Independent Test**: Run `pnpm ios`, `pnpm android`, `pnpm web` from a fresh checkout; complete US1 on each.

- [X] T105 [P] [US5] Verify SecureStore fallback path on web (localStorage) in `tokenStore.ts`; test `tests/lib/auth/tokenStoreWeb.test.ts`.
- [X] T106 [P] [US5] Web responsive audit at 360 / 768 / 1280px — record screenshots into `specs/012-expo-rn-frontend/screens/` and adjust any clipping.
- [X] T107 [US5] Lighthouse CI config (`mobile/lighthouserc.json`) running against the web export; thresholds LCP ≤ 2.5s, INP ≤ 200ms, CLS ≤ 0.1 (Constitution IV).
- [X] T108 [US5] iOS-specific: `Info.plist` photo library usage description for avatar picker (configured via `app.json`).
- [X] T109 [US5] Android-specific: confirm Reanimated worklet config + back-button predictable navigation.
- [X] T110 [US5] `pnpm build:web` runs in CI and uploads artifact (workflow update T012).

**Checkpoint**: All three platforms green in CI; P1 stories run on each.

---

## Phase 8: Polish, Cutover, and Cross-Cutting

- [X] T111 [P] Accessibility sweep: every interactive primitive has `accessibilityRole` + `accessibilityLabel`; run iOS VoiceOver + TalkBack on the auth + Trilha flows; record findings in `specs/012-expo-rn-frontend/a11y.md`.
- [X] T112 [P] Bundle audit per route via `expo-atlas`; assert no route chunk > 150 kB gzip on web (Constitution IV). Record `bundle-report.md`.
- [ ] T113 [P] Add Detox smoke project at `mobile/e2e/` with Android emulator config and one happy-path spec (auth → trilha) — post-MVP gate.
- [X] T114 [P] Performance pass: `expo-image` everywhere (no raw `<Image>` from `react-native`); explicit width/height on all media; verify no CLS regression.
- [X] T115 [P] Run `quickstart.md` end-to-end and tick the acceptance checkboxes; record outcomes.
- [X] T116 [P] Update `README.md` at repo root to add `mobile` quick-run table; mark KMP `mobile/` as frozen.
- [X] T117 [P] Update `.specify/memory/constitution.md`? — NO unless a new violation is found; instead append a note to `CLAUDE.md` "Recent Changes" already auto-added by `update-agent-context.sh`.
- [X] T118 KMP freeze: add `mobile/FROZEN.md` and CI guard refusing new files under `mobile/shared/src/commonMain/` until cutover commit.
- [X] T119 **Cutover commit** (only after US1 + US2 + US3 + US4 are demonstrably at parity — FR-017): delete `mobile/shared`, `mobile/androidApp`, `mobile/iosApp`, `mobile/webApp`, the KMP root build files, and any Gradle/KMP CI workflows. Update `CLAUDE.md` and `Makefile` to drop KMP commands. Single commit; PR titled `chore(mobile): retire KMP frontend in favor of mobile`.
- [X] T120 Security review: audit token storage paths, avatar upload MIME enforcement, refresh single-flight, and error-envelope leakage; record in `specs/012-expo-rn-frontend/security-review.md`.
- [ ] T121 Run `pnpm tsc --noEmit && pnpm lint && pnpm test --coverage` and verify ≥80% line coverage on `mobile/src/` (Constitution II floor).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: no dependencies; T001 → T002 sequential, then T003–T014 in parallel.
- **Phase 2 (Foundational)**: depends on Phase 1. Within Phase 2, blocks the rest of the feature.
  - T015–T021 in parallel.
  - T022 → T023 → T024 sequential (API client + refresh interceptor).
  - T025–T026 in parallel after T024.
  - T027 → T028 → T029 sequential (auth store + bootstrap).
  - T030–T032 in parallel after T028.
  - T033–T038 sequential within layouts but each depends on T032.
  - T039–T047 in parallel (all isolated component files).
- **Phase 3 (US2)**: depends on Phase 2. Built **before** US1 because Trilha requires an authed session.
- **Phase 4 (US1)**: depends on Phase 2 and at minimum US2's session machinery (`useSessionStore` + tokenStore wired).
- **Phase 5 (US3)** and **Phase 6 (US4)**: depend on Phase 2; can run in parallel with each other after US1+US2 land.
- **Phase 7 (US5)**: depends on at least one user story shipping (US1).
- **Phase 8**: T118 + T119 (cutover) require US1–US4 demonstrably at parity (FR-017); T120 + T121 can run any time after Phase 7.

### User Story Dependencies

- **US2 → US1**: practical dependency (Trilha needs an authed user). Listed before US1 in execution order even though both are P1.
- **US3**: independent of US1/US4; shares `useSessionStore` only.
- **US4**: independent; shares only navigation + theme.
- **US5**: cross-cutting; verifies the others on three platforms.

### Within Each User Story

- Contract tests written first → fail → implement API client/hooks → tests pass.
- Hooks before screens.
- Each screen has at least one component or integration test before merge.
- Pt-BR copy + a11y labels added during screen implementation, not as polish.

### Parallel Opportunities

- All Setup `[P]` tasks (T003–T014).
- All Foundational primitives `[P]` (T039–T047) — separate files, no order.
- All Zod schemas `[P]` (T019–T021).
- Per story: all `[P]` contract tests + `[P]` API/hook modules run in parallel before screens.
- US3 + US4 implementable by two engineers in parallel after US1 + US2 land.

---

## Parallel Example — User Story 2 contract test wave

```bash
# After T048 checkpoint:
Task: "T049 [P] [US2] Contract test register"
Task: "T050 [P] [US2] Contract test login"
Task: "T051 [P] [US2] Contract test refresh"
Task: "T052 [P] [US2] Contract test verifyEmail + resend"
Task: "T053 [P] [US2] Contract test passwordReset"
```

## Parallel Example — Foundational primitives wave

```bash
# After T038 checkpoint:
Task: "T039 [P] Button + test"
Task: "T040 [P] Card + test"
Task: "T041 [P] FormField + test"
Task: "T042 [P] OtpInput + test"
Task: "T043 [P] EmptyState + test"
Task: "T044 [P] MascotPlaceholder + test"
Task: "T045 [P] TopStatusBar + test"
Task: "T046 [P] BottomNavigation + test"
Task: "T047 [P] Skeleton/ScreenLoader/ErrorState/Toast + tests"
```

---

## Implementation Strategy

### MVP First

1. Phase 1 + Phase 2 (Setup + Foundational).
2. Phase 3 (US2 — auth + onboarding).
3. Phase 4 (US1 — Trilha core loop).
4. **STOP + validate** with quickstart.md smoke tests on real iOS + Android devices via Expo Go and on Web via `pnpm web`.
5. Demo MVP.

### Incremental Delivery

1. MVP (US1 + US2) shipped to internal testers via Expo Go QR.
2. US3 (Perfil) — closes the parity gap with the KMP app's account features.
3. US4 (Redação + Simulado) — restores tab depth.
4. US5 (Multiplatform hardening + Lighthouse) — ensures Web parity.
5. **Cutover** (T119) — delete `mobile/` once parity is signed off.

### Parallel Team Strategy

- Engineer A: Setup + Foundational; then US2.
- Engineer B: tokens + primitives during Phase 2; then US1.
- After US1 + US2 land:
  - Engineer A → US3.
  - Engineer B → US4.
  - Either engineer → US5 + Polish.

---

## Notes

- `[P]` = different files, no incomplete-task dependency.
- `[US#]` traces a task back to its spec user story.
- Test-first per Constitution II (NON-NEGOTIABLE) — no implementation merges until the matching contract or integration test fails first.
- Commit boundary suggestion: one commit per `[US#]` sub-section (e.g., "feat(auth-rn): T056–T058 auth feature module").
- Stop at any **Checkpoint** to validate the story independently before moving on.
- Do not merge anything under `mobile/` (KMP) during this feature; cutover is a single deletion commit (T119).
- Total tasks: **121**. Estimated buckets — Setup: 14, Foundational: 34, US2: 23, US1: 11, US3: 15, US4: 7, US5: 6, Polish + cutover: 11.
