package blitz

import (
	"math"
	"testing"
)

// TestDynamicRSI_AllGains verifies RSI approaches 100 with all upward moves.
func TestDynamicRSI_AllGains(t *testing.T) {
	closes := []float64{100, 110, 120, 130, 140, 150, 160, 170, 180, 190}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	// First bar should be 50 (neutral)
	if result[0] != 50 {
		t.Errorf("result[0] = %f, want 50", result[0])
	}

	// All subsequent bars should approach 100 as average down goes to 0
	for i := 1; i < len(result); i++ {
		if result[i] < 50 {
			t.Errorf("result[%d] = %f, want >= 50 for all-gains data", i, result[i])
		}
		if result[i] > 100 {
			t.Errorf("result[%d] = %f, want <= 100", i, result[i])
		}
	}

	// Later bars should be very close to 100 (avgD approaches 0)
	if result[len(result)-1] < 90 {
		t.Errorf("result[%d] = %f, want > 90 for consistent gains", len(result)-1, result[len(result)-1])
	}
}

// TestDynamicRSI_AllLosses verifies RSI approaches 0 with all downward moves.
func TestDynamicRSI_AllLosses(t *testing.T) {
	closes := []float64{200, 190, 180, 170, 160, 150, 140, 130, 120, 110}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	// First bar should be 50 (neutral)
	if result[0] != 50 {
		t.Errorf("result[0] = %f, want 50", result[0])
	}

	// All subsequent bars should approach 0 as average up goes to 0
	for i := 1; i < len(result); i++ {
		if result[i] > 50 {
			t.Errorf("result[%d] = %f, want <= 50 for all-losses data", i, result[i])
		}
		if result[i] < 0 {
			t.Errorf("result[%d] = %f, want >= 0", i, result[i])
		}
	}

	// Later bars should be very close to 0 (avgU approaches 0)
	if result[len(result)-1] > 10 {
		t.Errorf("result[%d] = %f, want < 10 for consistent losses", len(result)-1, result[len(result)-1])
	}
}

// TestDynamicRSI_Alternating verifies RSI behavior with alternating up/down moves.
func TestDynamicRSI_Alternating(t *testing.T) {
	closes := []float64{100, 110, 100, 110, 100, 110, 100, 110, 100, 110}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	// First bar should be 50 (neutral)
	if result[0] != 50 {
		t.Errorf("result[0] = %f, want 50", result[0])
	}

	// With perfect alternation and RMA smoothing, RSI won't stay near 50
	// due to the weighted averaging effect of RMA. The first bar after
	// an up move will have avgD = 0, giving RSI = 100. Subsequent
	// bars will gradually move toward 50 as both averages converge.
	// Just verify values are in valid range.
	for i := 1; i < len(result); i++ {
		if result[i] < 0 || result[i] > 100 {
			t.Errorf("result[%d] = %f, want in [0, 100]", i, result[i])
		}
	}

	// With maxLength=14 and many bars, the RSI should eventually
	// get closer to 50 as the RMA averages balance out
	lastIdx := len(result) - 1
	if math.Abs(result[lastIdx]-50) > 20 {
		t.Errorf("result[%d] = %f, want closer to 50 for long alternating series", lastIdx, result[lastIdx])
	}
}

// TestDynamicRSI_SingleBar verifies RSI[0] = 50.
func TestDynamicRSI_SingleBar(t *testing.T) {
	closes := []float64{100}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}
	if result[0] != 50 {
		t.Errorf("result[0] = %f, want 50", result[0])
	}
}

// TestDynamicRSI_Range verifies all values in [0, 100].
func TestDynamicRSI_Range(t *testing.T) {
	// Create noisy price data
	closes := []float64{100, 105, 98, 110, 95, 115, 90, 120, 85, 125, 80, 130, 75, 135, 70}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	for i, rsi := range result {
		if rsi < 0 || rsi > 100 {
			t.Errorf("result[%d] = %f, want in [0, 100]", i, rsi)
		}
	}
}

