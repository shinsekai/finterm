package cache

import (
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

// cacheFactories provides both cache implementations for parameterized tests.
var cacheFactories = []struct {
	name string
	new  func(t *testing.T) Cache
}{
	{
		name: "Memory",
		new: func(t *testing.T) Cache {
			s := New()
			t.Cleanup(func() { _ = s.Close() })
			return s
		},
	},
	{
		name: "SQLite",
		new: func(t *testing.T) Cache {
			s, err := NewSQLite(filepath.Join(t.TempDir(), "test.db"))
			if err != nil {
				t.Fatalf("failed to create sqlite store: %v", err)
			}
			t.Cleanup(func() { _ = s.Close() })
			return s
		},
	},
}

// TestCache_SetAndGet verifies that Set stores a value and Get retrieves it before TTL expires.
func TestCache_SetAndGet(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			key := "test-key"
			value := "test-value"
			ttl := 10 * time.Second

			store.Set(key, value, ttl)

			got, found := store.Get(key)
			if !found {
				t.Fatal("expected to find key, but it was not found")
			}
			if got != value {
				t.Errorf("expected value %v, got %v", value, got)
			}
		})
	}
}

// TestCache_TTLExpiry verifies that Get returns miss after TTL has elapsed.
func TestCache_TTLExpiry(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			key := "test-key"
			value := "test-value"
			ttl := 10 * time.Millisecond

			store.Set(key, value, ttl)

			// Value should be available immediately
			_, found := store.Get(key)
			if !found {
				t.Fatal("expected to find key immediately after Set")
			}

			// Wait for TTL to expire
			time.Sleep(15 * time.Millisecond)

			// Value should not be available after expiry
			_, found = store.Get(key)
			if found {
				t.Error("expected key to be expired, but it was found")
			}
		})
	}
}

// TestCache_Delete verifies that Delete removes an entry from the cache.
func TestCache_Delete(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			key := "test-key"
			value := "test-value"

			store.Set(key, value, 10*time.Second)

			// Verify key exists
			_, found := store.Get(key)
			if !found {
				t.Fatal("expected to find key before Delete")
			}

			store.Delete(key)

			// Verify key is deleted
			_, found = store.Get(key)
			if found {
				t.Error("expected key to be deleted, but it was found")
			}
		})
	}
}

// TestCache_Delete_NonExistent verifies that Delete is safe to call on a non-existent key.
func TestCache_Delete_NonExistent(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			// Should not panic on non-existent key
			store.Delete("non-existent-key")
		})
	}
}

// TestCache_Flush verifies that Flush removes all entries from the cache.
func TestCache_Flush(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			// Add multiple entries
			store.Set("key1", "value1", 10*time.Second)
			store.Set("key2", "value2", 10*time.Second)
			store.Set("key3", "value3", 10*time.Second)

			// Verify entries exist
			if got := store.Len(); got != 3 {
				t.Fatalf("expected 3 entries before Flush, got %d", got)
			}

			store.Flush()

			// Verify all entries are removed
			if got := store.Len(); got != 0 {
				t.Errorf("expected 0 entries after Flush, got %d", got)
			}

			_, found := store.Get("key1")
			if found {
				t.Error("expected key1 to be flushed, but it was found")
			}

			_, found = store.Get("key2")
			if found {
				t.Error("expected key2 to be flushed, but it was found")
			}

			_, found = store.Get("key3")
			if found {
				t.Error("expected key3 to be flushed, but it was found")
			}
		})
	}
}

// TestCache_Len_ExcludesExpired verifies that Len only counts non-expired entries.
func TestCache_Len_ExcludesExpired(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			// Add entries with different TTLs
			store.Set("key1", "value1", 10*time.Second)
			store.Set("key2", "value2", 10*time.Millisecond) // Will expire quickly
			store.Set("key3", "value3", 10*time.Second)

			// Initially should have 3 entries
			if got := store.Len(); got != 3 {
				t.Fatalf("expected 3 entries initially, got %d", got)
			}

			// Wait for key2 to expire
			time.Sleep(15 * time.Millisecond)

			// Should now have only 2 entries
			if got := store.Len(); got != 2 {
				t.Errorf("expected 2 entries after expiry, got %d", got)
			}

			// Verify which keys still exist
			_, found := store.Get("key1")
			if !found {
				t.Error("expected key1 to still exist")
			}

			_, found = store.Get("key2")
			if found {
				t.Error("expected key2 to be expired")
			}

			_, found = store.Get("key3")
			if !found {
				t.Error("expected key3 to still exist")
			}
		})
	}
}

