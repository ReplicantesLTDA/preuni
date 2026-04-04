# External REST API Contract

**Audience**: KMP client (Android, iOS, Web)
**Gateway base URL**: `https://api.preuni.com.br/v1`
**Auth**: Bearer JWT in `Authorization` header for all authenticated endpoints
**Content-Type**: `application/json` for all request/response bodies
**Error envelope**:
```json
{ "error": { "code": "ERROR_CODE", "message": "Human-readable message" } }
```

---

## Authentication (`/auth`)

### POST `/auth/register`
Register a new student account.

**Request**
```json
{
  "email": "student@example.com",
  "password": "minLength8",
  "display_name": "Ana Beatriz"
}
```
**Response 201**
```json
{
  "student_id": "uuid",
  "access_token": "jwt",
  "refresh_token": "opaque-token",
  "expires_in": 3600
}
```
**Errors**: `EMAIL_ALREADY_EXISTS` 409, `VALIDATION_ERROR` 422

---

### POST `/auth/login`
Authenticate with email + password.

**Request**
```json
{ "email": "student@example.com", "password": "..." }
```
**Response 200**
```json
{
  "student_id": "uuid",
  "access_token": "jwt",
  "refresh_token": "opaque-token",
  "expires_in": 3600
}
```
**Errors**: `INVALID_CREDENTIALS` 401, `EMAIL_NOT_VERIFIED` 403

---

### POST `/auth/refresh`
Exchange a refresh token for a new access token.

**Request**
```json
{ "refresh_token": "opaque-token" }
```
**Response 200** — same shape as login 200
**Errors**: `INVALID_TOKEN` 401, `TOKEN_EXPIRED` 401

---

### POST `/auth/logout`
Revoke the current refresh token. Auth required.

**Request**: empty body
**Response 204**

---

### POST `/auth/password/reset-request`
Trigger a password-reset email. No auth required.

**Request**
```json
{ "email": "student@example.com" }
```
**Response 202** — always succeeds (no enumeration)

---

### POST `/auth/password/reset`
Complete password reset using OTP from email.

**Request**
```json
{ "email": "student@example.com", "otp": "123456", "new_password": "..." }
```
**Response 204**
**Errors**: `INVALID_OTP` 400, `OTP_EXPIRED` 400

---

### POST `/auth/email/verify`
Verify email with OTP sent at registration.

**Request**
```json
{ "email": "student@example.com", "otp": "123456" }
```
**Response 204**
**Errors**: `INVALID_OTP` 400, `OTP_EXPIRED` 400

---

## Student Profile (`/students/me`)

### GET `/students/me`
Get the current student's profile. Auth required.

**Response 200**
```json
{
  "id": "uuid",
  "display_name": "Ana Beatriz",
  "avatar_url": "https://...",
  "xp_total": 2400,
  "streak_count": 14,
  "readiness_score": 62,
  "enrolled_tracks": ["uuid-math", "uuid-natural-sciences"],
  "created_at": "ISO-8601"
}
```

---

### PATCH `/students/me`
Update profile. Auth required.

**Request** (all fields optional)
```json
{ "display_name": "Ana", "avatar_url": "https://..." }
```
**Response 200** — updated student object (same as GET)

---

### GET `/students/me/achievements`
List all achievements (locked and unlocked). Auth required.

**Response 200**
```json
{
  "achievements": [
    {
      "id": "uuid",
      "code": "STREAK_7",
      "name": "Uma semana seguida",
      "description": "...",
      "icon_url": "https://...",
      "unlocked": true,
      "unlocked_at": "ISO-8601"
    }
  ]
}
```

---

## Subject Tracks (`/tracks`)

### GET `/tracks`
List all active subject tracks. No auth required.

**Response 200**
```json
{
  "tracks": [
    {
      "id": "uuid",
      "slug": "mathematics",
      "name": "Matemática",
      "description": "...",
      "icon_url": "https://...",
      "color_token": "track-math-500",
      "enem_area": "MATHEMATICS",
      "lesson_count": 98
    }
  ]
}
```

