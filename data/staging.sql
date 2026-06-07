-- 轻量 staging 库 — 用于多连接并存测试
PRAGMA foreign_keys = ON;

CREATE TABLE migrations (
    id          INTEGER PRIMARY KEY,
    version     TEXT    NOT NULL UNIQUE,
    applied_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE feature_flags (
    key         TEXT PRIMARY KEY,
    enabled     INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE audit_log (
    id         INTEGER PRIMARY KEY,
    action     TEXT    NOT NULL,
    actor      TEXT,
    payload    TEXT,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);

CREATE VIEW recent_audit AS
SELECT id, action, actor, created_at
FROM audit_log
ORDER BY created_at DESC
LIMIT 20;

INSERT INTO migrations (version, applied_at) VALUES
    ('202601010001_init',           '2026-01-01 00:01:00'),
    ('202601150002_users',          '2026-01-15 00:02:00'),
    ('202602010003_orders',         '2026-02-01 00:03:00'),
    ('202603010004_feature_flags',  '2026-03-01 00:04:00');

INSERT INTO feature_flags (key, enabled, description) VALUES
    ('dark_mode',        1, 'Enable dark theme toggle'),
    ('readonly_default', 0, 'Open new connections as read-only'),
    ('csv_export',       1, 'Allow CSV copy from result grid'),
    ('multi_connection', 1, 'Allow multiple open databases');

INSERT INTO audit_log (action, actor, payload, created_at) VALUES
    ('connection.open',  'dev', '{"file":"staging.sqlite3"}', '2026-06-01 09:00:00'),
    ('query.execute',    'dev', '{"sql":"SELECT * FROM feature_flags"}', '2026-06-01 09:05:00'),
    ('schema.inspect',   'dev', '{"table":"migrations"}', '2026-06-01 09:10:00'),
    ('connection.close', 'dev', '{"file":"staging.sqlite3"}', '2026-06-01 09:15:00');
