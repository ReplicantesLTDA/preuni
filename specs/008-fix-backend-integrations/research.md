# Research: Fix Backend Integration Issues (Home + Profile)

**Feature**: 008-fix-backend-integrations
**Date**: 2026-04-15
**Status**: Complete

---

## Finding 1: Platform bootstrap drift is a primary regression source

**Decision**: Move authenticated HTTP bootstrap to one shared factory consumed by Android, iOS, and Web entrypoints.

**Rationale**:
- Android currently builds the HTTP client with `getAccessToken` and installs token refresh interception.
- iOS and Web currently build the HTTP client without token injection/refresh wiring.
- Home/Profile depend on authenticated `GET /v1/students/me`; missing bearer propagation causes integration failures.

**Alternatives considered**:
- Keep platform-local wiring and patch iOS/Web only now: rejected; regressions likely reappear with future edits.
- Move platform-specific wrappers only: rejected; still duplicates auth wiring logic.

---

## Finding 2: Existing backend endpoint contract is sufficient for Home/Profile

**Decision**: Keep `GET /v1/students/me` and `PATCH /v1/students/me` as the integration contract for both sections.

**Rationale**:
- user-svc already exposes these endpoints with the payload fields required by Home/Profile (`display_name`, `username`, `email`, `xp_total`, `streak_count`, `readiness_score`, etc.).
- No new endpoint is required to satisfy current requirements.

**Alternatives considered**:
- Introduce a dedicated `/v1/home/dashboard` endpoint: rejected for this feature; unnecessary backend scope expansion.

---

## Finding 3: Timeout and retry behavior must be narrowed to transient failures

**Decision**: Use a 10-second request timeout for Home/Profile fetches and exponential backoff retries only for transient failures (network and 5xx).

**Rationale**:
- Specification requires timeout at 10 seconds and resilient retry behavior.
- Retrying non-transient 4xx failures wastes time and degrades UX.

**Alternatives considered**:
- Global 30-second timeout with fixed retry delay: rejected; violates acceptance behavior and slows failure feedback.
- No automatic retries, manual retry only: rejected; weaker resilience in unstable networks.

---

## Finding 4: Schema safety needs explicit validation at integration boundary

**Decision**: Validate required response fields before rendering Home/Profile models.

**Rationale**:
- Typed DTO deserialization catches many errors, but explicit boundary validation improves clarity for malformed/partial payloads.
- Requirement calls out schema validation to prevent UI corruption or crashes.

**Alternatives considered**:
- Trust DTO parsing only: rejected; less explicit and weaker diagnostics for partial contract drift.

---

## Finding 5: Observability must be privacy-safe

**Decision**: Add network telemetry for request/response timing and outcome, with PII/token redaction.

**Rationale**:
- Requirement includes request/response logging and timing metrics.
- Current generic logging does not provide structured metrics and risks exposing sensitive details.

**Alternatives considered**:
- Keep default Ktor INFO logging only: rejected; insufficient timing/attempt insight and redaction guarantees.
- Log full payloads for debugging: rejected; conflicts with PII redaction requirement.

---

## Finding 6: E2E coverage should span both client and service contract boundaries

**Decision**: Add two complementary test layers:
1. KMP integration tests (`MockEngine`) for client-side transport/contract behavior.
2. user-svc handler integration tests for endpoint payload and validation envelope behavior.

**Rationale**:
- Mobile tests catch auth header propagation, timeout/retry logic, and deserialization behavior.
- Backend integration tests catch server-side contract drift and validation regressions.
- Together they provide reliable pre-merge detection for Home/Profile integration issues.

**Alternatives considered**:
- UI-only compose tests: rejected; poor signal on network contract failures.
- Backend-only tests: rejected; would miss client wiring regressions across platforms.
