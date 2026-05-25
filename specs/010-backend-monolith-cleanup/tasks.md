# Tasks: Backend Folder Reorganization for Monolith Architecture

**Input**: Design documents from `/specs/010-backend-monolith-cleanup/`
**Prerequisites**: plan.md, spec.md, research.md, quickstart.md (no data-model.md / contracts/ — pure source-tree reorg)

**Tests**: No new tests written. Reuse the contract + integration tests authored in feature 009; they MUST pass post-reorg without assertion changes (FR-007). Test files are MOVED with their packages.

**Organization**: Tasks grouped by user story. Because this is a mechanical reorg, US1 (single coherent tree) is the load-bearing story; US2/US3/US4 are tightly coupled and largely complete by virtue of US1 finishing, but each has its own verification task.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Different files, no dependencies on incomplete tasks
- **[Story]**: US1, US2, US3, US4
- All paths absolute from repo root

## Path Conventions

Source: `backend/svc/{monolith,auth,user,mail,content,learning,simulation,dissertation,notification}/`
Target: `backend/app/` (single Go module) + `backend/pkg/` (unchanged shared infra).

---

## Phase 1: Setup (Pre-reorg safety)

**Purpose**: Establish rollback anchor + clean working tree before the structural move.

- [ ] T001 Verify clean working tree: `cd /Users/dwbessa/projects/preuni && git status --porcelain` returns empty
- [ ] T002 Stop docker stack: `docker compose -f /Users/dwbessa/projects/preuni/infra/docker-compose.yml stop monolith gateway`
- [ ] T003 Tag rollback anchor at HEAD: `git tag pre-monolith-reorg && git push origin pre-monolith-reorg` (push only after confirming the team is aware)
- [ ] T004 Confirm feature-009 baseline test suite passes pre-reorg: `cd /Users/dwbessa/projects/preuni/backend/svc/monolith && go test ./tests/contract/... ./tests/integration/... -count=1` → all green

---

## Phase 2: Foundational (Tree moves — blocking US1, US2, US3)

**Purpose**: Mechanical relocation of all source files. All `git mv` operations happen here; no edits to file contents yet.

**⚠️ CRITICAL**: All subsequent phases assume the new tree exists.

- [ ] T005 Rename monolith module directory: `cd /Users/dwbessa/projects/preuni && git mv backend/svc/monolith backend/app`
- [ ] T006 Create target dirs for auth domain: `mkdir -p /Users/dwbessa/projects/preuni/backend/app/internal/auth`
- [ ] T007 Move auth subpackages: `cd /Users/dwbessa/projects/preuni && for d in adapters domain handler ports repository router; do git mv backend/svc/auth/$d backend/app/internal/auth/$d; done`
- [ ] T008 Delete legacy auth cmd + Dockerfile + go.mod: `cd /Users/dwbessa/projects/preuni && git rm -r backend/svc/auth/cmd backend/svc/auth/Dockerfile backend/svc/auth/go.mod backend/svc/auth/go.sum && rmdir backend/svc/auth`
- [ ] T009 Create target dirs for user domain: `mkdir -p /Users/dwbessa/projects/preuni/backend/app/internal/user`
- [ ] T010 Move user subpackages: `cd /Users/dwbessa/projects/preuni && for d in domain handler repository router; do git mv backend/svc/user/$d backend/app/internal/user/$d; done`
- [ ] T011 Delete legacy user cmd + Dockerfile + go.mod: `cd /Users/dwbessa/projects/preuni && git rm -r backend/svc/user/cmd backend/svc/user/Dockerfile backend/svc/user/go.mod backend/svc/user/go.sum && rmdir backend/svc/user`
- [ ] T012 Delete Elixir mail tree entirely: `cd /Users/dwbessa/projects/preuni && git rm -r backend/svc/mail`
- [ ] T013 Delete stub services: `cd /Users/dwbessa/projects/preuni && git rm -r backend/svc/content backend/svc/learning backend/svc/simulation backend/svc/dissertation backend/svc/notification`
- [ ] T014 Verify `backend/svc/` is gone: `cd /Users/dwbessa/projects/preuni && rmdir backend/svc 2>/dev/null; test ! -d backend/svc`

