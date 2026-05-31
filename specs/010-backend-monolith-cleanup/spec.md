# Feature Specification: Backend Folder Reorganization for Monolith Architecture

**Feature Branch**: `010-backend-monolith-cleanup`
**Created**: 2026-05-25
**Status**: Draft
**Input**: User description: "i want to check if all backend folders are being used... by taking a quick look i can see that the /svc/mail is deprecated by now, we can remove it. we did some refactor to use a monolith architecture instead of microservices... (we can updated constitution) and i believe that some folder recreation is need, because mail is inside monolith, but user is outside... lets update folder distribuction to follow a monolith architecture, lets organize these folders"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Single coherent backend tree (Priority: P1)

As a backend developer joining the project, I can open the `backend/` directory and immediately see one canonical layout that reflects the way the system actually runs (a single monolith), so I do not have to reason about which folders are live, which are deprecated, and which are partial implementations.

**Why this priority**: The repository currently mixes the live monolith package with leftover scaffolding from the multi-service era (`svc/auth`, `svc/user` retained for rollback only; `svc/mail` deprecated; `svc/{content,learning,simulation,dissertation,notification}` stubs that were never wired into production). New contributors waste hours figuring out which tree to edit. Consolidating the layout is the single biggest reduction in onboarding friction we can ship right now.

**Independent Test**: A new developer follows the README, runs the backend, and can locate the source for any production endpoint within 60 seconds without asking a teammate which directory to open.

**Acceptance Scenarios**:

1. **Given** the reorganized repository, **When** a developer searches for the authentication code, **Then** it lives in a clearly named submodule of the monolith (alongside the rest of the backend domain code), not in a sibling top-level directory.
2. **Given** the reorganized repository, **When** a developer looks for the mail-sending logic, **Then** it sits next to the other backend domains in the same module, not in a deprecated standalone folder.
3. **Given** the reorganized repository, **When** a developer reads the top-level `backend/` listing, **Then** every directory present is either actively used by the running binary or explicitly labelled as a separate concern (shared packages, infrastructure, etc.) — no rotting legacy trees remain.

---

### User Story 2 - Deprecated source removed (Priority: P1)

As a platform operator, I can be confident that what is checked into `main` reflects what is deployed, so that "rollback" no longer means "rebuild the deleted Elixir service from git history".

**Why this priority**: The Elixir mail service was removed from production deployment in feature 009 but its source tree (`backend/svc/mail/`) is still on disk. Anything that lives in `main` is a maintenance burden (dependabot alerts, lint runs, accidental imports). Removing dead code is a constitution requirement.

**Independent Test**: After the cleanup, the repository contains zero references to the Elixir mail service in source, build configs, CI workflows, or documentation (apart from historical mention in the 009 spec).

**Acceptance Scenarios**:

1. **Given** the cleaned-up repository, **When** a developer searches the tree for the Elixir mail source, **Then** the search returns no results outside of historical specification documents.
2. **Given** the cleaned-up repository, **When** CI runs, **Then** no job attempts to build, test, or lint the removed Elixir mail service.
3. **Given** the cleaned-up repository, **When** a developer runs the local development command, **Then** no container or process for the removed mail service is started or referenced.

---

### User Story 3 - Stub services either claimed or removed (Priority: P2)

As a tech lead, I can look at the backend tree and immediately tell which domains have real implementations versus which are placeholders, so estimation and planning are based on what actually exists.

**Why this priority**: `content`, `learning`, `simulation`, `dissertation`, and `notification` exist today as `cmd/server/main.go` stubs with no real handlers. They imply "we have these services" when in fact we have empty shells. Either fold them into the monolith as named-but-empty domain packages (truthful "these are stubs we will fill in") or delete them outright. Either choice is better than the current ambiguity.

**Independent Test**: A product manager can read the backend layout and correctly state, in one minute, which capabilities currently work and which are not yet implemented.

**Acceptance Scenarios**:

1. **Given** the reorganized repository, **When** a stakeholder asks "is the dissertation feature ready?", **Then** the answer is unambiguous from the folder structure alone (either it has real source code, or its placeholder is clearly labelled as stub/skeleton).
2. **Given** the reorganized repository, **When** the backend is started, **Then** no orphan binaries are launched whose source is empty.

---

### User Story 4 - Constitution reflects current architecture (Priority: P2)

As an engineering team member, I can read the project constitution and find that its description of the backend matches what is actually built, so the constitution remains a credible source of standards.

