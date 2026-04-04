# Data Model: preuni.com.br – ENEM Prep Platform

**Branch**: `001-enem-prep-platform` | **Date**: 2026-04-03

Each microservice owns its own PostgreSQL schema. Tables are grouped by owning service. Cross-service access is via API only — no cross-schema queries in application code.

---

## auth-svc schema

### `credentials`

Stores authentication credentials only. Profile data is in `user-svc`.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK, default gen_random_uuid() | |
| `email` | VARCHAR(320) | UNIQUE, NOT NULL | Normalized to lowercase |
| `password_hash` | VARCHAR(256) | NOT NULL | bcrypt, cost ≥ 12 |
| `email_verified` | BOOLEAN | NOT NULL, default false | |
| `email_verified_at` | TIMESTAMPTZ | | |
| `created_at` | TIMESTAMPTZ | NOT NULL, default now() | |
| `updated_at` | TIMESTAMPTZ | NOT NULL, default now() | |

### `refresh_tokens`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `credential_id` | UUID | NOT NULL, FK → credentials(id) ON DELETE CASCADE | |
| `token_hash` | VARCHAR(256) | UNIQUE, NOT NULL | SHA-256 of the raw token |
| `expires_at` | TIMESTAMPTZ | NOT NULL | |
| `revoked_at` | TIMESTAMPTZ | | NULL = still valid |
| `issued_at` | TIMESTAMPTZ | NOT NULL, default now() | |

### `otp_codes`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `credential_id` | UUID | NOT NULL, FK → credentials(id) ON DELETE CASCADE | |
| `purpose` | VARCHAR(32) | NOT NULL | EMAIL_VERIFY, PASSWORD_RESET, LOGIN_OTP |
| `code_hash` | VARCHAR(256) | NOT NULL | SHA-256; never store plaintext |
| `expires_at` | TIMESTAMPTZ | NOT NULL | 15 min TTL |
| `used_at` | TIMESTAMPTZ | | NULL = not yet consumed |
| `created_at` | TIMESTAMPTZ | NOT NULL, default now() | |

---

## user-svc schema

### `students`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | Same UUID as `credentials.id` (shared across services) |
| `display_name` | VARCHAR(64) | NOT NULL | |
| `avatar_url` | TEXT | | S3-compatible URL |
| `xp_total` | INT | NOT NULL, default 0, CHECK ≥ 0 | Cumulative experience points |
| `streak_count` | INT | NOT NULL, default 0, CHECK ≥ 0 | Consecutive active days |
| `streak_last_active_date` | DATE | | NULL for new students |
| `readiness_score` | SMALLINT | CHECK 0–100 | Composite ENEM readiness %; recomputed async |
| `created_at` | TIMESTAMPTZ | NOT NULL, default now() | |
| `updated_at` | TIMESTAMPTZ | NOT NULL, default now() | |

### `track_enrollments`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `student_id` | UUID | NOT NULL, FK → students(id) ON DELETE CASCADE | |
| `track_id` | UUID | NOT NULL | ID from content-svc (no FK — cross-service) |
| `enrolled_at` | TIMESTAMPTZ | NOT NULL, default now() | |
| `unenrolled_at` | TIMESTAMPTZ | | NULL = currently enrolled |
| UNIQUE | | (student_id, track_id) where unenrolled_at IS NULL | One active enrollment per track |

### `achievements`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `code` | VARCHAR(64) | UNIQUE, NOT NULL | e.g., STREAK_7, FIRST_SIMULATION |
| `name` | VARCHAR(128) | NOT NULL | Display name |
| `description` | TEXT | NOT NULL | |
| `icon_url` | TEXT | NOT NULL | |
| `condition_type` | VARCHAR(32) | NOT NULL | STREAK_DAYS, SIMULATIONS_COMPLETED, TRACK_COMPLETED, XP_EARNED |
| `condition_value` | INT | NOT NULL | Threshold for unlock |