**Checkpoint**: `ls backend/` returns exactly `app/  go.work  go.work.sum  pkg/`.

---

## Phase 3: User Story 1 — Single coherent backend tree (Priority: P1) 🎯 MVP

**Goal**: Module path renamed, all imports rewritten, build green.

**Independent Test**: `cd backend/app && go build ./... && go test ./...` returns zero errors. Live curl on monolith returns 201/200 on the same endpoints as before.

### Implementation for User Story 1

- [ ] T015 [US1] Edit `/Users/dwbessa/projects/preuni/backend/app/go.mod`: change module path `github.com/preuni/svc/monolith` → `github.com/preuni/app`; remove `replace github.com/preuni/svc/auth => ../auth` and `replace github.com/preuni/svc/user => ../user`; keep `replace github.com/preuni/pkg => ../pkg`
- [ ] T016 [US1] Edit `/Users/dwbessa/projects/preuni/backend/go.work`: replace `use (...)` block to list ONLY `./pkg` and `./app`
- [ ] T017 [US1] Bulk import rewrite across all `.go` files under `backend/app/`: `cd /Users/dwbessa/projects/preuni && find backend/app -name '*.go' -exec sed -i '' 's|github.com/preuni/svc/monolith/|github.com/preuni/app/|g; s|github.com/preuni/svc/auth|github.com/preuni/app/internal/auth|g; s|github.com/preuni/svc/user|github.com/preuni/app/internal/user|g' {} +`
- [ ] T018 [US1] Verify import rewrite caught everything: `cd /Users/dwbessa/projects/preuni && grep -rn 'github.com/preuni/svc/' backend/app/ | grep -v _test.go || echo NONE` (expected: NONE)
- [ ] T019 [US1] Rewrite imports in test files too: `cd /Users/dwbessa/projects/preuni && grep -rn 'github.com/preuni/svc/' backend/app/` (any remaining hits must also be patched by re-running T017)
- [ ] T020 [US1] Run `go mod tidy` inside the new module: `cd /Users/dwbessa/projects/preuni/backend/app && go mod tidy`
- [ ] T021 [US1] Compile-check: `cd /Users/dwbessa/projects/preuni/backend/app && go build ./...` → zero errors
- [ ] T022 [US1] Vet-check: `cd /Users/dwbessa/projects/preuni/backend/app && go vet ./...` → zero issues
- [ ] T023 [US1] Run unit + contract tests (no DB needed): `cd /Users/dwbessa/projects/preuni/backend/app && go test -short ./...`
- [ ] T024 [US1] Run integration tests against running Postgres: `cd /Users/dwbessa/projects/preuni && docker compose -f infra/docker-compose.yml up -d postgres redis && cd backend/app && TEST_DB_URL="postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable" go test ./tests/integration/... -count=1`

**Checkpoint**: Module renamed; all tests pass against the moved tree.

---

## Phase 4: User Story 2 — Deprecated source removed (Priority: P1)

**Goal**: Zero traces of Elixir mail or standalone auth/user binaries in the repository.

**Independent Test**: Greps for Elixir source, `svc/mail` paths, and `mail` compose service all return empty.

### Implementation for User Story 2