**Why this priority**: The constitution and `CLAUDE.md` still reference a multi-service backend. A constitution that contradicts the running system erodes trust in the entire document — including the parts that genuinely guide quality. Updating it is small but high-leverage.

**Independent Test**: A new engineer reads the constitution end-to-end and the description of the backend matches what they see in the repository, with no contradictions.

**Acceptance Scenarios**:

1. **Given** the updated constitution, **When** a reader gets to the architecture section, **Then** it states that the backend is a Go monolith with shared infrastructure packages.
2. **Given** the updated constitution, **When** standards reference "the auth service" or "the mail service", **Then** those references are reworded to reflect domain packages within the monolith.

---

### Edge Cases

- **Rollback path during transition**: If reorganization happens before production cutover of feature 009 is verified, the legacy `svc/auth` + `svc/user` standalone binaries may still be needed. The reorg must either wait for production validation, or preserve a recoverable git tag pointing at the pre-reorg tree.
- **External references**: Any docs, scripts, dashboards, or CI workflows that hard-code paths like `backend/svc/auth/...` need updating atomically with the move.
- **Import paths**: Go module paths embed directory names. Moving a package changes its import path everywhere; every internal consumer needs updating in the same change.
- **Test fixtures**: Integration test helpers reference DB schemas (e.g., `auth.credentials`, `users.students`). Schema names are out of scope; only the Go source layout changes.
- **In-flight feature branches**: Any branch that touches `backend/svc/auth/` or `backend/svc/user/` will conflict heavily after the move. The reorg should be timed against the active branch list and announced.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The backend source tree MUST present a single, canonical home for each backend domain (authentication, students/users, mail, and any other production-active domain), with no domain duplicated between the active monolith and a legacy top-level folder. **Acceptance**: `find backend/svc -maxdepth 1 -type d` returns one entry per actually-deployed component plus the shared package directory; no duplicated domain.
- **FR-002**: The deprecated Elixir mail service tree MUST be removed from the repository, including source, build descriptors, container definitions, and Makefile / compose references. **Acceptance**: `grep -r "svc/mail" backend/ infra/ Makefile` returns no matches outside historical specs.
- **FR-003**: The existing standalone `backend/svc/auth/` and `backend/svc/user/` modules MUST be merged into the monolith module as internal domain subpackages. After the merge, neither folder exists at the old top-level location. Their standalone `cmd/server/main.go` binaries MUST be removed. Rollback to a pre-monolith deployment, if ever needed, is performed via a git tag pointing at the pre-reorg commit. **Acceptance**: `ls backend/svc/auth backend/svc/user` returns "no such file or directory"; auth + user handlers, repositories, domain types, ports, adapters, and router builders are reachable under the monolith module.
- **FR-004**: Stub services (`content`, `learning`, `simulation`, `dissertation`, `notification`) MUST be folded into the monolith as empty internal domain subpackages. Each subpackage gets a one-line `README.md` marking it "scaffolding only — no endpoints yet". The five corresponding `backend/svc/<stub>/` top-level directories MUST be deleted. **Acceptance**: For each of the five stub names, the empty domain subpackage exists inside the monolith module with a README, and the original top-level `backend/svc/<stub>/` directory no longer exists.
- **FR-005**: The Go workspace file (`backend/go.work`) MUST list only modules that are still part of the canonical layout. **Acceptance**: Every entry in `go.work` corresponds to a directory that exists and is reachable by the build.
- **FR-006**: Local development commands (Makefile targets, compose services, dev scripts) MUST continue to work without changes to their public names. **Acceptance**: `make dev`, `make run-monolith`, `make test-backend`, `make migrate` all execute successfully against the reorganized tree.
- **FR-007**: Existing end-to-end behavior of the monolith binary MUST be preserved exactly. No routes change, no HTTP contracts change, no envelope shapes change. **Acceptance**: The contract + integration test suite added in feature 009 (`backend/svc/monolith/tests/...`, wherever its successor path is) passes against the reorganized layout without modification to assertions.
- **FR-008**: The project constitution and `CLAUDE.md` MUST be updated to describe the backend as a Go monolith composed of domain packages, removing references that imply a multi-service deployment topology. **Acceptance**: A diff of the constitution shows the architecture section rewritten; CLAUDE.md "Project Structure" section shows only paths that exist post-reorg.
- **FR-009**: NGINX gateway configuration MUST route every `/v1/*` family that exists today to the monolith, removing dead upstream blocks for services that no longer exist. **Acceptance**: `nginx.conf` references no upstream that has no corresponding compose service.
- **FR-010**: All Go import paths referenced from anywhere in the repository (tests, mains, internal packages) MUST resolve after the reorganization. **Acceptance**: `go build ./...` and `go test ./... -short` from each remaining module produce zero errors.

