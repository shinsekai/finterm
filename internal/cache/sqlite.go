package cache

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite" // Pure-Go SQLite driver; registers with database/sql
)

// SQLiteStore is a persistent cache backed by SQLite with per-key TTL.
type SQLiteStore struct {
	db        *sql.DB
	mu        sync.RWMutex
	stop      chan struct{}
	closeOnce sync.Once
}

// NewSQLite opens or creates a SQLite-backed cache at the given path.
// It runs embedded migrations and starts a background cleanup goroutine.
func NewSQLite(path string) (*SQLiteStore, error) {
	// Ensure parent directory exists with 0700 permissions.
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("creating cache directory: %w", err)
		}
		_ = os.Chmod(dir, 0o700)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}

	// Run migrations.
	if _, err := db.Exec(migration001); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	// Set file permissions to 0600 if the file exists.
	if _, err := os.Stat(path); err == nil {
		_ = os.Chmod(path, 0o600)
	}

	s := &SQLiteStore{
		db:   db,
		stop: make(chan struct{}),
	}
	go s.cleanup()
	return s, nil
}

// Close stops the background cleanup goroutine and closes the database.
func (s *SQLiteStore) Close() error {
	var err error
	s.closeOnce.Do(func() {
		close(s.stop)
		err = s.db.Close()
	})
	return err
}

// Get retrieves the value for the given key.
// Returns the value and true if the key exists and has not expired.
// Implements lazy expiration: expired entries are deleted on access.
func (s *SQLiteStore) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	var valueBlob []byte
	var expiresAt int64
	err := s.db.QueryRow("SELECT value, expires_at FROM cache_entries WHERE key = ?", key).Scan(&valueBlob, &expiresAt)
	s.mu.RUnlock()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		return nil, false
	}

	if time.Now().UnixNano() > expiresAt {
		// Lazy expiration.
		s.Delete(key)
		return nil, false
	}

	var val interface{}
	if err := gob.NewDecoder(bytes.NewReader(valueBlob)).Decode(&val); err != nil {
		return nil, false
	}
	return val, true
}

// Set stores the value with the given key and TTL.
func (s *SQLiteStore) Set(key string, value interface{}, ttl time.Duration) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&value); err != nil {
		return
	}

	now := time.Now()
	expiresAt := now.Add(ttl).UnixNano()
	cachedAt := now.UnixNano()

	s.mu.Lock()
	_, _ = s.db.Exec(
		`INSERT INTO cache_entries (key, value, expires_at, cached_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET
		   value = excluded.value,
		   expires_at = excluded.expires_at,
		   cached_at = excluded.cached_at`,
		key, buf.Bytes(), expiresAt, cachedAt,
	)
	s.mu.Unlock()
}

// Delete removes the entry with the given key from the cache.
func (s *SQLiteStore) Delete(key string) {
	s.mu.Lock()
	_, _ = s.db.Exec("DELETE FROM cache_entries WHERE key = ?", key)
	s.mu.Unlock()
}

// DeleteByPrefix removes all entries whose key starts with the given prefix.
func (s *SQLiteStore) DeleteByPrefix(prefix string) {
	s.mu.Lock()
	_, _ = s.db.Exec("DELETE FROM cache_entries WHERE key LIKE ?", prefix+"%")
	s.mu.Unlock()
}

// Flush removes all entries from the cache.
func (s *SQLiteStore) Flush() {
	s.mu.Lock()
	_, _ = s.db.Exec("DELETE FROM cache_entries")
	s.mu.Unlock()
}

// Len returns the number of non-expired entries in the cache.
func (s *SQLiteStore) Len() int {
	s.mu.RLock()
	var count int
	now := time.Now().UnixNano()
	err := s.db.QueryRow("SELECT COUNT(*) FROM cache_entries WHERE expires_at > ?", now).Scan(&count)
	s.mu.RUnlock()
	if err != nil {
		return 0
	}
	return count
}

// GetMetadata returns metadata about a cached entry without retrieving the value.
func (s *SQLiteStore) GetMetadata(key string) (time.Time, bool) {
	s.mu.RLock()
	var cachedAt int64
	var expiresAt int64
	err := s.db.QueryRow("SELECT cached_at, expires_at FROM cache_entries WHERE key = ?", key).Scan(&cachedAt, &expiresAt)
	s.mu.RUnlock()

	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, false
		}
		return time.Time{}, false
	}

	if time.Now().UnixNano() > expiresAt {
		s.Delete(key)
		return time.Time{}, false
	}

	return time.Unix(0, cachedAt), true
}

// IsStale checks if a cached entry is older than the given staleness threshold.
func (s *SQLiteStore) IsStale(key string, stalenessThreshold time.Duration) bool {
	cachedAt, exists := s.GetMetadata(key)
	if !exists {
		return true
	}
	return time.Since(cachedAt) > stalenessThreshold
}

// GetWithMetadata returns both the cached value and its metadata.
func (s *SQLiteStore) GetWithMetadata(key string) (interface{}, time.Time, bool) {
	s.mu.RLock()
	var valueBlob []byte
	var cachedAt int64
	var expiresAt int64
	err := s.db.QueryRow("SELECT value, cached_at, expires_at FROM cache_entries WHERE key = ?", key).
		Scan(&valueBlob, &cachedAt, &expiresAt)
	s.mu.RUnlock()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, time.Time{}, false
		}
		return nil, time.Time{}, false
	}

	if time.Now().UnixNano() > expiresAt {
		s.Delete(key)
		return nil, time.Time{}, false
	}

	var val interface{}
	if err := gob.NewDecoder(bytes.NewReader(valueBlob)).Decode(&val); err != nil {
		return nil, time.Time{}, false
	}
	return val, time.Unix(0, cachedAt), true
}

// cleanup runs in a background goroutine and removes expired entries periodically.
func (s *SQLiteStore) cleanup() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeExpired()
		case <-s.stop:
			return
		}
	}
}

// removeExpired deletes all entries whose expires_at is in the past.
func (s *SQLiteStore) removeExpired() {
	s.mu.Lock()
	_, _ = s.db.Exec("DELETE FROM cache_entries WHERE expires_at <= ?", time.Now().UnixNano())
	s.mu.Unlock()
}
