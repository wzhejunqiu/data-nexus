-- analytics 库 — 事件统计场景，用于第二个/第三个连接测试
PRAGMA foreign_keys = ON;

CREATE TABLE events (
    id          INTEGER PRIMARY KEY,
    event_name  TEXT    NOT NULL,
    user_id     INTEGER,
    session_id  TEXT,
    properties  TEXT,
    occurred_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE daily_metrics (
    day             TEXT PRIMARY KEY,
    page_views      INTEGER NOT NULL DEFAULT 0,
    unique_users    INTEGER NOT NULL DEFAULT 0,
    avg_session_sec REAL    NOT NULL DEFAULT 0.0
);

CREATE INDEX idx_events_name ON events(event_name);
CREATE INDEX idx_events_occurred_at ON events(occurred_at);

CREATE VIEW top_events AS
SELECT event_name, COUNT(*) AS event_count
FROM events
GROUP BY event_name
ORDER BY event_count DESC;

INSERT INTO daily_metrics (day, page_views, unique_users, avg_session_sec) VALUES
    ('2026-06-01', 1204, 312, 186.5),
    ('2026-06-02', 1388, 341, 192.0),
    ('2026-06-03', 1102, 298, 174.3),
    ('2026-06-04', 1560, 402, 201.7),
    ('2026-06-05', 1499, 389, 198.2),
    ('2026-06-06', 980,  256, 165.8),
    ('2026-06-07', 1675, 421, 210.4);

INSERT INTO events (event_name, user_id, session_id, properties, occurred_at)
SELECT
    CASE (n % 5)
        WHEN 0 THEN 'page_view'
        WHEN 1 THEN 'table_browse'
        WHEN 2 THEN 'sql_execute'
        WHEN 3 THEN 'connection_open'
        ELSE 'export_csv'
    END,
    (n % 40) + 1,
    'sess_' || ((n % 12) + 1),
    '{"source":"demo","index":' || n || '}',
    datetime('2026-06-07 08:00:00', '+' || n || ' minutes')
FROM (
    SELECT value + 1 AS n
    FROM generate_series(0, 79)
);