### `student_achievements`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `student_id` | UUID | NOT NULL, FK → students(id) ON DELETE CASCADE | |
| `achievement_id` | UUID | NOT NULL, FK → achievements(id) | |
| `unlocked_at` | TIMESTAMPTZ | NOT NULL, default now() | |
| UNIQUE | | (student_id, achievement_id) | Can only unlock once |

### `xp_events`

Append-only log; `xp_total` in `students` is the materialized sum.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `student_id` | UUID | NOT NULL, FK → students(id) | |
| `amount` | INT | NOT NULL, CHECK > 0 | |
| `source` | VARCHAR(32) | NOT NULL | LESSON_COMPLETE, REVIEW_SESSION, SIMULATION_COMPLETE |
| `source_id` | UUID | | ID of the triggering entity |
| `earned_at` | TIMESTAMPTZ | NOT NULL, default now() | |

---

## content-svc schema

### `subject_tracks`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `slug` | VARCHAR(64) | UNIQUE, NOT NULL | e.g., mathematics, natural-sciences |
| `name` | VARCHAR(128) | NOT NULL | Portuguese display name |
| `description` | TEXT | NOT NULL | |
| `icon_url` | TEXT | NOT NULL | |
| `color_token` | VARCHAR(32) | NOT NULL | Design system token, e.g., track-math-500 |
| `enem_area` | VARCHAR(32) | NOT NULL | LANGUAGES, HUMAN_SCIENCES, NATURAL_SCIENCES, MATHEMATICS, WRITING |
| `order_index` | SMALLINT | NOT NULL | Display order |
| `is_active` | BOOLEAN | NOT NULL, default true | |

### `concepts`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `track_id` | UUID | NOT NULL, FK → subject_tracks(id) | |
| `name` | VARCHAR(256) | NOT NULL | e.g., "Funções de 1º grau" |
| `description` | TEXT | | |
| `order_index` | INT | NOT NULL | Within track |

### `lessons`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `concept_id` | UUID | NOT NULL, FK → concepts(id) | One lesson per concept |
| `track_id` | UUID | NOT NULL, FK → subject_tracks(id) | Denormalized for filtering |
| `title` | VARCHAR(256) | NOT NULL | |
| `description` | TEXT | | |
| `order_index` | INT | NOT NULL | Within track |
| `is_published` | BOOLEAN | NOT NULL, default false | |

### `exercises`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `lesson_id` | UUID | NOT NULL, FK → lessons(id) ON DELETE CASCADE | |
| `type` | VARCHAR(32) | NOT NULL | MULTIPLE_CHOICE, FILL_BLANK, MATCHING, ORDERING |
| `prompt` | TEXT | NOT NULL | Question text |
| `support_text` | TEXT | | Context or reading passage |
| `options` | JSONB | | Array of {id, text} for MCQ and MATCHING |
| `correct_answer` | TEXT | NOT NULL | Key answer (option id or fill-blank value) |
| `explanation` | TEXT | NOT NULL | Shown after answer |
| `difficulty` | VARCHAR(8) | NOT NULL | EASY, MEDIUM, HARD |
| `order_index` | SMALLINT | NOT NULL | Within lesson (max 10) |

---

## learning-svc schema

### `concept_states`

Per-student FSRS state. One row per (student_id × concept_id) pair.
Indexed on `(student_id, fsrs_due)` for efficient "due reviews" queries.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `student_id` | UUID | NOT NULL | From user-svc (no FK) |
| `concept_id` | UUID | NOT NULL | From content-svc (no FK) |
| `fsrs_stability` | DOUBLE PRECISION | NOT NULL, default 0 | Days at R = 0.9 |
| `fsrs_difficulty` | DOUBLE PRECISION | NOT NULL, default 5.0 | Range 1–10 |
| `fsrs_state` | VARCHAR(16) | NOT NULL, default 'NEW' | NEW, LEARNING, REVIEW, RELEARNING |
| `fsrs_due` | TIMESTAMPTZ | NOT NULL | Next review time |
| `fsrs_last_review` | TIMESTAMPTZ | | |
| `reps` | INT | NOT NULL, default 0 | Total review count |
| `lapses` | INT | NOT NULL, default 0 | Forgotten count |
| `created_at` | TIMESTAMPTZ | NOT NULL, default now() | |
| `updated_at` | TIMESTAMPTZ | NOT NULL, default now() | |
| UNIQUE | | (student_id, concept_id) | |

