package cache

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/stretchr/testify/require"
)

func newTestSQLiteStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLite(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestSQLiteStore_SetAndGet verifies basic Set/Get round-trip.
func TestSQLiteStore_SetAndGet(t *testing.T) {
	s := newTestSQLiteStore(t)

	s.Set("key1", "value1", 10*time.Second)

	got, found := s.Get("key1")
	require.True(t, found, "expected key to be found")
	require.Equal(t, "value1", got)
}

// TestSQLiteStore_ExpiredEntryReturnsNotFound verifies lazy expiration on read.
func TestSQLiteStore_ExpiredEntryReturnsNotFound(t *testing.T) {
	s := newTestSQLiteStore(t)

	s.Set("key1", "value1", 10*time.Millisecond)

	_, found := s.Get("key1")
	require.True(t, found, "expected key to be found immediately")

	time.Sleep(20 * time.Millisecond)

	_, found = s.Get("key1")
	require.False(t, found, "expected expired key to be not found")
}

// TestSQLiteStore_DeleteRemovesEntry verifies Delete removes a key.
func TestSQLiteStore_DeleteRemovesEntry(t *testing.T) {
	s := newTestSQLiteStore(t)

	s.Set("key1", "value1", 10*time.Second)
	_, found := s.Get("key1")
	require.True(t, found)

	s.Delete("key1")
	_, found = s.Get("key1")
	require.False(t, found, "expected key to be deleted")
}

// TestSQLiteStore_DeleteByPrefix verifies prefix-based deletion.
func TestSQLiteStore_DeleteByPrefix(t *testing.T) {
	s := newTestSQLiteStore(t)

	s.Set("prefix:a", "1", 10*time.Second)
	s.Set("prefix:b", "2", 10*time.Second)
	s.Set("other:c", "3", 10*time.Second)

	s.DeleteByPrefix("prefix:")

	_, found := s.Get("prefix:a")
	require.False(t, found)
	_, found = s.Get("prefix:b")
	require.False(t, found)

	got, found := s.Get("other:c")
	require.True(t, found)
	require.Equal(t, "3", got)
}

// TestSQLiteStore_PersistsAcrossRestart verifies un-expired entries survive process restart.
func TestSQLiteStore_PersistsAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persist.db")

	// Create store, set value, close.
	s1, err := NewSQLite(dbPath)
	require.NoError(t, err)
	s1.Set("key1", "value1", 10*time.Minute)
	require.NoError(t, s1.Close())

	// Re-open and read.
	s2, err := NewSQLite(dbPath)
	require.NoError(t, err)
	defer func() { _ = s2.Close() }()

	got, found := s2.Get("key1")
	require.True(t, found, "expected key to persist across restart")
	require.Equal(t, "value1", got)
}

// TestSQLiteStore_ConcurrentAccess verifies safe concurrent read/write under the race detector.
func TestSQLiteStore_ConcurrentAccess(t *testing.T) {
	s := newTestSQLiteStore(t)

	var wg sync.WaitGroup
	numGoroutines := 100
	numOps := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := Key("concurrent", strconv.Itoa(id), strconv.Itoa(j))
				s.Set(key, id*numGoroutines+j, 10*time.Second)
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := Key("concurrent", strconv.Itoa(id), strconv.Itoa(j))
				s.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

// TestSQLiteStore_CleanupGoroutineRemovesExpired verifies the background cleanup goroutine.
func TestSQLiteStore_CleanupGoroutineRemovesExpired(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "cleanup.db")

	s, err := NewSQLite(dbPath)
	require.NoError(t, err)

	// Use a very short TTL so entries expire quickly.
	s.Set("key1", "value1", 50*time.Millisecond)
	s.Set("key2", "value2", 50*time.Millisecond)
	s.Set("key3", "value3", 10*time.Minute)

	require.Equal(t, 3, s.Len())

	// Wait for expiry and at least one cleanup tick (60s is too long for tests,
	// so we manually trigger cleanup).
	time.Sleep(100 * time.Millisecond)
	s.removeExpired()

	require.Equal(t, 1, s.Len())

	_, found := s.Get("key1")
	require.False(t, found)
	_, found = s.Get("key2")
	require.False(t, found)

	got, found := s.Get("key3")
	require.True(t, found)
	require.Equal(t, "value3", got)

	require.NoError(t, s.Close())
}

