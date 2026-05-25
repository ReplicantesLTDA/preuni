# Security Review — Monolith Refactor

**Date**: 2026-05-25
**Scope**: `backend/svc/monolith/`, refactored handlers in `backend/svc/auth/`, `backend/svc/user/`

## Findings

### Secrets handling

- `INTERNAL_SERVICE_TOKEN` read from env at startup (`backend/svc/monolith/internal/config/config.go:36`). Never logged. Compared via `==` in `pkg/middleware/auth.go:69` — exact match, no leakage on failure path (returns generic "invalid internal token").
- `JWT_SIGNING_KEY` read from env (config.go:39). Never logged. Used only inside `domain.JWTService` and `pkgmw.RequireAuth([]byte(key))`.
- `SMTP_PASS` read from env (config.go:55). Passed to `smtp.PlainAuth` only. Never serialized into responses or logs. SMTP error messages wrap network errors with no credential context.

### Authn / authz

- All `/v1/auth/*` mutating endpoints retained existing protection model. `/logout`, `/password/change`, `/email/change/*`, `/account` (DELETE) still behind `RequireAuth`.
- All `/v1/students/*` routes behind `RequireAuth`.
- Internal endpoints (`/internal/students`, `/internal/email/send`) behind `InternalAuth(token)`. Live-validated: missing token → 401.
- Refresh route now reachable (T006) — JWT rotation logic in `RefreshTokenHandler` unchanged from pre-refactor code.

### Email enumeration

- Preserved: `/v1/auth/otp/request` and `/v1/auth/password/reset/request` always return 202 regardless of account existence (otp_login.go:35, password_reset.go:35). Sender invocation moved inside a goroutine that silently returns on lookup miss.

### Input validation

- Mail handler validator (`backend/svc/monolith/internal/mail/validator.go`) returns 422 with stable message strings matching the Elixir contract. No params reflected back into error bodies → no XSS surface.
- JSON decoder errors return generic 422 (no stack trace leakage).

### Error responses

- Standard envelope `{"error":{"code","message","field?"}}` via `pkg/middleware`. No internal error string surfaces in 5xx paths (`apperrors.Internal` wraps via `apperrors.As`).

### Resource exhaustion

- SMTP delivery is fire-and-forget via `go func()` inside the handler. Bounded by goroutine cost only; no explicit queue depth limit. **Risk**: an attacker that obtains the internal token could spam `/internal/email/send` and exhaust SMTP relay quota / file descriptors. Mitigation: gateway-side rate limiting + the internal-token requirement (token rotation is the recovery path).
- HTTP server uses `ReadHeaderTimeout: 10s` (`cmd/server/main.go:34`). No body-read timeout configured — Slowloris-style POST bodies could tie up handler goroutines. **Recommend**: add `srv.ReadTimeout` + `srv.WriteTimeout` in a follow-up.

### TLS

- SMTP path 465 uses `tls.Dial` with `MinVersion: TLS 1.2` and `ServerName` set (`sender.go:36`). Certificate validation is the default (verify chain + hostname).
- 587 path goes through `smtp.SendMail` which performs STARTTLS via the default config — verifies certs unless server lacks STARTTLS.

### Verdict

No new vulnerabilities introduced by the refactor. Pre-existing gaps logged above; none are blockers for the cutover.
