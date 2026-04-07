# Tasks: Onboarding Flow and UX Fixes

**Input**: Design documents from `specs/005-onboarding-ux-fixes/`
**Branch**: `005-onboarding-ux-fixes`

**Constitution note**: This project requires test-first for all new business logic. Test tasks must be written and confirmed FAILING before the corresponding implementation tasks.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on other incomplete tasks)
- **[Story]**: Which user story this task belongs to
- All KMP paths relative to `mobile/shared/src/`

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: API method shared by US1 (track re-selection) and US2 (onboarding wiring). Must exist before either story's store can call it.

**⚠️ CRITICAL**: US1 and US2 stores both depend on T001.

- [X] T001 Add `updateTracks(trackIds: List<String>): Result<Unit>` to `commonMain/kotlin/com/preuni/shared/data/user/UserApiClient.kt` — call `PATCH v1/students/me/onboarding` with body `{"enrolled_track_ids": trackIds}`; map non-success status codes via `toAppError()`

**Checkpoint**: Foundation ready — US1 and US2 implementation can begin.

---

## Phase 2: User Story 1 — Track Re-selection from Profile (Priority: P1) 🎯 MVP

**Goal**: A logged-in user can open Profile → "Alterar matérias", toggle tracks, save, and the new selection is persisted. An empty selection is rejected with a validation message.

**Independent Test**: `./gradlew :shared:desktopTest --tests "*ChangeTrackStoreTest*"` passes; on the running app, Profile tab has "Alterar matérias" button that opens a track-picker and successfully saves.

### Tests for User Story 1 — Write FIRST, confirm RED

> **TDD**: Write `ChangeTrackStoreTest.kt` first. Run `./gradlew :shared:desktopTest --tests "*ChangeTrackStoreTest*"`. It must FAIL before T003.

- [X] T002 [US1] Write `commonTest/kotlin/com/preuni/shared/presentation/profile/ChangeTrackStoreTest.kt` — fake `updateTracks` lambda; four `@Test` functions using `runTest` with `@BeforeTest setMain(UnconfinedTestDispatcher())` + `@AfterTest resetMain()`: (1) `load_populatesSelectedTrackIds` — initial `selectedTrackIds` matches what was passed to factory; (2) `save_validSelection_emitsSavedLabel` — toggle one track then Save → collect `Label.Saved`; (3) `save_emptySelection_emitsValidationError` — deselect all then Save → `state.error != null`, no label emitted, `updateTracks` never called; (4) `save_repositoryError_setsError` — fake throws `AppError.Unknown()` → `state.error is AppError.Unknown`, `state.isLoading == false`

### Implementation for User Story 1

- [X] T003 [P] [US1] Create `commonMain/kotlin/com/preuni/shared/presentation/profile/ChangeTrackStore.kt` — `State(selectedTrackIds: Set<String>, isLoading: Boolean, error: AppError?)`, intents `Load(initial: Set<String>)`, `ToggleTrack(id)`, `Save`; labels `Saved`, `ValidationError(msg)`; `CoroutineExecutor` calls `updateTracks` lambda; validation: emit `ValidationError` if `selectedTrackIds.isEmpty()` on Save
- [X] T004 [P] [US1] Create `commonMain/kotlin/com/preuni/shared/presentation/profile/ChangeTrackScreen.kt` — composable receiving `store: ChangeTrackStore`, `availableTracks: List<Track>`, `onBack: () -> Unit`; shows `FilterChip` list for each track (selected = `track.id in state.selectedTrackIds`), "Salvar" button (disabled when loading/empty), error text when `state.error != null`; collects `Label.Saved` label to call `onBack`
- [X] T005 [US1] Add `Config.ChangeTrack` (serializable data object) to `ProfileComponent.Config` sealed interface in `commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileComponent.kt`; add `Child.ChangeTrack(store: ChangeTrackStore, availableTracks: List<Track>)` to `Child` sealed interface; add `createChild` branch for `Config.ChangeTrack` that builds `ChangeTrackStore` pre-loaded with current enrolled track IDs (pass via `Config.ChangeTrack(initialIds: Set<String>)`); add `fun navigateToChangeTrack(initialIds: Set<String>)` method calling `navigation.push(Config.ChangeTrack(initialIds))`
- [X] T006 [US1] Add "Alterar matérias" `OutlinedButton` to `commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileScreen.kt` — place it in the action buttons section (after "Change email"); button calls `onChangeTrack` lambda param; add `onChangeTrack: () -> Unit` to `ProfileScreen` function signature
- [X] T007 [US1] Wire `ChangeTrack` route in `commonMain/kotlin/com/preuni/shared/PreuniApp.kt` — in `ProfileContent()`, add `is ProfileComponent.Child.ChangeTrack -> ChangeTrackScreen(store = child.store, availableTracks = child.availableTracks, onBack = component::navigateBack)`; update the `Config.Profile` branch's `ProfileScreen` call to pass `onChangeTrack = { component.navigateToChangeTrack(child.store.state.enrolledTrackIds ?: emptySet()) }` (read enrolled tracks from ProfileStore state)
- [X] T008 [US1] Confirm green: run `JAVA_HOME=$(/usr/libexec/java_home -v 23) ./gradlew :shared:desktopTest --tests "*ChangeTrackStoreTest*"` from `mobile/` — all 4 tests must PASS

