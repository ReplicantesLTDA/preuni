# Contract: Backend Monolith API

**Feature**: 009-backend-monolith-refactor
**Direction**: Clients -> API Gateway (NGINX) -> Monolith

---

## Global HTTP Contract

### Request ID

- Gateway forwards `X-Request-ID`.
- Backend services and the monolith MUST preserve request-id propagation so logs can be correlated end-to-end.

### Error envelope

All non-2xx responses use the shared JSON envelope:

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

## Authentication Contract

### JWT-protected endpoints

- Protected endpoints MUST require:
  - `Authorization: Bearer <access_token>`
- Missing/invalid token MUST return:
  - `401 Unauthorized` with the standard error envelope.

### Internal service endpoints

- Internal endpoints MUST require:
  - `Authorization: Bearer <internal_service_token>`
- Missing/invalid internal token MUST return:
  - `401 Unauthorized` with `error.code = "UNAUTHORIZED"`.

---

## Public API Endpoints (Gateway-facing)

### `POST /v1/auth/register`

**Request**:

```json
{
  "email": "ana@example.com",
  "password": "P@ssw0rd",
  "display_name": "Ana"
}
```

**Success**: `201 Created`

```json
{
  "student_id": "uuid",
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 3600
}
```

### `POST /v1/auth/login`

**Request**:

```json
{
  "email": "ana@example.com",
  "password": "P@ssw0rd"
}
```

**Success**: `200 OK` with `AuthResponse`.

### `POST /v1/auth/refresh`

**Request**:

```json
{
  "refresh_token": "..."
}
```

**Success**: `200 OK` with rotated `AuthResponse`.

Notes:
- A refresh handler exists in the current Go auth service code, but it is not wired into the router today; the monolith cutover MUST mount this endpoint to satisfy gateway expectations.

### `POST /v1/auth/email/verify`

**Request**:

```json
{
  "email": "ana@example.com",
  "otp": "123456"
}
```

**Success**: `204 No Content`

### `POST /v1/auth/otp/request`

**Request**:

```json
{ "email": "ana@example.com" }
```

**Success**: `202 Accepted` (always, to prevent email enumeration)

### `POST /v1/auth/otp/verify`

**Request**:

```json
{
  "email": "ana@example.com",
  "otp": "123456"
}
```

**Success**: `200 OK` with `AuthResponse`

### `POST /v1/auth/password/reset/request`

**Request**:

```json
{ "email": "ana@example.com" }
```

**Success**: `202 Accepted` (always, to prevent email enumeration)

### `POST /v1/auth/password/reset/confirm`

**Request**:

```json
{
  "email": "ana@example.com",
  "otp": "123456",
  "new_password": "NewP@ssw0rd"
}
```

**Success**: `204 No Content`

### `POST /v1/auth/logout` (JWT required)

**Request**:

```json
{ "refresh_token": "..." }
```

**Success**: `204 No Content`

### `POST /v1/auth/password/change` (JWT required)

**Request**:

```json
{
  "current_password": "P@ssw0rd",
  "new_password": "NewP@ssw0rd"
}
```

**Success**: `204 No Content`

### `POST /v1/auth/email/change/request` (JWT required)

**Request**:

```json
{ "new_email": "ana+new@example.com" }
```

**Success**: `202 Accepted`

### `POST /v1/auth/email/change/confirm` (JWT required)

**Request**:

```json
{
  "new_email": "ana+new@example.com",
  "otp": "123456"
}
```

**Success**: `204 No Content`

### `DELETE /v1/auth/account` (JWT required)

**Request (optional)**:

```json
{ "confirmation": "DELETE" }
```

**Success**: `204 No Content`

Notes:
- Current gateway routing may not forward `/v1/auth/account` unless the cutover routes all `/v1/auth/*` traffic to the monolith (or explicitly adds this path).

### `GET /v1/students/me` (JWT required)

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

### `PATCH /v1/students/me` (JWT required)

**Request**:

```json
{
  "display_name": "Ana Lima",
  "username": "ana01"
}
```

**Success**: `200 OK` with full `StudentResponse`.

---

## Internal Endpoints (service-to-service)

### `POST /internal/students`

**Purpose**: Provision a student profile after registration.

**Request**:

```json
{
  "student_id": "uuid",
  "display_name": "Ana",
  "username": "optional",
  "email": "ana@example.com"
}
```

**Success**: `201 Created`

```json
{ "id": "uuid" }
```

### `POST /internal/email/send`

**Request**:

```json
{
  "type": "WELCOME",
  "to": "ana@example.com",
  "params": {
    "display_name": "Ana",
    "verification_link": "https://..."
  }
}
```

**Success**: `202 Accepted`

```json
{ "status": "queued" }
```

**Validation failure**: `422 Unprocessable Entity`

```json
{ "error": "WELCOME requires display_name and verification_link params" }
```

**Delivery failure**: `500 Internal Server Error`

```json
{ "error": "delivery failed", "reason": "..." }
```

Supported types and required params:
- `WELCOME`: requires `display_name`, `verification_link`
- `EMAIL_VERIFY`: requires `otp`
- `OTP_LOGIN`: requires `otp`
- `EMAIL_CHANGE`: requires `otp`
- `PASSWORD_RESET`: requires `otp`
