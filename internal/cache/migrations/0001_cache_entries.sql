CREATE TABLE IF NOT EXISTS cache_entries (
    key        TEXT PRIMARY KEY,
    value      BLOB NOT NULL,
    expires_at INTEGER NOT NULL,  -- unix nanoseconds
    cached_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_cache_expires ON cache_entries(expires_at);
