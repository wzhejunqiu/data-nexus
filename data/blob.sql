-- BLOB 测试库 — 手动验收 BLOB 列展示、禁用编辑、不同字节数
PRAGMA foreign_keys = ON;

CREATE TABLE assets (
    id        INTEGER PRIMARY KEY,
    name      TEXT NOT NULL,
    mime_type TEXT,
    payload   BLOB,
    thumb     BLOB,
    note      TEXT
);

INSERT INTO assets (name, mime_type, payload, thumb, note) VALUES
    ('icon.png',  'image/png',  X'89504E470D0A1A0A',           X'89504E47',        '小 PNG 头'),
    ('photo.jpg', 'image/jpeg', X'FFD8FFE000104A464946000101',  NULL,               'JPEG + NULL thumb'),
    ('empty.bin', 'application/octet-stream', X'',             NULL,               '零长度 BLOB'),
    ('doc.pdf',   'application/pdf', X'255044462D312E34',       NULL,               'PDF 头 %PDF-1.4'),
    ('no-file',   'text/plain',   NULL,                          NULL,               'payload 为 NULL');

CREATE TABLE raw_chunks (
    id   INTEGER PRIMARY KEY,
    tag  TEXT NOT NULL,
    data BLOB NOT NULL
);

INSERT INTO raw_chunks (tag, data) VALUES
    ('tiny',   X'00112233445566778899'),
    ('small',  randomblob(256)),
    ('medium', randomblob(4096)),
    ('large',  randomblob(65536));
