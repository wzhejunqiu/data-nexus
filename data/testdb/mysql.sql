-- Test database seed for local MySQL (make test-db-up)
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS post_tags;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;
DROP VIEW IF EXISTS active_users;
SET FOREIGN_KEY_CHECKS = 1;

CREATE TABLE users (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE,
    name       VARCHAR(255) NOT NULL,
    role       VARCHAR(64)  NOT NULL DEFAULT 'user',
    avatar     BLOB,
    score      DOUBLE DEFAULT 0.0,
    is_active  TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id           INT AUTO_INCREMENT PRIMARY KEY,
    user_id      INT NOT NULL,
    title        VARCHAR(512) NOT NULL,
    body         TEXT,
    published    TINYINT(1) NOT NULL DEFAULT 0,
    view_count   INT NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE orders (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    user_id    INT NOT NULL,
    status     VARCHAR(64) NOT NULL DEFAULT 'pending',
    total      DOUBLE NOT NULL DEFAULT 0.0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE order_items (
    id           INT AUTO_INCREMENT PRIMARY KEY,
    order_id     INT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity     INT NOT NULL DEFAULT 1,
    unit_price   DOUBLE NOT NULL DEFAULT 0.0,
    FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE TABLE tags (
    id   INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE
);

CREATE TABLE post_tags (
    post_id INT NOT NULL,
    tag_id  INT NOT NULL,
    PRIMARY KEY (post_id, tag_id),
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
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
WHERE u.is_active = 1
GROUP BY u.id, u.email, u.name, u.role;

INSERT INTO users (email, name, role, score, is_active, created_at) VALUES
    ('alice@example.com', 'Alice Chen', 'admin', 98.5, 1, '2025-01-02 08:00:00'),
    ('bob@example.com',   'Bob Wang',   'user',  72.0, 1, '2025-01-03 09:15:00'),
    ('carol@example.com', 'Carol Li',   'user',  65.3, 1, '2025-01-04 10:30:00'),
    ('dave@example.com',  'Dave Zhang', 'editor', 81.2, 1, '2025-01-05 11:45:00'),
    ('eve@example.com',   'Eve Liu',    'user',  55.0, 0, '2025-01-06 12:00:00');

INSERT INTO users (email, name, role, score, is_active, created_at)
WITH RECURSIVE nums AS (
    SELECT 6 AS n
    UNION ALL
    SELECT n + 1 FROM nums WHERE n < 60
)
SELECT
    CONCAT('user', n, '@example.com'),
    CONCAT('User ', n),
    IF(n % 10 = 0, 'editor', 'user'),
    ROUND(20 + MOD(n * 137, 800) / 10, 1),
    IF(n % 7 = 0, 0, 1),
    DATE_ADD('2025-02-01', INTERVAL n HOUR)
FROM nums;

INSERT INTO tags (name) VALUES
    ('postgres'), ('mysql'), ('go'), ('react'), ('wails'), ('database');

INSERT INTO posts (user_id, title, body, published, view_count, created_at)
WITH RECURSIVE nums AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM nums WHERE n < 120
)
SELECT
    MOD(n - 1, 60) + 1,
    CONCAT('Post #', n, ': Sample title'),
    CONCAT('Sample body for post ', n, '. Used for browse, sort, and SQL testing.'),
    IF(n % 3 = 0, 0, 1),
    MOD(n * 17, 500),
    DATE_ADD('2025-03-01', INTERVAL n HOUR)
FROM nums;

INSERT INTO post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM posts p
JOIN tags t ON t.id = MOD(p.id - 1, 6) + 1
WHERE p.id <= 40;

INSERT INTO orders (user_id, status, total, created_at) VALUES
    (1, 'paid',    129.90, '2025-04-01 10:00:00'),
    (2, 'pending',  59.50, '2025-04-03 12:00:00'),
    (3, 'paid',    399.00, '2025-04-05 14:00:00');

INSERT INTO order_items (order_id, product_name, quantity, unit_price) VALUES
    (1, 'Data Nexus Pro License', 1, 129.90),
    (2, 'Sticker Pack',           5,  11.90),
    (3, 'Annual Subscription',    1, 399.00);
