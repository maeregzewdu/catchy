CREATE TABLE accounts (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL UNIQUE,
    smtp_host  TEXT NOT NULL,
    smtp_port  INTEGER NOT NULL,
    imap_host  TEXT NOT NULL,
    imap_port  INTEGER NOT NULL,
    username   TEXT NOT NULL,
    password   TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
