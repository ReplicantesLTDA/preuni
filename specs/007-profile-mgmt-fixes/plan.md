# Implementation Plan: Profile Management Fixes & Backend Integration

**Branch**: `007-profile-mgmt-fixes` | **Date**: 2026-04-08 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `specs/007-profile-mgmt-fixes/spec.md`

---

## Summary

Wire the existing Profile (Perfil) section UI to the backend and fix navigation/UX regressions. Backend endpoints already exist. Work is entirely in the KMP shared module and `PreuniApp.kt`. No backend changes needed. Five areas addressed in priority order: logout button (P1), navigation state reset (P1), back arrows (P2), safe-zone compliance (P2), backend integration for all four profile management actions (P3).

---

## Technical Context

**Language/Version**: Kotlin 2.1.20 (KMP shared module), Go 1.23 (backend — no changes)  
**Primary Dependencies**: Compose Multiplatform 1.8.0, Decompose 3.3.0, MVIKotlin 4.2.0, Ktor Client  
**Storage**: `SecureStorage` / `TokenStore` (local session only); no new persistent storage  
**Testing**: `kotlin.test` + `DefaultStoreFactory` on JVM target (`desktopTest`)  
**Target Platform**: Android + iOS (KMP shared UI)  
**Project Type**: Mobile app (Kotlin Multiplatform)  
**Performance Goals**: INP <= 200ms for all interactions; logout completes within 5s  
**Constraints**: Offline-safe logout (clear local tokens even if network call fails); safe-zone insets respected on all devices  
**Scale/Scope**: ~15 Kotlin files modified, 2 new store files, 2 new test files

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Test-first for new features | PASS | ChangePasswordStore and ChangeEmailStore have dedicated test files; ProfileStore tests extended |
| Unit tests required for business logic | PASS | All new store reducers/executors covered |
| Integration tests for auth flows | PASS | Existing auth integration test suite covers backend endpoints; store tests use fakes |
| Design system adherence | PASS | Using Icons.AutoMirrored.Filled.ArrowBack + IconButton (existing pattern from OtpLoginScreen:42); OutlinedButton for logout |
| Consistent feedback patterns | PASS | isLoading: Boolean + error: String? on each screen — follows existing EditPasswordScreen pattern |
| Accessible by default | PASS | Back arrows use contentDescription = "Voltar"; logout button is a standard Button |
| Mobile-first | PASS | Safe zone fix is the primary motivation of this feature |
| Coverage floor >= 80% | PASS | New stores fully unit-tested |
| No dead code | PASS | Config.ConfirmNewEmail removed and merged into Config.ChangeEmail store phase |

Post-design re-check: No violations introduced.

---

## Project Structure

### Documentation (this feature)

```text
specs/007-profile-mgmt-fixes/
├── plan.md              ← this file
├── research.md          ← Phase 0 output
├── data-model.md        ← Phase 1 output
├── quickstart.md        ← Phase 1 output
├── contracts/
│   └── auth-repository-extensions.md
└── tasks.md             ← Phase 2 output (created by /speckit.tasks)
```

### Source Code (modified files)

```text
mobile/shared/src/commonMain/kotlin/com/preuni/shared/

  domain/auth/
  └── AuthRepository.kt           ← add changePassword, changeEmailRequest,
                                     changeEmailConfirm, deleteAccount

  data/auth/
  ├── AuthApiClient.kt             ← add 4 new API methods + DTOs
  └── AuthRepositoryImpl.kt        ← implement 4 new AuthRepository methods

  presentation/auth/
  └── OtpLoginScreen.kt            ← add .statusBarsPadding() to Column

  presentation/profile/
  ├── ChangePasswordStore.kt       ← NEW
  ├── ChangeEmailStore.kt          ← NEW
  ├── ProfileStore.kt              ← add authRepository, fix deleteAccount, add Label.UsernameSaved
  ├── ProfileComponent.kt          ← add authRepository, resetToRoot(), shared store
                                     instances, remove Config.ConfirmNewEmail
  ├── ProfileScreen.kt             ← add onLogout + Sair button
  ├── EditUsernameScreen.kt        ← add back arrow + isLoading/error params
  ├── EditPasswordScreen.kt        ← add back arrow
  ├── ChangeEmailScreen.kt         ← add onBack param + back arrow
  └── ConfirmNewEmailScreen.kt     ← add onBack param + back arrow

  presentation/main/
  └── MainComponent.kt             ← call profileComponent.resetToRoot() in selectTab()

  PreuniApp.kt                     ← wire all new stores, callbacks, label collectors

mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/profile/
  ├── ChangePasswordStoreTest.kt   ← NEW
  ├── ChangeEmailStoreTest.kt      ← NEW
  └── ProfileStoreTest.kt          ← add account-deletion and UsernameSaved tests
```

**Structure Decision**: Single KMP shared module. All changes in `mobile/shared` and `PreuniApp.kt`. No new modules or build targets required.

---

## Implementation Reference

### US1 — Logout Button

1. `ProfileScreen.kt`: Add `onLogout: () -> Unit` param. Add `OutlinedButton(onClick = onLogout, modifier = Modifier.fillMaxWidth()) { Text("Sair") }` above the Delete Account button section.
2. `ProfileComponent.kt`: Add `fun logout() = onLogout()`.
3. `PreuniApp.kt`: Pass `onLogout = component::logout` to ProfileScreen.

### US2 — Navigation State Reset

4. `ProfileComponent.kt`: Add `fun resetToRoot() = navigation.navigate { listOf(Config.Profile) }`.
5. `MainComponent.kt`: In `selectTab()`, add guard:
   ```
   if (_selectedTab.value == BottomTab.PROFILE && tab != BottomTab.PROFILE) {
       profileComponent.resetToRoot()
   }
   ```