---

### POST `/tracks/{trackId}/enroll`
Enroll in a track. Auth required.

**Response 200**
```json
{ "enrolled_at": "ISO-8601" }
```
**Errors**: `ALREADY_ENROLLED` 409, `TRACK_NOT_FOUND` 404

---

### DELETE `/tracks/{trackId}/enroll`
Unenroll from a track. Auth required.

**Response 204**

---

### GET `/tracks/{trackId}/lessons`
List lessons for a track with the student's progress. Auth required.

**Query params**: `?status=NOT_STARTED|IN_PROGRESS|COMPLETED`

**Response 200**
```json
{
  "track_id": "uuid",
  "lessons": [
    {
      "id": "uuid",
      "title": "Funções de 1º grau",
      "order_index": 1,
      "status": "COMPLETED",
      "completed_at": "ISO-8601"
    }
  ]
}
```

---

## Lessons (`/lessons`)

### GET `/lessons/{lessonId}`
Get lesson detail with exercises. Auth required.

**Response 200**
```json
{
  "id": "uuid",
  "title": "Funções de 1º grau",
  "concept_id": "uuid",
  "exercises": [
    {
      "id": "uuid",
      "type": "MULTIPLE_CHOICE",
      "prompt": "Qual é o valor de f(2) para f(x) = 3x + 1?",
      "options": [
        { "id": "A", "text": "5" },
        { "id": "B", "text": "7" },
        { "id": "C", "text": "9" },
        { "id": "D", "text": "4" }
      ],
      "order_index": 0
    }
  ],
  "progress": { "status": "IN_PROGRESS", "last_exercise_index": 2 }
}
```
Note: `correct_answer` and `explanation` are NOT included in this response — returned only via the answer endpoint.

---

### POST `/lessons/{lessonId}/exercises/{exerciseId}/answer`
Submit an answer. Auth required.

**Request**
```json
{ "answer": "B" }
```
**Response 200**
```json
{
  "correct": true,
  "correct_answer": "B",
  "explanation": "f(2) = 3(2) + 1 = 7",
  "lesson_complete": false,
  "next_exercise_index": 3
}
```
When `lesson_complete: true`, the response also includes:
```json
{
  "lesson_complete": true,
  "xp_earned": 20,
  "concept_state": "LEARNING"
}
```

---

## Learning & Reviews (`/learning`)

### GET `/learning/review-session`
Get the current day's due review batch (max 20 concepts). Auth required.

**Response 200**
```json
{
  "session_id": "uuid",
  "due_count": 12,
  "concepts": [
    {
      "concept_id": "uuid",
      "concept_name": "Funções de 1º grau",
      "track_id": "uuid",
      "exercise": {
        "id": "uuid",
        "type": "MULTIPLE_CHOICE",
        "prompt": "...",
        "options": [...]
      }
    }
  ]
}
```
Returns `{"due_count": 0, "concepts": []}` when nothing is due.

---

### POST `/learning/reviews/{conceptId}/answer`
Submit a review answer and receive updated FSRS rating. Auth required.

**Request**
```json
{ "answer": "B", "session_id": "uuid" }
```
**Response 200**
```json
{
  "correct": true,
  "correct_answer": "B",
  "explanation": "...",
  "fsrs_rating": 3,
  "next_review_in_days": 7,
  "session_complete": false
}
```
When `session_complete: true`:
```json
{
  "session_complete": true,
  "concepts_strengthened": 12,
  "xp_earned": 30
}
```

---

## Simulations (`/simulations`)

### POST `/simulations`
Start a new ENEM simulation. Auth required.

**Response 201**
```json
{
  "id": "uuid",
  "status": "IN_PROGRESS",
  "started_at": "ISO-8601",
  "question_count": 180
}
```
**Errors**: `SIMULATION_ALREADY_IN_PROGRESS` 409

---

### GET `/simulations/{simulationId}`
Get simulation state with all questions and saved answers. Auth required.

