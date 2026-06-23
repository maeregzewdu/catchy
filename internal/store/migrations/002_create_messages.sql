CREATE TABLE messages (
    id          TEXT PRIMARY KEY,
    account_id  TEXT REFERENCES accounts(id) ON DELETE CASCADE,
    source      TEXT NOT NULL,
    message_id  TEXT,
    folder      TEXT NOT NULL,
    subject     TEXT,
    from_addr   TEXT NOT NULL,
    to_addrs    TEXT NOT NULL DEFAULT '[]',
    cc_addrs    TEXT NOT NULL DEFAULT '[]',
    bcc_addrs   TEXT NOT NULL DEFAULT '[]',
    body_text   TEXT,
    body_html   TEXT,
    raw_mime    BLOB,
    is_read     INTEGER NOT NULL DEFAULT 0,
    is_starred  INTEGER NOT NULL DEFAULT 0,
    sent_at     DATETIME,
    received_at DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_account_folder ON messages(account_id, folder);
CREATE INDEX idx_messages_source ON messages(source);
CREATE INDEX idx_messages_received_at ON messages(received_at DESC);
