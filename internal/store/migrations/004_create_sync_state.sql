CREATE TABLE sync_state (
    account_id   TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    folder       TEXT NOT NULL,
    uid_next     INTEGER NOT NULL DEFAULT 0,
    uid_validity INTEGER NOT NULL DEFAULT 0,
    last_sync    DATETIME,
    PRIMARY KEY (account_id, folder)
);
