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
- [ ] T015 [P] [US1] Refine register layout/copy consistency in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/auth/RegisterScreen.kt`
- [ ] T016 [P] [US1] Refine verify-email layout/copy consistency in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/auth/VerifyEmailScreen.kt`
- [ ] T017 [P] [US1] Refine OTP login layout/copy consistency in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/auth/OtpLoginScreen.kt`
- [ ] T018 [US1] Keep entry → auth → onboarding → main routing coherent in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/RootComponent.kt`
- [ ] T019 [US1] Run the “Core Flow Smoke Test” checklist in `specs/011-app-design-refresh/quickstart.md`

**Checkpoint**: US1 can be demoed without touching main navigation.

---

## Phase 4: User Story 2 — Clear Core Navigation (Priority: P1)

**Goal**: After login, the student can orient quickly and move between Trilha, Redação, Amigos, Liga, Perfil, and reach Ajustes from Perfil.

**Independent Test**: Log in → bottom tabs match wireframe destinations → switching tabs keeps location obvious and consistent.

- [ ] T020 [US2] Update bottom tab enum to wireframe destinations in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/navigation/BottomNavigation.kt`
- [ ] T021 [US2] Update bottom nav rendering for 5 tabs + accessibility labels in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/navigation/BottomNavigation.kt`
- [ ] T022 [US2] Update default tab + selection logic for new tabs in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/main/MainComponent.kt`
- [ ] T023 [US2] Update main scaffold to render new tab destinations in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`
- [ ] T024 [P] [US2] Add friends placeholder screen using `EmptyState` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/friends/FriendsScreen.kt`
- [ ] T025 [P] [US2] Add league placeholder screen using `EmptyState` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/league/LeagueScreen.kt`
- [ ] T026 [US2] Integrate friends destination into main tab routing in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`
- [ ] T027 [US2] Integrate league destination into main tab routing in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`
- [ ] T028 [US2] Ensure top status bar is included on core tabs (wireframe scaffold) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`
- [ ] T029 [US2] Add settings entry point from profile (gear or row action) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileScreen.kt`
- [ ] T030 [US2] Add settings screen placeholder using `EmptyState` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/settings/SettingsScreen.kt`
- [ ] T031 [US2] Wire profile → settings navigation (Decompose child) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileComponent.kt`
- [ ] T032 [US2] Render settings child in profile content switch in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`
- [ ] T033 [US2] Run the “Navigation Audit” checklist in `specs/011-app-design-refresh/quickstart.md`

**Checkpoint**: US2 delivers the wireframe navigation + consistent scaffold.

---

## Phase 5: User Story 3 — Guided Writing Journey (Priority: P2)

**Goal**: Student can enter Redação, understand the current step, and follow a calm guided flow.

**Independent Test**: Open Redação → see stage/progress + next action → move between steps without losing context.

- [ ] T034 [US3] Replace old “Simulados em breve” message with Redação entry scaffold in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/simulate/SimulateScreen.kt`
- [ ] T035 [US3] Add minimal step model for the guided journey in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/simulate/WritingStep.kt`
- [ ] T036 [US3] Add progress UI (current step + next action) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/simulate/SimulateScreen.kt`
- [ ] T037 [US3] Add locked/upcoming explanation state (no hidden actions) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/simulate/SimulateScreen.kt`
- [ ] T038 [US3] Ensure copy is Portuguese, calming, and scannable in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/simulate/SimulateScreen.kt`
- [ ] T039 [US3] Validate “Guided Writing Journey” acceptance scenarios in `specs/011-app-design-refresh/spec.md`

**Checkpoint**: US3 is usable even with mock/no backend.

---

## Phase 6: User Story 4 — Social Identity and Status (Priority: P3)

**Goal**: Friends, league, profile, and settings feel motivating, readable, and consistent (including empty states).

**Independent Test**: Open Amigos/Liga/Perfil/Ajustes → each screen communicates purpose, status, and next action clearly.

- [ ] T040 [P] [US4] Add friendly empty-state copy + layout polish in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/friends/FriendsScreen.kt`
- [ ] T041 [P] [US4] Add friendly empty-state copy + layout polish in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/league/LeagueScreen.kt`
- [ ] T042 [US4] Make profile sections more scannable using `SectionHeader` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileScreen.kt`
- [ ] T043 [US4] Ensure long usernames/labels don’t clip on profile in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileScreen.kt`
- [ ] T044 [US4] Ensure settings sections are scannable (grouped, readable) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/settings/SettingsScreen.kt`
- [ ] T045 [US4] Ensure empty/low-data states feel encouraging (not unfinished) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/EmptyState.kt`
- [ ] T046 [US4] Validate edge cases section items in `specs/011-app-design-refresh/spec.md`

**Checkpoint**: Social + identity screens feel “real” even with placeholders.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Consistency, accessibility, and cleanup across the whole refreshed flow.

- [ ] T047 [P] Normalize any remaining English strings touched by this feature in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/`
- [ ] T048 Ensure all icons/actions have meaningful accessibility labels in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/`
- [ ] T049 Replace ad-hoc spacing with spacing tokens on all modified screens in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/`
- [ ] T050 Remove dead code paths caused by tab remap (e.g., unused cases/imports) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`
- [ ] T051 Run the full manual checklist in `specs/011-app-design-refresh/quickstart.md`

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