- [ ] T025 [US2] Update `/Users/dwbessa/projects/preuni/infra/docker-compose.yml`: remove any `auth`, `user`, `mail` service blocks; remove the `monolith_svc` Dockerfile path's `svc/` prefix → `app/Dockerfile`; gateway `depends_on` already targets `monolith`
- [ ] T026 [US2] Update `/Users/dwbessa/projects/preuni/backend/app/Dockerfile`: COPY paths drop `svc/` (was `COPY svc/monolith/ ./svc/monolith/` → `COPY app/ ./app/`); WORKDIR + build command target `./app/cmd/server`
- [ ] T027 [US2] Update `/Users/dwbessa/projects/preuni/infra/nginx/nginx.conf`: delete `upstream auth_svc`, `upstream user_svc`, and any other `*_svc` blocks whose compose service no longer exists. Keep `upstream monolith_svc`.
- [ ] T028 [US2] Update `/Users/dwbessa/projects/preuni/Makefile`: every `cd backend/svc/monolith` → `cd backend/app`; remove `mail`, `auth`, `user` from `run-backend` / `stop-backend` service lists
- [ ] T029 [US2] Grep audit — no `backend/svc/` references: `cd /Users/dwbessa/projects/preuni && grep -rn 'backend/svc/' Makefile infra/ backend/ 2>/dev/null | grep -v '/specs/' || echo NONE` (expected: NONE)
- [ ] T030 [US2] Grep audit — no Elixir source files: `cd /Users/dwbessa/projects/preuni && find backend -name '*.ex' -o -name '*.exs' -o -name 'mix.exs'` (expected: empty output)
- [ ] T031 [US2] Grep audit — no Elixir mail mentions outside specs: `cd /Users/dwbessa/projects/preuni && grep -rln 'Elixir\|Phoenix\|Swoosh\|mix phx' backend/ infra/ Makefile 2>/dev/null | grep -v specs/ || echo NONE` (expected: NONE)

**Checkpoint**: All three greps return empty; CI builds reproduce.

---

## Phase 5: User Story 3 — Stub services folded as empty domain packages (Priority: P2)

**Goal**: Each of the five stub domains has an `internal/<domain>/` directory with a one-line README. The original `svc/<stub>/` directories are gone (already done in T013).

**Independent Test**: `ls backend/app/internal/{content,learning,simulation,dissertation,notification}/README.md` lists all five files; each file is one line stating "scaffolding only — no endpoints yet".

### Implementation for User Story 3

- [ ] T032 [P] [US3] Create stub package marker: `mkdir -p /Users/dwbessa/projects/preuni/backend/app/internal/content && echo 'Scaffolding only — no endpoints yet. Implement when the content feature lands.' > /Users/dwbessa/projects/preuni/backend/app/internal/content/README.md`
- [ ] T033 [P] [US3] Same for learning: `mkdir -p /Users/dwbessa/projects/preuni/backend/app/internal/learning && echo 'Scaffolding only — no endpoints yet. Implement when the learning feature lands.' > /Users/dwbessa/projects/preuni/backend/app/internal/learning/README.md`
- [ ] T034 [P] [US3] Same for simulation: `mkdir -p /Users/dwbessa/projects/preuni/backend/app/internal/simulation && echo 'Scaffolding only — no endpoints yet. Implement when the simulation feature lands.' > /Users/dwbessa/projects/preuni/backend/app/internal/simulation/README.md`
- [ ] T035 [P] [US3] Same for dissertation: `mkdir -p /Users/dwbessa/projects/preuni/backend/app/internal/dissertation && echo 'Scaffolding only — no endpoints yet. Implement when the dissertation feature lands.' > /Users/dwbessa/projects/preuni/backend/app/internal/dissertation/README.md`
- [ ] T036 [P] [US3] Same for notification: `mkdir -p /Users/dwbessa/projects/preuni/backend/app/internal/notification && echo 'Scaffolding only — no endpoints yet. Implement when the notification feature lands.' > /Users/dwbessa/projects/preuni/backend/app/internal/notification/README.md`
- [ ] T037 [US3] Stage all five READMEs: `cd /Users/dwbessa/projects/preuni && git add backend/app/internal/{content,learning,simulation,dissertation,notification}/README.md`
- [ ] T038 [US3] Verify directory count: `cd /Users/dwbessa/projects/preuni && ls backend/app/internal/ | sort` shows exactly: `adapters`, `auth`, `config`, `content`, `dissertation`, `learning`, `mail`, `notification`, `router`, `simulation`, `user`

**Checkpoint**: Five stub packages exist with truthful README markers; no Go files in them so they do not pollute `go test`.

---

## Phase 6: User Story 4 — Constitution + CLAUDE.md updated (Priority: P2)

