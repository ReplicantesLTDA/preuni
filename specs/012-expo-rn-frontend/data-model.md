# Phase 1 — Data Model

**Scope**: The frontend introduces no persistent entities. It re-presents entities owned by the backend monolith and keeps a small slice of local-only state.

## Server-owned entities (read by client)

| Entity | Source endpoint | Client representation |
|--------|-----------------|------------------------|
| `Session` | `POST /v1/auth/{register,login,refresh,otp/verify}` | `{ accessToken, refreshToken, accessTokenExpiresAt }` — held in `tokenStore` + Zustand session store |
| `Student` | `GET /v1/students/me` | `{ id, email, displayName, username, avatarUrl, xpTotal, streakCount, readinessScore, onboardingCompleted }` |
| `Credentials state` | derived from session + `Student.emailVerified` flow | `{ isAuthed, emailVerified, onboardingCompleted }` — drives route gating |
| `Track / Subject path` | content endpoints | parity port of KMP `Track` shape; refined when feature 006 endpoints are wired |
| `RedaçãoPrompt`, `RedaçãoSubmission` | redacao endpoints | as exposed by backend |
| `Simulado`, `SimuladoAttempt` | simulation endpoints | as exposed by backend |

All shapes are declared as **Zod schemas** in `mobile/src/types/` and inferred to TS types. Schemas are the contract — any 200 response that fails to parse fails the test and surfaces a typed error envelope to the caller.

### Zod schema sketch (`mobile/src/types/auth.ts`)

```ts
export const SessionSchema = z.object({
  accessToken: z.string(),
  refreshToken: z.string(),
  accessTokenExpiresAt: z.string().datetime(),
});

export const StudentSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  displayName: z.string(),
  username: z.string().nullable(),
  avatarUrl: z.string().url().nullable(),
  xpTotal: z.number().int().nonnegative(),
  streakCount: z.number().int().nonnegative(),
  readinessScore: z.number().int().min(0).max(100),
  onboardingCompleted: z.boolean(),
});

export const ApiErrorEnvelope = z.object({
  error: z.object({
    code: z.string().optional(),
    field: z.string().optional(),
    message: z.string(),
  }),
});
```

## Client-only state

| Key | Storage | Purpose |
|-----|---------|---------|
| `access_token`, `refresh_token`, `access_token_expires_at` | `expo-secure-store` | session restore + refresh interceptor |
| `welcome_seen` | `expo-secure-store` | gate first-launch welcome screen |
| `active_track_id` | `expo-secure-store` | last-selected subject track (parity with feature 006) |
| Session machine `{ status, student }` | Zustand (`useSessionStore`) | drives the route-gate layout |
| UI flags (theme override, debug menu open, etc.) | Zustand (`useUiStore`, optional) | ephemeral |

## State transitions — auth/session machine

```
            ┌────────────┐
            │  loading   │  ← root mount; reads SecureStore
            └─────┬──────┘
   no tokens │     │ tokens valid
             ▼     ▼
         ┌──────┐  ┌──────┐  401 on any request
         │ anon │  │authed│ ───────────────┐
         └─┬────┘  └──┬───┘                │
   register/login    │                     │
           │         ▼                     │
           │   refresh succeeds            │
           │         │                     │
           │   refresh fails               │
           │         ▼                     │
           └──────► clear tokens ◄─────────┘
```

- `loading → authed`: SecureStore returns a non-expired access token.
- `loading → anon`: no tokens or refresh fails on cold start.
- `authed → anon`: refresh fails OR user signs out OR account is deleted.
- `anon → authed`: successful register, login, OTP-login, or post-verify auto-login (FR-007).

## Validation rules (client-side, before hitting the API)

| Field | Rule (Zod) |
|-------|------------|
| `email` | `z.string().email()` |
| `password` (register) | `z.string().min(8).max(128)` + at least one letter + one digit |
| `otp` | `z.string().regex(/^\d{6}$/)` |
| `displayName` | `z.string().min(2).max(64)` |
| `username` | `z.string().regex(/^[a-z0-9_]{3,32}$/)` (when surfaced) |
| Avatar file | image MIME, ≤ 5 MB |

Server-side validation remains the source of truth; client validation is a UX layer that prevents the obvious bad submit.

## Relationships

- A `Session` belongs to exactly one `Student` (by `student.id` claim in JWT).
- A `Student` has many `Tracks`, `RedaçãoSubmissions`, `SimuladoAttempts` (already modeled server-side; the client does not re-model these joins, it just renders lists).

This document is intentionally short — re-modeling backend entities on the client would duplicate truth. Schemas + endpoints in [`contracts/api-client.md`](./contracts/api-client.md) are the binding artifact.
