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
# shape as the Create duplicate-username gap fixed earlier). 78.1% ->
# 78.7% after more canceled-context fault injection (essay.ListByUser,
# gamification.EnsureCurrentWeekEntry/WeeklyLeaderboard/WeekClose,
# social.ListFriends, user.Update) and direct unit tests for
# StudentRepository's unexported contains/join/itoa helpers' previously
# unreachable branches (Update() always calls them with found/non-empty
# input). 78.7% -> 79.6% after cracking the tx-based gap: reconciler.go
# actually opens its own tx via r.db.Begin(ctx) (canceled-context works
# directly); streak.go's GetForUpdate/RecordSubmission take a caller-
# supplied pgx.Tx, solved by pool.Begin(ctx) + immediate Rollback(ctx)
# then calling the function with that now-closed tx -- pgx genuinely
# returns "tx is closed". Also proved canceled-context fault injection
# works at the HTTP layer via req.WithContext(canceledCtx), used for
# OnboardingHandler/ListEssays/GetStreak/ListFriends. Also added
# TestIntegration_Reconciler_MissingCorrectionsRowIsInternalError (a
# completed job with no matching corrections row -- real inconsistent-
# state edge case). 79.6% -> 80.2% after extending both techniques
# further: essay.Repository.Submit opens its own tx (same shape as
# reconciler.go), so canceled-context fault injection worked directly;
# HTTP-layer canceled-context tests added for DeleteStudent, DataExport,
# UpdateStudent (a real PATCH body, so the DB call fails, not JSON
# decoding), gamification's WeeklyLeaderboard, and MyMedals. 80.2% ->
# 80.6% after RefreshTokenHandler's unknown-token branch (a real
# nonexistent refresh_token string, no context tricks) and HTTP-layer
# canceled-context tests for ChangePasswordHandler, SubmitEssayHandler,
# the auth-domain DeleteAccountHandler, and RegisterHandler (a brand-new
# email, so this hits credRepo.Create's generic-failure branch, not the
# already-covered duplicate-email 409). 80.6% -> 80.8% after covering
# router.buildSender's two branches (Noop fallback vs. real
# *mail.SMTPSender) -- a pure config-branching function, zero coverage
# before. Remaining gaps mostly need a *second* DB call to fail after a
# first one succeeds (single-cancel-before-request can't target that),
# or breaking crypto/rand/HMAC internals (not real fault injection) --
# diminishing returns confirmed across two dedicated passes... until
# 80.8% -> 81.9% via a genuinely new deterministic technique: a pgx
# QueryTracer (nthQueryFailTracer, tests/integration/nth_query_fail_*.go)
# that returns an already-canceled context on exactly the Nth query,
# confirmed by reading pgx v5.9.2's Conn.Query source and empirically
# verified against real Postgres before use (each new test also run 3x
# individually with zero flakiness). Reaches "first DB call succeeds,
# second fails" branches plain canceled-context injection structurally
# cannot: RefreshTokenHandler's FindByID/Store/Revoke-fails,
# LogoutHandler's Revoke-fails, OnboardingHandler's FindByID-fails,
# essay.getGrade's Internal(err) branch, VerifyEmailHandler's
# MarkUsed/MarkEmailVerified-fails. This technique generalizes to any
# other remaining "second-call" gap in the codebase. 81.9% -> 83.4%
# after applying the same technique to ChangePasswordHandler,
# ChangeEmailConfirmHandler, PasswordResetHandler (confirm),
# OTPLoginVerifyHandler, and DeleteAccountHandler's second/third/fourth-
# call failure branches (12 sub-cases, each verified 3x for flakiness).
# Floor set to 83 for headroom.
#
# Usage: ./check-coverage.sh (run from backend/app/)

set -euo pipefail

COVERAGE_FLOOR="${COVERAGE_FLOOR:-83}"

cd "$(dirname "${BASH_SOURCE[0]}")/../app"

go test -p 1 -coverpkg=./... -coverprofile=/tmp/preuni-backend-coverage.out ./...

TOTAL=$(go tool cover -func=/tmp/preuni-backend-coverage.out | tail -1 | awk '{print $NF}' | tr -d '%')

echo "Go monolith coverage: ${TOTAL}% (floor: ${COVERAGE_FLOOR}%)"

if (( $(echo "${TOTAL} < ${COVERAGE_FLOOR}" | bc -l) )); then
    echo "FAIL: coverage ${TOTAL}% is below the ${COVERAGE_FLOOR}% floor"
    exit 1
fi

echo "OK: coverage meets the floor"