**Goal**: Constitution and `CLAUDE.md` describe the backend as a Go monolith. No statements contradict the actual tree.

**Independent Test**: A diff of the two files shows the architecture sections rewritten; reading them end-to-end matches what `ls backend/` returns.

### Implementation for User Story 4

- [ ] T039 [US4] Rewrite the backend architecture section of `/Users/dwbessa/projects/preuni/.specify/memory/constitution.md`: state that the backend is a single Go binary (`backend/app/`) composed of domain packages under `internal/`; remove references to "the auth service" / "the mail service" as separate processes; bump version (e.g. 1.0.0 → 1.1.0) and update `Last Amended` to 2026-05-25
- [ ] T040 [US4] Rewrite the "Backend" section of `/Users/dwbessa/projects/preuni/CLAUDE.md` "Active Technologies": collapse the dual entries (monolith + legacy auth/user + Elixir mail) into one paragraph stating "Single Go 1.24 binary at `backend/app/`; domains under `app/internal/<domain>/`; mail handled in-process via SMTP"
- [ ] T041 [US4] Rewrite the "Project Structure" section of `/Users/dwbessa/projects/preuni/CLAUDE.md`: replace the `backend/svc/{...}` tree with the new `backend/{go.work, pkg/, app/}` layout from `specs/010-backend-monolith-cleanup/plan.md`
- [ ] T042 [US4] Update the "Commands" section of `/Users/dwbessa/projects/preuni/CLAUDE.md`: all `cd backend/svc/...` → `cd backend/app`; remove the `Run Elixir mail service` block
- [ ] T043 [US4] Update or create `/Users/dwbessa/projects/preuni/backend/README.md` with the two-line layout summary + pointer to `app/README.md` for running locally
- [ ] T044 [US4] Validation — read both files end-to-end and confirm no statement contradicts `ls /Users/dwbessa/projects/preuni/backend/`

**Checkpoint**: Constitution + `CLAUDE.md` reflect the actual tree. No "the auth service" / "Elixir mail" references remain outside historical specs.

---

## Phase 7: Polish & Validation

**Purpose**: End-to-end live verification matching the feature-009 baseline, then commit.

- [ ] T045 Build container image fresh: `cd /Users/dwbessa/projects/preuni/infra && docker compose build monolith` → success
- [ ] T046 Bring stack up: `cd /Users/dwbessa/projects/preuni/infra && docker compose up -d postgres redis monolith gateway`
- [ ] T047 Smoke `/health`: `curl -sS http://localhost:8088/health` → `ok`
- [ ] T048 Smoke gateway register: `EMAIL="qs010+$(date +%s)@preuni.com.br"; curl -sS -w "\n%{http_code}\n" -X POST http://localhost:8080/v1/auth/register -H "Content-Type: application/json" -d "{\"email\":\"$EMAIL\",\"password\":\"P@ssw0rd123\",\"display_name\":\"Reorg\"}"` → 201 + AuthResponse
- [ ] T049 Smoke `/v1/students/me` with the returned access token → 200 + StudentResponse
- [ ] T050 Smoke internal mail send → 202: `curl -sS -X POST http://localhost:8088/internal/email/send -H "Authorization: Bearer $INTERNAL_SERVICE_TOKEN" -H "Content-Type: application/json" -d '{"type":"EMAIL_VERIFY","to":"dwbessa@gmail.com","params":{"otp":"010101"}}'`
- [ ] T051 Perf spot-check on `/v1/auth/login`: 100 sequential curls; confirm p95 ≤ ~5 ms (feature-009 baseline was 3.44 ms; constitution gate p95 < 500 ms)
- [ ] T052 Final layout audit: `ls /Users/dwbessa/projects/preuni/backend/` returns exactly `app/  go.work  go.work.sum  pkg/`
- [ ] T053 Final svc audit: `test ! -d /Users/dwbessa/projects/preuni/backend/svc && echo "svc gone"`
- [ ] T054 Final blame audit on a sample moved file: `cd /Users/dwbessa/projects/preuni && git log --follow backend/app/internal/auth/handler/register.go | head` shows pre-reorg commits (proves `git mv` preserved blame)
- [ ] T055 Commit as single PR: `git add -A && git commit -m "refactor: collapse backend into app/ module" -m "Merges svc/{auth,user,mail,monolith,content,learning,simulation,dissertation,notification} into a single Go module at backend/app/. Elixir mail tree removed. Stub services preserved as empty internal packages. HTTP contracts unchanged."`
- [ ] T056 Stop stack to leave a clean environment: `docker compose -f /Users/dwbessa/projects/preuni/infra/docker-compose.yml stop monolith gateway postgres redis`

