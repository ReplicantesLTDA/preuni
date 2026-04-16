# Data Model: Backend Integration Fixes (Home + Profile)

**Feature**: 008-fix-backend-integrations
**Date**: 2026-04-15

No new persistent tables or storage keys are introduced. This feature hardens runtime integration models and state transitions.

---

## Existing API Entity: `StudentResponse` (user-svc)

Returned by:
- `GET /v1/students/me`
- `PATCH /v1/students/me`

```json
{
  "id": "string",
  "display_name": "string",
  "username": "string",
  "email": "string",
  "avatar_url": "string|null",
  "xp_total": 0,
  "streak_count": 0,
  "readiness_score": 0.0,
  "onboarding_completed": true
}
```

### Validation rules at client boundary

- `id` must be non-blank.
- `display_name` must be non-blank.
- `username` must be non-blank and match app/backend username constraints.
- `email` must be non-blank.
- `xp_total` must be `>= 0`.
- `streak_count` must be `>= 0`.
- `readiness_score` must be between `0.0` and `1.0`.

Invalid payloads are treated as integration errors and never rendered directly.

---

## Existing Domain Entity: `Student`

```kotlin
Student(
  id: String,
  displayName: String,
  username: String,
  email: String,
  avatarUrl: String?,
  xpTotal: Long,
  streakCount: Int,
  readinessScore: Double,
  onboardingCompleted: Boolean
)
```

Used by both Home and Profile stores as the source-of-truth view model.

---

## New Runtime Entity: `RequestTelemetry`

Tracks per-attempt request metrics (non-persistent):

```text
RequestTelemetry {
  route: String
  method: String
  statusCode: Int?        // null on transport failure
  attempt: Int
  durationMs: Long
  outcome: enum { SUCCESS, TRANSIENT_FAILURE, NON_TRANSIENT_FAILURE, TIMEOUT }
}
```

PII and credentials are excluded from telemetry payload.

---

## New Runtime Entity: `RetryDecision`

```text
RetryDecision {
  shouldRetry: Boolean
  nextDelayMs: Long       // exponential backoff
  reason: String
}
```

Rules:
- Retry only network failures and 5xx responses.
- Do not retry 4xx validation/auth errors.
- Stop after configured max attempts.

---

## Store State Changes

### `HomeStore.State` (existing; behavior extended)

```kotlin
State(
  student: Student? = null,
  isLoading: Boolean = false,
  error: AppError? = null
)
```

Transitions:
- `Load` -> fetch with timeout/retry -> success sets `student`, failure sets `error`
- `Retry` -> re-runs full fetch flow with fresh request

### `ProfileStore.State` (existing; behavior extended)

```kotlin
State(
  student: Student? = null,
  isLoading: Boolean = false,
  error: AppError? = null,
  showDeleteConfirmation: Boolean = false
)
```

Transitions:
- `LoadProfile` -> same hardened fetch flow as Home
- profile update intents -> validated backend response refreshes `student`
- failures always keep previous valid `student` snapshot and expose actionable `error`
