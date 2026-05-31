# Research: Backend Folder Reorganization for Monolith

**Feature**: 010-backend-monolith-cleanup
**Date**: 2026-05-25
**Status**: Complete (no open clarifications — see [spec.md](./spec.md) and `checklists/requirements.md`)

---

## Finding 1: One Go module beats nine

**Decision**: Collapse `backend/svc/*` (9 modules) into a single Go module `backend/app/` with domains under `internal/<domain>/`.

**Rationale**:
- The deployed binary is one. Nine modules with `replace` directives across them only made sense during the multi-service phase.
- A single module eliminates `replace` indirection, lets `go test ./...` from one root cover everything, and removes the need to update `go.work` whenever a new domain is added.
- `internal/` blocks accidental external imports, which is what we want — domains should not be consumed outside the app.

**Alternatives considered**:
- Keep `backend/svc/monolith/` and just delete the empty siblings: rejected — leaves the "svc" naming that implies "one of many services".
- Collapse all the way to `backend/` as the module root: rejected — leaves no place for `pkg/` (shared infra) without a same-module dependency cycle risk; mixing shared infra and app code in the same module is less hygienic than the two-module split.

---

## Finding 2: `git mv` over copy + delete

**Decision**: Every relocation uses `git mv`. No `cp` + `git rm`.

**Rationale**:
- Git rename detection (default threshold 50%) follows files moved with `git mv`. Blame survives.
- A delete-then-add diff confuses code review and obliterates history for anyone running `git log --follow`.

**Alternatives considered**:
- Use `git filter-repo` to rewrite history with the new paths: rejected — destructive; coordinating force-pushes across the team for a structural cleanup is not worth it.

---

## Finding 3: Single coordinated PR, not staged migration

**Decision**: One PR moves everything. No interim "both trees coexist" state.

**Rationale**:
- Reorg is mechanical: `git mv` + import-path replace + Dockerfile/compose edits. The diff is huge but conceptually trivial.
- A staged migration would mean keeping both `svc/monolith` and `app/` building simultaneously, which doubles the work and surfaces inconsistency bugs for no benefit.
- The window for breaking active branches is shorter with one PR.

**Alternatives considered**:
- Phase by domain: rejected — each domain's import-path change cascades; the second domain's PR would conflict with the first's.

---

## Finding 4: Module path becomes `github.com/preuni/app`

**Decision**: Rename module `github.com/preuni/svc/monolith` → `github.com/preuni/app`. All `github.com/preuni/svc/{auth,user}/*` imports become `github.com/preuni/app/internal/{auth,user}/*`.

**Rationale**:
- Matches the new directory layout. Confusion between `svc/` (the folder) and `svc` (the module prefix) goes away.
- Internal subpackages get the `internal/` privacy guarantee for free.

**Alternatives considered**:
- Keep `github.com/preuni/svc/monolith` and just move files: rejected — module path no longer matches directory; `go.mod` reads as a lie.

---

## Finding 5: Five stubs → empty internal packages with README markers

**Decision**: For `content`, `learning`, `simulation`, `dissertation`, `notification`, create `backend/app/internal/<domain>/README.md` saying "scaffolding only — no endpoints yet". No `.go` files in those packages.

**Rationale**:
- Reserves the namespace truthfully. When the time comes to implement, the location is obvious.
- A package with only `README.md` does not pollute `go build` output.
- Avoids the temptation to keep empty `package <domain>` files which would clutter test runs.

**Alternatives considered**:
- Delete the five stubs entirely until needed: rejected — small benefit to the tree, and developers writing new code lose the implicit "this is where it goes" signal.
- Keep a `doc.go` per stub instead of `README.md`: rejected — Go doc.go conventions are about godoc, not about "this is intentionally empty"; README is clearer for non-Go readers.

---

## Finding 6: Constitution describes the backend as a monolith now

**Decision**: Rewrite the backend section of `.specify/memory/constitution.md` and `CLAUDE.md` to describe a single Go binary composed of domain packages. Remove references to "the auth service" / "the mail service" as separate processes.

**Rationale**:
- Constitution is the single source of truth for engineering standards. If it lies about the architecture, every standard it sets loses credibility by association.
- `CLAUDE.md` is auto-loaded into every AI session; stale paths confuse tooling more than they help.

**Alternatives considered**:
- Leave the constitution and update only `CLAUDE.md`: rejected — same problem, smaller scope.

---

## Finding 7: NGINX upstreams pruned

**Decision**: Delete every `upstream <name>_svc { server <name>:NNNN; }` block whose service no longer exists in `infra/docker-compose.yml`.

**Rationale**:
- An nginx upstream pointing at a non-existent service is dead config. It loads fine but is misleading; future operators must verify each one is/isn't used.
- The gateway already routes `/v1/auth/*` and `/v1/students/*` to `monolith_svc` (feature 009 cutover). The other upstreams are unused.

**Alternatives considered**:
- Keep upstreams for documentation: rejected — comments document, not unused code.
