#!/usr/bin/env bash
# check-migration-downgrades.sh — enforces Constitution VI's migration
# reversibility rule for infra/migrations/ (the Go-owned schemas).
#
# infra/migrations/*.sql are plain numbered files applied in order by a
# bash loop (backend-ci.yml), not a migration tool with native up/down
# support like Alembic. The convention adopted in v2.3.0: any migration
# NNN_name.sql ships a sibling NNN_name.down.sql with the reverse DDL.
#
# Pre-existing files are grandfathered via infra/migrations/.grandfathered
# (frozen at v2.3.0's ratification) — not retroactively blocked. Editing
# one of them requires removing it from that list and adding its downgrade
# in the same PR, which is exactly what makes this check start requiring
# one for that file.
#
# Usage: ./check-migration-downgrades.sh (run from repo root or anywhere)

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

MIGRATIONS_DIR="infra/migrations"
GRANDFATHER_LIST="$MIGRATIONS_DIR/.grandfathered"

fail=0

while IFS= read -r -d '' up_file; do
    rel="${up_file#"$MIGRATIONS_DIR"/}"

    # Skip down files themselves and the standalone grant script (not a
    # per-schema numbered migration).
    case "$rel" in
        *.down.sql) continue ;;
        grant-*.sql) continue ;;
    esac

    if grep -qxF "$rel" "$GRANDFATHER_LIST" 2>/dev/null; then
        continue
    fi

    down_file="${up_file%.sql}.down.sql"
    if [ ! -f "$down_file" ]; then
        echo "FAIL: $rel has no matching $(basename "$down_file")"
        echo "  Either add the downgrade, or if this is a pre-existing"
        echo "  migration you're not touching, it should already be in"
        echo "  $GRANDFATHER_LIST — something is inconsistent."
        fail=1
    fi
done < <(find "$MIGRATIONS_DIR" -name '*.sql' -print0 | sort -z)

if [ "$fail" -ne 0 ]; then
    echo
    echo "FAIL: one or more migrations are missing a downgrade (Constitution VI)."
    exit 1
fi

echo "OK: every non-grandfathered migration has a downgrade."
