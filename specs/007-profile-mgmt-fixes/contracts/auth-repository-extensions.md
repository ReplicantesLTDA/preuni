# Contract: AuthRepository Extensions

**Feature**: 007-profile-mgmt-fixes  
**Layer**: Domain interface (`mobile/shared/.../domain/auth/AuthRepository.kt`)  
**Direction**: Mobile app → Backend (auth-svc via NGINX gateway)

All new methods require a valid `Authorization: Bearer <access_token>` header, which `ApiClient` adds automatically via its `defaultRequest` block.

---

## New Methods on `AuthRepository`

### `changePassword(currentPassword: String, newPassword: String): Result<Unit>`

**Backend endpoint**: `POST /v1/auth/password/change`

Request body:
```json
{ "current_password": "...", "new_password": "..." }
```

Responses:
- `204 No Content` → `Result.success(Unit)`
- `422 Unprocessable Entity` (wrong current password) → `Result.failure(AppError.Validation(field="current_password", message=...))`
- `401 Unauthorized` → `Result.failure(AppError.Unauthorized())`

Side effect on success: backend revokes all refresh tokens for this user. The mobile app does NOT need to re-login; the current access token remains valid until it expires.

---

### `changeEmailRequest(newEmail: String): Result<Unit>`

**Backend endpoint**: `POST /v1/auth/email/change/request`

Request body:
```json
{ "new_email": "..." }
```

Responses:
- `204 No Content` → `Result.success(Unit)` (OTP sent to new email)
- `409 Conflict` (email already in use) → `Result.failure(AppError.Conflict(message=...))`
- `422 Unprocessable Entity` → `Result.failure(AppError.Validation(...))`

---

### `changeEmailConfirm(newEmail: String, otp: String): Result<Unit>`

**Backend endpoint**: `POST /v1/auth/email/change/confirm`

Request body:
```json
{ "new_email": "...", "otp": "..." }
```

Responses:
- `204 No Content` → `Result.success(Unit)` (email updated)
- `422 Unprocessable Entity` (invalid/expired OTP) → `Result.failure(AppError.Validation(field="otp", message=...))`

---

### `deleteAccount(): Result<Unit>`

**Backend endpoint**: `DELETE /v1/auth/account`

No request body.

Responses:
- `204 No Content` → `Result.success(Unit)` (account anonymized, tokens revoked)
- `401 Unauthorized` → `Result.failure(AppError.Unauthorized())`

Side effect on success: mobile app must call `tokenStore.clear()` and navigate to the Auth screen.

---

## New Store Interfaces

### `ChangePasswordStore`

```kotlin
interface ChangePasswordStore : Store<Intent, State, Label> {
    data class State(
        val isLoading: Boolean = false,
        val error: String? = null,
    )
    sealed interface Intent {
        data class Submit(val currentPassword: String, val newPassword: String) : Intent
        data object ClearError : Intent
    }
    sealed interface Label {
        data object Saved : Label
    }
}
```

### `ChangeEmailStore`

```kotlin
interface ChangeEmailStore : Store<Intent, State, Label> {
    enum class Phase { REQUEST, CONFIRM }
    data class State(
        val newEmail: String = "",
        val phase: Phase = Phase.REQUEST,
        val isLoading: Boolean = false,
        val error: String? = null,
        val resendCooldown: Int = 0,
    )
    sealed interface Intent {
        data class RequestCode(val newEmail: String) : Intent
        data class ConfirmCode(val otp: String) : Intent
        data object ResendCode : Intent
        data object ClearError : Intent
    }
    sealed interface Label {
        data class EmailChanged(val newEmail: String) : Label
    }
}
```
