# Contract: API Client

**Module**: `mobile/src/lib/api/`

## Shape

```ts
// mobile/src/lib/api/client.ts
export interface ApiClient {
  request<TOut>(input: {
    method: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';
    path: string;                   // "/v1/auth/login"
    body?: unknown;                 // serialized via JSON.stringify
    schema: z.ZodType<TOut>;        // 2xx body validated against this
    auth?: boolean;                 // default true; attaches Bearer access token
    signal?: AbortSignal;
  }): Promise<TOut>;
}
```

## Behavior

1. **Base URL** from `EXPO_PUBLIC_API_BASE_URL` (defaults to `http://localhost:8080`). Wired into Expo Router root layout.
2. **Auth header**: `Authorization: Bearer <accessToken>` when `auth !== false` and `tokenStore.getAccessToken()` returns a value.
3. **Refresh interceptor**: on `401` with a refresh token present:
   - Lock so concurrent 401s don't fan out.
   - `POST /v1/auth/refresh` with `{ refreshToken }`.
   - On success: persist new tokens, retry the original request **once**.
   - On failure: clear tokens, set `useSessionStore.setState({ status: 'anon' })`, throw `AppError.Unauthorized`.
4. **Error envelope**: non-2xx body is parsed against `ApiErrorEnvelope` (Zod). Unparseable bodies fall back to `AppError.Unknown('Algo deu errado. Tente novamente.')`.
5. **Mapping**: status → `AppError` variant:
   - `401` → `AppError.Unauthorized`
   - `403` → `AppError.Forbidden(message)`
   - `409` → `AppError.Conflict(message)`
   - `422` → `AppError.Validation(field, message)`
   - other → `AppError.Unknown(message)`
6. **Logging**: every request emits `{ method, path, status, durationMs }` to the dev console; redacted in production.
7. **Avatar upload** uses `fetch` with `multipart/form-data` (Zod schema for response only); not routed through the JSON path.

## Endpoint surface (must compile against these)

`POST /v1/auth/register`, `POST /v1/auth/login`, `POST /v1/auth/refresh`, `POST /v1/auth/logout`, `POST /v1/auth/email/verify`, `POST /v1/auth/email/verify-resend`, `POST /v1/auth/otp/{request,verify}`, `POST /v1/auth/password/reset/{request,confirm}`, `GET /v1/students/me`, `PATCH /v1/students/me`, `DELETE /v1/students/me`, `PUT /v1/students/me/avatar`, `POST /v1/students/me/avatar/confirm`, `PATCH /v1/students/me/onboarding`, `GET /v1/students/me/data-export`.

## TanStack Query integration

- Query keys are namespaced by feature: `['student','me']`, `['trilha','home', trackId]`, etc.
- Mutations invalidate the obvious neighbors (e.g. completing an activity → `invalidateQueries({ queryKey: ['student','me'] })` + `['trilha','home']`) — satisfies FR-010 without a manual refresh.
- Default options: `retry: 1` on GETs, `retry: 0` on mutations, `staleTime: 30_000`.

## Acceptance

- [ ] A failing test exists for each `AppError` mapping branch (401, 403, 409, 422, generic).
- [ ] A `msw` test demonstrates: 401 → refresh succeeds → original retry returns 200.
- [ ] A `msw` test demonstrates: 401 → refresh fails → session goes to `anon` and user routed to `/(auth)/login`.
- [ ] All schemas are Zod and exported from `mobile/src/types/`.
