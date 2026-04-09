# Tasks: Profile Management Fixes & Backend Integration

**Input**: Design documents from `specs/007-profile-mgmt-fixes/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Tests**: Included for new MVIKotlin stores (`ChangePasswordStore`, `ChangeEmailStore`) and ProfileStore additions per Constitution (unit tests required for all business logic).

**Organization**: Tasks grouped by user story — each phase is independently deliverable and testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no mutual dependencies)
- **[Story]**: Maps to user story from spec.md (US1–US5)

---

## Phase 1: Setup

**Purpose**: Verify baseline before any changes.

- [X] T001 Confirm current branch is `007-profile-mgmt-fixes` and run `cd mobile/shared && ./gradlew desktopTest` to establish a passing baseline

---

## Phase 2: Foundational — Auth Data Layer Extension

**Purpose**: Extend the auth data layer with the 4 missing API methods. Required by US5 (P3 backend integration). US1–US4 (pure UI/nav fixes) are independent of this phase and can proceed in parallel.

**Note**: No user story label — these are shared infrastructure tasks.

- [X] T002 Add method signatures `changePassword(currentPassword: String, newPassword: String): Result<Unit>`, `changeEmailRequest(newEmail: String): Result<Unit>`, `changeEmailConfirm(newEmail: String, otp: String): Result<Unit>`, and `deleteAccount(): Result<Unit>` to `mobile/shared/src/commonMain/kotlin/com/preuni/shared/domain/auth/AuthRepository.kt`

- [X] T003 Add serializable request DTOs (`ChangePasswordRequest`, `ChangeEmailRequestBody`, `ChangeEmailConfirmRequest`) and 4 new methods — `changePassword()` (POST v1/auth/password/change), `changeEmailRequest()` (POST v1/auth/email/change/request), `changeEmailConfirm()` (POST v1/auth/email/change/confirm), `deleteAccount()` (DELETE v1/auth/account) — to `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/auth/AuthApiClient.kt`

- [X] T004 Implement the 4 new `AuthRepository` methods in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/auth/AuthRepositoryImpl.kt` by delegating to the corresponding `AuthApiClient` methods added in T003

**Checkpoint**: Auth data layer complete. US5 store tasks can now begin. US1–US4 may have been running in parallel.

---

## Phase 3: User Story 1 — Logout Button (Priority: P1) 🎯 MVP

**Goal**: A logged-in user can tap "Sair" in the Perfil tab to sign out and be redirected to the welcome/login screen.

**Independent Test**: Navigate to Perfil tab → tap "Sair" → confirm app returns to login/welcome screen and session is cleared.

- [X] T005 [US1] Add `fun logout() = onLogout()` public method to `ProfileComponent` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileComponent.kt`

- [X] T006 [US1] Add `onLogout: () -> Unit` parameter to `ProfileScreen` and add a full-width `OutlinedButton(onClick = onLogout) { Text("Sair") }` between the HorizontalDivider and the Delete Account button in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileScreen.kt`

- [X] T007 [US1] Wire `onLogout = component::logout` in the `ProfileComponent.Child.Profile` branch of `ProfileContent` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`

**Checkpoint**: Tap "Sair" in Perfil → redirected to login. Logout works offline (tokens cleared even if network fails — already handled by `AuthRepositoryImpl.logout()`).

---

## Phase 4: User Story 2 — Profile Navigation State Reset (Priority: P1)

**Goal**: Returning to the Perfil tab always shows the main Profile overview, regardless of which sub-page the user was on before switching tabs.

**Independent Test**: Open Change Username sub-page → switch to Aprender tab → return to Perfil tab → confirm main Profile overview is shown (not Change Username).

- [X] T008 [US2] Add `fun resetToRoot() = navigation.navigate { listOf(Config.Profile) }` to `ProfileComponent` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileComponent.kt`