### Key Entities *(directory-level)*

Target layout after the reorganization:

```text
backend/
├── go.work                    # lists pkg + app only
├── pkg/                       # shared infra: config, logger, errors, middleware
└── app/                       # the monolith (renamed from backend/svc/monolith/)
    ├── go.mod
    ├── cmd/server/            # main entrypoint
    ├── internal/
    │   ├── auth/              # ex backend/svc/auth (handlers, domain, repo, ports, adapters, router)
    │   ├── user/              # ex backend/svc/user
    │   ├── mail/              # already inside; stays
    │   ├── content/           # empty stub + README
    │   ├── learning/          # empty stub + README
    │   ├── simulation/        # empty stub + README
    │   ├── dissertation/      # empty stub + README
    │   ├── notification/      # empty stub + README
    │   ├── adapters/          # cross-domain adapters (e.g. in-process student/email)
    │   ├── config/
    │   └── router/
    └── tests/                 # contract + integration tests
```

- **Active backend tree (`backend/`)**: Holds `pkg/` (shared infra) and `app/` (the monolith). Only two top-level entries inside `backend/`. No `svc/` directory at all.
- **Monolith module (`backend/app/`)**: The single deployable Go binary. Each backend domain (auth, user, mail, content, learning, simulation, dissertation, notification) is an `internal/<domain>/` subpackage of this module.
- **Shared packages (`backend/pkg/`)**: Cross-cutting infrastructure (config, logger, errors, middleware). Unchanged.
- **Removed trees**: `backend/svc/auth/`, `backend/svc/user/`, `backend/svc/mail/`, `backend/svc/content/`, `backend/svc/learning/`, `backend/svc/simulation/`, `backend/svc/dissertation/`, `backend/svc/notification/`, and `backend/svc/monolith/` (renamed to `backend/app/`). The entire `backend/svc/` directory disappears.
- **Infrastructure (`infra/`)**: Compose, NGINX, migrations. Updated to drop references to removed services and to point at `backend/app/` for builds.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new backend engineer can locate the source for any production HTTP endpoint within 60 seconds of opening the repository, without asking a teammate. Measured by onboarding shadowing of the next new hire (or a teammate volunteering for a timed exercise).
- **SC-002**: The repository contains zero source files for the Elixir mail service after the reorganization. Measured by `find backend -name "*.ex" -o -name "*.exs"` returning zero matches.
- **SC-003**: All existing integration and contract tests pass against the reorganized tree without changing assertions or test setup beyond import path renames. Measured by CI run of feature-009 test files.
- **SC-004**: The constitution and `CLAUDE.md` contain no statements that contradict the running architecture (verified by a documentation review checklist filled in during the change).
- **SC-005**: The `backend/svc/` directory is removed entirely. `backend/` contains exactly `go.work`, `pkg/`, and `app/`. Measured by `ls backend/` showing only those entries.
- **SC-006**: After the reorganization, the local dev loop (`make dev` from a clean checkout to a successful `/v1/auth/register → 201` curl) takes no longer than it did before the change. Measured by a one-time wall-clock comparison.

## Assumptions

- Production cutover of feature 009 (gateway routing to monolith) is either already validated, or the reorganization will be timed to happen after that validation. Standalone `svc/auth` + `svc/user` binaries are deleted; rollback to multi-service deployment, if ever needed, is performed by checking out the pre-reorg git tag (created as part of this work).
- Database schemas (`auth.*`, `users.*`) stay as they are. This change is about Go source layout, not data model.
- Existing client applications (mobile + web) are unaffected because the HTTP surface does not change.
- The five stub services (`content`, `learning`, `simulation`, `dissertation`, `notification`) have no production traffic today; choosing between "fold into monolith as empty packages" vs "delete entirely" is purely a documentation-of-intent decision.
- Git history is preserved via `git mv` (not delete + add) so blame survives the move.
- A single coordinated PR is preferred over a multi-PR migration, since the change is mostly mechanical renames and the team is small.

## Out of Scope

- Splitting the monolith into multiple binaries again.
- Implementing the stub domains (content, learning, simulation, dissertation, notification) — only deciding their folder fate.
- Refactoring handler-level code or repository-level code inside the moved packages.
- Database schema changes.
- Client (mobile/web) changes.
