-- Data Nexus demo database — 用于手动测试表结构、分页、SQL、索引、视图等
PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id         INTEGER PRIMARY KEY,
    email      TEXT    NOT NULL UNIQUE,
    name       TEXT    NOT NULL,
    role       TEXT    NOT NULL DEFAULT 'user',
    avatar     BLOB,
    score      REAL    DEFAULT 0.0,
    is_active  INTEGER NOT NULL DEFAULT 1,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE posts (
    id           INTEGER PRIMARY KEY,
    user_id      INTEGER NOT NULL REFERENCES users(id),
    title        TEXT    NOT NULL,
    body         TEXT,
    published    INTEGER NOT NULL DEFAULT 0,
    view_count   INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE orders (
    id         INTEGER PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    status     TEXT    NOT NULL DEFAULT 'pending',
    total      REAL    NOT NULL DEFAULT 0.0,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE order_items (
    id           INTEGER PRIMARY KEY,
    order_id     INTEGER NOT NULL REFERENCES orders(id),
    product_name TEXT    NOT NULL,
    quantity     INTEGER NOT NULL DEFAULT 1,
    unit_price   REAL    NOT NULL DEFAULT 0.0
);

CREATE TABLE tags (
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE post_tags (
    post_id INTEGER NOT NULL REFERENCES posts(id),
    tag_id  INTEGER NOT NULL REFERENCES tags(id),
    PRIMARY KEY (post_id, tag_id)
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_published ON posts(published);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

CREATE VIEW active_users AS
SELECT
    u.id,
    u.email,
    u.name,
    u.role,
    COUNT(p.id) AS post_count
FROM users u
LEFT JOIN posts p ON p.user_id = u.id
WHERE u.is_active = 1
GROUP BY u.id, u.email, u.name, u.role;

INSERT INTO users (email, name, role, avatar, score, is_active, created_at) VALUES
    ('alice@example.com',   'Alice Chen',   'admin', X'89504E470D0A1A0A', 98.5, 1, '2025-01-02 08:00:00'),
    ('bob@example.com',     'Bob Wang',     'user',  NULL,                  72.0, 1, '2025-01-03 09:15:00'),
    ('carol@example.com',   'Carol Li',     'user',  NULL,                  65.3, 1, '2025-01-04 10:30:00'),
    ('dave@example.com',    'Dave Zhang',   'editor',X'FFD8FFE000104A464946', 81.2, 1, '2025-01-05 11:45:00'),
    ('eve@example.com',     'Eve Liu',      'user',  NULL,                  55.0, 0, '2025-01-06 12:00:00'),
    ('frank@example.com',   'Frank Wu',     'user',  NULL,                  44.8, 1, '2025-01-07 13:20:00'),
    ('grace@example.com',   'Grace Xu',     'user',  NULL,                  88.1, 1, '2025-01-08 14:35:00'),
    ('henry@example.com',   'Henry Sun',    'user',  NULL,                  33.6, 1, '2025-01-09 15:50:00'),
    ('iris@example.com',    'Iris Zhao',    'editor',NULL,                  77.4, 1, '2025-01-10 16:05:00'),
    ('jack@example.com',    'Jack Ma',      'user',  NULL,                  91.0, 1, '2025-01-11 17:20:00'),
    ('kate@example.com',    'Kate Lin',     'user',  NULL,                  62.7, 1, '2025-01-12 18:35:00'),
    ('leo@example.com',     'Leo Yang',     'user',  NULL,                  58.9, 1, '2025-01-13 19:50:00'),
    ('mia@example.com',     'Mia Hu',       'user',  NULL,                  47.2, 0, '2025-01-14 20:05:00'),
    ('noah@example.com',    'Noah Gao',     'user',  NULL,                  69.5, 1, '2025-01-15 21:20:00'),
    ('olivia@example.com',  'Olivia Tang',  'user',  NULL,                  74.3, 1, '2025-01-16 22:35:00'),
    ('paul@example.com',    'Paul Feng',    'user',  NULL,                  39.1, 1, '2025-01-17 23:50:00'),
    ('quinn@example.com',   'Quinn He',     'user',  NULL,                  83.6, 1, '2025-01-18 08:05:00'),
    ('rose@example.com',    'Rose Jin',     'user',  NULL,                  56.4, 1, '2025-01-19 09:20:00'),
    ('sam@example.com',     'Sam Cai',      'user',  NULL,                  71.8, 1, '2025-01-20 10:35:00'),
    ('tina@example.com',    'Tina Song',    'user',  NULL,                  64.0, 1, '2025-01-21 11:50:00');

-- 批量生成更多用户（用于分页测试，共 60 行）
INSERT INTO users (email, name, role, score, is_active, created_at)
SELECT
    'user' || n || '@example.com',
    'User ' || n,
    CASE WHEN n % 10 = 0 THEN 'editor' ELSE 'user' END,
    ROUND(20 + (n * 1.37) % 80, 1),
    CASE WHEN n % 7 = 0 THEN 0 ELSE 1 END,
    datetime('2025-02-01', '+' || n || ' hours')
FROM (
    SELECT 21 + value AS n
    FROM generate_series(0, 39)
);

INSERT INTO tags (name) VALUES
    ('sqlite'),
    ('go'),
    ('react'),
    ('wails'),
    ('database'),
    ('tutorial'),
    ('release'),
    ('performance');

INSERT INTO posts (user_id, title, body, published, view_count, created_at)
SELECT
    ((n - 1) % 60) + 1,
    'Post #' || n || ': Exploring SQLite feature ' || ((n - 1) % 8 + 1),
    'Sample body for post ' || n || '. Used for browse, sort, and SQL testing.',
    CASE WHEN n % 3 = 0 THEN 0 ELSE 1 END,
    (n * 17) % 500,
    datetime('2025-03-01', '+' || n || ' hours')
FROM (
    SELECT value + 1 AS n
    FROM generate_series(0, 119)
);

INSERT INTO post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM posts p
JOIN tags t ON t.id = ((p.id - 1) % 8) + 1
WHERE p.id <= 40;

INSERT INTO orders (user_id, status, total, created_at) VALUES
    (1,  'paid',      129.90, '2025-04-01 10:00:00'),
    (1,  'shipped',   249.00, '2025-04-02 11:00:00'),
    (2,  'pending',    59.50, '2025-04-03 12:00:00'),
    (3,  'cancelled',  19.99, '2025-04-04 13:00:00'),
    (4,  'paid',      399.00, '2025-04-05 14:00:00'),
    (5,  'refunded',   88.00, '2025-04-06 15:00:00'),
    (6,  'paid',      156.75, '2025-04-07 16:00:00'),
    (7,  'shipped',   210.20, '2025-04-08 17:00:00'),
    (8,  'pending',    45.00, '2025-04-09 18:00:00'),
    (9,  'paid',      320.00, '2025-04-10 19:00:00'),
    (10, 'paid',       99.99, '2025-04-11 20:00:00'),
    (11, 'shipped',   175.50, '2025-04-12 21:00:00'),
    (12, 'pending',    33.33, '2025-04-13 22:00:00');

INSERT INTO order_items (order_id, product_name, quantity, unit_price) VALUES
    (1,  'Data Nexus Pro License', 1, 129.90),
    (2,  'SQLite Book',            2,  49.00),
    (2,  'Go Mug',                 1, 151.00),
    (3,  'Sticker Pack',           5,  11.90),
    (4,  'Trial Plan',             1,  19.99),
    (5,  'Annual Subscription',    1, 399.00),
    (6,  'Support Package',        1,  88.00),
    (7,  'Team Seat',              3,  52.25),
    (8,  'Workshop Ticket',        2, 105.10),
    (9,  'Merch Bundle',           1,  45.00),
    (10, 'Enterprise Plan',        1, 320.00),
    (11, 'Addon Pack',             1,  99.99),
    (12, 'Training Session',       1, 175.50),
    (13, 'Debug Kit',              1,  33.33);
