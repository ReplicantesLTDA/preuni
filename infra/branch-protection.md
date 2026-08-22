# Branch protection: `main`

Constitution Principle VI (Engineering Workflow, NON-NEGOTIABLE): `main` is
protected — no direct pushes, no force-push, no bypassing status checks; the
only way code reaches `main` is a pull request with all required CI checks
green plus at least one human approval.

## Required settings

- **Require a pull request before merging** — direct pushes to `main` disabled.
- **Require approvals**: at least 1, including for changes authored by an AI
  assistant — Governance section has no exception for AI-authored code.
- **Require status checks to pass before merging**, with these checks
  required (from `.github/workflows/`):
  - `backend-ci / ci`
  - `correction-service-ci / ci`
  - `mobile-ci / ci`
- **Require branches to be up to date before merging.**
- **Do not allow force pushes.**
- **Do not allow deletions.**

## Apply via `gh`

```bash
gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  "/repos/{owner}/{repo}/branches/main/protection" \
  -f "required_status_checks[strict]=true" \
  -f "required_status_checks[contexts][]=backend-ci / ci" \
  -f "required_status_checks[contexts][]=correction-service-ci / ci" \
  -f "required_status_checks[contexts][]=mobile-ci / ci" \
  -F "enforce_admins=true" \
  -F "required_pull_request_reviews[required_approving_review_count]=1" \
  -F "restrictions="
```

Replace `{owner}/{repo}` with the actual GitHub org/repo. This repo's
"main branch for PRs" is currently `001-enem-prep-platform` (per the
project's branching convention — see recent commit history); re-run this
against whichever branch is designated the merge target once that's settled,
and again against `main` once the project cuts over.

## Verifying it's live

```bash
gh api "/repos/{owner}/{repo}/branches/main/protection" | jq '.required_status_checks, .required_pull_request_reviews'
```
