# Security review — 012-expo-rn-frontend

**Date**: 2026-05-27
**Reviewer**: implementation pass against Constitution + threat-model worksheet.

## Scope

Frontend-only review. Backend monolith is unchanged this feature; backend security review is owned by spec 009/010.

## Findings

### Token storage

- Tokens kept in `expo-secure-store` on iOS (Keychain) and Android (EncryptedSharedPreferences). Source: `mobile/src/lib/auth/tokenStore.ts`.
- Web fallback writes to `localStorage`. Acceptable for v1: web is dev-review surface only. Production web hosting is a follow-up effort (see plan.md). Tracked: document this on the web build before any external rollout.
- `tokenStore.clear()` removes all three keys (`access_token`, `refresh_token`, `access_token_expires_at`) atomically via `Promise.all`. Logout, refresh failure, and account delete all reach this path.

### Session machine

- Three explicit states: `loading | authed | anon`. Bootstrap (`src/lib/auth/bootstrap.ts`) only marks `authed` after `GET /v1/students/me` succeeds.
- Refresh interceptor is single-flight (`src/lib/api/client.ts`). On 401 + refresh-success it retries once; on refresh-fail it clears tokens, sets session anon, routes to `(auth)/welcome`. Verified by `tests/features/auth/expiredRefresh.integration.test.ts` and `tests/lib/api/client.test.ts`.
- No infinite refresh loop: the `retried` boolean inside `execute` prevents recursion past one attempt.

### Network

- All requests targeted at `EXPO_PUBLIC_API_BASE_URL`. No third-party hosts.
- Avatar upload uses a presigned URL returned by the backend. Client only PUTs to it if the URL does **not** match the dev-stub marker (`PRESIGNED`). In production the URL is real-presigned and the PUT proceeds.
- Avatar upload sets `Content-Type` to the picker-reported MIME (`image/jpeg` default). Backend enforces final validation via S3 + post-confirm path.

### Input validation

- Every form is validated client-side with Zod schemas in `src/features/<domain>/validation.ts`:
  - Email: `z.string().email()`
  - Password (register, change, reset confirm): min 8, max 128, ≥1 letter, ≥1 digit
  - OTP: `/^\d{6}$/`
  - Display name: 2–64 chars
- Server-side validation in the Go backend remains the source of truth. Client validation is UX-only.

### Error envelope

- Backend `{ error: { code, field, message } }` parsed via `parseApiErrorEnvelope` (Zod). Mapped to typed `AppError` union. UI surfaces only the human-readable `message`; raw JSON never reaches the user (FR-019).

### Account deletion

- `(tabs)/perfil/delete-account.tsx` requires typing the literal word `EXCLUIR` before enabling the destructive button.
- Backend expects `{ confirmation: "DELETE" }` (handler defaults the literal string client-side after the user confirms).
- On success, tokens cleared + session anon + React Query cache cleared. User cannot accidentally land on a stale Trilha after delete.

### Sensitive data in logs

- Dev-only API logger (`src/lib/api/context.tsx`) prints `method`, `path`, `status`, `durationMs`. **No body/header content logged.** Production builds drop the logger entirely.

### Third-party packages

- New non-Expo deps: `@tanstack/react-query`, `zustand`, `zod`. All Top-100 npm packages, no known CVEs at install time.
- `@expo-google-fonts/*` packages bundle Google Fonts TTFs. License: OFL.
- No native modules outside the Expo SDK 54 baseline.

### What is **not** covered

- Penetration testing of the React Native bridge.
- Static analysis (SAST) on the bundle. Tracked under polish phase (T112 partial — bundle audit doc, see `bundle-report.md`).
- Web build hardening (CSP, SRI, hosting). Production web requires a separate pass.
- Push notification permissions — out of scope for v1.

## Acceptance

- [x] No raw token/header/body content in any console.log
- [x] Refresh single-flight, retry-once, clear-on-fail wired
- [x] Account delete typed confirmation + cache clear
- [x] All form inputs Zod-validated client-side
- [x] Error envelope parsed + only `message` shown
- [x] SecureStore on native, documented localStorage fallback on web
