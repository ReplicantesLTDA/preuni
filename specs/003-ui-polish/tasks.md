# Tasks: Friendly UI Polish — Core Frontend

**Input**: Design documents from `/specs/003-ui-polish/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓, quickstart.md ✓

**Organization**: Tasks grouped by user story. Each phase is independently testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to
- No test tasks — not requested in spec; visual review via `make run-web` is the gate

---

## Phase 1: Setup

**Purpose**: Create the new `ui/` package structure. No dependency changes — Material 3 is already on the classpath.

- [X] T001 Create directory `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/theme/` (create placeholder file to initialize the package)
- [X] T002 Create directory `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/` (create placeholder file to initialize the package)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build the `PreuniTheme` + component library. ALL user story phases depend on this phase being complete first.

**⚠️ CRITICAL**: No screen work can begin until this phase is complete.

- [X] T003 [P] Create `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/theme/Color.kt` — define the full Material 3 light color scheme with violet primary (`#6750A4`), using `lightColorScheme()` from `androidx.compose.material3`. Include all 13 role tokens from research.md §3.
- [X] T004 [P] Create `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/theme/Shape.kt` — define `PreuniShapes` using `Shapes(extraSmall=12dp, small=16dp, medium=20dp, large=28dp, extraLarge=50dp)` from `androidx.compose.material3`. The `small` (16dp) value is used by `PreuniTextField`; `extraLarge` (50dp) is the pill shape for buttons.
- [X] T005 [P] Create `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/theme/SubjectTrackColors.kt` — define `data class TrackColorScheme(container: Color, onContainer: Color)`, `data class SubjectTrack(id, name, shortName, emoji, colorScheme)`, and `val SubjectTracks: List<SubjectTrack>` with all 5 ENEM areas and their WCAG AA verified color values from data-model.md.
- [X] T006 Create `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/theme/PreuniTheme.kt` — `@Composable fun PreuniTheme(content: @Composable () -> Unit)` that calls `MaterialTheme(colorScheme = preuniLightColorScheme, shapes = preuniShapes, content = content)`. Depends on T003 and T004.
- [X] T007 [P] Create `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/PreuniButton.kt` — implement per contract in `specs/003-ui-polish/contracts/PreuniButton.md`. Wraps `Button` from Material 3 with `modifier = modifier.fillMaxWidth().heightIn(min = 52.dp)`. When `isLoading = true`, shows `CircularProgressIndicator(modifier = Modifier.size(20.dp), strokeWidth = 2.dp)` and passes `enabled = false` to the underlying `Button`. Shape is inherited from `MaterialTheme.shapes.extraLarge` — no hardcoded shape value in the component.
- [X] T008 [P] Create `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/PreuniTextField.kt` — implement per contract in `specs/003-ui-polish/contracts/PreuniTextField.md`. Wraps `OutlinedTextField` with `shape = MaterialTheme.shapes.small`, `modifier = modifier.fillMaxWidth()`, and `supportingText` slot that always renders (passing `null` when no error) to prevent layout shift. No hardcoded shape/color values.
- [X] T009 [P] Create `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/SubjectTrackCard.kt` — implement per contract in `specs/003-ui-polish/contracts/SubjectTrackCard.md`. Uses `Card` with `colors = CardDefaults.cardColors(containerColor = track.colorScheme.container)`, `shape = MaterialTheme.shapes.medium`, `modifier = modifier.fillMaxWidth()`. When `selected = true`, draws a 2dp primary-color border using `border = BorderStroke(2.dp, MaterialTheme.colorScheme.primary)`. Row layout: emoji `Text` (headlineSmall) + `Column` with `shortName` (titleMedium) in `onContainer` color.
- [X] T010 Apply `PreuniTheme { ... }` wrapper in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt` — wrap the top-level content block (the `when (val child = ...)` block) with `PreuniTheme`. Import `com.preuni.shared.ui.theme.PreuniTheme`. Depends on T006.

**Checkpoint**: After T010, run `make run-web`. The login screen should show rounded text fields (16dp corners) and the primary color should be violet. Buttons remain pill-shaped (M3 default was already pill).

---

## Phase 3: User Story 1 — Polished Auth Screens (Priority: P1) 🎯 MVP

**Goal**: All auth screens (Login, Register, VerifyEmail, OtpLogin) use `PreuniButton` + `PreuniTextField` with consistent visual treatment, warm layout, and clear hierarchy.

**Independent Test**: Open `http://localhost:8088`. Navigate through the auth flow without logging in. Verify rounded inputs, pill buttons, consistent spacing, and inline error display on each of the 4 auth screens.

