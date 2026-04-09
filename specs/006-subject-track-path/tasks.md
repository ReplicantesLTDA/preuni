# Tasks: Subject Track Path

**Input**: Design documents from `specs/006-subject-track-path/`
**Branch**: `006-subject-track-path`

**Constitution note**: This project requires test-first for all new business logic. Test tasks must be written and confirmed FAILING before the corresponding implementation tasks.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on other incomplete tasks)
- **[Story]**: Which user story this task belongs to
- All KMP paths relative to `mobile/shared/src/`

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: `TokenStore` active-track helpers and `ModuleNode` model used by both US1 and US2.

**⚠️ CRITICAL**: US1 and US2 both depend on T001 and T002.

- [X] T001 Add `getActiveTrackId(): String?`, `setActiveTrackId(trackId: String)`, `clearActiveTrackId()` to `commonMain/kotlin/com/preuni/shared/data/auth/TokenStore.kt` — use private const `KEY_ACTIVE_TRACK_ID = "active_track_id"` and the existing `storage` field; follow the `welcomeSeen` / `markWelcomeSeen` pattern already in this file
- [X] T002 [P] Create `commonMain/kotlin/com/preuni/shared/domain/learn/ModuleNode.kt` — data class with fields `id: String`, `index: Int`, `title: String`, `state: NodeState`, `xpReward: Int = 10`; sealed interface `NodeState { LOCKED, ACTIVE, COMPLETED }`; also add top-level `fun generateStubModules(count: Int = 15): List<ModuleNode>` that returns a list where index 0 is `ACTIVE` and the rest are `LOCKED`, titles "Módulo {index+1}"

**Checkpoint**: Foundation ready — US1 and US2 implementation can begin.

---

## Phase 2: User Story 1 — Fix "Começar" Button & Single-Select Onboarding (Priority: P1) 🎯 MVP

**Goal**: A user can select exactly one subject during onboarding and tap "Começar" to immediately land on the Learn tab — even if the backend sync fails.

**Independent Test**: `./gradlew :shared:desktopTest --tests "*OnboardingStoreTest*"` passes; on the running app, selecting one subject and tapping "Começar" navigates to the Main screen with the Learn tab active.

### Tests for User Story 1 — Write FIRST, confirm RED

> **TDD**: Rewrite `OnboardingStoreTest.kt` for single-select. Run `./gradlew :shared:desktopTest --tests "*OnboardingStoreTest*"`. Tests must FAIL before T004.

- [X] T003 Rewrite `commonTest/kotlin/com/preuni/shared/presentation/onboarding/OnboardingStoreTest.kt` — replace all multi-select assertions with single-select: (1) `selectTrack_setsSelectedTrackId` — dispatch `SelectTrack("math")` → `state.selectedTrackId == "math"`; (2) `selectTrack_replacesExisting` — dispatch `SelectTrack("A")` then `SelectTrack("B")` → `state.selectedTrackId == "B"`; (3) `complete_noSelection_emitsValidationError` — dispatch `Complete` with no selection → `Label.ValidationError` emitted; (4) `complete_withSelection_savesLocallyAndEmitsCompleted` — fake `setActiveTrackId` called with selected id, `Label.Completed` emitted; (5) `complete_backendFailure_stillEmitsCompleted` — fake `completeOnboarding` throws, `Label.Completed` still emitted (fire-and-forget)

### Implementation for User Story 1

- [X] T004 Rewrite `OnboardingStore.State` in `commonMain/kotlin/com/preuni/shared/presentation/onboarding/OnboardingStore.kt` — replace `selectedTrackIds: Set<String>` with `selectedTrackId: String?`; replace `Intent.ToggleTrack` with `Intent.SelectTrack(trackId: String)`; update `Executor`: on `SelectTrack` dispatch `Msg.TrackSelected(trackId)` which replaces state; on `Complete`, if `selectedTrackId == null` publish `Label.ValidationError`; otherwise call `setActiveTrackId(selectedTrackId)` first, publish `Label.Completed` immediately, then launch coroutine for `completeOnboarding([selectedTrackId])` fire-and-forget (swallow errors)
- [X] T005 Update `commonMain/kotlin/com/preuni/shared/presentation/onboarding/TrackSelectionScreen.kt` — change `FilterChip` `selected` from `track.id in state.selectedTrackIds` to `track.id == state.selectedTrackId`; change `onClick` from `Intent.ToggleTrack` to `Intent.SelectTrack(track.id)`; update "Começar" enabled condition from `selectedTrackIds.isNotEmpty()` to `selectedTrackId != null`
- [X] T006 Update `OnboardingComponent` in `commonMain/kotlin/com/preuni/shared/presentation/onboarding/OnboardingComponent.kt` — add `setActiveTrackId: (String) -> Unit` parameter to the factory/constructor; thread it through to `OnboardingStoreFactory`
- [X] T007 Wire `setActiveTrackId` in `commonMain/kotlin/com/preuni/shared/presentation/RootComponent.kt` — in the `Config.Onboarding` branch of `createChild`, pass `setActiveTrackId = { tokenStore.setActiveTrackId(it) }` to `OnboardingComponent`
- [X] T008 Confirm green: run `JAVA_HOME=$(/usr/libexec/java_home -v 23) ./gradlew :shared:desktopTest --tests "*OnboardingStoreTest*"` from `mobile/` — all 5 tests must PASS