- [X] T009 [US2] Update `selectTab()` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/main/MainComponent.kt` to call `profileComponent.resetToRoot()` before updating `_selectedTab` when the current tab is `BottomTab.PROFILE` and the new tab is different:
  ```kotlin
  fun selectTab(tab: BottomTab) {
      if (_selectedTab.value == BottomTab.PROFILE && tab != BottomTab.PROFILE) {
          profileComponent.resetToRoot()
      }
      _selectedTab.value = tab
  }
  ```

**Checkpoint**: Profile navigation stack resets on tab switch. Leaving and returning to Perfil always lands on main overview.

---

## Phase 5: User Story 3 — Back Navigation in Profile Sub-pages (Priority: P2)

**Goal**: Every Profile sub-page has a visible back arrow that returns the user to the main Profile overview.

**Independent Test**: Open any Profile sub-page → tap back arrow → confirm return to main Profile overview.

- [X] T010 [P] [US3] Add `IconButton(onClick = onBack, modifier = Modifier.align(Alignment.Start)) { Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Voltar") }` at the top of the Column (before the title `Text`) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/EditUsernameScreen.kt`

- [X] T011 [P] [US3] Add the same `IconButton(ArrowBack, "Voltar")` at the top of the Column in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/EditPasswordScreen.kt`

- [X] T012 [P] [US3] Add `onBack: () -> Unit` parameter and `IconButton(ArrowBack, "Voltar")` at the top of the Column in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ChangeEmailScreen.kt`

- [X] T013 [P] [US3] Add `onBack: () -> Unit` parameter and `IconButton(ArrowBack, "Voltar")` at the top of the Column in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ConfirmNewEmailScreen.kt`

- [X] T014 [US3] Pass `onBack = component::navigateBack` to `ChangeEmailScreen` and `ConfirmNewEmailScreen` calls in the `ProfileContent` composable in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`

**Checkpoint**: All 4 sub-pages (EditUsername, EditPassword, ChangeEmail, ConfirmNewEmail) show a tappable back arrow. System back button also works (already handled by `handleBackButton = true` on the child stack).

---

## Phase 6: User Story 4 — Safe Zone Compliance (Priority: P2)

**Goal**: All back arrows and navigation controls are rendered below the status bar, notch, and camera cutout on every device.

**Independent Test**: On a device with a notch or Dynamic Island, open the Login-with-Code screen and all Profile sub-pages — confirm back arrows are fully visible and tappable below the status bar.

- [X] T015 [P] [US4] Add `.statusBarsPadding()` to the `Column` modifier chain in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/auth/OtpLoginScreen.kt`, before `.padding(horizontal = 24.dp)`: change to `Modifier.fillMaxSize().statusBarsPadding().padding(horizontal = 24.dp)`

- [X] T016 [P] [US4] Add `.statusBarsPadding()` before `.padding(horizontal = 24.dp, vertical = 32.dp)` on the outer `Column` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileScreen.kt`

- [X] T017 [P] [US4] Add `.statusBarsPadding()` before `.padding(horizontal = 24.dp, vertical = 32.dp)` on the outer `Column` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/EditUsernameScreen.kt`

- [X] T018 [P] [US4] Add `.statusBarsPadding()` before `.padding(horizontal = 24.dp, vertical = 32.dp)` on the outer `Column` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/EditPasswordScreen.kt`

- [X] T019 [P] [US4] Add `.statusBarsPadding()` before `.padding(horizontal = 24.dp, vertical = 32.dp)` on the outer `Column` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ChangeEmailScreen.kt`

- [X] T020 [P] [US4] Add `.statusBarsPadding()` before `.padding(horizontal = 24.dp, vertical = 32.dp)` on the outer `Column` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ConfirmNewEmailScreen.kt`

- [X] T021 [P] [US4] Add `.statusBarsPadding()` before `.padding(horizontal = 24.dp, vertical = 32.dp)` on the outer `Column` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/DeleteAccountScreen.kt`

**Checkpoint**: All 7 screens apply status bar inset. Navigation controls no longer overlap system UI on notched devices.

---

## Phase 7: User Story 5 — Backend Integration (Priority: P3)

**Goal**: Change username, change password, change email (OTP flow), and delete account all communicate with the backend and provide success/error feedback.

**Independent Test per action**:
- Username: Change username → confirm new name appears in Profile overview after navigating back
- Password: Change password with wrong current → error shown; change with correct current → success
- Email: Submit new email → OTP code screen appears; enter correct OTP → Profile shows updated email
- Delete: Type DELETE + confirm → account deleted, redirected to login

**⚠️ Constitution requires test-first for new stores: write failing tests BEFORE implementing.**

### Tests (write first — they must fail before T024/T025 are implemented)

- [X] T022 [US5] Write unit tests for `ChangePasswordStore` in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/profile/ChangePasswordStoreTest.kt`:
  - `Submit with valid credentials → network success → Label.Saved published, isLoading = false`
  - `Submit with wrong current password (422) → error state set, Label.Saved not published`
  - `Submit with network error → AppError.NetworkError in state`
  - `ClearError intent → error = null`

- [X] T023 [P] [US5] Write unit tests for `ChangeEmailStore` in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/profile/ChangeEmailStoreTest.kt`:
  - `RequestCode with valid email → success → phase = CONFIRM, resendCooldown = 60`
  - `RequestCode with conflict (409) → error set, phase stays REQUEST`
  - `ConfirmCode with valid OTP → success → Label.EmailChanged(newEmail) published`
  - `ConfirmCode with invalid OTP (422) → error set, Label not published`
  - `ResendCode → resets cooldown = 60 and re-calls requestCode`

- [X] T027 [P] [US5] Add 3 failing test cases to `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/profile/ProfileStoreTest.kt`:
  - `ConfirmDeleteAccount → calls authRepository.deleteAccount() (not userRepository.anonymize())`
  - `ConfirmDeleteAccount → success → Label.AccountDeleted published`
  - `UpdateUsername → success → Label.UsernameSaved published`

### Implementation

- [X] T024 [US5] Create `ChangePasswordStore` (interface + `ChangePasswordStoreFactory`) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ChangePasswordStore.kt`. Constructor takes `authRepository: AuthRepository`. Intent: `Submit(currentPassword: String, newPassword: String)`, `ClearError`. State: `isLoading: Boolean`, `error: String?`. Label: `Saved`. Executor calls `authRepository.changePassword(current, new)`.

- [X] T025 [P] [US5] Create `ChangeEmailStore` (interface + `ChangeEmailStoreFactory`) in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ChangeEmailStore.kt`. Constructor takes `authRepository: AuthRepository`. `Phase` enum: `REQUEST`, `CONFIRM`. State: `newEmail`, `phase`, `isLoading`, `error: String?`, `resendCooldown: Int`. Intents: `RequestCode(newEmail)`, `ConfirmCode(otp)`, `ResendCode`, `ClearError`. Label: `EmailChanged(newEmail: String)`. Executor handles resend cooldown via coroutine `repeat` tick. `RequestCode` success → `phase = CONFIRM`, `resendCooldown = 60`. `ConfirmCode` success → publish `Label.EmailChanged`.

- [X] T026 [US5] Update `ProfileStoreFactory` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileStore.kt`:
  - Add `authRepository: AuthRepository` constructor parameter
  - In `deleteAccount()` executor: replace `repo.anonymize()` with `authRepository.deleteAccount()`
  - After successful `updateProfile(username = ...)` in `updateProfile()`: publish `Label.UsernameSaved`
  - Add `data object UsernameSaved : Label` to the `Label` sealed interface

- [X] T028 [US5] Refactor `ProfileComponent` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileComponent.kt`:
  - Add `authRepository: AuthRepository` constructor parameter
  - Replace per-`createChild` `ProfileStore` creation with a class-level `val profileStore = ProfileStoreFactory(storeFactory, userRepository, authRepository).create()`
  - Add class-level `val changePasswordStore = ChangePasswordStoreFactory(storeFactory, authRepository).create()`
  - Add class-level `val changeEmailStore = ChangeEmailStoreFactory(storeFactory, authRepository).create()`
  - Remove `Config.ConfirmNewEmail` config and `navigateToConfirmNewEmail()` method (phase is now managed by `ChangeEmailStore.State.phase`)
  - Update `Child` sealed interface: `EditUsername(val store: ProfileStore)`, `EditPassword(val store: ChangePasswordStore)`, `ChangeEmail(val store: ChangeEmailStore)`, `DeleteAccount(val store: ProfileStore)`
  - Update `createChild` to return the class-level store instances for each config

- [X] T029 [US5] Pass `authRepository = authRepository` in the `ProfileComponent(...)` constructor call inside `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/main/MainComponent.kt`

- [X] T030 [US5] Fully rewire `ProfileContent` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt` for all backend-integrated branches:
  - `Child.EditUsername`: pass `currentUsername = child.store.state.student?.username ?: ""` and `isLoading = child.store.state.isLoading`; on `onSave(username)` dispatch `ProfileStore.Intent.UpdateUsername(username)`; collect `ProfileStore.Label.UsernameSaved` → `component.navigateBack()`
  - `Child.EditPassword`: pass `store = child.store`; on `onSave(current, new)` dispatch `ChangePasswordStore.Intent.Submit(current, new)`; pass `error = child.store.state.error`; collect `ChangePasswordStore.Label.Saved` → `component.navigateBack()`; pass `isLoading = child.store.state.isLoading`
  - `Child.ChangeEmail`: if `child.store.state.phase == REQUEST` render `ChangeEmailScreen` wired to `RequestCode`; else render `ConfirmNewEmailScreen` wired to `ConfirmCode`/`ResendCode`; collect `ChangeEmailStore.Label.EmailChanged` → `component.navigateBack()`; wire `onBack = component::navigateBack` for both composables
  - `Child.DeleteAccount`: on `onConfirm` dispatch `ProfileStore.Intent.ConfirmDeleteAccount`; pass `isLoading = child.store.state.isLoading`; collect `ProfileStore.Label.AccountDeleted` — this triggers `component.logout()` which propagates through `MainComponent.onLogout` → `RootComponent` → `navigation.replaceAll(Config.Auth)`

- [X] T031 [US5] Run `cd mobile/shared && ./gradlew desktopTest` and confirm all 3 new test files pass (ChangePasswordStoreTest, ChangeEmailStoreTest, ProfileStoreTest additions)

**Checkpoint**: All four profile management actions work end-to-end with the backend. Success/error states displayed. Logout after account deletion navigates to login screen.

---

## Phase 8: Polish & Cross-Cutting

**Purpose**: Final validation and cleanup.

- [X] T032 Run full test suite `cd mobile/shared && ./gradlew desktopTest` and confirm no regressions from any phase

- [X] T033 [P] Review `PreuniApp.kt` ProfileContent for any remaining `TODO`/stub callbacks (empty `{}` lambdas, hardcoded `""` values) and confirm none remain

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: No dependencies — start immediately (or in parallel with US1–US4)
- **US1 (Phase 3)**: Independent of Foundation — can start before T002–T004
- **US2 (Phase 4)**: Independent of Foundation — can start before T002–T004
- **US3 (Phase 5)**: Independent of Foundation — can start before T002–T004
- **US4 (Phase 6)**: Independent of Foundation and all other stories — all 7 tasks parallelizable
- **US5 (Phase 7)**: Requires Foundation (T002–T004) complete before T024–T030
- **Polish (Phase 8)**: Requires all desired phases complete

### Story-Level Dependencies

| Story | Depends on | Can be parallelized with |
|-------|-----------|--------------------------|
| US1 | None | US2, US3, US4, Foundation |
| US2 | None | US1, US3, US4, Foundation |
| US3 | None | US1, US2, US4, Foundation |
| US4 | None | US1, US2, US3, Foundation |
| US5 | Foundation (T002–T004) | — |

### Within US5

- T022, T023, T027 (test tasks) can run in parallel — different files
- T024, T025 (store creation) can run in parallel — different files; depend on T022/T023 being written first
- T026 depends on T002 (AuthRepository interface)
- T028 depends on T024, T025, T026
- T029 depends on T028
- T030 depends on T028, T029

---

## Parallel Opportunities

### Phase 2 + US1–US4 simultaneously

```
# All of these can run in parallel from the start:
Foundation: T002 → T003 → T004
US1:        T005 → T006 → T007
US2:        T008 → T009
US3:        T010, T011, T012, T013 (all parallel) → T014
US4:        T015, T016, T017, T018, T019, T020, T021 (all parallel)
```

### Within US5 (after Foundation complete)

```
# Write all tests in parallel:
T022 (ChangePasswordStoreTest)
T023 (ChangeEmailStoreTest)
T027 (ProfileStoreTest additions)

# Then implement stores in parallel:
T024 (ChangePasswordStore)
T025 (ChangeEmailStore)
T026 (ProfileStore fix)

# Then sequential:
T028 (ProfileComponent refactor) → T029 (MainComponent) → T030 (PreuniApp wiring) → T031 (run tests)
```

---

## Implementation Strategy

### MVP First (US1 only — 3 tasks, urgent)

1. Complete T001 (baseline)
2. Complete T005 → T006 → T007 (logout button — 3 tasks)
3. **STOP and TEST**: Tap "Sair" → redirected to login
4. Deploy to testers immediately — logout is urgent

### Recommended Delivery Order

1. T001 (baseline) + T002–T004 (foundation) in parallel with T005–T007 (US1 logout)
2. T008–T009 (US2 nav reset) + T010–T014 (US3 back arrows) — quick wins
3. T015–T021 (US4 safe zone) — all parallel, fast
4. T022–T031 (US5 backend integration) — most complex, requires Foundation

### Total Task Count

- **Phase 1**: 1 task
- **Phase 2**: 3 tasks
- **Phase 3 (US1)**: 3 tasks
- **Phase 4 (US2)**: 2 tasks
- **Phase 5 (US3)**: 5 tasks
- **Phase 6 (US4)**: 7 tasks
- **Phase 7 (US5)**: 12 tasks
- **Phase 8**: 2 tasks
- **Total**: **35 tasks**

---

## Notes

- `[P]` tasks touch different files — safe to implement simultaneously
- Each user story phase is independently demonstrable before starting the next
- T005–T007 (logout) should be the first thing delivered given its urgent priority
- US4 (safe zone) tasks are all mechanical modifier additions — parallelizable across the whole team
- US5 test tasks (T022, T023, T027) must be written and confirmed failing before T024/T025/T026
- `Config.ConfirmNewEmail` is deleted in T028 — confirm T013/T014 (ConfirmNewEmailScreen changes from US3) are done first so the composable is ready before it's wired in T030
