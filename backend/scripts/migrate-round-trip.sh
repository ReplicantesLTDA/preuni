#!/usr/bin/env bash
# migrate-round-trip.sh — proves migration reversibility for infra/migrations/
# schemas that have full up/down coverage (Constitution VI).
#
# Unlike check-migration-downgrades.sh (which only checks that *.down.sql
# files exist), this actually EXECUTES the round-trip on a real Postgres:
# apply every up in order, apply every down in reverse order, apply every
# up again. A schema with any grandfathered (down-less) migration is
# skipped -- you can't round-trip a chain with a gap in it. As schemas
# gain full down coverage over time, this script starts proving them
# automatically, with no further script changes needed.
#
# Requires: psql on PATH, PGPASSWORD/-h/-U/-d matching backend-ci.yml's
# Postgres service (defaults below match local dev via docker-compose).
#
# Usage: ./migrate-round-trip.sh (run from repo root or anywhere)

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

MIGRATIONS_DIR="infra/migrations"
GRANDFATHER_LIST="$MIGRATIONS_DIR/.grandfathered"

PGHOST="${PGHOST:-localhost}"
PGUSER="${PGUSER:-preuni}"
PGDATABASE="${PGDATABASE:-preuni}"
export PGPASSWORD="${PGPASSWORD:-preuni}"

psql_exec() {
    psql -h "$PGHOST" -U "$PGUSER" -d "$PGDATABASE" -v ON_ERROR_STOP=1 -f "$1"
}

any_ran=0

for schema_dir in "$MIGRATIONS_DIR"/*/; do
    schema="$(basename "$schema_dir")"
    ups=()
    while IFS= read -r -d '' f; do
        ups+=("$f")
    done < <(find "$schema_dir" -maxdepth 1 -name '*.sql' ! -name '*.down.sql' ! -name 'grant-*.sql' -print0 | sort -z)

    [ "${#ups[@]}" -eq 0 ] && continue

    fully_paired=1
    for up in "${ups[@]}"; do
        down="${up%.sql}.down.sql"
        if [ ! -f "$down" ]; then
            fully_paired=0
            break
        fi
    done

    if [ "$fully_paired" -ne 1 ]; then
        echo "SKIP $schema: has a grandfathered (down-less) migration, can't round-trip yet"
        continue
    fi

    echo "ROUND-TRIP $schema: ${#ups[@]} migration(s)"
    any_ran=1

    for up in "${ups[@]}"; do
        echo "  up   -> $up"
        psql_exec "$up"
    done

    for ((i = ${#ups[@]} - 1; i >= 0; i--)); do
        down="${ups[$i]%.sql}.down.sql"
        echo "  down -> $down"
        psql_exec "$down"
    done

    for up in "${ups[@]}"; do
        echo "  up   -> $up (re-apply)"
        psql_exec "$up"
    done
done

if [ "$any_ran" -eq 0 ]; then
    echo "No schema has full up/down coverage yet -- nothing to round-trip."
    echo "This is expected until grandfathered migrations gain downgrades."
fi