**Checkpoint**: US1 complete. Selecting one subject and tapping "Começar" immediately navigates to Main.

---

## Phase 3: User Story 2 — S-Curve Gamified Learning Path (Priority: P1)

**Goal**: The "Aprender" tab shows a scrollable S-curve path of star-shaped module nodes with bezier connecting lines — rendered from stub data even with no backend content.

**Independent Test**: `./gradlew :shared:desktopTest --tests "*LearnStoreTest*"` passes; on the running app, the Learn tab shows a vertical zigzag path of at least 15 star nodes.

### Tests for User Story 2 — Write FIRST, confirm RED

> **TDD**: Write `LearnStoreTest.kt` first. Run `./gradlew :shared:desktopTest --tests "*LearnStoreTest*"`. Must FAIL before T010.

- [X] T009 Write `commonTest/kotlin/com/preuni/shared/presentation/learn/LearnStoreTest.kt` — four `@Test` with `runTest` + `@BeforeTest setMain(UnconfinedTestDispatcher())` + `@AfterTest resetMain()`: (1) `load_withActiveTrack_populatesModules` — fake `getActiveTrackId` returns "math", fake `getTracks` returns list with math track; dispatch `Load`; `advanceUntilIdle()`; assert `state.activeTrackId == "math"` and `state.modules.size == 15` and `state.modules[0].state == NodeState.ACTIVE`; (2) `load_noActiveTrack_leavesModulesEmpty` — fake returns null; assert `state.activeTrackId == null` and `state.modules.isEmpty()`; (3) `tapNode_active_emitsOpenLesson` — set up state with active node at index 0; dispatch `TapNode("module-0")`; collect `Label.OpenLesson("module-0")`; (4) `tapNode_locked_emitsShowLockedMessage` — dispatch `TapNode` for locked node; collect `Label.ShowLockedMessage`

### Implementation for User Story 2

- [X] T010 [P] Create `commonMain/kotlin/com/preuni/shared/presentation/learn/LearnStore.kt` — `State(activeTrackId: String?, activeTrack: Track?, modules: List<ModuleNode>, isLoading: Boolean, error: AppError?)`; intents `Load`, `TapNode(nodeId: String)`, `Retry`; labels `OpenLesson(moduleId: String)`, `ShowLockedMessage`; `CoroutineExecutor`: on `Load`, read `getActiveTrackId()`, if null leave modules empty; else call `getTracks()` to find the track, call `generateStubModules(15)`, dispatch state; on `TapNode`, find node — if `ACTIVE` publish `OpenLesson`; if `LOCKED` publish `ShowLockedMessage`; inject `getActiveTrackId: () -> String?` and `getTracks: suspend () -> Result<List<Track>>`
- [X] T011 [P] Create `commonMain/kotlin/com/preuni/shared/presentation/learn/LearnScreen.kt` — composable `LearnScreen(store: LearnStore, onOpenLesson: (moduleId: String) -> Unit)`; collect state and labels; top bar showing `state.activeTrack?.name ?: "Aprender"`; main content: `LazyColumn` where each item is a `ModuleNodeItem(node, horizontalFraction, onTap)`; horizontal fraction alternates `if (index % 2 == 0) 0.68f else 0.10f`; behind the list items draw a `Canvas` with bezier connectors between consecutive node centers; each `ModuleNodeItem` draws a 56.dp star shape using `Canvas` — five-point star path; fill and border colour based on `NodeState`; `ACTIVE` node shows a pulsing `animateFloat` ring; `LOCKED` node shows a `Lock` icon overlay (16.dp); label collector: `OpenLesson` → `onOpenLesson(id)`; `ShowLockedMessage` → `LaunchedEffect` showing `SnackbarHostState.showSnackbar("Complete os módulos anteriores primeiro")`
- [X] T012 [P] Create `commonMain/kotlin/com/preuni/shared/presentation/learn/PlaceholderLessonScreen.kt` — composable `PlaceholderLessonScreen(moduleTitle: String, onBack: () -> Unit)`; `Scaffold` with `TopAppBar` (back arrow, title = moduleTitle); centered `Text("Em breve — conteúdo chegando!")` in body
- [X] T013 Wire `LearnStore` into `MainComponent` in `commonMain/kotlin/com/preuni/shared/presentation/main/MainComponent.kt` — add `learnStore: LearnStore` field created via `LearnStoreFactory(storeFactory, getActiveTrackId = { tokenStore.getActiveTrackId() }, getTracks = { contentRepository.getTracks() }).create()`; add `PlaceholderLessonScreen` child stack or simple boolean state `lessonModuleId: StateFlow<String?>` for lesson navigation; add `fun openLesson(moduleId: String)` and `fun closeLesson()`; inject `tokenStore: TokenStore` and `contentRepository: ContentRepository` (already available via RootComponent)
- [X] T014 Wire `LearnScreen` in `commonMain/kotlin/com/preuni/shared/PreuniApp.kt` — in `MainContent()`, replace the existing `LearnScreen()` no-arg call with `LearnScreen(store = mainComponent.learnStore, onOpenLesson = mainComponent::openLesson)`; add conditional: if `lessonModuleId != null` render `PlaceholderLessonScreen(moduleTitle = "Módulo", onBack = mainComponent::closeLesson)` over the tab content
- [X] T015 Confirm green: run `JAVA_HOME=$(/usr/libexec/java_home -v 23) ./gradlew :shared:desktopTest --tests "*LearnStoreTest*"` from `mobile/` — all 4 tests must PASS