### US3 — Back Arrows

6. `EditUsernameScreen.kt`: Add `IconButton(onClick = onBack) { Icon(ArrowBack, "Voltar") }` before the title Text.
7. `EditPasswordScreen.kt`: Same pattern.
8. `ChangeEmailScreen.kt`: Add `onBack: () -> Unit` param + same `IconButton`.
9. `ConfirmNewEmailScreen.kt`: Add `onBack: () -> Unit` param + same `IconButton`.
10. `PreuniApp.kt`: Pass `onBack = component::navigateBack` to ChangeEmailScreen and ConfirmNewEmailScreen.

### US4 — Safe Zone

11. `OtpLoginScreen.kt`: Add `.statusBarsPadding()` to Column modifier before `.padding(horizontal = 24.dp)`.
12. All profile sub-page screens and ProfileScreen: Add `.statusBarsPadding()` before `.padding(horizontal = 24.dp, vertical = 32.dp)`.

### US5 — Backend Integration

#### 5a. New AuthApiClient methods (AuthApiClient.kt)

- `changePassword(current: String, new: String): Result<Unit>` — POST v1/auth/password/change
- `changeEmailRequest(newEmail: String): Result<Unit>` — POST v1/auth/email/change/request
- `changeEmailConfirm(newEmail: String, otp: String): Result<Unit>` — POST v1/auth/email/change/confirm
- `deleteAccount(): Result<Unit>` — DELETE v1/auth/account

#### 5b. Domain + repository wiring

13. `AuthRepository.kt`: Declare all four methods.
14. `AuthRepositoryImpl.kt`: Delegate to apiClient.

#### 5c. ProfileStore — authRepository + fixes

15. `ProfileStoreFactory`: Add `authRepository: AuthRepository` param. Fix `deleteAccount()` to call `authRepository.deleteAccount()`. Add `Label.UsernameSaved` published on `updateProfile` success.

#### 5d. New stores

16. `ChangePasswordStore.kt` (new): Intent `Submit(current, new)` -> `authRepository.changePassword()` -> `Label.Saved`. State: `isLoading`, `error`.
17. `ChangeEmailStore.kt` (new): Two-phase store (REQUEST / CONFIRM). Intents: `RequestCode(email)`, `ConfirmCode(otp)`, `ResendCode`, `ClearError`. `resendCooldown` countdown via coroutine. Labels: `EmailChanged(newEmail)`.

#### 5e. ProfileComponent refactor

18. `ProfileComponent.kt`:
    - Add `authRepository: AuthRepository` constructor param.
    - Create `profileStore`, `changePasswordStore`, `changeEmailStore` at class level (shared instances).
    - Remove `Config.ConfirmNewEmail` — phase managed inside ChangeEmailStore.
    - Child types: `EditUsername(store: ProfileStore)`, `EditPassword(store: ChangePasswordStore)`, `ChangeEmail(store: ChangeEmailStore)`, `DeleteAccount(store: ProfileStore)`.

#### 5f. MainComponent — pass authRepository

19. `MainComponent.kt`: Pass `authRepository = authRepository` to ProfileComponent constructor.

#### 5g. PreuniApp.kt full wiring

20. `PreuniApp.kt` — ProfileContent:
    - `Child.Profile`: `onLogout = component::logout`.
    - `Child.EditUsername`: Dispatch `ProfileStore.Intent.UpdateUsername(username)` on save. Collect `ProfileStore.Label.UsernameSaved` -> `component.navigateBack()`. Pass current username from store state.
    - `Child.EditPassword`: Dispatch `ChangePasswordStore.Intent.Submit(current, new)` on save. Collect `ChangePasswordStore.Label.Saved` -> `component.navigateBack()`.
    - `Child.ChangeEmail`: If `store.state.phase == REQUEST` show `ChangeEmailScreen(onSubmit = { store.accept(RequestCode(it)) })`, else show `ConfirmNewEmailScreen(onSubmit = { store.accept(ConfirmCode(it)) }, onResend = { store.accept(ResendCode) })`. Collect `ChangeEmailStore.Label.EmailChanged` -> `component.navigateBack()`.
    - `Child.DeleteAccount`: Dispatch `ProfileStore.Intent.ConfirmDeleteAccount` on confirm. Collect `ProfileStore.Label.AccountDeleted` -> RootComponent handles final navigation via the existing `onLogout` chain.

---

## Test Plan

### ChangePasswordStoreTest.kt (new)

- Submit -> network success -> Label.Saved published, isLoading = false
- Submit -> wrong current password (422) -> error set, Label.Saved NOT published
- Submit -> network error -> AppError.NetworkError in state
- ClearError -> error = null

### ChangeEmailStoreTest.kt (new)

- RequestCode with valid email -> success -> phase = CONFIRM, resendCooldown = 60
- RequestCode -> conflict (409) -> error set, phase stays REQUEST
- ConfirmCode with valid OTP -> success -> Label.EmailChanged published
- ConfirmCode with wrong OTP (422) -> error set, Label NOT published
- ResendCode -> resets cooldown = 60, re-triggers network call

### ProfileStoreTest.kt (additions)

- ConfirmDeleteAccount -> calls authRepository.deleteAccount() (verified via fake)
- ConfirmDeleteAccount -> success -> Label.AccountDeleted published
- UpdateUsername -> success -> Label.UsernameSaved published

---

## Complexity Tracking

No constitution violations. No new abstractions beyond the minimum required stores.
