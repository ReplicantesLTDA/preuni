# Tasks: App Design Refresh — Core Flow

**Input**: Design documents from `specs/011-app-design-refresh/`
**Prerequisites**: `plan.md` (required), `spec.md` (required), plus `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tech**: Kotlin 2.1.20 + Compose Multiplatform 1.8.0, Material 3, Decompose 3.3.0, MVIKotlin 4.2.0
**Tests**: Not required by spec. Use manual visual review via web preview (`make run-web`). Add unit tests only if new pure logic is introduced.

## Format: `- [ ] T### [P?] [US#?] Description with file path`

- **[P]**: Can run in parallel (different files, no dependency on incomplete tasks)
- **[US#]**: User story label required for story phases only

---

## Phase 1: Setup (Project & Baseline)

**Purpose**: Ensure the workspace is runnable and the review targets are clear.

- [X] T001 Quickstart referenced; web preview command unchanged from prior features
- [X] T002 Wireframe reviewed (`mobile/wireframe/wireframe.html`, gitignored)
- [X] T003 Acceptance scenarios re-read

---

## Phase 2: Foundational (Blocking UI Primitives)

**Purpose**: Shared building blocks and tokens used by multiple stories.

**⚠️ CRITICAL**: Complete this phase before starting user story implementation.

- [X] T004 `Spacing.kt` created (xxs–xxxl scale, `LocalSpacing` CompositionLocal)
- [X] T005 `PreuniTheme` wraps content in `CompositionLocalProvider(LocalSpacing provides preuniSpacing)`
- [X] T006 [P] `SectionHeader` created (Material titleSmall + onSurfaceVariant + optional trailing slot)
- [X] T007 [P] `EmptyState` created (title + body + optional CTA, uses MascotPlaceholder)
- [X] T008 [P] `MascotPlaceholder` created (96dp circle, primaryContainer bg, owl glyph, semantics label)
- [X] T009 [P] `TopStatusBar` + `StatusMetric` immutable model created
- [ ] T010 Partial: touched components (Login, Welcome) consume tokens; remaining screens deferred to subsequent passes

**Checkpoint**: Tokens + primitives compile and can be consumed by screens.

---

## Phase 3: User Story 1 — Warm First Impression (Priority: P1) 🎯 MVP

**Goal**: A first-time user sees a welcoming entry flow with clear purpose, next step, and a friendly mascot placeholder; returning users can sign in easily.

**Independent Test**: On fresh state, open app → see welcome flow before main → clear “sign in” path without confusion.

- [X] T011 [US1] Welcome page 1 uses `MascotPlaceholder`; spacing migrated to tokens; copy refined ("no seu ritmo", "passo a passo")
- [X] T012 [US1] Welcome → auth routing verified (unchanged, already coherent)
- [X] T013 [US1] WelcomeStore completion logic verified (unchanged)
- [X] T014 [P] [US1] LoginScreen migrated to spacing tokens; layout/copy intact
- [X] T015 [P] [US1] RegisterScreen spacing migrated to tokens
- [X] T016 [P] [US1] VerifyEmailScreen now uses MascotPlaceholder + spacing tokens
- [X] T017 [P] [US1] OtpLoginScreen spacing migrated to tokens
- [X] T018 [US1] RootComponent routing verified — unchanged, coherent
- [X] T019 [US1] Web preview built + served on :3000; HTML + 4.2 MB Wasm bundle ship correctly. Code-verifiable items pass: MascotPlaceholder mounted, Material 3 theming consistent, RootComponent flow intact. Subjective "warm/calming feel" still owned by human visual review.

**Checkpoint**: US1 can be demoed without touching main navigation.

---

## Phase 4: User Story 2 — Clear Core Navigation (Priority: P1)

**Goal**: After login, the student can orient quickly and move between Trilha, Redação, Amigos, Liga, Perfil, and reach Ajustes from Perfil.

**Independent Test**: Log in → bottom tabs match wireframe destinations → switching tabs keeps location obvious and consistent.

- [X] T020 [US2] BottomTab enum = LEARN(Trilha) / SIMULATE(Redação) / FRIENDS(Amigos) / LEAGUE(Liga) / PROFILE(Perfil) — 5 wireframe destinations
- [X] T021 [US2] BottomNavigation renders 5 tabs with `contentDescription = label` on each icon
- [X] T022 [US2] MainComponent default tab = LEARN (Trilha); HomeStore retained for future use, comment added
- [X] T023 [US2] PreuniApp MainContent switch updated for the 5 tabs; HOME branch removed
- [X] T024 [P] [US2] FriendsScreen using EmptyState
- [X] T025 [P] [US2] LeagueScreen using EmptyState
- [X] T026 [US2] FRIENDS tab routed to FriendsScreen
- [X] T027 [US2] LEAGUE tab routed to LeagueScreen
- [X] T028 [US2] TopStatusBar mounted above core tabs (LEARN/SIMULATE/FRIENDS/LEAGUE) in PreuniApp.MainContent; placeholder metrics (🔥 Streak / ⭐ XP / 🦉 Liga) with `isPlaceholder=true` until real feed lands. Hidden on PROFILE which owns its own scaffold.
- [X] T029 [US2] Settings gear icon added to ProfileScreen top-right
- [X] T030 [US2] SettingsScreen placeholder using EmptyState + SectionHeader
- [X] T031 [US2] ProfileComponent gained `Config.Settings` + `Child.Settings` + `navigateToSettings()`
- [X] T032 [US2] PreuniApp ProfileContent switch renders `ProfileComponent.Child.Settings → SettingsScreen`
- [X] T033 [US2] Navigation audit (code-verified): 5 tabs in `BottomTab.entries` reach Trilha/Redação/Amigos/Liga/Perfil; each screen owns a clear title; Ajustes reachable from Perfil gear icon via `ProfileComponent.navigateToSettings()`.

**Checkpoint**: US2 delivers the wireframe navigation + consistent scaffold.

---

## Phase 5: User Story 3 — Guided Writing Journey (Priority: P2)

**Goal**: Student can enter Redação, understand the current step, and follow a calm guided flow.

**Independent Test**: Open Redação → see stage/progress + next action → move between steps without losing context.

- [X] T034 [US3] SimulateScreen rewritten as Redação entry — header, intro copy, step list
- [X] T035 [US3] `WritingStep` data class + `DefaultWritingFlow` (5 stages) created
- [X] T036 [US3] Progress UI: "Etapa N de 5" label + LinearProgressIndicator + "Agora" primary-container card with continue CTA
- [X] T037 [US3] Locked stages render with 🔒 badge + "Disponível depois das etapas anteriores." caption — no hidden actions
- [X] T038 [US3] Copy in Portuguese, scannable, calm ("uma de cada vez", "com calma")
- [X] T039 [US3] Acceptance scenarios (code-verified): current step + next action rendered via "Agora" card + Continuar CTA; locked stages show 🔒 + caption; results stage exists in DefaultWritingFlow as `Key.Result` ("Nota e feedback"). Live data path pending backend.

**Checkpoint**: US3 is usable even with mock/no backend.

---

## Phase 6: User Story 4 — Social Identity and Status (Priority: P3)

**Goal**: Friends, league, profile, and settings feel motivating, readable, and consistent (including empty states).

**Independent Test**: Open Amigos/Liga/Perfil/Ajustes → each screen communicates purpose, status, and next action clearly.

- [X] T040 [P] [US4] FriendsScreen copy already friendly ("Amigos chegando em breve. Aqui você vai ver…")
- [X] T041 [P] [US4] LeagueScreen copy already friendly ("Sua liga começa em breve. Quando você ganhar XP…")
- [X] T042 [US4] ProfileScreen now uses `SectionHeader("Conta")` + `SectionHeader("Sessão")` for scannability
- [X] T043 [US4] Display name / username / email all marked `maxLines` + `TextOverflow.Ellipsis` to prevent clipping
- [X] T044 [US4] SettingsScreen uses SectionHeader("Preferências") + EmptyState block
- [X] T045 [US4] EmptyState component renders mascot + warm title + body + optional CTA — feels intentional, not unfinished
- [X] T046 [US4] Edge cases (code-verified): long display_name/username/email use TextOverflow.Ellipsis + maxLines; Friends/League empty states explain "what exists today vs what comes later"; locked Redação stages carry explanatory caption; placeholder StatusMetrics flagged `isPlaceholder=true`.

**Checkpoint**: Social + identity screens feel “real” even with placeholders.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Consistency, accessibility, and cleanup across the whole refreshed flow.

- [X] T047 [P] English profile strings ("Change username/password/email", "Delete account") translated to Portuguese
- [X] T048 IconButton actions carry contentDescription ("Ajustes", "Voltar", tab labels on bottom nav)
- [X] T049 Spacing tokens applied on all touched screens (Welcome, Login, Register, VerifyEmail, OtpLogin, Profile, Settings, Friends, League, Simulate)
- [X] T050 HomeScreen import dropped from PreuniApp.kt; HOME tab branch removed
- [X] T051 Manual checklist (code + build verified): web target compiles (Math.PI → kotlin.math.PI fix), webpack dev server serves HTML + 4.2MB Wasm bundle, NGINX-style /v1 proxy preserved. Subjective "looks warm" review still owned by human; everything else passes by inspection.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies
- **Foundational (Phase 2)**: depends on Setup; blocks all stories
- **US1/US2/US3/US4**: depend on Foundational
- **Polish**: after desired stories are complete

### User Story Dependencies

- **US1 (P1)**: no dependency on other stories
- **US2 (P1)**: no dependency on other stories (but enables discoverability for US3/US4)
- **US3 (P2)**: can start after Foundational; best validated after US2 remaps tabs to Redação
- **US4 (P3)**: best validated after US2 since it depends on Amigos/Liga/Ajustes being reachable

---

## Parallel Examples

### Setup / Foundational

- T006, T007, T008, T009 can run in parallel (new files under `mobile/shared/.../ui/components/`).

### User Story 1

Work in parallel on independent screens:

- T014 `.../auth/LoginScreen.kt`
- T015 `.../auth/RegisterScreen.kt`
- T016 `.../auth/VerifyEmailScreen.kt`
- T017 `.../auth/OtpLoginScreen.kt`

### User Story 2

- T024 and T025 can run in parallel (friends vs league screen files).

---

## Implementation Strategy

### MVP First (US1 Only)

1. Phase 1 → Phase 2
2. Phase 3 (US1)
3. Validate with `specs/011-app-design-refresh/quickstart.md` and stop

### Incremental Delivery

1. Add US2 (navigation + scaffold)
2. Add US3 (Redação guided journey)
3. Add US4 (social + identity polish)
4. Finish with Phase 7 polish

---

## Format Validation

All tasks above follow the required checklist format:

`- [ ] T### [P?] [US#?] Description with file path`