**Checkpoint**: US2 complete. Learn tab shows the S-curve path of 15 star nodes.

---

## Phase 4: User Story 3 — Switch Subject from Profile (Priority: P2)

**Goal**: Profile → "Alterar matérias" enforces single-select; saving updates the active track locally and switches to the Learn tab.

**Independent Test**: `./gradlew :shared:desktopTest --tests "*ChangeTrackStoreTest*"` passes; on the running app, changing subject from Profile lands on the Learn tab showing the new subject's path.

### Tests for User Story 3 — Write FIRST, confirm RED

> **TDD**: Update `ChangeTrackStoreTest.kt` for single-select. Must FAIL before T017.

- [X] T016 Update `commonTest/kotlin/com/preuni/shared/presentation/profile/ChangeTrackStoreTest.kt` — replace multi-select assertions: (1) `load_populatesSelectedTrackId` — initial `selectedTrackId` matches the one passed to factory; (2) `selectTrack_replacesExisting` — select A then B → `selectedTrackId == "B"`; (3) `save_validSelection_emitsSaved` — select one track → Save → collect `Label.Saved`; (4) `save_noSelection_emitsValidationError` — dispatch Save with null selection → `state.error != null`, no label; (5) `save_callsSetActiveTrackId` — verify fake `setActiveTrackId` called with selected id on successful save

### Implementation for User Story 3

- [X] T017 Update `ChangeTrackStore.kt` in `commonMain/kotlin/com/preuni/shared/presentation/profile/ChangeTrackStore.kt` — replace `selectedTrackIds: Set<String>` with `selectedTrackId: String?`; rename `Intent.ToggleTrack` to `Intent.SelectTrack(trackId: String)` (replace semantics); update `Executor.save()`: validate `selectedTrackId != null` (else emit `ValidationError`); call `setActiveTrackId(selectedTrackId)` locally; call `updateTracks([selectedTrackId])` fire-and-forget; publish `Label.Saved`; add `setActiveTrackId: (String) -> Unit` injection parameter
- [X] T018 Update `ChangeTrackScreen.kt` in `commonMain/kotlin/com/preuni/shared/presentation/profile/ChangeTrackScreen.kt` — change chip `selected` from `track.id in state.selectedTrackIds` to `track.id == state.selectedTrackId`; change `onClick` from `Intent.ToggleTrack` to `Intent.SelectTrack(track.id)`; update "Salvar" enabled condition; update initial load to pass current active track id as initial selection
- [X] T019 Expose `selectTab(tab: BottomTab)` publicly in `commonMain/kotlin/com/preuni/shared/presentation/main/MainComponent.kt` — it already exists as `fun selectTab(tab: BottomTab)` internally; verify it's accessible from `ProfileComponent`; no change needed if already public
- [X] T020 Update `ProfileComponent` in `commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileComponent.kt` — add `onSwitchToLearn: () -> Unit` constructor parameter; in `createChild(Config.ChangeTrack)`, pass `setActiveTrackId = { tokenStore.setActiveTrackId(it) }` to `ChangeTrackStoreFactory`; in the `Label.Saved` handler inside the ChangeTrack child branch, call `onSwitchToLearn()` after navigating back
- [X] T021 Wire `onSwitchToLearn` in `commonMain/kotlin/com/preuni/shared/PreuniApp.kt` — where `ProfileScreen` / `ProfileComponent` is used in `MainContent()`, pass `onSwitchToLearn = { mainComponent.selectTab(BottomTab.LEARN) }`; also pass `tokenStore` if needed to `ProfileComponent` for `setActiveTrackId` injection
- [X] T022 Confirm green: run `JAVA_HOME=$(/usr/libexec/java_home -v 23) ./gradlew :shared:desktopTest --tests "*ChangeTrackStoreTest*"` from `mobile/` — all 5 tests must PASS