- [X] T011 [US1] Update `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/auth/LoginScreen.kt` — replace `OutlinedTextField` calls with `PreuniTextField`, replace `Button` with `PreuniButton`. Add a brand header section above the form: `Text("PreUni", style = MaterialTheme.typography.displaySmall, color = MaterialTheme.colorScheme.primary)` with a `Text("Prepare-se para o ENEM.", style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)` subtitle. Add `Spacer(Modifier.height(48.dp))` between brand header and form. Imports: `com.preuni.shared.ui.components.PreuniButton`, `com.preuni.shared.ui.components.PreuniTextField`.
- [X] T012 [US1] Update `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/auth/RegisterScreen.kt` — replace `OutlinedTextField` calls with `PreuniTextField`, replace `Button` with `PreuniButton`. Ensure each field's `isError` + `errorMessage` parameters are wired to the existing `AppError.Validation` check pattern from the store state. Maintain the existing back navigation button.
- [X] T013 [US1] Update `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/auth/VerifyEmailScreen.kt` — replace `OutlinedTextField` with `PreuniTextField`, replace `Button` with `PreuniButton`. Add a centered icon placeholder (use `Text("📧", style = MaterialTheme.typography.displayMedium)`) above the instruction text. Instruction text: `"Enviamos um código para ${email}. Insira-o abaixo."` in `bodyMedium` / `onSurfaceVariant`.
- [X] T014 [US1] Update `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/auth/OtpLoginScreen.kt` — replace `OutlinedTextField` with `PreuniTextField`, replace `Button` with `PreuniButton`. Add a `Text("🔑", style = MaterialTheme.typography.displayMedium)` icon and a `Text("Entre sem senha", style = MaterialTheme.typography.headlineMedium)` heading above the input field.

**Checkpoint**: All 4 auth screens are visually polished. Inputs are rounded; buttons are pill-shaped; each screen has clear hierarchy and breathing space. Verify inline error display by submitting an empty login form.

---

## Phase 4: User Story 2 — Welcoming Home / Dashboard (Priority: P2)

**Goal**: Home screen greets the user by name, displays streak/XP badges in brand colors, shows a grid of the 5 ENEM subject track cards, and has a pill-shaped CTA.

**Independent Test**: Log in with a registered account. Home screen shows greeting, streak/XP badges, 5 subject track cards in their distinct colors, and a pill CTA. Verify at 360px viewport width.

- [X] T015 [US2] Update `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/home/HomeScreen.kt` — add imports for `SubjectTrackCard`, `SubjectTracks`, `PreuniButton`. Replace the `Button(onClick = onStartLearning)` inside the CTA card with `PreuniButton(text = "Aprender agora", onClick = onStartLearning)`. Below the readiness card (or CTA card if readiness is 0), add a `Text("Matérias", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)` heading followed by a `Column` with `verticalArrangement = Arrangement.spacedBy(8.dp)` containing `SubjectTracks.forEach { track -> SubjectTrackCard(track = track, onClick = {}) }`. Add `Spacer(Modifier.height(16.dp))` before the matérias heading.
- [X] T016 [P] [US2] Update the `StreakBadge` private composable in `HomeScreen.kt` — change `containerColor` from `MaterialTheme.colorScheme.tertiaryContainer` to `MaterialTheme.colorScheme.primaryContainer`. Change text from `"$count"` to `"$count dias"` with label `"🔥 Sequência"` above the count in `labelSmall` style. This makes the streak badge visually prominent with brand color.
- [X] T017 [P] [US2] Update the `XpPill` private composable in `HomeScreen.kt` — wrap in `Card` with `shape = MaterialTheme.shapes.extraLarge` (pill) explicitly using `CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.secondaryContainer)`. Keep the `"⭐ $xp XP"` text in `labelLarge`.

**Checkpoint**: Log in and verify: personalized greeting visible, streak badge in primary color, XP pill is pill-shaped, 5 ENEM track cards display in correct distinct colors, CTA button is pill-shaped and full-width.

---

## Phase 5: User Story 3 — Consistent Bottom Navigation (Priority: P3)

**Goal**: Bottom navigation has correct icons per tab (not placeholder Home icons) and active tab uses brand violet.

**Independent Test**: Log in and tap through all 4 tabs. Each tab icon is distinct and thematically relevant. Active tab is violet; inactive tabs are muted. No layout shift when switching tabs.

