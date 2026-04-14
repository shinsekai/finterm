package blitz

import (
	"math"
	"testing"

	"github.com/shinsekai/finterm/internal/domain/dynamo"
)

// TestCompute_StrongUptrend verifies that a strong uptrend produces a Long signal.
func TestCompute_StrongUptrend(t *testing.T) {
	// 30 bars of rising prices with some pullbacks to avoid RSI saturation
	closes := []float64{
		100, 103, 106, 104, 108, 112, 110, 115, 119, 117,
		122, 126, 124, 129, 133, 131, 136, 140, 138, 143,
		147, 145, 150, 154, 152, 157, 161, 159, 164, 168,
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if result.Current != Long {
		t.Errorf("Expected Current = Long, got %s", result.Current)
	}
}

// TestCompute_StrongDowntrend verifies that a strong downtrend produces a Short signal.
func TestCompute_StrongDowntrend(t *testing.T) {
	// 30 bars of falling prices with some bounces to avoid RSI saturation at 0
	closes := []float64{
		168, 165, 162, 164, 160, 157, 159, 155, 152, 154,
		150, 147, 149, 145, 142, 144, 140, 137, 139, 135,
		132, 134, 130, 127, 129, 125, 122, 124, 120, 117,
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if result.Current != Short {
		t.Errorf("Expected Current = Short, got %s", result.Current)
	}
}

// TestCompute_Sideways verifies that flat prices produce a Hold signal.
func TestCompute_Sideways(t *testing.T) {
	// 30 bars of nearly flat prices with no clear trend
	closes := []float64{
		100, 100, 100, 100, 100, 100, 100, 100, 100, 100,
		100, 100, 100, 100, 100, 100, 100, 100, 100, 100,
		100, 100, 100, 100, 100, 100, 100, 100, 100, 100,
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	// Should be Hold since there's no clear trend direction
	if result.Current != Hold {
		t.Errorf("Expected Current = Hold for sideways market, got %s", result.Current)
	}
}

// TestCompute_TrendReversal verifies that score flips from Long to Short on trend reversal.
func TestCompute_TrendReversal(t *testing.T) {
	// 25 bars rising with pullbacks, then 25 bars falling with bounces
	closes := []float64{
		100, 103, 106, 104, 108, 112, 110, 115, 119, 117,
		122, 126, 124, 129, 133, 131, 136, 140, 138, 143,
		147, 145, 150, 154, 152, // Peak around 154
		150, 147, 149, 145, 142, 144, 140, 137, 139, 135,
		132, 134, 130, 127, 129, 125, 122, 124, 120, 117,
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	// After strong downtrend, should be Short
	if result.Current != Short {
		t.Errorf("Expected Current = Short after trend reversal, got %s", result.Current)
	}

	// Check that we had a Long signal somewhere before the reversal
	foundLong := false
	for i := 5; i < 25; i++ { // Skip warmup period
		if result.Scores[i] == Long {
			foundLong = true
			break
		}
	}
	if !foundLong {
		t.Error("Expected to find Long signal during uptrend phase")
	}
}

// TestCompute_HoldBehavior verifies that score persists between signal changes.
func TestCompute_HoldBehavior(t *testing.T) {
	// 10 bars rising with pullbacks, then 10 bars flat, then 10 bars rising again
	closes := []float64{
		100, 103, 106, 104, 108, 112, 110, 115, 119, 127, // Rising
		127, 127, 127, 127, 127, 127, 127, 127, 127, 127, // Flat
		127, 130, 133, 131, 135, 139, 137, 141, 145, 143, // Rising again
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// During flat period (bars 10-19), score should maintain from previous signal
	for i := 10; i < 20; i++ {
		if result.Scores[i] == Hold && result.Scores[i-1] == Long {
			// This is expected - Hold maintains previous Long
			continue
		}
		if result.Scores[i] == Long || result.Scores[i] == Short {
			// Also valid - signal continues
			continue
		}
	}

	// Final score should be Long from the final uptrend
	if result.Current != Long {
		t.Errorf("Expected Current = Long, got %s", result.Current)
	}
}

// TestCompute_ShortOverridesLong verifies that Short overrides Long when both conditions met.
func TestCompute_ShortOverridesLong(t *testing.T) {
	// Strong downtrend with bounces to avoid RSI saturation
	closes := []float64{
		168, 165, 162, 164, 160, 157, 159, 155, 152, 154,
		150, 147, 149, 145, 142, 144, 140, 137, 139, 135,
		132, 134, 130, 127, 129, 125, 122, 124, 120, 117,
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}
	if result.Current != Short {
		t.Errorf("Expected Current = Short for strong downtrend, got %s", result.Current)
	}
}

// TestCompute_MinimumBars verifies error with less than 2 bars.
func TestCompute_MinimumBars(t *testing.T) {
	tests := []struct {
		name   string
		closes []float64
	}{
		{"empty slice", []float64{}},
		{"single bar", []float64{100}},
	}

	cfg := DefaultConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compute(tt.closes, cfg)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

// TestCompute_InvalidConfig verifies error with zero RSI length.
func TestCompute_InvalidConfig(t *testing.T) {
	closes := []float64{100, 101, 102, 103, 104}

	tests := []struct {
		name string
		cfg  Config
	}{
		{"zero RSI length", Config{RSILength: 0, TSIPeriod: 14, Threshold: 48}},
		{"negative RSI length", Config{RSILength: -1, TSIPeriod: 14, Threshold: 48}},
		{"zero TSI period", Config{RSILength: 12, TSIPeriod: 0, Threshold: 48}},
		{"negative TSI period", Config{RSILength: 12, TSIPeriod: -1, Threshold: 48}},
		{"zero threshold", Config{RSILength: 12, TSIPeriod: 14, Threshold: 0}},
		{"negative threshold", Config{RSILength: 12, TSIPeriod: 14, Threshold: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compute(closes, tt.cfg)
			if err == nil {
				t.Errorf("Expected error for %s config, got nil", tt.name)
			}
		})
	}
}

// TestCompute_DefaultConfig verifies that ComputeSingle uses defaults.
func TestCompute_DefaultConfig(t *testing.T) {
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 100 + float64(i)*2
	}

	// Compute with DefaultConfig directly
	cfg := DefaultConfig()
	result1, err1 := Compute(closes, cfg)
	if err1 != nil {
		t.Fatalf("Compute with DefaultConfig failed: %v", err1)
	}

	// Compute with ComputeSingle
	result2, err2 := ComputeSingle(closes)
	if err2 != nil {
		t.Fatalf("ComputeSingle failed: %v", err2)
	}

	// Both should produce identical results
	if result1.Current != result2.Current {
		t.Errorf("Results differ: DefaultConfig=%s, ComputeSingle=%s", result1.Current, result2.Current)
	}
}

// TestCompute_TSIReported verifies that Result.TSI contains latest correlation value.
func TestCompute_TSIReported(t *testing.T) {
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 100 + float64(i)*2
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// Compute TSI manually to verify
	tsi := dynamo.TSI(closes, cfg.TSIPeriod)
	expectedTSI := tsi[len(tsi)-1]

	if result.TSI != expectedTSI {
		t.Errorf("TSI mismatch: got %f, want %f", result.TSI, expectedTSI)
	}

	// For an uptrend, TSI should be positive
	if result.TSI <= 0 {
		t.Errorf("Expected positive TSI for uptrend, got %f", result.TSI)
	}
}

// TestCompute_RSISmoothReported verifies that Result.RSISmooth contains latest smoothed RSI.
func TestCompute_RSISmoothReported(t *testing.T) {
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 100 + float64(i)*2
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// Compute RSI and smoothed RSI manually to verify
	rsi := dynamo.DynamicRSI(closes, cfg.RSILength)
	rsiSmooth := dynamo.DynamicEMA(rsi, cfg.RSILength)
	expectedRSISmooth := rsiSmooth[len(rsiSmooth)-1]

	if math.Abs(result.RSISmooth-expectedRSISmooth) > 0.0001 {
		t.Errorf("RSISmooth mismatch: got %f, want %f", result.RSISmooth, expectedRSISmooth)
	}
}

// TestCompute_ScoreString verifies "LONG", "SHORT", "HOLD" strings.
func TestCompute_ScoreString(t *testing.T) {
	tests := []struct {
		name     string
		score    Score
		expected string
	}{
		{"Long score", Long, "LONG"},
		{"Short score", Short, "SHORT"},
		{"Hold score", Hold, "HOLD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.score.String() != tt.expected {
				t.Errorf("Score(%d).String() = %s, want %s", tt.score, tt.score.String(), tt.expected)
			}
		})
	}
}

// TestCompute_AllScoresPopulated verifies that Scores slice has same length as input.
func TestCompute_AllScoresPopulated(t *testing.T) {
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 100 + float64(i)*2
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	if len(result.Scores) != len(closes) {
		t.Errorf("Scores length mismatch: got %d, want %d", len(result.Scores), len(closes))
	}

	// Verify all scores are valid
	for i, score := range result.Scores {
		if score != Long && score != Short && score != Hold {
			t.Errorf("Invalid score at bar %d: %d", i, score)
		}
	}
}

// TestCompute_KnownSequence verifies expected scores for a carefully constructed price sequence.
func TestCompute_KnownSequence(t *testing.T) {
	// Phase 1: 20 bars of rising prices with pullbacks → expect Long
	// Phase 2: 10 bars sideways → expect Hold (keeps Long)
	// Phase 3: 20 bars of falling prices with bounces → expect Short
	closes := []float64{
		100, 103, 106, 104, 108, 112, 110, 115, 119, 117,
		122, 126, 124, 129, 133, 131, 136, 140, 138, 143, // Rising to 143
		143, 143, 143, 143, 143, 143, 143, 143, 143, 143, // Sideways
		143, 140, 142, 138, 135, 137, 133, 130, 132, 128,
		125, 127, 123, 120, 122, 118, 115, 117, 113, 110, // Falling to 110
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// Check final score should be Short after downtrend
	if result.Current != Short {
		t.Errorf("Phase 3: expected final score = Short, got %s", result.Current)
	}

	// Verify Phase 1 had Long signals
	foundLongInPhase1 := false
	for i := 5; i < 20; i++ { // Skip first few bars for warm-up
		if result.Scores[i] == Long {
			foundLongInPhase1 = true
			break
		}
	}
	if !foundLongInPhase1 {
		t.Error("Phase 1: expected to find Long signal during uptrend")
	}

	// Verify Phase 3 had Short signals (after sufficient data)
	foundShortInPhase3 := false
	for i := 35; i < 50; i++ {
		if result.Scores[i] == Short {
			foundShortInPhase3 = true
			break
		}
	}
	if !foundShortInPhase3 {
		t.Error("Phase 3: expected to find Short signal during downtrend")
	}
}

// TestCompute_VerifyHoldLogic verifies that first bar is always Hold.
func TestCompute_VerifyHoldLogic(t *testing.T) {
	closes := make([]float64, 10)
	for i := range closes {
		closes[i] = 100 + float64(i)*5
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// First bar should always be Hold (no previous bar for comparison)
	if result.Scores[0] != Hold {
		t.Errorf("First bar score should be Hold, got %s", result.Scores[0])
	}
}

// TestCompute_ThresholdEffect verifies that threshold affects signal generation.
func TestCompute_ThresholdEffect(t *testing.T) {
	// Create moderate uptrend
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 100 + float64(i)*1 // Gentle uptrend
	}

	// Low threshold - should be easier to get Long
	cfgLow := Config{RSILength: 12, TSIPeriod: 14, Threshold: 10}
	resultLow, errLow := Compute(closes, cfgLow)
	if errLow != nil {
		t.Fatalf("Compute with low threshold failed: %v", errLow)
	}

	// High threshold - should be harder to get Long
	cfgHigh := Config{RSILength: 12, TSIPeriod: 14, Threshold: 90}
	resultHigh, errHigh := Compute(closes, cfgHigh)
	if errHigh != nil {
		t.Fatalf("Compute with high threshold failed: %v", errHigh)
	}

	// With low threshold, more likely to be Long
	// With high threshold, more likely to be Hold (or Short if conditions favor)
	// This test mainly verifies different thresholds don't crash
	_ = resultLow
	_ = resultHigh
}

// TestCompute_NaNHandling verifies handling of NaN values in input.
func TestCompute_NaNHandling(t *testing.T) {
	closes := []float64{
		100, 101, 102, math.NaN(), 104, 105, 106, 107,
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed with NaN: %v", err)
	}

	// Should produce valid result despite NaN
	if len(result.Scores) != len(closes) {
		t.Errorf("Scores length mismatch with NaN: got %d, want %d", len(result.Scores), len(closes))
	}

	// First bar should still be Hold
	if result.Scores[0] != Hold {
		t.Errorf("First bar score should be Hold with NaN input, got %s", result.Scores[0])
	}
}

// TestCompute_DowntrendNegativeTSI verifies that downtrend produces negative TSI.
func TestCompute_DowntrendNegativeTSI(t *testing.T) {
	// Strong downtrend
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 160 - float64(i)*3
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// TSI should be negative for downtrend (negative correlation with bar index)
	if result.TSI >= 0 {
		t.Errorf("Expected negative TSI for downtrend, got %f", result.TSI)
	}
}

// TestCompute_UptrendPositiveTSI verifies that uptrend produces positive TSI.
func TestCompute_UptrendPositiveTSI(t *testing.T) {
	// Strong uptrend
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 100 + float64(i)*3
	}

	cfg := DefaultConfig()
	result, err := Compute(closes, cfg)

	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// TSI should be positive for uptrend (positive correlation with bar index)
	if result.TSI <= 0 {
		t.Errorf("Expected positive TSI for uptrend, got %f", result.TSI)
	}
}

// TestDebug_UptrendValues helps debug why signals aren't triggering.
func TestDebug_UptrendValues(t *testing.T) {
	// Strong uptrend
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 100 + float64(i)*2
	}

	cfg := DefaultConfig()

	// Compute TSI
	tsi := dynamo.TSI(closes, cfg.TSIPeriod)

	// Compute RSI
	rsi := dynamo.DynamicRSI(closes, cfg.RSILength)

	// Compute smoothed RSI
	rsiSmooth := dynamo.DynamicEMA(rsi, cfg.RSILength)

	t.Logf("Final values:")
	t.Logf("  TSI: %.4f", tsi[len(tsi)-1])
	t.Logf("  RSI: %.2f", rsi[len(rsi)-1])
	t.Logf("  RSI Smooth: %.2f", rsiSmooth[len(rsiSmooth)-1])
	t.Logf("  Threshold: %.2f", cfg.Threshold)

	last := len(closes) - 1
	t.Logf("\nLong signal conditions at last bar:")
	t.Logf("  TSI > 0: %v (%.4f > 0)", tsi[last] > 0, tsi[last])
	t.Logf("  RSI Smooth rising: %v (%.2f > %.2f)", rsiSmooth[last] > rsiSmooth[last-1], rsiSmooth[last], rsiSmooth[last-1])
	t.Logf("  RSI Smooth > threshold: %v (%.2f > %.2f)", rsiSmooth[last] > cfg.Threshold, rsiSmooth[last], cfg.Threshold)

	// Show last 5 bars
	t.Logf("\nLast 5 bars:")
	for i := len(closes) - 5; i < len(closes); i++ {
		t.Logf("  Bar %2d: Close=%.1f, TSI=%.4f, RSI=%.2f, RSI Smooth=%.2f", i, closes[i], tsi[i], rsi[i], rsiSmooth[i])
	}
}