Index: `CREATE INDEX idx_concept_states_due ON concept_states (student_id, fsrs_due) WHERE fsrs_state != 'NEW';`

### `review_logs`

Append-only log of every review answer. Used for FSRS parameter optimization in v2.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `student_id` | UUID | NOT NULL | |
| `concept_id` | UUID | NOT NULL | |
| `rating` | SMALLINT | NOT NULL, CHECK 1–4 | 1=Again, 2=Hard, 3=Good, 4=Easy |
| `state_before` | VARCHAR(16) | NOT NULL | FSRS state before this review |
| `stability_before` | DOUBLE PRECISION | NOT NULL | |
| `difficulty_before` | DOUBLE PRECISION | NOT NULL | |
| `scheduled_days` | INT | | Days since last review (elapsed interval) |
| `reviewed_at` | TIMESTAMPTZ | NOT NULL, default now() | |

### `lesson_progress`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `student_id` | UUID | NOT NULL | |
| `lesson_id` | UUID | NOT NULL | From content-svc |
| `status` | VARCHAR(16) | NOT NULL, default 'NOT_STARTED' | NOT_STARTED, IN_PROGRESS, COMPLETED |
| `last_exercise_index` | SMALLINT | NOT NULL, default 0 | Resume position |
| `started_at` | TIMESTAMPTZ | | |
| `completed_at` | TIMESTAMPTZ | | |
| UNIQUE | | (student_id, lesson_id) | |

---

## simulation-svc schema

### `simulation_questions`

Content-only table; not student-specific.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `enem_area` | VARCHAR(32) | NOT NULL | LANGUAGES, HUMAN_SCIENCES, NATURAL_SCIENCES, MATHEMATICS |
| `enem_year` | SMALLINT | | NULL = original/curated |
| `prompt` | TEXT | NOT NULL | |
| `support_texts` | JSONB | | Array of {type, content} (text, image_url, table) |
| `options` | JSONB | NOT NULL | Array of {id: 'A'|'B'|'C'|'D'|'E', text} |
| `correct_option` | CHAR(1) | NOT NULL, CHECK IN ('A','B','C','D','E') | |
| `competency_tags` | TEXT[] | | ENEM competency codes |
| `difficulty` | VARCHAR(8) | NOT NULL | EASY, MEDIUM, HARD |
| `order_index` | SMALLINT | NOT NULL | Position in full simulation (1–180) |
| `is_active` | BOOLEAN | NOT NULL, default true | |

### `simulations`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `student_id` | UUID | NOT NULL | |
| `status` | VARCHAR(16) | NOT NULL, default 'IN_PROGRESS' | IN_PROGRESS, COMPLETED, ABANDONED |
| `started_at` | TIMESTAMPTZ | NOT NULL, default now() | |
| `completed_at` | TIMESTAMPTZ | | |
| `result` | JSONB | | Populated on completion; {area → {score, total, competency_breakdown}} |

### `simulation_answers`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `simulation_id` | UUID | NOT NULL, FK → simulations(id) ON DELETE CASCADE | |
| `question_id` | UUID | NOT NULL, FK → simulation_questions(id) | |
| `selected_option` | CHAR(1) | CHECK IN ('A','B','C','D','E') | NULL = skipped/unanswered |
| `answered_at` | TIMESTAMPTZ | | |
| UNIQUE | | (simulation_id, question_id) | |