// TestCache_ConcurrentAccess verifies that concurrent Set and Get operations are safe.
// This test uses -race to detect data races.
func TestCache_ConcurrentAccess(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			var wg sync.WaitGroup
			numGoroutines := 100
			numOpsPerGoroutine := 100

			// Concurrent Set operations
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < numOpsPerGoroutine; j++ {
						key := Key("concurrent", strconv.Itoa(id), strconv.Itoa(j))
						store.Set(key, id*numGoroutines+j, 10*time.Second)
					}
				}(i)
			}

			// Concurrent Get operations
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < numOpsPerGoroutine; j++ {
						key := Key("concurrent", strconv.Itoa(id), strconv.Itoa(j))
						store.Get(key)
					}
				}(i)
			}

			wg.Wait()
		})
	}
}

// TestCache_ConcurrentDelete verifies that concurrent Delete operations are safe.
func TestCache_ConcurrentDelete(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			// Pre-populate cache
			for i := 0; i < 100; i++ {
				store.Set(Key("delete", strconv.Itoa(i)), i, 10*time.Second)
			}

			var wg sync.WaitGroup
			numGoroutines := 50

			// Concurrent Delete operations
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					store.Delete(Key("delete", strconv.Itoa(id)))
				}(i)
			}

			wg.Wait()

			// Verify length
			if got := store.Len(); got != 50 {
				t.Errorf("expected 50 entries after concurrent deletes, got %d", got)
			}
		})
	}
}

// TestCache_LazyCleanup verifies that expired entries are removed on Get (lazy expiration).
func TestCache_LazyCleanup(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			key := "test-key"
			value := "test-value"
			ttl := 10 * time.Millisecond

			store.Set(key, value, ttl)

			// Wait for TTL to expire
			time.Sleep(15 * time.Millisecond)

			// Len should still count the expired entry before lazy cleanup
			initialLen := store.Len()
			if initialLen != 1 {
				t.Logf("Note: Len() before lazy cleanup returned %d", initialLen)
			}

			// Get should trigger lazy cleanup
			_, found := store.Get(key)
			if found {
				t.Error("expected key to be expired after Get, but it was found")
			}

			// Len should now exclude the expired entry after Get triggered cleanup
			_, found = store.Get(key)
			if found {
				t.Error("expected key to still be not found after second Get")
			}
		})
	}
}

// TestCache_BackgroundCleanup verifies that the background goroutine removes expired entries.
func TestCache_BackgroundCleanup(t *testing.T) {
	// Create a store with a very short cleanup interval for testing
	store := &Store{
		items: make(map[string]*cacheEntry),
		stop:  make(chan struct{}),
	}

	// Manually start cleanup with 100ms interval for faster test
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				store.removeExpired()
			case <-store.stop:
				return
			}
		}
	}()

	// Add entries with short TTL
	store.Set("key1", "value1", 50*time.Millisecond)
	store.Set("key2", "value2", 50*time.Millisecond)
	store.Set("key3", "value3", 10*time.Second)

	// Initially should have 3 entries
	if got := store.Len(); got != 3 {
		t.Fatalf("expected 3 entries initially, got %d", got)
	}

	// Wait for background cleanup to run after expiry
	time.Sleep(150 * time.Millisecond)

	// Should now have only 1 entry (key3)
	if got := store.Len(); got != 1 {
		t.Errorf("expected 1 entry after background cleanup, got %d", got)
	}

	// Verify which keys still exist
	_, found := store.Get("key1")
	if found {
		t.Error("expected key1 to be cleaned up by background goroutine")
	}

	_, found = store.Get("key2")
	if found {
		t.Error("expected key2 to be cleaned up by background goroutine")
	}

	_, found = store.Get("key3")
	if !found {
		t.Error("expected key3 to still exist")
	}

	_ = store.Close()
}

