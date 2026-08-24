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

The nested `-f`/`-F` flag form GitHub's docs show does not actually
validate against the current API schema (`strict` arrives as the string
`"true"`, not a boolean; an empty `restrictions=` isn't accepted as
null) — confirmed by trying it directly. Use a JSON payload instead:

```bash
cat > /tmp/branch-protection.json <<'JSON'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["backend-ci / ci", "correction-service-ci / ci", "mobile-ci / ci"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1
  },
  "restrictions": null
}
JSON

gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  "/repos/{owner}/{repo}/branches/main/protection" \
  --input /tmp/branch-protection.json

rm /tmp/branch-protection.json
```

Replace `{owner}/{repo}` with the actual GitHub org/repo. `dev` is this
repo's actual integration branch (GitHub's configured default branch,
the target of every feature PR) and stays unprotected-in-the-formal-
sense — day-to-day work lands there via ordinary review, not this gate.
`main` was unused until `035-self-hosted-prod-deploy` revived it as the
release branch that triggers production deploys (`docs/decisions/
0002-self-hosted-nas-production.md`); this command's target of `main`
is now correct and current, not aspirational.

## Verifying it's live

```bash
gh api "/repos/{owner}/{repo}/branches/main/protection" | jq '.required_status_checks, .required_pull_request_reviews'
```