**Checkpoint**: US3 complete. Changing subject from Profile immediately switches to the Learn tab with the new subject's path.

---

## Phase 5: User Story 4 — Bottom Navigation Polish (Priority: P2)

**Goal**: Bottom nav works correctly after subject selection — Learn tab is the default after onboarding, and tab switches preserve scroll state.

**Independent Test**: On the running app, tapping between Learn and Profile tabs does not reset the Learn path scroll position.

- [X] T023 [US4] Set LEARN as default selected tab in `commonMain/kotlin/com/preuni/shared/presentation/main/MainComponent.kt` — change `private val _selectedTab = MutableStateFlow(BottomTab.HOME)` to `BottomTab.LEARN` as the initial value
- [X] T024 [US4] Preserve Learn tab scroll state — in `LearnScreen.kt`, capture `rememberLazyListState()` and hoist it so it survives recomposition when switching tabs; pass it into `LazyColumn(state = listState)`; since `MainComponent` manages the tab, the composable will stay in the composition tree and the state will be preserved automatically via `remember` — verify this works on Android by navigating away and back

**Checkpoint**: US4 complete. Navigating away from Learn and back preserves scroll position.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T025 Run full `JAVA_HOME=$(/usr/libexec/java_home -v 23) ./gradlew :shared:desktopTest` from `mobile/` — entire desktopTest suite must be green (no regressions in LoginStoreTest, HomeStoreTest, ProfileStoreTest, WelcomeStoreTest, etc.)
- [X] T026 [P] Verify `ModuleNode.generateStubModules()` correctness — assert first node is ACTIVE, all others LOCKED, count matches requested, titles are "Módulo 1" through "Módulo {count}" (manual code review + desktopTest)
- [X] T027 [P] Verify `TokenStore` active track key — read `TokenStore.kt` and confirm `KEY_ACTIVE_TRACK_ID` is distinct from `KEY_WELCOME_SEEN`, `KEY_ACCESS_TOKEN`, `KEY_REFRESH_TOKEN`; no collision

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: Start immediately — T001 and T002 are independent
- **US1 (Phase 2)**: Depends on T001 (TokenStore); T003 (test) → T004 + T005 (parallel) → T006 → T007 → T008
- **US2 (Phase 3)**: Depends on T001 (TokenStore) + T002 (ModuleNode); T009 (test) → T010 + T011 + T012 (parallel) → T013 → T014 → T015
- **US3 (Phase 4)**: Depends on T001 (TokenStore); T016 (test) → T017 + T018 (parallel) → T019 → T020 → T021 → T022
- **US4 (Phase 5)**: Depends on US2 (LearnScreen); T023 + T024 independent of each other
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Independence

- **US1** and **US2** share T001+T002 foundation but are otherwise independent
- **US3** depends on T001 but is independent of US2 (can be worked in parallel after T001)
- **US4** depends on US2 (needs LearnScreen to exist)

### Parallel Opportunities Within Each Story

**US2**: T010 (LearnStore) + T011 (LearnScreen) + T012 (PlaceholderLessonScreen) can be written simultaneously
**US3**: T017 (ChangeTrackStore) + T018 (ChangeTrackScreen) can be written simultaneously
**Polish**: T026 + T027 can be verified simultaneously

---

## Implementation Strategy

### MVP (US1 + US2 — unblocks every new user)

1. T001, T002 (foundation)
2. T003 → T004 + T005 → T006 → T007 → T008 (fix Começar)
3. T009 → T010 + T011 + T012 → T013 → T014 → T015 (S-curve path)
4. **VALIDATE**: New user can register, pick subject, land on S-curve Learn screen
5. Commit and demo

### Full Delivery

1. Foundation (T001–T002)
2. US1 (T003–T008) — fix onboarding
3. US2 (T009–T015) — gamified path screen
4. US3 (T016–T022) — subject switching
5. US4 (T023–T024) — nav polish
6. Polish (T025–T027)

---

## Notes

- Java 23 required for all Gradle commands: `JAVA_HOME=$(/usr/libexec/java_home -v 23)`
- All new stores follow `CoroutineExecutor` + label pattern (RegisterStore / VerifyEmailStore reference)
- `LearnScreen` star path drawn entirely with Compose `Canvas` — no image assets
- `generateStubModules()` is the fallback until a real content endpoint exists
- The bezier connector in `LearnScreen` must be drawn in a single `Canvas` pass that spans the full scrollable height — use `drawBehind` modifier or a layered `Box`
- **IMPORTANT** For completed tasks, mark as `[X]` in this file