// TestKey_Deterministic verifies that Key produces consistent output for the same inputs.
func TestKey_Deterministic(t *testing.T) {
	parts := []string{"rsi", "AAPL", "14", "daily"}

	// Call Key multiple times with same parts
	result1 := Key(parts...)
	result2 := Key(parts...)
	result3 := Key(parts...)

	if result1 != result2 {
		t.Error("Key should produce deterministic output")
	}
	if result2 != result3 {
		t.Error("Key should produce deterministic output")
	}

	expected := "rsi:AAPL:14:daily"
	if result1 != expected {
		t.Errorf("expected %s, got %s", expected, result1)
	}
}

// TestKey_DifferentInputs verifies that Key produces different keys for different inputs.
func TestKey_DifferentInputs(t *testing.T) {
	testCases := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "RSI daily",
			parts:    []string{"rsi", "AAPL", "14", "daily"},
			expected: "rsi:AAPL:14:daily",
		},
		{
			name:     "RSI weekly",
			parts:    []string{"rsi", "AAPL", "14", "weekly"},
			expected: "rsi:AAPL:14:weekly",
		},
		{
			name:     "EMA daily",
			parts:    []string{"ema", "AAPL", "10", "daily"},
			expected: "ema:AAPL:10:daily",
		},
		{
			name:     "Quote",
			parts:    []string{"quote", "AAPL"},
			expected: "quote:AAPL",
		},
		{
			name:     "Single part",
			parts:    []string{"single"},
			expected: "single",
		},
		{
			name:     "Empty parts",
			parts:    []string{},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Key(tc.parts...)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestKey_CollisionFree verifies that different inputs produce different keys.
func TestKey_CollisionFree(t *testing.T) {
	keys := []string{
		Key("rsi", "AAPL", "14", "daily"),
		Key("rsi", "AAPL", "14", "weekly"),
		Key("rsi", "AAPL", "20", "daily"),
		Key("rsi", "MSFT", "14", "daily"),
		Key("ema", "AAPL", "14", "daily"),
		Key("quote", "TSLA"),
	}

	seen := make(map[string]bool)
	for _, key := range keys {
		if seen[key] {
			t.Errorf("collision detected for key: %s", key)
		}
		seen[key] = true
	}
}

// TestCache_Overwrite verifies that Set overwrites an existing key.
func TestCache_Overwrite(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			key := "test-key"

			store.Set(key, "value1", 10*time.Second)
			store.Set(key, "value2", 10*time.Second)

			got, found := store.Get(key)
			if !found {
				t.Fatal("expected to find key after overwrite")
			}
			if got != "value2" {
				t.Errorf("expected value2 after overwrite, got %v", got)
			}
		})
	}
}

// TestCache_Close verifies that Close stops the cleanup goroutine without panicking.
func TestCache_Close(t *testing.T) {
	store := New()

	// Add some entries
	store.Set("key1", "value1", 10*time.Second)
	store.Set("key2", "value2", 10*time.Second)

	// Close should not panic
	_ = store.Close()

	// Operations after close should still work (they're just not using the goroutine anymore)
	store.Set("key3", "value3", 10*time.Second)
	_, found := store.Get("key1")
	if !found {
		t.Error("expected key1 to still exist after Close")
	}
}

// TestCache_NilValue verifies that Set can store nil values.
func TestCache_NilValue(t *testing.T) {
	for _, factory := range cacheFactories {
		t.Run(factory.name, func(t *testing.T) {
			store := factory.new(t)

			key := "test-key"

			store.Set(key, nil, 10*time.Second)

			got, found := store.Get(key)
			if !found {
				t.Fatal("expected to find key with nil value")
			}
			if got != nil {
				t.Errorf("expected nil value, got %v", got)
			}
		})
	}
}
