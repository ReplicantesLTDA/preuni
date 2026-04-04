# Research: preuni.com.br – ENEM Prep Platform

**Branch**: `001-enem-prep-platform` | **Date**: 2026-04-03

## 1. Spaced-Repetition Algorithm: FSRS vs SM-2

**Decision**: **FSRS (Free Spaced Repetition Scheduler) v6**

**Rationale**:
- FSRS outperforms SM-2 in 91.9% of default decks; FSRS-6 shows 99.6% superiority in predictive accuracy
- 20–30% fewer reviews needed to maintain the same retention level
- Both Go and Kotlin have official open-source implementations maintained by the `open-spaced-repetition` org:
  - Go: `github.com/open-spaced-repetition/go-fsrs/v3`
  - Kotlin: `github.com/open-spaced-repetition/FSRS-Kotlin`
- Default pre-trained parameters (trained on 700 M Anki reviews from 20 K users) perform well for new students with no prior data — no cold-start problem
- Per-user parameter optimization kicks in meaningfully only after ≥ 1,000 reviews; defer optimization to v2

**Per-concept FSRS state fields** (stored in `ConceptState` table, one row per student × concept):

| Field | Type | Purpose |
|-------|------|---------|
| `stability` | float64 | Days until retrievability drops to 0.9; strength of memory trace |
| `difficulty` | float64 (1–10) | Inherent material complexity; affects how fast stability grows |
| `state` | enum | NEW → LEARNING → REVIEW ↔ RELEARNING |
| `due` | timestamp | Next scheduled review time |
| `last_review` | timestamp | When it was last reviewed |
| `reps` | int | Total number of reviews |
| `lapses` | int | Times the concept was forgotten (state reverted to RELEARNING) |

**Initial stability by first-answer grade** (pre-trained defaults):
- Forgot (Again): 0.40 days → schedule for same-day repeat
- Hard: 1.18 days
- Good: 3.17 days
- Easy: 15.69 days

**Alternatives considered**:
- SM-2 (1987): simpler but 20–30% less efficient; no official Kotlin library; rejected
- Custom algorithm: deferred to v2 (insufficient review data at launch)

---

## 2. Frontend: Kotlin Multiplatform (KMP) + Compose Multiplatform

**Decision**: KMP 2.x + Compose Multiplatform 1.8+ for Android and iOS (production); Kotlin/Wasm web target (Beta, early adopter)

**Rationale**:
- iOS Compose Multiplatform reached **stable** status in May 2025 (Compose 1.8.0) — safe for production
- Android has been stable; KMP shared business logic widely proven in production
- Web target (Kotlin/Wasm) is in Beta; suitable for early adopter launch but requires active monitoring

**Navigation**: **Decompose** (by arkivanov)
- Provides shared navigation logic in `commonMain` — not tied to UI framework
- Better separation of navigation state from UI than Voyager
- Works with all KMP targets (Android, iOS, Wasm)
- Integrates cleanly with MVIKotlin for business logic

**State management**: **Decompose + MVIKotlin**
- MVIKotlin provides deterministic MVI (Model-View-Intent) pattern: Store → State + Action/Intent
- Time-travel debugging available in development builds
- Both libraries are Kotlin-pure, no Android lifecycle dependency in shared code

**HTTP client**: Ktor Client (multiplatform)
**Local persistence**: SQLDelight (multiplatform SQL)
**Dependency injection**: Koin (multiplatform)

**⚠ Constitution Bundle-Size Note** (see Complexity Tracking in plan.md):
Kotlin/Wasm + Compose for Web produces a monolithic WASM binary typically 2–5 MB gzipped — far above the 150 kB per-chunk rule. This violates the letter but not the spirit of the constitution (the rule targets unnecessary JS chunk bloat). WASM is cached aggressively by browsers and counted as a single resource, not a route chunk. Mitigation: minimize non-essential dependencies, enable Kotlin release optimizations (IR + tree-shaking), defer non-critical initialization.

**Alternatives considered**:
- React Native: rejected — abandons Kotlin/JVM type safety and shared business logic
- Flutter: rejected — Dart ecosystem, less control over native rendering
- Native iOS/Android separate: rejected — doubles codebase and effort

---

## 3. Backend: Go Microservices Architecture

**Decision**: REST APIs (internal and external, v1), NGINX API gateway, JWT validated at gateway

**Rationale** (for a 2–5 engineer team, research-backed):
- REST everywhere is simpler to debug, iterate on, and operate at this team size; gRPC setup cost (proto tooling, build infra) adds friction without measurable benefit at v1 scale
- gRPC migration path preserved for high-throughput internal paths in v2 (proto definitions still written from day one in `/proto/` directory for future readiness)
- NGINX chosen over Kong/Envoy: proven, minimal operational overhead, handles routing + rate limiting + SSL termination, no dedicated API platform engineer required

**JWT strategy**: Validate signature + expiry at gateway; forward decoded `X-User-ID` and `X-User-Email` headers to downstream services. Services trust the gateway for external requests. Service-to-service calls use a shared internal service token (not user JWTs).

**PostgreSQL connection pooling**: `pgxpool` per service (pool size 20–30). Add PgBouncer in transaction mode only when total connection count approaches PostgreSQL `max_connections`. Do not add prematurely.

**Monorepo layout**:
```
backend/
├── pkg/          # Shared: logging (zap), config, middleware, errors
├── svc/          # One directory per microservice
├── proto/        # Protobuf definitions (for future gRPC migration)
└── go.work       # Go workspace for multi-module monorepo
```

