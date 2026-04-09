# Quickstart: Profile Management Fixes & Backend Integration

**Feature**: 007-profile-mgmt-fixes  
**Date**: 2026-04-08

---

## What this feature does

Fixes three categories of bugs in the Profile (Perfil) section and adds the missing backend integration for all account management actions.

**Bug fixes**:
1. Logout button added to the Perfil screen (urgent — none existed before)
2. Profile sub-pages no longer retain navigation state when the user switches tabs
3. All profile sub-pages now have a visible back arrow
4. All screens with top navigation controls now respect device safe zones (notch, Dynamic Island, status bar)

**Backend integration** (all endpoints already exist — only mobile wiring was missing):
- Change username → `PATCH /v1/students/me`
- Change password → `POST /v1/auth/password/change`
- Change email (two-phase OTP flow) → `POST /v1/auth/email/change/request` + confirm
- Delete account → `DELETE /v1/auth/account`

---

## How to run tests

```bash
# Run KMP shared module tests (fast, JVM target)
cd mobile/shared && ./gradlew desktopTest

# Run specific test class
cd mobile/shared && ./gradlew desktopTest --tests "*.ChangePasswordStoreTest"
cd mobile/shared && ./gradlew desktopTest --tests "*.ChangeEmailStoreTest"
cd mobile/shared && ./gradlew desktopTest --tests "*.ProfileStoreTest"
```

---

## Key design decisions

| Decision | Rationale |
|----------|-----------|
| `ChangePasswordStore` + `ChangeEmailStore` as separate files | Follows `ChangeTrackStore` pattern; AuthRepository is separate from UserRepository |
| `Config.ConfirmNewEmail` removed; phase lives in `ChangeEmailStore.State` | Email change is one logical flow; splitting into two Decompose configs complicates store lifetime |
| `ProfileStore` shared instance at `ProfileComponent` class level | EditUsername and DeleteAccount sub-pages need to read/write profile state |
| `navigation.navigate { listOf(Config.Profile) }` for resetToRoot | Decompose 3.x standard; replaces entire back stack with root config |
| `Modifier.statusBarsPadding()` (not `safeContentPadding`) | Only the status bar (top inset) needs compensation here; bottom is handled by Scaffold |
| Account deletion uses `DELETE /v1/auth/account` (auth-svc), not `/v1/students/me` | auth-svc endpoint revokes tokens AND anonymizes credentials; user-svc alone leaves tokens active |

---

## Files changed at a glance

| File | Change type |
|------|-------------|
| `domain/auth/AuthRepository.kt` | Add 4 method declarations |
| `data/auth/AuthApiClient.kt` | Add 4 API methods + request DTOs |
| `data/auth/AuthRepositoryImpl.kt` | Implement 4 new methods |
| `presentation/auth/OtpLoginScreen.kt` | statusBarsPadding() on Column |
| `presentation/profile/ChangePasswordStore.kt` | NEW |
| `presentation/profile/ChangeEmailStore.kt` | NEW |
| `presentation/profile/ProfileStore.kt` | authRepository dep, deleteAccount fix, UsernameSaved label |
| `presentation/profile/ProfileComponent.kt` | authRepository, resetToRoot(), shared stores, remove ConfirmNewEmail config |
| `presentation/profile/ProfileScreen.kt` | onLogout + Sair button |
| `presentation/profile/EditUsernameScreen.kt` | Back arrow + store params |
| `presentation/profile/EditPasswordScreen.kt` | Back arrow |
| `presentation/profile/ChangeEmailScreen.kt` | onBack + back arrow |
| `presentation/profile/ConfirmNewEmailScreen.kt` | onBack + back arrow |
| `presentation/main/MainComponent.kt` | resetToRoot() on tab switch + authRepository forwarding |
| `PreuniApp.kt` | Full ProfileContent rewiring |
| `test/.../ChangePasswordStoreTest.kt` | NEW |
| `test/.../ChangeEmailStoreTest.kt` | NEW |
| `test/.../ProfileStoreTest.kt` | Add 3 test cases |