- [X] T018 [US3] Update `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/navigation/BottomNavigation.kt` — replace placeholder icons. Add import for `androidx.compose.material.icons.filled.School`, `androidx.compose.material.icons.filled.Quiz`, `androidx.compose.material.icons.filled.EditNote`. Update `BottomTab` enum: `HOME` keeps `Icons.Filled.Home`; `LEARN` changes to `Icons.Filled.School`; `SIMULATE` changes to `Icons.Filled.Quiz`; `PROFILE` keeps `Icons.Filled.Person`. Remove the two placeholder comments. If `Icons.Filled.School` or `Icons.Filled.Quiz` are not available in the current material-icons-core dependency, use `Icons.Filled.MenuBook` for LEARN and `Icons.Filled.Assignment` for SIMULATE — both are in `material-icons-extended` already pulled by Compose.
- [X] T019 [US3] Verify `NavigationBar` in `BottomNavigation.kt` inherits correct M3 colors from `PreuniTheme` — active indicator uses `NavigationBarItemDefaults.colors()` which by default uses `colorScheme.secondaryContainer` for the indicator. To use brand violet for the active icon tint, explicitly set `colors = NavigationBarItemDefaults.colors(indicatorColor = MaterialTheme.colorScheme.secondaryContainer, selectedIconColor = MaterialTheme.colorScheme.onSecondaryContainer)`. This matches M3 spec and provides clear active state distinction.

**Checkpoint**: Navigate all 4 tabs. Each has a distinct, thematically appropriate icon. Active tab has a pill-shaped indicator in brand violet container color.

---

## Phase 6: User Story 4 — Onboarding Flow (Priority: P4)

**Goal**: Onboarding screens use `PreuniButton` for CTAs, the track-selection page (page 4) uses `SubjectTrackCard`, and the overall layout feels warm and focused.

**Independent Test**: Create a new account, verify email, then walk through all 4 onboarding pages. Each page has one focus, pill CTAs, and page 4 shows all 5 subject track cards with multi-select. "Começar" only activates when ≥ 1 track is selected.

- [X] T020 [P] [US4] Update `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/onboarding/OnboardingScreen.kt` — replace both `Button` calls (Next and Complete) with `PreuniButton`. The Next button: `PreuniButton(text = "Próximo", onClick = { store.accept(OnboardingStore.Intent.NextPage) }, modifier = Modifier.width(160.dp))`. The Complete button: `PreuniButton(text = "Começar", onClick = { store.accept(OnboardingStore.Intent.Complete) }, enabled = state.selectedTrackIds.isNotEmpty() && !state.isLoading)`. Keep `TextButton` for the "Voltar" back action — it intentionally has lower visual weight.
- [X] T021 [US4] Update the track-selection page (page index 3) in `OnboardingScreen.kt` — replace the placeholder `Box` with a `Column(modifier = Modifier.fillMaxSize().verticalScroll(rememberScrollState()))` containing: `Text("Escolha suas matérias", style = MaterialTheme.typography.headlineMedium)`, `Spacer(Modifier.height(8.dp))`, `Text("Você pode escolher mais de uma.", style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)`, `Spacer(Modifier.height(16.dp))`, then `SubjectTracks.forEach { track -> SubjectTrackCard(track = track, selected = state.selectedTrackIds.contains(track.id), onClick = { store.accept(OnboardingStore.Intent.ToggleTrack(track.id)) }); Spacer(Modifier.height(8.dp)) }`. Add `ToggleTrack(trackId: String)` intent to `OnboardingStore` if not already present (check `OnboardingStore.kt` first).
- [X] T022 [P] [US4] Update the `OnboardingPage` content pages (indices 0-2) in `OnboardingScreen.kt` — add a large centered emoji per page above the title. Page 0: `"🎓"`, Page 1: `"🔁"`, Page 2: `"📝"`. Use `Text(p.emoji, style = MaterialTheme.typography.displayLarge, modifier = Modifier.align(Alignment.CenterHorizontally))` before the title. Add the `emoji` field to the `OnboardingPage` private data class.
- [X] T023 [P] [US4] Update the page indicator dots in `OnboardingScreen.kt` — replace the current `Surface` dots with a `Row(horizontalArrangement = Arrangement.spacedBy(6.dp))` of `Box` composables: active dot is `Box(Modifier.width(24.dp).height(8.dp).clip(MaterialTheme.shapes.extraLarge).background(MaterialTheme.colorScheme.primary))`; inactive dot is `Box(Modifier.size(8.dp).clip(CircleShape).background(MaterialTheme.colorScheme.outlineVariant))`. This gives a pill-shaped active indicator (elongated dot) matching Headspace-style paging UX.

**Checkpoint**: Walk through all 4 onboarding pages with a new account. Each page shows an emoji, focused copy, and pill CTA. Page 4 shows all 5 track cards; selecting tracks enables "Começar".

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, cleanup, and consistency pass across all modified screens.