**Microservices**:

| Service | Lang | Responsibility |
|---------|------|----------------|
| `gateway` | Go | Request routing, JWT validation, rate limiting |
| `auth-svc` | Go | Registration, login, token refresh, password reset |
| `user-svc` | Go | Student profiles, XP, streaks, achievements |
| `content-svc` | Go | Subject tracks, lessons, exercises (read-heavy, content CMS) |
| `learning-svc` | Go | Lesson progress, FSRS scheduling, review sessions |
| `simulation-svc` | Go | ENEM simulations, questions, results |
| `dissertation-svc` | Go | Prompts, submissions, AI feedback orchestration |
| `notification-svc` | Go | Push notifications (FCM/APNs) for streak reminders |
| `mail-svc` | Elixir | Transactional email (registration, OTP, password reset) |

**Alternatives considered**:
- gRPC internally: deferred to v2 (added complexity outweighs benefit at current team size)
- Kong gateway: rejected (requires dedicated operator; overkill for v1)
- Single monolith: considered for v1 simplicity but rejected because domain boundaries are clear and microservices from the start avoid expensive future splits; FSRS, dissertation AI, and email are naturally separate

---

## 4. Mailing Service: Elixir + Phoenix + Swoosh

**Decision**: Phoenix 1.7 + Swoosh for all transactional email

**Rationale**:
- Swoosh is the standard Elixir email library; supports adapter pattern (Mailgun, SendGrid, Postmark, SMTP) — no vendor lock-in
- Elixir's concurrency model (OTP/GenServer) is well-suited for email queuing and retry logic under load
- Service exposes a simple internal REST endpoint (POST `/internal/email/send`) called by other services
- Email types in scope for v1: registration confirmation, OTP (2FA), password reset, streak reminder digest (future)

**Communication with Go services**: HTTP REST. The mail service accepts a JSON payload describing the email type, recipient, and template variables. It renders HTML/text from EEx templates and dispatches via the configured adapter.

**Alternatives considered**:
- AWS SES directly from Go: rejected — loses the unified template management and retry logic Elixir provides
- Go-based mail (gomail): considered but email reliability and templating are better served by Elixir's OTP fault tolerance

---

## 5. AI Dissertation Feedback

**Decision**: **Anthropic Claude API** (claude-sonnet-4-6 or claude-haiku-4-5 for cost/speed balance)

**Rationale**:
- Dissertation feedback requires nuanced Portuguese language analysis across five official criteria — a task well-suited to a large language model
- Claude API provides a structured JSON response mode, making extraction of per-competency scores reliable
- claude-haiku-4-5 meets the 60-second feedback delivery target (SC-006) at low cost; claude-sonnet-4-6 as fallback for complex essays
- Feedback prompt includes: the essay text, the five ENEM competency rubric descriptions, and an instruction to respond with JSON containing scores (0–200 per criterion) and improvement notes per criterion

**Prompt strategy**:
- System prompt: ENEM official competency rubric for each of Competências 1–5
- User message: essay text + prompt theme
- Response format: structured JSON with `competency_N_score`, `competency_N_feedback`, `total_score`, `general_notes`
- Temperature: 0.3 (consistent grading, not creative)

**Alternatives considered**:
- OpenAI GPT-4o: viable but less cost-competitive at scale; Portuguese language support comparable
- Human review (immediate): out of scope for v1 per spec Assumptions
- Rule-based analysis: rejected — insufficient for nuanced Portuguese prose evaluation

---

## 6. Storage Architecture

**Decision**:

| Store | Technology | Use Case |
|-------|-----------|----------|
| Primary DB | PostgreSQL 16 | All persistent domain data |
| Cache / Sessions | Redis 7 | JWT refresh token store, rate-limit counters, review session state |
| Object Storage | S3-compatible (AWS S3 / MinIO local) | Essay supporting images, audio, profile avatars |

**PostgreSQL specifics**:
- Each microservice owns its own schema (not separate DB per service — operational overhead not justified at v1 scale). Cross-schema joins are forbidden; services communicate via API.
- Spaced repetition queries use a covering index on `(student_id, due)` in `concept_state` table for fast "due reviews" lookups.
- EXPLAIN plan review required for all queries involving joins (per constitution).

**Alternatives considered**:
- Separate PostgreSQL instance per service: deferred to v2 (operational complexity)
- MongoDB for exercises: rejected — relational structure benefits from SQL joins and JSONB handles flexible exercise options within PostgreSQL

---

## Summary of Key Decisions

| Topic | Decision | Primary Reason |
|-------|----------|---------------|
| Spaced repetition | FSRS v6 | 20-30% efficiency gain; official Go + Kotlin libs |
| KMP navigation | Decompose | Shared navigation in commonMain; KMP-native |
| KMP state | MVIKotlin | Deterministic MVI; testable pure Kotlin |
| Service communication | REST (internal + external, v1) | Simpler for small team; gRPC-ready via proto files |
| API gateway | NGINX | Proven, minimal ops, scales gradually |
| JWT auth | Validate at gateway | Centralized policy; avoids key distribution |
| DB connection | pgxpool per service | Sufficient for v1; PgBouncer added when needed |
| Email | Elixir + Swoosh | OTP fault tolerance; adapter flexibility |
| AI feedback | Claude API | Best Portuguese language understanding; structured JSON |
| Primary DB | PostgreSQL 16 | Single operational stack; JSONB for flexible fields |