**Checkpoint**: US1 complete. A user on Profile sees "Alterar matérias", can pick tracks, save, and return.

---

## Phase 3: User Story 2 — Welcome Screen on First Install (Priority: P2)

**Goal**: First-time install shows 3 welcome slides (pre-login). Subsequent launches (logged in or out) skip directly to Main or Auth. A `"welcome_seen"` flag in `SecureStorage` controls this.

**Independent Test**: `./gradlew :shared:desktopTest --tests "*WelcomeStoreTest*"` passes; on the app, clearing session data and relaunching shows the slides; relaunching again skips them.

### Tests for User Story 2 — Write FIRST, confirm RED

> **TDD**: Write `WelcomeStoreTest.kt` first. Run `./gradlew :shared:desktopTest --tests "*WelcomeStoreTest*"`. It must FAIL before T010.

- [ ] T009 [US2] Write `commonTest/kotlin/com/preuni/shared/presentation/welcome/WelcomeStoreTest.kt` — three `@Test` functions with `runTest` + `@BeforeTest setMain(UnconfinedTestDispatcher())` + `@AfterTest resetMain()`: (1) `nextPage_advancesPageIndex` — dispatch `NextPage` × 2 → `state.pageIndex == 2`; (2) `nextPage_onLastPage_emitsCompleted` — dispatch `NextPage` × 3 (0→1→2→complete) → collect `Label.Completed`; (3) `skip_emitsCompletedImmediately` — dispatch `Skip` → collect `Label.Completed`

### Implementation for User Story 2

