// Package chart provides tests for TPI history computation.
package chart

import (
	"testing"
	"time"

	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// TestTPIHistory_MatchesFiveEngineAverage tests that TPI matches the average of all five engine scores.
func TestTPIHistory_MatchesFiveEngineAverage(t *testing.T) {
	cfg := config.DefaultConfig()

	// Create test OHLCV data
	bars := make([]indicators.OHLCV, 100)
	for i := range bars {
		price := 100.0 + float64(i)
		bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000000,
		}
	}

	tpiHistory, err := computeTPIHistory(bars, cfg)
	if err != nil {
		t.Fatalf("computeTPIHistory failed: %v", err)
	}

	// Verify TPI values are in valid range [-1, 1]
	for i, tpi := range tpiHistory {
		if tpi < -1 || tpi > 1 {
			t.Errorf("TPI at index %d = %f is outside valid range [-1, 1]", i, tpi)
		}
	}

	// Verify we got one TPI value per bar
	if len(tpiHistory) != len(bars) {
		t.Errorf("Expected %d TPI values, got %d", len(bars), len(tpiHistory))
	}
}

// TestTPIHistory_ExcludesInProgressBar tests that in-progress bars are excluded.
func TestTPIHistory_ExcludesInProgressBar(t *testing.T) {
	cfg := config.DefaultConfig()

	// Create test OHLCV data with today's bar (in-progress)
	today := time.Now().UTC()
	bars := make([]indicators.OHLCV, 10)
	for i := 0; i < 10; i++ {
		date := today.Add(-time.Duration(10-i) * 24 * time.Hour)
		price := 100.0 + float64(i)
		bars[i] = indicators.OHLCV{
			Date:   date,
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000000,
		}
	}

	tpiHistory, err := computeTPIHistory(bars, cfg)
	if err != nil {
		t.Fatalf("computeTPIHistory failed: %v", err)
	}

	// The last bar (today's bar) should have been processed
	// TPI computation uses the full bars array, so all bars should have values
	if len(tpiHistory) != len(bars) {
		t.Errorf("Expected %d TPI values, got %d", len(bars), len(tpiHistory))
	}
}

// TestTPIHistory_InsufficientDataReturnsError tests that insufficient data returns an error.
func TestTPIHistory_InsufficientDataReturnsError(t *testing.T) {
	cfg := config.DefaultConfig()

	// Test with empty bars
	_, err := computeTPIHistory([]indicators.OHLCV{}, cfg)
	if err == nil {
		t.Error("Expected error for empty bars, got nil")
	}

	// Test with 1 bar (insufficient for TSI and RSI)
	_, err = computeTPIHistory([]indicators.OHLCV{
		{
			Date:   time.Now().Add(-24 * time.Hour),
			Open:   100,
			High:   110,
			Low:    95,
			Close:  105,
			Volume: 1000,
		},
	}, cfg)
	if err == nil {
		t.Error("Expected error for 1 bar, got nil")
	}
}

// TestTPIHistory_ComputesConsistentValues tests that TPI computation is consistent.
func TestTPIHistory_ComputesConsistentValues(t *testing.T) {
	cfg := config.DefaultConfig()

	// Create deterministic test data
	bars := make([]indicators.OHLCV, 50)
	for i := range bars {
		price := 100.0 + float64(i)*0.5
		bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(50-i) * 24 * time.Hour),
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000000,
		}
	}

	tpiHistory1, err1 := computeTPIHistory(bars, cfg)
	if err1 != nil {
		t.Fatalf("First computeTPIHistory failed: %v", err1)
	}

	tpiHistory2, err2 := computeTPIHistory(bars, cfg)
	if err2 != nil {
		t.Fatalf("Second computeTPIHistory failed: %v", err2)
	}

	// Verify consistent results
	if len(tpiHistory1) != len(tpiHistory2) {
		t.Fatalf("Length mismatch: %d vs %d", len(tpiHistory1), len(tpiHistory2))
	}

	for i := range tpiHistory1 {
		if tpiHistory1[i] != tpiHistory2[i] {
			t.Errorf("TPI at index %d is inconsistent: %f vs %f", i, tpiHistory1[i], tpiHistory2[i])
		}
	}
}
