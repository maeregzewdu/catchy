CREATE TABLE attachments (
    id         TEXT PRIMARY KEY,
    message_id TEXT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    filename   TEXT NOT NULL,
    mime_type  TEXT NOT NULL,
    size       INTEGER NOT NULL,
    path       TEXT NOT NULL
);

CREATE INDEX idx_attachments_message_id ON attachments(message_id);
