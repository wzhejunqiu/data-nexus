#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${1:?compose file required}"
shift

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

echo "Seeding PostgreSQL (users, posts, orders, tags, active_users view)..."
"$@" -f "$COMPOSE_FILE" exec -T postgres \
	psql -v ON_ERROR_STOP=1 -U test -d testdb \
	< "$ROOT/data/testdb/postgres.sql"

echo "Seeding MySQL (users, posts, orders, tags, active_users view)..."
"$@" -f "$COMPOSE_FILE" exec -T mysql \
	mysql -utest -ptest testdb \
	< "$ROOT/data/testdb/mysql.sql"

echo "Test data seeded in testdb on both databases."
