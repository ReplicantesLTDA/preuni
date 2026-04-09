# Data Model: Profile Management Fixes & Backend Integration

**Feature**: 007-profile-mgmt-fixes  
**Date**: 2026-04-08

No new persistent data entities are introduced by this feature. All data is already modelled in existing domain types. This document records which fields are read/written by the new operations and the state shapes of the two new stores.

---

## Existing Domain Types Used

### `Student` (read + username write)

```
Student {
  id: String
  displayName: String
  username: String         ← written by Change Username action
  email: String            ← written by Change Email action
  avatarUrl: String?
  xpTotal: Long
  streakCount: Int
  readinessScore: Double
  onboardingCompleted: Boolean
}
```

Updated via `PATCH /v1/students/me` (username/displayName) and `POST /v1/auth/email/change/confirm` (email). `ProfileStore.State.student` is refreshed on success.

---

### `AuthSession` / `TokenStore` (written on logout + account deletion)

```
TokenStore keys: userId, userEmail, accessToken, refreshToken, activeTrackId
```

All keys are cleared on logout and on account deletion. No new keys added.

---

## New Store State Shapes

### `ChangePasswordStore.State`

```
State {
  currentPassword: String   — input field
  newPassword: String       — input field
  confirmPassword: String   — input field
  isLoading: Boolean        — true during API request
  error: String?            — error message to display
}
```

**Transitions**:
- `Submit` intent → isLoading = true → on success: publish `Label.Saved`, isLoading = false → on failure: error = message, isLoading = false

---

### `ChangeEmailStore.State`

```
Phase: enum { REQUEST, CONFIRM }

State {
  newEmail: String          — email entered in phase REQUEST
  otp: String               — OTP entered in phase CONFIRM
  phase: Phase              — controls which composable is rendered
  isLoading: Boolean
  error: String?
  resendCooldown: Int       — seconds remaining before resend is allowed (counts down 60→0)
}
```

**Transitions**:
- `RequestCode(email)` → isLoading = true → on success: phase = CONFIRM, resendCooldown = 60, isLoading = false → on failure: error = message
- `ConfirmCode(otp)` → isLoading = true → on success: publish `Label.EmailChanged`, isLoading = false → on failure: error = message
- `ResendCode` → resets resendCooldown = 60, re-dispatches RequestCode(newEmail)
- `TickCooldown` — internal tick: resendCooldown = max(0, resendCooldown - 1)

---

## Validation Rules (unchanged)

All existing `AuthValidator` rules apply:
- Username: lowercase letters, digits, `-`, `_`; length 3–30
- Email: standard RFC 5322 format
- Password: minimum 6 characters; strength meter shown during registration (not re-applied on change)
- Current password: non-empty (correctness validated server-side)
- OTP: exactly 6 digits
