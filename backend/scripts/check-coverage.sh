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
# RefreshTokenRepository.RevokeAllForCredential completely untested, and
# then 69.2% after covering GET /v1/students/me/data-export (0% before),
# and now 70.8% after covering OTPLoginVerifyHandler and PasswordResetHandler
# (confirm side)'s success paths, seeding a known OTP row directly since
# their request-side handlers generate one asynchronously in a goroutine.
# 70.8% -> 71.8% after covering those same request-side handlers' success
# branch too (email found + verified -> OTP generated), polling for the
# background goroutine's DB write instead of racing it. 71.9% -> 72.1%
# after covering RefreshTokenHandler's success path (rotate + re-issue),
# previously only exercised via its missing-field validation branch.
# 72.1% -> 72.3% after covering ReconcileOnce's "still pending, past
# GradingTimeout" default branch, previously only its completed/failed
# job-status branches were tested. 72.5% -> 73.2% after deleting four
# genuinely dead functions with zero call sites anywhere in src/ or
# tests/ (auth/domain: OTPCode.IsExpired, OTPCode.IsUsed,
# Credentials.UpdatePassword; testhelper: MustEnv) -- confirmed dead via
# grep before removal, not just low-coverage. 73.2% -> 73.6% after
# covering PasswordResetHandler (unknown email, weak new password) and
# LoginHandler (unknown email, unverified email, wrong password) branches
# that only had their happy path or missing-field validation tested.
# 73.6% -> 74.2% after covering ChangePasswordHandler (wrong current
# password, weak new password), OTPLoginVerifyHandler (unknown email),
# and ChangeEmailConfirmHandler (no OTP ever requested) branches.
# 74.2% -> 74.8% after a batched round covering VerifyEmailHandler
# (unknown email), RegisterHandler (weak password, empty display name,
# duplicate email), LogoutHandler (success + unknown-token branches,
# entirely untested before), DeleteAccountHandler (empty-body decode-
# fails branch), and SubmitEssayHandler (missing-field validation).
# 74.8% -> 75.2% after another batch: mail.render's unsupported-type
# branch, WeeklyLeaderboard's default-tier branch, the OTP-login request
# goroutine's unverified-email no-op branch, and submitting an essay for
# a credential with no backing users.students row (streak.GetForUpdate's
# not-found branch -- simulates the register.go non-fatal provisioning-
# failure path). 75.2% -> 75.5% after adding pure-domain unit tests for
# NewCredentials, HashPassword (incl. its previously-untested bcrypt-
# 72-byte-limit error branch, a real gap between what ValidatePassword
# allows (255 chars) and what bcrypt accepts), and tierIndex's unknown-
# tier fallback. 75.5% -> 76.7% after covering internal/config (0%
# before, pure env-var parsing -- Load/getEnv/envInt/require), the
# in-process email adapter's Send (success/validation-failure/
# underlying-failure), defaultUsername's short-id branch, and Create's
# duplicate-username 409 branch (isDuplicateKeyError, previously
# unreachable via any real registration flow). 76.7% -> 78.1% after
# reaching real apperrors.Internal(err) branches across the auth/
# gamification/social repositories via an already-canceled
# context.Context -- genuine pgx behavior (the same error a real client
# disconnect or request timeout produces in production), not a mock.
# Covered: CredentialsRepository, OTPRepository, RefreshTokenRepository,
# gamification's AwardStreakMedals/ListMedals, social's friendship repo.
# Also added CredentialsRepository.UpdateEmail's duplicate-email 409
# branch (previously unreachable via any real change-email flow, same
# shape as the Create duplicate-username gap fixed earlier).
# Floor set to 77 for headroom.
#
# Usage: ./check-coverage.sh (run from backend/app/)

set -euo pipefail

COVERAGE_FLOOR="${COVERAGE_FLOOR:-77}"

cd "$(dirname "${BASH_SOURCE[0]}")/../app"

go test -p 1 -coverpkg=./... -coverprofile=/tmp/preuni-backend-coverage.out ./...

TOTAL=$(go tool cover -func=/tmp/preuni-backend-coverage.out | tail -1 | awk '{print $NF}' | tr -d '%')

echo "Go monolith coverage: ${TOTAL}% (floor: ${COVERAGE_FLOOR}%)"

if (( $(echo "${TOTAL} < ${COVERAGE_FLOOR}" | bc -l) )); then
    echo "FAIL: coverage ${TOTAL}% is below the ${COVERAGE_FLOOR}% floor"
    exit 1
fi

echo "OK: coverage meets the floor"
