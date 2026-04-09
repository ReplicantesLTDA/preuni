# API Contracts: Onboarding Flow and UX Fixes

## Existing endpoint reused for track re-selection

### PATCH /v1/students/me/onboarding

**Used by**: `ChangeTrackStore` (track re-selection from Profile) and `OnboardingComponent` (initial track selection)  
**Auth**: Bearer JWT (via NGINX gateway)  
**Already registered** in `backend/svc/user/cmd/server/main.go`

**Request body**:
```json
{
  "enrolled_track_ids": ["math", "sciences", "languages"]
}
```

**Success response** — `200 OK`:
```json
{
  "id": "uuid",
  "display_name": "Ana",
  "username": "ana",
  "email": "ana@example.com",
  "xp_total": 0,
  "streak_count": 0,
  "readiness_score": 0.0,
  "onboarding_completed": true
}
```

**Error responses**:
- `401 Unauthorized` — invalid or missing JWT
- `500 Internal Server Error` — database failure

**Notes**: The endpoint is idempotent. Calling it multiple times with different `enrolled_track_ids` is the intended re-selection mechanism. Track enrollment rows are not yet written to a separate table (backend TODO), but the API call succeeds and returns the updated student.

---

## New method added to UserApiClient (mobile only)

### UserApiClient.updateTracks(trackIds: List<String>): Result<Unit>

Wraps `PATCH /v1/students/me/onboarding`. Returns `Result.success(Unit)` on 200, maps non-success status codes to `AppError` via the existing `toAppError()` function.

**No new backend work required.**