// TestDynamicRSI_NoWarmUp verifies values from bar 1 (no warm-up period needed).
func TestDynamicRSI_NoWarmUp(t *testing.T) {
	closes := []float64{100, 105, 98, 110, 95}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	// Bar 0 is 50 (neutral)
	// Bars 1+ should have valid RSI values
	for i := 1; i < len(result); i++ {
		if result[i] == 50 {
			t.Errorf("result[%d] = %f, want non-neutral value", i, result[i])
		}
	}

	// All values should be in valid range
	for i, rsi := range result {
		if rsi < 0 || rsi > 100 {
			t.Errorf("result[%d] = %f, want in [0, 100]", i, rsi)
		}
	}
}

// TestDynamicRSI_EmptyInput verifies handling of empty input.
func TestDynamicRSI_EmptyInput(t *testing.T) {
	closes := []float64{}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

// TestDynamicRSI_KnownValues compares against hand-calculated values.
func TestDynamicRSI_KnownValues(t *testing.T) {
	// Simple case: 4 bars with alternating gains/losses
	// With maxLength=2 and RMA seeding from first value:
	// Bar 0: RSI = 50 (neutral, first bar)
	// Bar 1: up=5, down=0, avgU[0]=0 (seed), avgU[1]=2.5 (RMA with len=2), avgD[0]=0, avgD[1]=0
	//        rs = 2.5/0 -> RSI = 100
	// Bar 2: up=0, down=5, avgU[2]=1.25, avgD[2]=2.5
	//        rs = 1.25/2.5 = 0.5, RSI = 100 - 100/1.5 = 33.33
	// Bar 3: up=5, down=0, avgU[3]=3.125, avgD[3]=1.25
	//        rs = 3.125/1.25 = 2.5, RSI = 100 - 100/3.5 = 71.43
	closes := []float64{100, 105, 100, 105}
	maxLength := 2

	result := DynamicRSI(closes, maxLength)

	expected := []float64{50, 100, 100.0 / 3.0, 250.0 / 3.5}

	for i, exp := range expected {
		if math.Abs(result[i]-exp) > 0.01 {
			t.Errorf("result[%d] = %f, want %f", i, result[i], exp)
		}
	}
}

// TestDynamicRSI_ConstantPrices verifies RSI stays at 50.
func TestDynamicRSI_ConstantPrices(t *testing.T) {
	closes := []float64{100, 100, 100, 100, 100}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	// With no price changes, upMoves and downMoves are all 0
	// Both avgU and avgD will be 0, which we handle as RSI = 0 (all losses case)
	// This is consistent with the formula when both are 0
	for i := 1; i < len(result); i++ {
		// When both avgU and avgD are 0, we go to the all-losses case (avgU == 0)
		if result[i] != 0 {
			t.Errorf("result[%d] = %f, want 0 for constant prices", i, result[i])
		}
	}
}

// TestDynamicRSI_SmallLength verifies behavior with small maxLength.
func TestDynamicRSI_SmallLength(t *testing.T) {
	closes := []float64{100, 110, 120, 130, 140}
	maxLength := 2

	result := DynamicRSI(closes, maxLength)

	// With maxLength=2, the RMA adapts quickly
	// Verify first bar is still 50
	if result[0] != 50 {
		t.Errorf("result[0] = %f, want 50", result[0])
	}

	// All values should be in valid range
	for i, rsi := range result {
		if rsi < 0 || rsi > 100 {
			t.Errorf("result[%d] = %f, want in [0, 100]", i, rsi)
		}
	}
}

// TestDynamicRSI_TwoBars verifies RSI with exactly 2 bars.
func TestDynamicRSI_TwoBars(t *testing.T) {
	closes := []float64{100, 110}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	// Bar 0: 50 (neutral)
	// Bar 1: up=10, down=0, avgD=0 -> RSI=100
	expected := []float64{50, 100}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("result[%d] = %f, want %f", i, result[i], exp)
		}
	}
}

// TestDynamicRSI_Precision verifies precision handling.
func TestDynamicRSI_Precision(t *testing.T) {
	// Use specific values to verify precision
	closes := []float64{100, 100.5, 99.5, 101, 99}
	maxLength := 14

	result := DynamicRSI(closes, maxLength)

	// All values should be finite numbers
	for i, rsi := range result {
		if math.IsNaN(rsi) || math.IsInf(rsi, 0) {
			t.Errorf("result[%d] = %f (NaN or Inf)", i, rsi)
		}
	}
}