// TestSQLiteStore_CorruptDBFallsBack verifies that a truncated DB file is detected on open.
func TestSQLiteStore_CorruptDBFallsBack(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "corrupt.db")

	// Write a truncated/corrupted file.
	require.NoError(t, os.WriteFile(dbPath, []byte("not a sqlite db"), 0o600))

	_, err := NewSQLite(dbPath)
	require.Error(t, err, "expected error opening corrupted database")
}

// TestSQLiteStore_FilePermissions verifies DB file is created with 0600 and parent dir with 0700.
func TestSQLiteStore_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "subdir", "finterm.db")

	s, err := NewSQLite(dbPath)
	require.NoError(t, err)
	require.NoError(t, s.Close())

	info, err := os.Stat(dbPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "expected db file mode 0600")

	dirInfo, err := os.Stat(filepath.Dir(dbPath))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm(), "expected parent dir mode 0700")
}

// TestSQLiteStore_ClosedStoreRejectsOperations verifies that operations on a closed store are safe.
func TestSQLiteStore_ClosedStoreRejectsOperations(t *testing.T) {
	s := newTestSQLiteStore(t)
	require.NoError(t, s.Close())

	// Operations should not panic; they may be no-ops.
	s.Set("key", "value", 10*time.Second)
	_, found := s.Get("key")
	require.False(t, found)
	s.Delete("key")
	s.Flush()
	_ = s.Len()
}

// TestSQLiteStore_GobRoundTripGlobalQuote verifies gob encoding/decoding for GlobalQuote.
func TestSQLiteStore_GobRoundTripGlobalQuote(t *testing.T) {
	s := newTestSQLiteStore(t)

	quote := &alphavantage.GlobalQuote{
		Symbol: "AAPL",
		Price:  "150.00",
		Volume: "1000000",
		Open:   "149.00",
		High:   "151.00",
		Low:    "148.50",
		Change: "1.00",
	}

	s.Set("quote:AAPL", quote, 10*time.Minute)

	got, found := s.Get("quote:AAPL")
	require.True(t, found)
	gq, ok := got.(*alphavantage.GlobalQuote)
	require.True(t, ok)
	require.Equal(t, "AAPL", gq.Symbol)
	require.Equal(t, "150.00", gq.Price)
}

// TestSQLiteStore_GobRoundTripTimeSeriesDaily verifies gob encoding/decoding for TimeSeriesDaily.
func TestSQLiteStore_GobRoundTripTimeSeriesDaily(t *testing.T) {
	s := newTestSQLiteStore(t)

	ts := &alphavantage.TimeSeriesDaily{
		MetaData: alphavantage.TimeSeriesMetadata{
			Symbol:      "AAPL",
			Information: "Daily Time Series",
		},
		TimeSeries: map[string]alphavantage.TimeSeriesEntry{
			"2024-01-01": {
				Open:   "150.00",
				High:   "155.00",
				Low:    "149.00",
				Close:  "152.00",
				Volume: "1000000",
			},
		},
	}

	s.Set("ts:AAPL", ts, 10*time.Minute)

	got, found := s.Get("ts:AAPL")
	require.True(t, found)
	result, ok := got.(*alphavantage.TimeSeriesDaily)
	require.True(t, ok)
	require.Equal(t, "AAPL", result.MetaData.Symbol)
	require.Equal(t, "152.00", result.TimeSeries["2024-01-01"].Close)
}

// TestSQLiteStore_GobRoundTripCryptoDaily verifies gob encoding/decoding for CryptoDaily.
func TestSQLiteStore_GobRoundTripCryptoDaily(t *testing.T) {
	s := newTestSQLiteStore(t)

	crypto := &alphavantage.CryptoDaily{
		MetaData: alphavantage.CryptoMetadata{
			DigitalCode: "BTC",
			MarketCode:  "USD",
		},
		TimeSeries: map[string]alphavantage.CryptoEntry{
			"2024-01-01": {
				Open:   "40000.00",
				High:   "41000.00",
				Low:    "39000.00",
				Close:  "40500.00",
				Volume: "50000",
			},
		},
	}

	s.Set("crypto:BTC", crypto, 10*time.Minute)

	got, found := s.Get("crypto:BTC")
	require.True(t, found)
	result, ok := got.(*alphavantage.CryptoDaily)
	require.True(t, ok)
	require.Equal(t, "BTC", result.MetaData.DigitalCode)
	require.Equal(t, "40500.00", result.TimeSeries["2024-01-01"].Close)
}
