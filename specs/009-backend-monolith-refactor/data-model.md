# Data Model: Backend Monolith Refactor

**Feature**: 009-backend-monolith-refactor
**Date**: 2026-05-25

This feature does not introduce new persistent tables. It consolidates service binaries and rewrites the mail component in Go while preserving the existing HTTP contracts.

---

## Persistent Storage (unchanged)

### PostgreSQL migrations (source of truth)

Auth schema migrations (in `infra/migrations/auth/`):
- `001_create_credentials.sql`
- `002_create_refresh_tokens.sql`
- `003_create_otp_codes.sql`

User schema migrations (in `infra/migrations/user/`):
- `001_create_students.sql`
- `002_create_track_enrollments.sql`
- `003_create_achievements.sql`
- `004_create_xp_events.sql`
- `006_add_onboarding_completed.sql`
- `007_add_email_to_students.sql`

Redis is present in local infra but this feature does not add new Redis keys.

---

## Existing API Entity: `AuthResponse` (auth)

Returned by:
- `POST /v1/auth/register` (`201 Created`)
- `POST /v1/auth/login` (`200 OK`)
- `POST /v1/auth/refresh` (`200 OK`)

```json
{
  "student_id": "string",
  "access_token": "string",
  "refresh_token": "string",
  "expires_in": 3600
}
```

Notes:
- `student_id` is a shared identifier used across auth and user services.
- `expires_in` is the access token lifetime in seconds.

---

## Existing API Entity: `StudentResponse` (user)

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
  "onboarding_completed": false
}
```

---

## Existing Error Envelope (all services)

```json
{
  "error": {
    "code": "VALIDATION_ERROR|UNAUTHORIZED|FORBIDDEN|CONFLICT|INTERNAL_ERROR|NOT_FOUND",
    "message": "human readable message",
    "field": "optional_field_name"
  }
}
```

---

## Internal Contract Entity: `CreateStudentRequest`

`POST /internal/students` (internal-token protected)

```json
{
  "student_id": "string",
  "display_name": "string",
  "username": "string",
  "email": "string"
}
```

Notes:
- `username` is optional; if absent/blank, user-svc generates a default.

---

## Internal Contract Entity: `EmailSendRequest`

`POST /internal/email/send` (internal-token protected)

```json
{
  "type": "WELCOME|EMAIL_VERIFY|OTP_LOGIN|EMAIL_CHANGE|PASSWORD_RESET",
  "to": "recipient@example.com",
  "params": {
    "otp": "123456",
    "display_name": "optional",
    "verification_link": "optional"
  }
}
```

Parameter requirements (as validated by current mail behavior):
- `WELCOME`: requires `display_name` and `verification_link`
- `EMAIL_VERIFY`: requires `otp`
- `OTP_LOGIN`: requires `otp`
- `EMAIL_CHANGE`: requires `otp`
- `PASSWORD_RESET`: requires `otp`