---

## Dependencies & Execution Order

### Phase dependencies

- **Phase 1 (Setup)**: prerequisite for everything; safety tag + clean tree
- **Phase 2 (Foundational tree moves)**: MUST be complete before Phases 3-6 — every later phase assumes the new paths exist
- **Phase 3 (US1) — module + imports**: depends on Phase 2; blocks Phases 4-7 (no test passes until imports compile)
- **Phase 4 (US2) — deprecated removal**: depends on Phase 3 (Dockerfile/compose paths reference the renamed module)
- **Phase 5 (US3) — stub READMEs**: depends on Phase 2; parallelizable with Phase 4
- **Phase 6 (US4) — docs**: depends on Phase 2; parallelizable with Phases 4-5
- **Phase 7 (Polish)**: depends on Phases 3-6

### User story dependencies

- US1 (P1) — module rename: load-bearing; everything else depends on its tree state
- US2 (P1) — deprecated removal: depends on US1 only for the Dockerfile/compose path edits (T026)
- US3 (P2) — stub READMEs: independent after Phase 2
- US4 (P2) — docs: independent after Phase 2 (can be written in parallel with US2/US3)

### Within each story

- File renames precede file edits (Phase 2 before Phase 3)
- Module path edit (T015) precedes import rewrite (T017)
- Import rewrite precedes `go mod tidy` (T020)
- `go mod tidy` precedes build/test (T021-T024)

### Parallel opportunities

- T032-T036 (five stub READMEs): all [P] — different files
- Phases 4, 5, 6 can be split across three developers after Phase 3 lands

---

## Parallel Example: Stub READMEs

```bash
# All five can run together:
Task: "Create README at backend/app/internal/content/README.md"
Task: "Create README at backend/app/internal/learning/README.md"
Task: "Create README at backend/app/internal/simulation/README.md"
Task: "Create README at backend/app/internal/dissertation/README.md"
Task: "Create README at backend/app/internal/notification/README.md"
```

---

## Implementation Strategy

### Single PR

This is a mechanical refactor. Strategy:

1. Phase 1: tag rollback anchor; verify baseline
2. Phase 2: all `git mv` + `git rm` in one shot
3. Phase 3: edit module path + bulk-rewrite imports; build/test until green
4. Phases 4-6: tidy infra, READMEs, docs (parallelizable)
5. Phase 7: live smoke; commit; merge

### MVP scope

US1 + US2 (both P1). US1 makes the tree work; US2 makes it stop lying about what's deployed. US3 + US4 are documentation/cleanliness — important but not strictly required for the binary to keep serving traffic. Defer to a follow-up commit if any blocker emerges, but they're 6 file edits total and should ship in the same PR.

### Rollback

If anything goes wrong post-merge:

```bash
git revert <merge-commit>
# OR
git checkout pre-monolith-reorg  # tag from T003
```

---

## Notes

- `[P]` = different files, no incomplete-task deps
- All `git mv` operations must run inside the repo root (not via `cd` into subdirs) so the move is recorded relative to the repo
- The `sed -i ''` syntax in T017 is macOS-specific; on Linux drop the `''` after `-i`
- Bulk import rewrite catches most cases; rerun the grep audit (T018, T019) until clean
- Coverage gate (constitution ≥ 80%) is not affected — no tests deleted, only moved
- Feature 009's contract + integration test files move with the rest of the tree (`backend/svc/monolith/tests/` → `backend/app/tests/`)
- Commit in a single change so reviewers see one atomic refactor