**Response 200**
```json
{
  "id": "uuid",
  "status": "IN_PROGRESS",
  "started_at": "ISO-8601",
  "questions": [
    {
      "id": "uuid",
      "order_index": 1,
      "enem_area": "MATHEMATICS",
      "prompt": "...",
      "support_texts": [...],
      "options": [{ "id": "A", "text": "..." }],
      "saved_answer": "B"
    }
  ],
  "dissertation": {
    "prompt_id": "uuid",
    "prompt_text": "...",
    "support_texts": [...],
    "saved_content": "partial text...",
    "submission_id": "uuid"
  }
}
```

---

### PATCH `/simulations/{simulationId}/answers/{questionId}`
Save or update an answer. Auth required.

**Request**
```json
{ "selected_option": "C" }
```
**Response 200**
```json
{ "saved": true }
```

---

### POST `/simulations/{simulationId}/submit`
Submit the completed simulation for scoring. Auth required.
Triggers async result computation; `result` will be populated when status becomes `COMPLETED`.

**Response 202**
```json
{ "status": "COMPLETED", "result": { ... } }
```
`result` shape:
```json
{
  "scores_by_area": {
    "MATHEMATICS": { "score": 620, "total_questions": 45, "correct": 28 },
    "NATURAL_SCIENCES": { "score": 580, "total_questions": 45, "correct": 26 }
  },
  "total_score": 2300,
  "weakest_competencies": ["C3", "C4"]
}
```

---

### GET `/simulations`
List the student's simulations. Auth required.

**Query params**: `?status=IN_PROGRESS|COMPLETED`

**Response 200**
```json
{
  "simulations": [
    {
      "id": "uuid",
      "status": "COMPLETED",
      "started_at": "ISO-8601",
      "completed_at": "ISO-8601",
      "total_score": 2300
    }
  ]
}
```

---

## Dissertation (`/dissertation`)

### GET `/dissertation/prompts`
List available dissertation prompts. Auth required.

**Query params**: `?area=Meio+Ambiente&page=1&per_page=20`

**Response 200**
```json
{
  "prompts": [
    {
      "id": "uuid",
      "theme": "Desafios da reciclagem no Brasil",
      "thematic_area": "Meio Ambiente",
      "enem_year": null,
      "is_authentic_enem": false
    }
  ],
  "total": 45
}
```

---

### POST `/dissertation/submissions`
Submit a dissertation for feedback. Auth required.

**Request**
```json
{
  "prompt_id": "uuid",
  "content": "Essay text here...",
  "simulation_id": null
}
```
**Response 202**
```json
{
  "submission_id": "uuid",
  "status": "SUBMITTED",
  "estimated_feedback_seconds": 30
}
```

---

### GET `/dissertation/submissions/{submissionId}`
Get a submission with its feedback (if ready). Auth required.

**Response 200**
```json
{
  "id": "uuid",
  "prompt_id": "uuid",
  "content": "Essay text...",
  "word_count": 450,
  "status": "FEEDBACK_READY",
  "submitted_at": "ISO-8601",
  "feedback": {
    "c1_score": 160, "c1_notes": "...",
    "c2_score": 120, "c2_notes": "...",
    "c3_score": 140, "c3_notes": "...",
    "c4_score": 100, "c4_notes": "...",
    "c5_score": 120, "c5_notes": "...",
    "total_score": 640,
    "overall_notes": "..."
  }
}
```

---

### GET `/dissertation/submissions`
List the student's dissertation submissions. Auth required.

**Response 200**
```json
{
  "submissions": [
    {
      "id": "uuid",
      "prompt_id": "uuid",
      "theme": "Desafios da reciclagem no Brasil",
      "status": "FEEDBACK_READY",
      "total_score": 640,
      "submitted_at": "ISO-8601"
    }
  ]
}
```

---

## Versioning

- All breaking changes introduce a new major version (`/v2`)
- Non-breaking additions (new optional fields, new endpoints) do not require version bump
- Deprecated endpoints return `Sunset` and `Link` headers with migration path
