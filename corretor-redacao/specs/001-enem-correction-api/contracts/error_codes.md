# Typed Error Taxonomy (v2.0.0)

All API error responses carry an `error_code` (machine-readable, stable across versions) and a
`message` (pt-BR, may evolve). Codes are the **only** contract; messages are advisory.

## HTTP-layer errors

| `error_code` | HTTP | Surfaced by | Meaning |
|---|---|---|---|
| `unauthenticated` | 401 | every protected endpoint | Missing or invalid access token. |
| `token_expired` | 401 | every protected endpoint | Access token expired; client should refresh. |
| `invalid_credentials` | 401 | `POST /auth/login` | Email/password mismatch. |
| `invalid_token` | 401 | `POST /auth/refresh`, `POST /auth/verify-email` | Refresh or verification token unknown / expired / used. |
| `email_unverified` | 403 | every correction endpoint | Account exists but email is not verified. |
| `parental_consent_required` | 403 | every correction endpoint | User is under 18 and no consent record is on file. |
| `not_found` | 404 | `GET /corrections/{id}`, `POST /corrections/{id}/reevaluations` | Correction not found, or not owned by caller (indistinguishable by design). |
| `email_already_registered` | 409 | `POST /auth/register` | Email already exists. |
| `quota_exhausted` | 429 | `POST /corrections`, `POST /corrections/{id}/reevaluations` | Monthly quota consumed. Response includes `quota_reset_at`. |
| `rate_limited` | 429 | any endpoint | Per-IP / per-user rate limit. |

## Pre-validation errors (synchronous, do **not** enqueue, do **not** consume quota)

| `error_code` | HTTP | Meaning |
|---|---|---|
| `length_too_short` | 400 | Essay below 500 chars / 7 lines (whichever is shorter). |
| `length_too_long` | 400 | Essay above 3500 chars / 50 lines (whichever is larger). |
| `language_mismatch` | 400 | Essay is not pt-BR (pt-PT explicitly rejected). |
| `theme_missing_context` | 400 | Prompt theme has title only, missing contextualization. |
| `theme_missing_title` | 400 | Prompt theme has context only. |
| `payload_too_large` | 413 | Request body exceeds the configured limit (defense in depth). |
| `bad_request` | 400 | Generic shape violation (JSON parse, missing required keys). |

## Asynchronous-failure error codes (surfaced inside `CorrectionFailed`)

These appear in the body of `GET /corrections/{id}` with `status = "failed"`. The HTTP status of
the GET is always 200 (the resource exists; it just failed). `quota_consumed` distinguishes
user-attributable failures from provider/internal failures.

| `error_code` | `quota_consumed` | Meaning |
|---|---|---|
| `provider_rate_limited` | false | LLM provider returned a rate-limit signal. |
| `provider_timeout` | false | LLM call exceeded the worker's timeout budget. |
| `provider_unavailable` | false | LLM provider unreachable or returned 5xx. |
| `schema_violation` | false | LLM output failed JSON Schema validation twice in a row (Constitution IV). |
| `internal_error` | false | Worker / DB internal error. |
| `language_mismatch` | true | Language could not be confirmed pt-BR by the LLM-side check (slipped through pre-validation). |
| `theme_missing_context` | true | Theme rejected by the LLM-side check. |
| `length_too_short` | true | Length rejected post-claim (race-with-edit). |
| `length_too_long` | true | Length rejected post-claim. |

**Rule (FR-036)**: `quota_consumed = false` if and only if the failure cause is
provider/internal. User-attributable failures consume quota.

## Notes for clients

- All error responses are JSON. `Content-Type: application/json`.
- Messages are pt-BR; do not parse them. Use `error_code`.
- New error codes MAY appear in a minor contract bump; clients SHOULD treat unknown codes as
  retryable iff the HTTP status is 5xx or 429.
