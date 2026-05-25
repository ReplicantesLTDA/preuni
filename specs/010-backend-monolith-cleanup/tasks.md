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

- [X] T001 Working tree clean (pending feature-009 + spec-010 work committed first; tree empty after)
- [X] T002 Stack stopped (monolith + gateway)
- [X] T003 Rollback tag `pre-monolith-reorg` created locally (push deferred to team announcement)
- [X] T004 Baseline tests pass: monolith contract suite green pre-reorg

---

## Phase 2: Foundational (Tree moves — blocking US1, US2, US3)

**Purpose**: Mechanical relocation of all source files. All `git mv` operations happen here; no edits to file contents yet.

**⚠️ CRITICAL**: All subsequent phases assume the new tree exists.

- [X] T005 `git mv backend/svc/monolith backend/app`
- [X] T006 `backend/app/internal/auth/` exists (created by git mv)
- [X] T007 Moved auth subpackages (adapters, domain, handler, ports, repository, router) into `backend/app/internal/auth/`
- [X] T008 Removed `backend/svc/auth/{cmd,Dockerfile,go.mod,go.sum,.dockerignore}` + dir
- [X] T009 `backend/app/internal/user/` exists
- [X] T010 Moved user subpackages (domain, handler, repository, router) into `backend/app/internal/user/`
- [X] T011 Removed `backend/svc/user/{cmd,Dockerfile,go.mod,go.sum,.dockerignore}` + dir
- [X] T012 Removed Elixir `backend/svc/mail/` tree
- [X] T013 Removed all five stub `backend/svc/{content,learning,simulation,dissertation,notification}/`
- [X] T014 `backend/svc/` deleted; `ls backend/` = `app/  go.work  go.work.sum  pkg/`

**Checkpoint**: `ls backend/` returns exactly `app/  go.work  go.work.sum  pkg/`.

---

## Phase 3: User Story 1 — Single coherent backend tree (Priority: P1) 🎯 MVP

**Goal**: Module path renamed, all imports rewritten, build green.

**Independent Test**: `cd backend/app && go build ./... && go test ./...` returns zero errors. Live curl on monolith returns 201/200 on the same endpoints as before.

### Implementation for User Story 1

- [X] T015 [US1] Module path `github.com/preuni/svc/monolith` → `github.com/preuni/app`; `replace` for auth/user dropped; pkg replace kept
- [X] T016 [US1] `backend/go.work` lists only `./pkg` + `./app`
- [X] T017 [US1] Bulk import rewrite via sed across `backend/app/**/*.go`
- [X] T018 [US1] `grep -rn 'github.com/preuni/svc/' backend/app/` returns nothing
- [X] T019 [US1] Test files clean (rewrite caught everything)
- [X] T020 [US1] `go mod tidy` ran successfully; added testify/uuid as direct deps
- [X] T021 [US1] `go build ./...` zero errors
- [X] T022 [US1] `go vet ./...` zero issues
- [X] T023 [US1] `go test -short ./...` all green (mail, contract, integration, user handler/domain)
- [X] T024 [US1] Integration tests green against running Postgres

**Checkpoint**: Module renamed; all tests pass against the moved tree.

---

## Phase 4: User Story 2 — Deprecated source removed (Priority: P1)

**Goal**: Zero traces of Elixir mail or standalone auth/user binaries in the repository.

**Independent Test**: Greps for Elixir source, `svc/mail` paths, and `mail` compose service all return empty.

### Implementation for User Story 2

- [X] T025 [US2] `infra/docker-compose.yml` slimmed: removed auth/user/content/learning/simulation/dissertation/notification/mail blocks; monolith dockerfile path now `app/Dockerfile`
- [X] T026 [US2] `backend/app/Dockerfile` COPY paths drop `svc/`; only pkg + app copied; builds `./app/cmd/server`
- [X] T027 [US2] `infra/nginx/nginx.conf` rewritten: only `upstream monolith_svc`; only `/v1/auth/` + `/v1/students/` + `/health` locations
- [X] T028 [US2] Makefile `cd backend/svc/monolith` → `cd backend/app`; run-backend/stop-backend now list only `monolith gateway`
- [X] T029 [US2] Grep `backend/svc/` outside specs returns nothing
- [X] T030 [US2] `find backend -name '*.ex*'` returns nothing
- [X] T031 [US2] No Elixir/Phoenix/Swoosh references outside specs

