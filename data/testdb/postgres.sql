-- Test database seed for manual PostgreSQL acceptance (import into a real instance)
DROP TABLE IF EXISTS post_tags CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP VIEW IF EXISTS active_users;

CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    email      TEXT    NOT NULL UNIQUE,
    name       TEXT    NOT NULL,
    role       TEXT    NOT NULL DEFAULT 'user',
    avatar     BYTEA,
    score      DOUBLE PRECISION DEFAULT 0.0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE posts (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER NOT NULL REFERENCES users(id),
    title        TEXT    NOT NULL,
    body         TEXT,
    published    BOOLEAN NOT NULL DEFAULT FALSE,
    view_count   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    status     TEXT    NOT NULL DEFAULT 'pending',
    total      DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id           SERIAL PRIMARY KEY,
    order_id     INTEGER NOT NULL REFERENCES orders(id),
    product_name TEXT    NOT NULL,
    quantity     INTEGER NOT NULL DEFAULT 1,
    unit_price   DOUBLE PRECISION NOT NULL DEFAULT 0.0
);

CREATE TABLE tags (
    id   SERIAL PRIMARY KEY,
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

CREATE VIEW active_users AS
SELECT
    u.id,
    u.email,
    u.name,
    u.role,
    COUNT(p.id) AS post_count
FROM users u
LEFT JOIN posts p ON p.user_id = u.id
WHERE u.is_active = TRUE
GROUP BY u.id, u.email, u.name, u.role;

INSERT INTO users (email, name, role, score, is_active, created_at) VALUES
    ('alice@example.com', 'Alice Chen', 'admin', 98.5, TRUE, '2025-01-02 08:00:00'),
    ('bob@example.com',   'Bob Wang',   'user',  72.0, TRUE, '2025-01-03 09:15:00'),
    ('carol@example.com', 'Carol Li',   'user',  65.3, TRUE, '2025-01-04 10:30:00'),
    ('dave@example.com',  'Dave Zhang', 'editor', 81.2, TRUE, '2025-01-05 11:45:00'),
    ('eve@example.com',   'Eve Liu',    'user',  55.0, FALSE, '2025-01-06 12:00:00');

INSERT INTO users (email, name, role, score, is_active, created_at)
SELECT
    'user' || n || '@example.com',
    'User ' || n,
    CASE WHEN n % 10 = 0 THEN 'editor' ELSE 'user' END,
    ROUND((20 + (n * 1.37) % 80)::numeric, 1),
    CASE WHEN n % 7 = 0 THEN FALSE ELSE TRUE END,
    TIMESTAMP '2025-02-01' + (n || ' hours')::interval
FROM generate_series(6, 60) AS n;

INSERT INTO tags (name) VALUES
    ('postgres'), ('mysql'), ('go'), ('react'), ('wails'), ('database');

INSERT INTO posts (user_id, title, body, published, view_count, created_at)
SELECT
    ((n - 1) % 60) + 1,
    'Post #' || n || ': Sample title',
    'Sample body for post ' || n || '. Used for browse, sort, and SQL testing.',
    CASE WHEN n % 3 = 0 THEN FALSE ELSE TRUE END,
    (n * 17) % 500,
    TIMESTAMP '2025-03-01' + (n || ' hours')::interval
FROM generate_series(1, 120) AS n;

INSERT INTO post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM posts p
JOIN tags t ON t.id = ((p.id - 1) % 6) + 1
WHERE p.id <= 40;

INSERT INTO orders (user_id, status, total, created_at) VALUES
    (1, 'paid',    129.90, '2025-04-01 10:00:00'),
    (2, 'pending',  59.50, '2025-04-03 12:00:00'),
    (3, 'paid',    399.00, '2025-04-05 14:00:00');

INSERT INTO order_items (order_id, product_name, quantity, unit_price) VALUES
    (1, 'Data Nexus Pro License', 1, 129.90),
    (2, 'Sticker Pack',           5,  11.90),
    (3, 'Annual Subscription',    1, 399.00);