---

## dissertation-svc schema

### `dissertation_prompts`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `theme` | VARCHAR(256) | NOT NULL | Short title |
| `thematic_area` | VARCHAR(64) | NOT NULL | e.g., Direitos Humanos, Meio Ambiente |
| `enem_year` | SMALLINT | | NULL = curated/original |
| `is_authentic_enem` | BOOLEAN | NOT NULL, default false | |
| `prompt_text` | TEXT | NOT NULL | Full contextualized instruction |
| `support_texts` | JSONB | NOT NULL | Array of {type, title?, content} |
| `is_active` | BOOLEAN | NOT NULL, default true | |
| `created_at` | TIMESTAMPTZ | NOT NULL, default now() | |

### `dissertation_submissions`

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `student_id` | UUID | NOT NULL | |
| `prompt_id` | UUID | NOT NULL, FK → dissertation_prompts(id) | |
| `simulation_id` | UUID | | NULL = standalone practice |
| `content` | TEXT | NOT NULL | Student's written essay |
| `word_count` | SMALLINT | NOT NULL | Computed on save |
| `status` | VARCHAR(16) | NOT NULL, default 'DRAFT' | DRAFT, SUBMITTED, FEEDBACK_READY |
| `submitted_at` | TIMESTAMPTZ | | |
| `created_at` | TIMESTAMPTZ | NOT NULL, default now() | |
| `updated_at` | TIMESTAMPTZ | NOT NULL, default now() | |

### `dissertation_feedback`

One row per submission. Created by `dissertation-svc` after AI processing.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | UUID | PK | |
| `submission_id` | UUID | UNIQUE, NOT NULL, FK → dissertation_submissions(id) | |
| `c1_score` | SMALLINT | NOT NULL, CHECK 0–200 | Competência 1: Domínio da norma culta |
| `c1_notes` | TEXT | NOT NULL | |
| `c2_score` | SMALLINT | NOT NULL, CHECK 0–200 | Competência 2: Compreensão do tema |
| `c2_notes` | TEXT | NOT NULL | |
| `c3_score` | SMALLINT | NOT NULL, CHECK 0–200 | Competência 3: Seleção e organização de informações |
| `c3_notes` | TEXT | NOT NULL | |
| `c4_score` | SMALLINT | NOT NULL, CHECK 0–200 | Competência 4: Mecanismos linguísticos |
| `c4_notes` | TEXT | NOT NULL | |
| `c5_score` | SMALLINT | NOT NULL, CHECK 0–200 | Competência 5: Proposta de intervenção |
| `c5_notes` | TEXT | NOT NULL | |
| `total_score` | SMALLINT | NOT NULL, CHECK 0–1000 | Sum of c1–c5 |
| `overall_notes` | TEXT | NOT NULL | |
| `generated_at` | TIMESTAMPTZ | NOT NULL, default now() | |

---

## State Transitions

### ConceptState (FSRS)

```
NEW ──[first answer]──► LEARNING ──[learning steps done]──► REVIEW
                                                               │
                              RELEARNING ◄──[rated Again]──────┘
                                   │
                              REVIEW ◄──[learning steps done again]
```

### LessonProgress

```
NOT_STARTED ──[first exercise answered]──► IN_PROGRESS ──[all exercises done]──► COMPLETED
```

### Simulation

```
IN_PROGRESS ──[all MCQ + dissertation submitted]──► COMPLETED
            ──[explicit abandon]──────────────────► ABANDONED
```

### DissertationSubmission

```
DRAFT ──[student submits]──► SUBMITTED ──[AI feedback returned]──► FEEDBACK_READY
```

---

## Cross-Service ID References

Services reference each other's entities by UUID without database-level foreign keys. Consistency is maintained by:
- API contract validation at service boundaries
- Soft delete (never hard-delete content referenced by learning/simulation data)
- student IDs are the same UUID across all services (issued by `auth-svc` at registration)