- [ ] T010 [P] [US2] Create `commonMain/kotlin/com/preuni/shared/presentation/welcome/WelcomeStore.kt` — `State(pageIndex: Int = 0)`, intents `NextPage`, `PreviousPage`, `Skip`; label `Completed`; const `TOTAL_PAGES = 3`; `NextPage` on last page (index 2) publishes `Label.Completed` instead of advancing; `Skip` publishes `Label.Completed` directly; `PreviousPage` coerces to 0
- [ ] T011 [P] [US2] Create `commonMain/kotlin/com/preuni/shared/presentation/welcome/WelcomeScreen.kt` — composable receiving `store: WelcomeStore`, `onCompleted: () -> Unit`; `HorizontalPager` with 3 pages using the same `OnboardingPage` data (move the `PAGES` list constant verbatim from `OnboardingScreen.kt` to this file); page indicator dots (same pattern as `OnboardingScreen`); "Próximo" button on pages 0-1, "Começar" on page 2; "Pular" `TextButton` on every page; collects `Label.Completed` to call `onCompleted`
- [ ] T012 [P] [US2] Create `commonMain/kotlin/com/preuni/shared/presentation/welcome/WelcomeComponent.kt` — `ComponentContext` wrapper; holds `WelcomeStore`; exposes `onCompleted: () -> Unit`
- [ ] T013 [US2] Add welcome-seen helpers to `commonMain/kotlin/com/preuni/shared/data/auth/TokenStore.kt` — add private const `KEY_WELCOME_SEEN = "welcome_seen"`; add `fun welcomeSeen(): Boolean = storage.get(KEY_WELCOME_SEEN) != null`; add `fun markWelcomeSeen() = storage.put(KEY_WELCOME_SEEN, "true")`
- [ ] T014 [US2] Add `Config.Welcome` (serializable data object) to `RootComponent` in `commonMain/kotlin/com/preuni/shared/presentation/RootComponent.kt`; update `initialConfig` property: if `!tokenStore.welcomeSeen()` → `Config.Welcome`; else if `tokenStore.load() == null` → `Config.Auth`; else → `Config.Main`; add `Child.Welcome(component: WelcomeComponent)` to `Child` sealed interface; add `Config.Welcome` branch in `createChild` that builds `WelcomeComponent` with `onCompleted = { tokenStore.markWelcomeSeen(); navigation.replaceAll(Config.Auth) }`
- [ ] T015 [US2] Strip welcome pages from `OnboardingScreen.kt` — remove the `PAGES` list (moved to `WelcomeScreen.kt`); remove the `if (page < 3) { ... } else { ... }` pager branch; the pager now always renders the track-selection page; change `pageCount = { 4 }` to `pageCount = { 1 }`; change `repeat(4)` dot indicator to `repeat(1)`; remove `NextPage`/`PreviousPage` intent dispatches from nav row (only "Começar" button remains on the single page); update `OnboardingStore.kt` to set `TOTAL_PAGES = 1`
- [ ] T016 [US2] Wire `WelcomeComponent` in `commonMain/kotlin/com/preuni/shared/PreuniApp.kt` — add import for `WelcomeScreen` and `WelcomeComponent`; add `is RootComponent.Child.Welcome -> WelcomeScreen(store = child.component.store, onCompleted = child.component::onCompleted)` branch in the `when (val child = ...)` block; wire real `completeOnboarding` in `RootComponent.createChild` for `Config.Onboarding` — replace the default stub with `completeOnboarding = { ids -> userRepository.updateTracks(ids) }` (requires adding `updateTracks` wrapper to `UserRepositoryImpl` or calling `UserApiClient.updateTracks` directly)
- [ ] T017 [US2] Confirm green: run `JAVA_HOME=$(/usr/libexec/java_home -v 23) ./gradlew :shared:desktopTest --tests "*WelcomeStoreTest*"` from `mobile/` — all 3 tests must PASS

**Checkpoint**: US2 complete. Welcome slides appear once per device. Subsequent launches skip to Auth or Main correctly.

---

## Phase 4: User Story 3 — Home Page Loads Dashboard (Priority: P3)

**Goal**: `HomeStore` auto-retries 3× with 1 s delay before emitting the error state, so transient profile-load failures (e.g., profile creation lag after registration) resolve silently.

**Independent Test**: `./gradlew :shared:desktopTest --tests "*HomeStoreTest*"` passes including the new retry scenarios.

### Tests for User Story 3 — Write FIRST, confirm RED

> **TDD**: Add retry test cases to the existing `HomeStoreTest.kt` before modifying `HomeStore.kt`.

- [ ] T018 [US3] Add retry test cases to `commonTest/kotlin/com/preuni/shared/presentation/home/HomeStoreTest.kt` — two new `@Test` functions: (1) `load_transientFailure_retriesAndSucceeds` — fake repo fails first 2 calls then succeeds; dispatch `Load`; `advanceUntilIdle()`; assert `state.student != null` and `state.error == null`; (2) `load_persistentFailure_setsErrorAfterThreeAttempts` — fake repo always fails; dispatch `Load`; `advanceUntilIdle()`; assert `state.error != null` and `state.student == null`

### Implementation for User Story 3

