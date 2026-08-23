#!/usr/bin/env bash
# check-coverage.sh — run the Go monolith's full test suite with coverage
# and fail if it drops below the tracked floor.
#
# Constitution Principle II: coverage floor is 90%, reached incrementally
# (Principle VI: small PRs). COVERAGE_FLOOR below is today's *provisional*
# baseline, not yet 90%. Every PR must not lower the number below whatever
# COVERAGE_FLOOR currently is; bump it upward as coverage improves, never
# downward without a documented reason.
#
# -coverpkg=./... matters: most real exercise of essay/streak/social/
# gamification handler+repository code happens through tests/integration/
# (a different package hitting the full HTTP router), not same-package unit
# tests. Without -coverpkg, `go test ./...` only attributes coverage within
# each test binary's own package and drastically understates the real
# number — first measured wrong in backend-ci.yml (19.0%/15% floor); the
# corrected measurement against a local live Postgres (2026-08-22, this
# session) started at 55.2%, then 59.6%, 60.5%, then 64.0% after adding
# validation-error-path tests across auth (change_password, change_email,
# delete_account, otp, password_reset, refresh -- previously untested
# branches on handlers that already existed pre-014) and the user domain's
# onboarding/anonymize lifecycle endpoints, then 64.4% after covering
# essay.Repository.scanSubmission's not-found branch (GET /v1/essays/{id}
# for an unknown id), and now 68.6% after adding success-path tests for
# change_password, change_email (request+confirm), delete_account, and
# email/verify-resend -- these handlers previously only had their
# missing-field validation branch tested (auth_validation_test.go), never
# their actual success path, which had left CredentialsRepository's
# UpdatePasswordHash/UpdateEmail/Anonymize and
# RefreshTokenRepository.RevokeAllForCredential completely untested.
# Floor set to 67 for headroom.
#
# Usage: ./check-coverage.sh (run from backend/app/)

set -euo pipefail

COVERAGE_FLOOR="${COVERAGE_FLOOR:-67}"

cd "$(dirname "${BASH_SOURCE[0]}")/../app"

go test -p 1 -coverpkg=./... -coverprofile=/tmp/preuni-backend-coverage.out ./...

TOTAL=$(go tool cover -func=/tmp/preuni-backend-coverage.out | tail -1 | awk '{print $NF}' | tr -d '%')

echo "Go monolith coverage: ${TOTAL}% (floor: ${COVERAGE_FLOOR}%)"

if (( $(echo "${TOTAL} < ${COVERAGE_FLOOR}" | bc -l) )); then
    echo "FAIL: coverage ${TOTAL}% is below the ${COVERAGE_FLOOR}% floor"
    exit 1
fi

echo "OK: coverage meets the floor"
