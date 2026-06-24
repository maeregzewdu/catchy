-- Full-text search index over all messages.
-- Uses content=messages so only token data is stored, not duplicate text.

CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    subject,
    from_addr,
    to_addrs,
    body_text,
    content=messages,
    content_rowid=rowid
);

-- Populate FTS index from existing messages.
INSERT INTO messages_fts(messages_fts) VALUES('rebuild');

CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
  INSERT INTO messages_fts(rowid, subject, from_addr, to_addrs, body_text)
  VALUES (new.rowid, COALESCE(new.subject,''), COALESCE(new.from_addr,''),
          COALESCE(new.to_addrs,''), COALESCE(new.body_text,''));
END;

CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
  INSERT INTO messages_fts(messages_fts, rowid, subject, from_addr, to_addrs, body_text)
  VALUES ('delete', old.rowid, COALESCE(old.subject,''), COALESCE(old.from_addr,''),
          COALESCE(old.to_addrs,''), COALESCE(old.body_text,''));
END;

CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
  INSERT INTO messages_fts(messages_fts, rowid, subject, from_addr, to_addrs, body_text)
  VALUES ('delete', old.rowid, COALESCE(old.subject,''), COALESCE(old.from_addr,''),
          COALESCE(old.to_addrs,''), COALESCE(old.body_text,''));
  INSERT INTO messages_fts(rowid, subject, from_addr, to_addrs, body_text)
  VALUES (new.rowid, COALESCE(new.subject,''), COALESCE(new.from_addr,''),
          COALESCE(new.to_addrs,''), COALESCE(new.body_text,''));
END;
