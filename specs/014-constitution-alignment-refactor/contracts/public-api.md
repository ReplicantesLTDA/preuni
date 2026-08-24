# Public API Contract (Go monolith, client-facing)

Wire format: snake_case (per constitution). Auth: existing JWT bearer
(`internal/auth`), unchanged by this feature. Base path `/v1` behind NGINX,
per existing gateway convention.

## Essay (`/v1/essays`)

| Method | Path | Purpose | Auth |
|---|---|---|---|
| POST | `/v1/essays` | Submit an essay against a prompt theme. Returns 202 immediately (FR-002). Enforces quota (FR-001) — 429 `quota_exhausted` if free-tier already submitted today. | required |
| GET | `/v1/essays/{id}` | Poll submission status/result. `status: pending\|graded\|failed`. | required, owner-only |
| GET | `/v1/essays` | List caller's own submissions, reverse chronological. | required |

### `POST /v1/essays` request

```json
{ "prompt_theme_title": "string", "prompt_theme_context": "string", "essay_text": "string" }
```

### `POST /v1/essays` response — 202

```json
{ "id": "uuid", "status": "pending", "submitted_at": "2026-08-22T12:00:00Z" }
```

### `GET /v1/essays/{id}` response — graded

```json
{
  "id": "uuid",
  "status": "graded",
  "overall_score": 880,
  "competencies": [
    { "competency": 1, "score": 160, "justification_pt_br": "string", "excerpt": "string" }
  ],
  "graded_at": "2026-08-22T12:03:00Z"
}
```

### Error codes (new, additive to existing taxonomy)

| `error_code` | HTTP | Meaning |
|---|---|---|
| `quota_exhausted` | 429 | Free-tier daily submission already used; response includes `quota_reset_at` (next UTC midnight). |
| `grading_failed` | 200 (in body, `status: failed`) | Correction pipeline failed; `error_code`/`error_message_pt_br` mirror the correction service's typed taxonomy (see data-model.md). |

## Streak (`/v1/streaks`)

| Method | Path | Purpose |
|---|---|---|
| GET | `/v1/streaks/me` | Caller's current streak, longest streak, last submission day. |

## Social (`/v1/friends`)

| Method | Path | Purpose |
|---|---|---|
| POST | `/v1/friends/requests` | Send a friend request (`{ "addressee_id": "uuid" }`). |
| POST | `/v1/friends/requests/{id}/accept` | Accept a pending request. |
| DELETE | `/v1/friends/{id}` | Remove an accepted friendship. |
| GET | `/v1/friends` | List accepted friends with their current streak + latest grade. |

Visibility: only friends' `current_streak` and latest `essay_grades` are
exposed; non-friends return 404, not a redacted 200 (avoids confirming a
user exists — consistent with the existing `not_found` pattern in the
correction service's error taxonomy).

## Ranking (`/v1/ranking`)

| Method | Path | Purpose |
|---|---|---|
| GET | `/v1/ranking/weekly` | Current week's global leaderboard, caller's tier by default; supports a `tier` query param. |
| GET | `/v1/ranking/me` | Caller's current rank, tier, and weekly score. |

## Medals (`/v1/medals`)

| Method | Path | Purpose |
|---|---|---|
| GET | `/v1/medals/me` | Caller's earned medals with `type` + `earned_at`. |
