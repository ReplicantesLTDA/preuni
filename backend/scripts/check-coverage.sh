#!/usr/bin/env bash
# check-coverage.sh — run the Go monolith's full test suite with coverage
# and fail if it drops below the tracked floor.
#
# Constitution Principle II: coverage floor is 90%, reached incrementally
# (Principle VI: small PRs). COVERAGE_FLOOR below is today's *provisional*
# baseline, not yet 90% — specs/014-constitution-alignment-refactor/tasks.md
# T060 (Polish) raises it to 90% once the essay/streak/social/gamification
# test suites (US1-US3) land. Every PR must not lower the number below
# whatever COVERAGE_FLOOR currently is; bump it upward as coverage improves,
# never downward without a documented reason.
#
# Usage: ./check-coverage.sh (run from backend/app/)

set -euo pipefail

COVERAGE_FLOOR="${COVERAGE_FLOOR:-30}"

cd "$(dirname "${BASH_SOURCE[0]}")/../app"

go test -p 1 -coverprofile=/tmp/preuni-backend-coverage.out ./...

TOTAL=$(go tool cover -func=/tmp/preuni-backend-coverage.out | tail -1 | awk '{print $NF}' | tr -d '%')

echo "Go monolith coverage: ${TOTAL}% (floor: ${COVERAGE_FLOOR}%)"

if (( $(echo "${TOTAL} < ${COVERAGE_FLOOR}" | bc -l) )); then
    echo "FAIL: coverage ${TOTAL}% is below the ${COVERAGE_FLOOR}% floor"
    exit 1
fi

echo "OK: coverage meets the floor"
