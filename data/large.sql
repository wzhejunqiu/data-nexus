-- large 库 — 10 万行性能测试（手动验收：make dev -- --db data/large.db）
PRAGMA foreign_keys = ON;

CREATE TABLE large_rows (
    id      INTEGER PRIMARY KEY,
    payload TEXT NOT NULL
);

INSERT INTO large_rows (payload)
WITH RECURSIVE cnt(x) AS (
    SELECT 1
    UNION ALL
    SELECT x + 1 FROM cnt WHERE x < 100000
)
SELECT 'row-' || x FROM cnt;

CREATE INDEX idx_large_rows_payload ON large_rows(payload);
