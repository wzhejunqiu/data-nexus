package sqlite

import (
	"fmt"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

const schemaVersion = 1

const schemaSQL = `
CREATE TABLE IF NOT EXISTS connections (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    driver_type      TEXT NOT NULL,
    config_json      TEXT NOT NULL,
    secrets_backend  TEXT,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    last_used_at     TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_connections_last_used ON connections (last_used_at DESC);

CREATE TABLE IF NOT EXISTS connection_groups (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    parent_id  TEXT REFERENCES connection_groups(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_connection_groups_parent ON connection_groups (parent_id);
CREATE INDEX IF NOT EXISTS idx_connection_groups_sort ON connection_groups (parent_id, sort_order);

CREATE TABLE IF NOT EXISTS group_members (
    group_id     TEXT NOT NULL REFERENCES connection_groups(id) ON DELETE CASCADE,
    member_type  TEXT NOT NULL,
    member_id    TEXT NOT NULL,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, member_type, member_id)
);

CREATE INDEX IF NOT EXISTS idx_group_members_sort ON group_members (group_id, sort_order);

CREATE TABLE IF NOT EXISTS free_connections (
    connection_id TEXT PRIMARY KEY REFERENCES connections(id) ON DELETE CASCADE,
    sort_order    INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_free_connections_sort ON free_connections (sort_order);

CREATE TABLE IF NOT EXISTS sidebar_root_items (
    item_type  TEXT NOT NULL,
    item_id    TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    PRIMARY KEY (item_type, item_id)
);

CREATE INDEX IF NOT EXISTS idx_sidebar_root_sort ON sidebar_root_items (sort_order);

CREATE TABLE IF NOT EXISTS app_connection_state (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
`

func migrateDB(db *Store) error {
	var version int
	if err := db.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return model.ErrInternal(err.Error())
	}
	if version >= schemaVersion {
		return nil
	}
	if _, err := db.db.Exec(schemaSQL); err != nil {
		return model.ErrInternal(err.Error())
	}
	if err := db.ensureDefaultGroup(); err != nil {
		return err
	}
	if _, err := db.db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, schemaVersion)); err != nil {
		return model.ErrInternal(err.Error())
	}
	return nil
}
