# Contract: Home + Profile Backend Integration

**Feature**: 008-fix-backend-integrations
**Direction**: Mobile clients (Android/iOS/Web) -> API Gateway -> user-svc

---

## Authentication Contract

- All Home/Profile data requests MUST include:
  - `Authorization: Bearer <access_token>`
- Missing/invalid token MUST return:
  - `401 Unauthorized` with standard error envelope.
- Client MUST support token refresh and single retry of the original request when session can be silently refreshed.

---

## Endpoint Contract

### `GET /v1/students/me`

**Purpose**: Load profile summary used by both Home and Profile sections.

**Success**: `200 OK`

```json
{
  "id": "uuid",
  "display_name": "Ana Lima",
  "username": "ana01",
  "email": "ana@example.com",
  "avatar_url": null,
  "xp_total": 120,
  "streak_count": 4,
  "readiness_score": 0.63,
  "onboarding_completed": true
}
```

**Error envelope**:

```json
{
  "error": {
    "code": "VALIDATION_ERROR|UNAUTHORIZED|FORBIDDEN|CONFLICT|INTERNAL_ERROR|NOT_FOUND",
    "message": "human readable message",
    "field": "optional_field_name"
  }
}
```

### `PATCH /v1/students/me`

**Purpose**: Update profile fields and return canonical profile object.

**Request**:

```json
{
  "display_name": "Ana Lima",
  "username": "ana01"
}
```

**Success**: `200 OK` with full `StudentResponse` object (same schema as `GET /v1/students/me`).

**Validation failure**: `422` with `error.field` populated (for example `username`).

---

## Reliability Contract (Client Behavior)

- Timeout for Home/Profile fetch paths: **10 seconds** maximum.
- Retry policy:
  - Retry transient failures (network + 5xx) with exponential backoff.
  - Do **not** retry 4xx validation/auth errors.
  - Manual retry must trigger a fresh backend request.

---

## Observability Contract

For each Home/Profile request attempt, client logs/metrics MUST include:
- route/method
- status class or transport failure
- duration (ms)
- attempt index

Client logs MUST redact:
- access tokens
- refresh tokens
- emails and other direct PII fields in payload/body logs

---

## E2E Verification Matrix

| Scenario | Expected Contract Outcome |
|----------|----------------------------|
| Authenticated Home load | `GET /v1/students/me` returns `200` with complete schema and renders in <= 3s |
| Authenticated Profile load | Same endpoint contract and timing behavior as Home |
| Network timeout (>10s) | Client fails within 10s per-request timeout, surfaces retry action |
| Transient failure (NetworkError) | Auto-retries up to 3 attempts with exponential backoff (500ms → 1s → 2s), then shows error |
| Transient failure then recovery | On success within MAX_ATTEMPTS, data renders without manual retry |
| Persistent transient failure | After 3 failed attempts, error state surfaces with manual retry affordance |
| Non-transient 4xx (401, 403, 422) | No automatic retry; error state surfaces immediately |
| Malformed payload (blank required field) | Client validation rejects response; integration-safe error state shown (no crash) |
| Missing token | `401` response path exercised and user session recovery flow triggered |
| Manual retry | Fresh `GET /v1/students/me` request fired; success clears error state |

## Implemented Retry Policy

- **Transient errors** (retry): `AppError.NetworkError` (connectivity loss, HTTP timeout)
- **Non-transient errors** (no retry): `AppError.Unauthorized`, `AppError.Forbidden`,
  `AppError.Validation`, `AppError.Conflict`, `AppError.Unknown` (malformed payload)
- **Max attempts**: 3 (initial + 2 retries)
- **Backoff**: 500ms → 1000ms → 2000ms (exponential, capped at 5s)

## Telemetry Contract (Client)

Each Home/Profile request emits a `RequestTelemetry` record containing:
- `route`: URL path (e.g. `/v1/students/me`)
- `method`: HTTP method
- `statusCode`: HTTP response code or null on transport failure
- `attempt`: 0-indexed attempt number
- `durationMs`: elapsed time for the attempt
- `outcome`: `SUCCESS | TRANSIENT_FAILURE | NON_TRANSIENT_FAILURE | TIMEOUT`

**Redaction guarantees**: Authorization header values and response body content are
never present in telemetry records.
