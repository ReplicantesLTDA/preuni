# Research: Profile Management Fixes & Backend Integration

**Feature**: 007-profile-mgmt-fixes  
**Date**: 2026-04-08  
**Status**: Complete — all unknowns resolved from code analysis

---

## Finding 1: Backend endpoints already exist

**Decision**: No backend changes required.  
**Rationale**: All required API endpoints are already implemented and deployed:

| Action | Endpoint | Service |
|--------|----------|---------|
| Logout | `POST /v1/auth/logout` | auth-svc |
| Update username | `PATCH /v1/students/me` | user-svc |
| Change password | `POST /v1/auth/password/change` | auth-svc |
| Request email change | `POST /v1/auth/email/change/request` | auth-svc |
| Confirm email change | `POST /v1/auth/email/change/confirm` | auth-svc |
| Delete account | `DELETE /v1/auth/account` | auth-svc (calls user-svc internally) |

**Important**: Account deletion must use `DELETE /v1/auth/account` (auth-svc), NOT `DELETE /v1/students/me` (user-svc). The auth-svc endpoint revokes all tokens AND anonymizes credentials, then calls user-svc internally. The current mobile code incorrectly calls user-svc directly via `UserApiClient.anonymize()`.

---

## Finding 2: Mobile data layer gaps

**Decision**: Add four new methods to `AuthApiClient` + `AuthRepository` interface + `AuthRepositoryImpl`.  
**Rationale**: The following are missing:
- `changePassword(currentPassword, newPassword): Result<Unit>`
- `changeEmailRequest(newEmail): Result<Unit>`
- `changeEmailConfirm(newEmail, otp): Result<Unit>`
- `deleteAccount(): Result<Unit>`

**Username update already exists**: `UserApiClient.updateProfile(username)` → `PATCH /v1/students/me` is fully implemented. `ProfileStore.UpdateUsername` dispatches to it correctly. Only the wiring in `PreuniApp.kt` is missing (currently calls `component.navigateBack()` without dispatching the intent).

---

## Finding 3: ProfileComponent navigation stack is not reset on tab switch

**Decision**: Add `fun resetToRoot()` to `ProfileComponent` and call it from `MainComponent.selectTab()` when navigating away from the PROFILE tab.  
**Rationale**: `MainComponent` uses `MutableStateFlow<BottomTab>` for tab switching — it does not destroy or recreate the `ProfileComponent`. The ProfileComponent's `StackNavigation` back stack survives tab switches, causing the user to land on sub-pages instead of the main Profile overview when returning.

**Implementation**: `navigation.navigate { listOf(Config.Profile) }` resets the stack to just the root. Call this in `MainComponent.selectTab()` when `_selectedTab.value == BottomTab.PROFILE && tab != BottomTab.PROFILE`.

---

## Finding 4: Visible back arrows are missing from profile sub-pages

**Decision**: Add `IconButton(Icons.AutoMirrored.Filled.ArrowBack)` to EditUsernameScreen, EditPasswordScreen, ChangeEmailScreen, and ConfirmNewEmailScreen. Add `onBack: () -> Unit` parameter to ChangeEmailScreen and ConfirmNewEmailScreen (currently missing).  
**Rationale**: `OtpLoginScreen` (auth) already demonstrates the correct pattern at line 42. The `handleBackButton = true` flag on ProfileComponent's childStack handles Android system back, but there is no visible on-screen back arrow in any profile sub-page screen.  
**Alternatives considered**: Using a shared `TopAppBar` composable — rejected as over-engineering for this scope; a simple `IconButton` at the top of each screen is sufficient.

---

## Finding 5: Safe zone compliance — statusBarsPadding required

**Decision**: Apply `Modifier.statusBarsPadding()` to the outermost scrollable Column of every screen that renders navigation controls near the top of the screen.  
**Rationale**: All affected screens use `.padding(horizontal = 24.dp)` with no top-inset compensation. On Android, `Modifier.statusBarsPadding()` is available via `androidx.compose.foundation.layout`. In Compose Multiplatform 1.8.0 on iOS, CMP maps `WindowInsets.statusBars` to the iOS `safeAreaInsets.top`, so the same API works cross-platform.  
**Affected screens**: `OtpLoginScreen` (back arrow at line 42 with no top padding), `EditUsernameScreen`, `EditPasswordScreen`, `ChangeEmailScreen`, `ConfirmNewEmailScreen`. `ProfileScreen` and `DeleteAccountScreen` do not have navigation controls at the very top so they are lower priority but should also be fixed for consistency.

---

## Finding 6: New stores needed for password and email change

**Decision**: Create `ChangePasswordStore` and `ChangeEmailStore` as new files following the `ChangeTrackStore` pattern.  
**Rationale**: Password change requires `AuthRepository` (not `UserRepository`), and email change is a two-phase flow (request → OTP confirm) with a 60-second resend cooldown. These concerns are distinct from `ProfileStore`'s data-loading role.

**ChangeEmailStore design**: Manages both the request phase and the OTP confirmation phase via a `phase: Phase` state field (`REQUEST` vs `CONFIRM`). This avoids a two-component Decompose config for what is conceptually one flow. The `Config.ConfirmNewEmail` Decompose config is removed; `Config.ChangeEmail` renders different UI based on store phase.

**Alternatives considered**: Adding password/email change directly to ProfileStore — rejected because it would bloat the store and create a mixed responsibility between profile data display and mutating auth credentials.

---

## Finding 7: ProfileStore needs AuthRepository for account deletion

**Decision**: Add `authRepository: AuthRepository` to `ProfileStoreFactory` constructor. Change `deleteAccount()` in ProfileStore to call `authRepository.deleteAccount()` instead of `userRepository.anonymize()`.  
**Rationale**: Account deletion (`DELETE /v1/auth/account`) is an auth-svc operation; calling user-svc's `DELETE /v1/students/me` directly bypasses token revocation and leaves auth credentials in the database.

---

## Finding 8: ProfileStore instance should be shared across sub-page children

**Decision**: Create the `ProfileStore` instance at `ProfileComponent` class level (not inside `createChild`), and pass the same instance to `Child.Profile`, `Child.EditUsername`, and `Child.DeleteAccount`.  
**Rationale**: `EditUsernameScreen` needs the store to dispatch `UpdateUsername` and receive `UsernameSaved` labels. `DeleteAccountScreen` needs the store to dispatch `ConfirmDeleteAccount`. With a shared instance, sub-pages can read current state (e.g., current username) and dispatch intents, and the Profile screen reflects changes immediately when the user navigates back.

---

## Finding 9: Logout exposure from ProfileComponent

**Decision**: Add `fun logout() = onLogout()` public method to `ProfileComponent`. Add `onLogout: () -> Unit` parameter to `ProfileScreen`.  
**Rationale**: `ProfileComponent` receives `onLogout` in its constructor (private) but does not expose it publicly. `ProfileScreen` currently has no logout callback. The logout lambda is already fully wired through `MainComponent → RootComponent → navigation.replaceAll(Config.Auth)`.