- [ ] T019 [US3] Add auto-retry to `commonMain/kotlin/com/preuni/shared/presentation/home/HomeStore.kt` — in the `Executor.load()` private function, add a retry loop (max 3 attempts, `delay(1_000)` between attempts using `kotlinx.coroutines.delay`); only dispatch `Msg.ErrorReceived` after the 3rd failed attempt; all other retry-internal failures are swallowed; `Msg.Loading` is dispatched only once (before the first attempt)
- [ ] T020 [US3] Confirm green: run `JAVA_HOME=$(/usr/libexec/java_home -v 23) ./gradlew :shared:desktopTest --tests "*HomeStoreTest*"` from `mobile/` — all tests including the two new retry tests must PASS

**Checkpoint**: US3 complete. Newly registered users who hit a brief user-service lag no longer see an error state on Home.

---

## Phase 5: Polish & Cross-Cutting Concerns

- [ ] T021 Run full `JAVA_HOME=$(/usr/libexec/java_home -v 23) ./gradlew :shared:desktopTest` from `mobile/` — entire desktopTest suite must be green (no regressions in LoginStoreTest, OnboardingStoreTest, ProfileStoreTest, etc.)
- [ ] T022 [P] Verify `RootComponent` serialization — ensure `Config.Welcome` has `@Serializable data object` annotation and Decompose's `childStack` retains correct state across process death (check by reading `RootComponent.kt` and confirming the `Config.serializer()` call compiles)
- [ ] T023 [P] Update `mobile/webApp/webpack.config.d/dev-proxy.js` comment if needed — confirm proxy still routes `/v1/*` to NGINX on `:8080` with no change needed (no-op verification)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: Start immediately — T001 only
- **US1 (Phase 2)**: Depends on T001; T002 (test) → T003/T004 (parallel) → T005 → T006 → T007 → T008
- **US2 (Phase 3)**: Depends on T001 (for T016 wiring); T009 (test) → T010/T011/T012 (parallel) → T013 → T014 → T015 → T016 → T017
- **US3 (Phase 4)**: Fully independent of US1/US2; T018 (test) → T019 → T020
- **Polish (Phase 5)**: Depends on all user stories complete

### User Story Independence

- **US1** and **US3** are fully independent of each other — can proceed in parallel after T001
- **US2** depends only on T001 (for the `completeOnboarding` wiring in T016)
- **US3** has no dependencies on T001 or any other story

### Parallel Opportunities Within Each Story

**US1**: T003 (ChangeTrackStore) + T004 (ChangeTrackScreen) can be written simultaneously  
**US2**: T010 (WelcomeStore) + T011 (WelcomeScreen) + T012 (WelcomeComponent) can be written simultaneously  
**Polish**: T022 + T023 can be verified simultaneously

---

## Implementation Strategy

### MVP (US1 only — most urgent bug)

1. T001 — add `updateTracks` to `UserApiClient`
2. T002 → T003 + T004 → T005 → T006 → T007 → T008
3. **VALIDATE**: Users can change tracks from Profile
4. Commit and demo

### Full Delivery (all 3 stories)

1. T001 (foundation)
2. US1 (T002–T008) — unblocks the track re-selection bug
3. US3 (T018–T020) — independent, can overlap with US2
4. US2 (T009–T017) — welcome flow restructure
5. Polish (T021–T023)

---

## Notes

- Java 23 (`/usr/libexec/java_home -v 23`) is required for all Gradle/desktopTest commands — Java 25 is incompatible with Kotlin 2.1.20
- All new stores must use `@BeforeTest setMain(UnconfinedTestDispatcher())` + `@AfterTest resetMain()` (established pattern in this project)
- `ChangeTrackStore` and `WelcomeStore` use the same `CoroutineExecutor` + label pattern as `RegisterStore` / `VerifyEmailStore`
- `TrackSelectionScreen.kt` already loads tracks via `ContentRepository.getTracks()` — `ChangeTrackScreen.kt` should use the same approach for consistency
- **IMPORTANT** For completed tasks, mark as `[X]` in this file
