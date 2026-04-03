// Package cache provides a generic, thread-safe, in-memory cache with per-key TTL.
package cache

import (
	"strings"
	"sync"
	"time"
)

// Store represents a thread-safe in-memory cache with per-key TTL support.
type Store struct {
	mu    sync.RWMutex
	items map[string]*cacheEntry
	stop  chan struct{}
}

// cacheEntry holds a cached value with its expiration timestamp.
type cacheEntry struct {
	value      interface{}
	expiryTime time.Time
}

// New creates and returns a new cache Store with a background cleanup goroutine.
// The cleanup goroutine runs every 60 seconds and removes expired entries.
// Stop the cleanup goroutine by calling Close().
func New() *Store {
	s := &Store{
		items: make(map[string]*cacheEntry),
		stop:  make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Close stops the background cleanup goroutine.
func (s *Store) Close() {
	close(s.stop)
}

// Get retrieves the value for the given key.
// Returns the value and true if the key exists and has not expired.
// Returns nil and false if the key does not exist or has expired (lazy expiration).
func (s *Store) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	entry, exists := s.items[key]
	s.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiryTime) {
		// Lazy expiration: remove expired entry on access
		s.Delete(key)
		return nil, false
	}

	return entry.value, true
}

// Set stores the value with the given key and TTL.
// The value will expire after the specified TTL has elapsed.
func (s *Store) Set(key string, value interface{}, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = &cacheEntry{
		value:      value,
		expiryTime: time.Now().Add(ttl),
	}
}

// Delete removes the entry with the given key from the cache.
// It is safe to call Delete on a non-existent key.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
}

// Flush removes all entries from the cache.
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string]*cacheEntry)
}

// Len returns the number of non-expired entries in the cache.
// Expired entries are not counted.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, entry := range s.items {
		if now.Before(entry.expiryTime) {
			count++
		}
	}
	return count
}

// cleanup runs in a background goroutine and removes expired entries periodically.
// It runs every 60 seconds until Close() is called.
func (s *Store) cleanup() {
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

// removeExpired iterates through all entries and removes those that have expired.
// This is called periodically by the cleanup goroutine.
func (s *Store) removeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.items {
		if now.After(entry.expiryTime) {
			delete(s.items, key)
		}
	}
}

// Key builds a consistent, collision-free cache key from the provided parts.
// Parts are joined with a colon delimiter.
// Example: Key("rsi", "AAPL", "14", "daily") → "rsi:AAPL:14:daily"
func Key(parts ...string) string {
	return strings.Join(parts, ":")
}