- [ ] T024 [P] Verify 360px viewport: open each screen in browser DevTools at 360px width and confirm no horizontal scroll, no clipped elements, and all touch targets ≥ 44dp. Screens to check: Login, Register, VerifyEmail, OtpLogin, Home, Onboarding (all 4 pages).
- [X] T025 [P] Audit for raw values: search all modified `.kt` files for hardcoded hex colors (`0xFF...`) outside of `SubjectTrackColors.kt` and hardcoded dp values for shape radii outside of `Shape.kt`. Fix any violations to use theme tokens.
- [ ] T026 Run the quickstart.md validation checklist in `specs/003-ui-polish/quickstart.md` — complete steps 2 through 7 and confirm all `[ ]` items pass.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 directory creation — **BLOCKS all user story phases**
- **US1 Auth Screens (Phase 3)**: Depends on Foundational phase (T003–T010)
- **US2 Home (Phase 4)**: Depends on Foundational phase; T015 depends on `SubjectTrackCard` (T009)
- **US3 Navigation (Phase 5)**: Depends on Foundational phase only — no dependency on US1 or US2
- **US4 Onboarding (Phase 6)**: Depends on Foundational phase; T021 depends on `SubjectTrackCard` (T009)
- **Polish (Phase 7)**: Depends on all user story phases complete

### User Story Dependencies

- **US1 (P1)**: After Foundational — independent of US2, US3, US4
- **US2 (P2)**: After Foundational (specifically T009 for SubjectTrackCard) — independent of US1, US3, US4
- **US3 (P3)**: After Foundational — fully independent of all other stories
- **US4 (P4)**: After Foundational (specifically T009 for SubjectTrackCard) — independent of US1, US2, US3

### Within Each User Story

- Foundational components first (T003–T010) → Screen updates (T011+)
- Screen tasks within a story are independent of each other and can run in parallel

### Parallel Opportunities

- **Phase 2**: T003, T004, T005 can run in parallel (different files); T006, T007, T008, T009 can run in parallel after T003/T004; T010 depends on T006
- **Phase 3**: T011, T012, T013, T014 are all independent (different files) — run in parallel
- **Phase 4**: T015 depends on T009; T016 and T017 are edits to same file as T015 — run sequentially after T015
- **Phase 5**: T018 and T019 are edits to same file — run sequentially
- **Phase 6**: T020 and T021, T022, T023 are all edits to `OnboardingScreen.kt` — T020 first, then T021/T022/T023 in parallel (different code sections of the same file — be careful of conflicts; may need sequential)

---

## Parallel Example: Foundational Phase

```
Parallel batch 1 (no dependencies):
  T003 Color.kt
  T004 Shape.kt
  T005 SubjectTrackColors.kt

Parallel batch 2 (depends on T003 + T004):
  T006 PreuniTheme.kt  ← sequential (depends on T003, T004)

Parallel batch 3 (depends on T006 for theme, T005 for SubjectTrack):
  T007 PreuniButton.kt
  T008 PreuniTextField.kt
  T009 SubjectTrackCard.kt

T010 PreuniApp.kt  ← sequential (depends on T006)
```

## Parallel Example: User Story 1 (Auth Screens)

```
After Foundational complete:
  T011 LoginScreen.kt
  T012 RegisterScreen.kt
  T013 VerifyEmailScreen.kt
  T014 OtpLoginScreen.kt
  (all 4 in parallel — independent files)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T002)
2. Complete Phase 2: Foundational (T003–T010) — **critical**
3. Complete Phase 3: US1 Auth Screens (T011–T014)
4. **STOP and VALIDATE**: Run `make run-web`, walk through all auth screens, verify pill buttons + rounded inputs + no raw color values
5. Deliver polished auth experience as the first increment

### Incremental Delivery

1. Phase 1+2 → Theme + components ready
2. Phase 3 → Auth screens polished (MVP! First impression fixed)
3. Phase 4 → Home dashboard polished (session retention improved)
4. Phase 5 → Navigation polished (cohesion across app)
5. Phase 6 → Onboarding polished (activation experience complete)
6. Phase 7 → Cross-cutting verification

---

## Notes

- No new dependencies required — Material 3 + compose-foundation already in classpath
- All hardcoded color values belong only in `SubjectTrackColors.kt` (the 5 ENEM track colors); everywhere else use `MaterialTheme.colorScheme.*` tokens
- Shape values belong only in `Shape.kt`; everywhere else use `MaterialTheme.shapes.*` tokens
- When editing a screen, remove the old direct `Button`/`OutlinedTextField` imports and add the new component imports
- `make run-web` is the primary preview tool — no backend needed for visual work
- Commit after each phase checkpoint