**Checkpoint**: All three greps return empty; CI builds reproduce.

---

## Phase 5: User Story 3 — Stub services folded as empty domain packages (Priority: P2)

**Goal**: Each of the five stub domains has an `internal/<domain>/` directory with a one-line README. The original `svc/<stub>/` directories are gone (already done in T013).

**Independent Test**: `ls backend/app/internal/{content,learning,simulation,dissertation,notification}/README.md` lists all five files; each file is one line stating "scaffolding only — no endpoints yet".

### Implementation for User Story 3

- [X] T032 [P] [US3] `backend/app/internal/content/README.md` created
- [X] T033 [P] [US3] `backend/app/internal/learning/README.md` created
- [X] T034 [P] [US3] `backend/app/internal/simulation/README.md` created
- [X] T035 [P] [US3] `backend/app/internal/dissertation/README.md` created
- [X] T036 [P] [US3] `backend/app/internal/notification/README.md` created
- [X] T037 [US3] Staged all five READMEs
- [X] T038 [US3] `backend/app/internal/` lists exactly 11 dirs (adapters, auth, config, content, dissertation, learning, mail, notification, router, simulation, user)

**Checkpoint**: Five stub packages exist with truthful README markers; no Go files in them so they do not pollute `go test`.

---

## Phase 6: User Story 4 — Constitution + CLAUDE.md updated (Priority: P2)

**Goal**: Constitution and `CLAUDE.md` describe the backend as a Go monolith. No statements contradict the actual tree.

**Independent Test**: A diff of the two files shows the architecture sections rewritten; reading them end-to-end matches what `ls backend/` returns.

### Implementation for User Story 4

- [X] T039 [US4] Constitution gained Architecture section (monolith); version bumped 1.0.0 → 1.1.0; Last Amended 2026-05-25
- [X] T040 [US4] CLAUDE.md "Backend" section rewritten as single Go monolith block
- [X] T041 [US4] CLAUDE.md "Project Structure" rewritten to new `backend/{pkg,app}` tree
- [X] T042 [US4] CLAUDE.md "Commands" updated: all `cd backend/svc/...` → `cd backend/app`; Elixir block removed
- [X] T043 [US4] `backend/README.md` created with two-line layout summary + pointer to `app/README.md`
- [X] T044 [US4] Constitution + CLAUDE.md describe the actual tree (no contradictions remain)

**Checkpoint**: Constitution + `CLAUDE.md` reflect the actual tree. No "the auth service" / "Elixir mail" references remain outside historical specs.

---

## Phase 7: Polish & Validation

**Purpose**: End-to-end live verification matching the feature-009 baseline, then commit.

- [X] T045 `docker compose build monolith` → image built
- [X] T046 Stack up: postgres + redis + monolith + gateway all healthy
- [X] T047 `/health` → 200 `ok`
- [X] T048 Gateway register `/v1/auth/register` → 201 + AuthResponse
- [X] T049 `/v1/students/me` (Bearer JWT) → 200 + StudentResponse (proves in-process student provisioning still works post-reorg)
- [X] T050 `/internal/email/send` with internal token → 202 `{"status":"queued"}`
- [X] T051 `/v1/auth/login` x100 perf: avg 2.91ms, p50 2.67ms, p95 3.65ms, p99 25.79ms (constitution gate < 500ms ✓)
- [X] T052 `ls backend/` = `README.md  app  go.work  go.work.sum  pkg`
- [X] T053 `backend/svc/` does not exist
- [X] T054 `git log --follow backend/app/internal/auth/handler/register.go` shows pre-reorg commits (blame preserved)
- [X] T055 Single commit: `refactor: collapse backend into app/ module`
- [X] T056 Stack stopped (clean environment)

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
