#!/usr/bin/env bash
# migrate.sh — Run golang-migrate for all services or a specific one.
#
# Usage:
#   ./migrate.sh [up|down] [service]
#
# Examples:
#   ./migrate.sh up              # Run all migrations up
#   ./migrate.sh up auth         # Run only auth migrations up
#   ./migrate.sh down user       # Roll back user migrations one step

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/../migrations"

PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-preuni}"
PGPASSWORD="${PGPASSWORD:-preuni}"
PGDB="${PGDB:-preuni}"

DIRECTION="${1:-up}"
TARGET_SERVICE="${2:-}"

SERVICES=(auth user content learning simulation dissertation)

run_migrate() {
    local svc="$1"
    local schema="$svc"
    # users schema uses 'users' not 'user'
    if [ "$svc" = "user" ]; then schema="users"; fi

    local db_url="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/${PGDB}?search_path=${schema}&sslmode=disable"

    echo "→ Migrating $svc ($DIRECTION)..."
    migrate \
        -path "$MIGRATIONS_DIR/$svc" \
        -database "$db_url" \
        "$DIRECTION"
    echo "  ✓ $svc done"
}

if [ -n "$TARGET_SERVICE" ]; then
    run_migrate "$TARGET_SERVICE"
else
    for svc in "${SERVICES[@]}"; do
        run_migrate "$svc"
    done
fi

echo "All migrations complete."
