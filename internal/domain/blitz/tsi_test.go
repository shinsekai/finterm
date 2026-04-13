package blitz

import (
	"math"
	"testing"
)

// TestPearsonCorrelation_PerfectPositive verifies r = 1.0 for perfectly correlated data.
func TestPearsonCorrelation_PerfectPositive(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{1, 2, 3, 4, 5}
	period := 5

	result := PearsonCorrelation(x, y, period)

	// Only bar 4 should have a valid correlation
	// Bars 0-3 should be 0 (insufficient data)
	for i := 0; i < period-1; i++ {
		if result[i] != 0 {
			t.Errorf("result[%d] = %f, want 0 (insufficient data)", i, result[i])
		}
	}

	// Perfect positive correlation: r = 1.0
	if math.Abs(result[4]-1.0) > 0.0001 {
		t.Errorf("result[4] = %f, want 1.0", result[4])
	}
}

// TestPearsonCorrelation_PerfectNegative verifies r = -1.0 for perfectly anti-correlated data.
func TestPearsonCorrelation_PerfectNegative(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{5, 4, 3, 2, 1}
	period := 5

	result := PearsonCorrelation(x, y, period)

	// Perfect negative correlation: r = -1.0
	if math.Abs(result[4]-(-1.0)) > 0.0001 {
		t.Errorf("result[4] = %f, want -1.0", result[4])
	}
}

// TestPearsonCorrelation_NoCorrelation verifies r for weakly correlated data.
func TestPearsonCorrelation_NoCorrelation(t *testing.T) {
	// Using data with weak correlation (not perfectly uncorrelated)
	// x = [1, 2, 3, 4, 5]
	// y = [3, 4, 2, 5, 1] has moderate positive correlation
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{3, 4, 2, 5, 1}
	period := 5

	result := PearsonCorrelation(x, y, period)

	// The correlation is weak (approximately 0.3-0.5 range)
	// The test verifies we're not getting perfect correlation
	if math.Abs(result[4]) > 0.6 {
		t.Errorf("result[4] = %f, want weak correlation (< 0.6)", result[4])
	}

	// Also verify it's not perfectly correlated
	if math.Abs(result[4]-1.0) < 0.01 {
		t.Errorf("result[4] = %f, should not be perfectly correlated", result[4])
	}
	if math.Abs(result[4]+1.0) < 0.01 {
		t.Errorf("result[4] = %f, should not be perfectly anti-correlated", result[4])
	}
}

// TestPearsonCorrelation_ConstantData verifies r = 0 with constant data (no div by zero).
func TestPearsonCorrelation_ConstantData(t *testing.T) {
	x := []float64{5, 5, 5, 5, 5}
	y := []float64{1, 2, 3, 4, 5}
	period := 5

	result := PearsonCorrelation(x, y, period)

	// Constant x: sumX2 = 0, denominator = 0, r = 0
	if result[4] != 0 {
		t.Errorf("result[4] = %f, want 0 (constant data)", result[4])
	}
}

// TestPearsonCorrelation_InsufficientBars verifies bars before period return 0.
func TestPearsonCorrelation_InsufficientBars(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{1, 2, 3, 4, 5}
	period := 5

	result := PearsonCorrelation(x, y, period)

	// Bars 0-3 should be 0 (insufficient data)
	for i := 0; i < period-1; i++ {
		if result[i] != 0 {
			t.Errorf("result[%d] = %f, want 0 (insufficient data)", i, result[i])
		}
	}

	// Bar 4 should be valid (perfect correlation)
	if math.Abs(result[4]-1.0) > 0.0001 {
		t.Errorf("result[4] = %f, want 1.0", result[4])
	}
}

// TestPearsonCorrelation_WindowSliding verifies correlation updates as window moves.
func TestPearsonCorrelation_WindowSliding(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5, 6, 7}
	y := []float64{1, 2, 3, 4, 5, 6, 7}
	period := 3

	result := PearsonCorrelation(x, y, period)

	// Bars 0-1 should be 0 (insufficient data)
	if result[0] != 0 || result[1] != 0 {
		t.Errorf("Bars 0-1 should be 0, got %f, %f", result[0], result[1])
	}

	// Bars 2-6 should all have r = 1.0 (perfect correlation in each window)
	for i := 2; i < len(result); i++ {
		if math.Abs(result[i]-1.0) > 0.0001 {
			t.Errorf("result[%d] = %f, want 1.0", i, result[i])
		}
	}
}

// TestPearsonCorrelation_EmptyInput verifies handling of empty input.
func TestPearsonCorrelation_EmptyInput(t *testing.T) {
	x := []float64{}
	y := []float64{1, 2, 3}
	period := 5

	result := PearsonCorrelation(x, y, period)

	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

// TestPearsonCorrelation_ScaledData verifies correlation is scale-invariant.
func TestPearsonCorrelation_ScaledData(t *testing.T) {
	x1 := []float64{1, 2, 3, 4, 5}
	y1 := []float64{1, 2, 3, 4, 5}

	x2 := []float64{10, 20, 30, 40, 50}
	y2 := []float64{100, 200, 300, 400, 500}

	period := 5

	result1 := PearsonCorrelation(x1, y1, period)
	result2 := PearsonCorrelation(x2, y2, period)

	// Both should have r = 1.0
	if math.Abs(result1[4]-1.0) > 0.0001 {
		t.Errorf("result1[4] = %f, want 1.0", result1[4])
	}
	if math.Abs(result2[4]-1.0) > 0.0001 {
		t.Errorf("result2[4] = %f, want 1.0", result2[4])
	}
}

// TestPearsonCorrelation_ShiftedData verifies correlation is shift-invariant.
func TestPearsonCorrelation_ShiftedData(t *testing.T) {
	x1 := []float64{1, 2, 3, 4, 5}
	y1 := []float64{1, 2, 3, 4, 5}

	x2 := []float64{11, 12, 13, 14, 15}
	y2 := []float64{101, 102, 103, 104, 105}

	period := 5

	result1 := PearsonCorrelation(x1, y1, period)
	result2 := PearsonCorrelation(x2, y2, period)

	// Both should have r = 1.0
	if math.Abs(result1[4]-1.0) > 0.0001 {
		t.Errorf("result1[4] = %f, want 1.0", result1[4])
	}
	if math.Abs(result2[4]-1.0) > 0.0001 {
		t.Errorf("result2[4] = %f, want 1.0", result2[4])
	}
}

// TestPearsonCorrelation_BothConstant verifies r = 0 when both series are constant.
func TestPearsonCorrelation_BothConstant(t *testing.T) {
	x := []float64{5, 5, 5, 5, 5}
	y := []float64{3, 3, 3, 3, 3}
	period := 5

	result := PearsonCorrelation(x, y, period)

	// Both constant: sumX2 = 0, sumY2 = 0, denominator = 0, r = 0
	if result[4] != 0 {
		t.Errorf("result[4] = %f, want 0 (both constant)", result[4])
	}
}

// TestPearsonCorrelation_DifferentLengths verifies handling of different length slices.
func TestPearsonCorrelation_DifferentLengths(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5, 6, 7}
	y := []float64{1, 2, 3}
	period := 3

	result := PearsonCorrelation(x, y, period)

	// Result length should match x (shorter of the two)
	if len(result) != len(x) {
		t.Errorf("len(result) = %d, want %d", len(result), len(x))
	}

	// Bars 0-1 should be 0 (insufficient data)
	if result[0] != 0 || result[1] != 0 {
		t.Errorf("Bars 0-1 should be 0, got %f, %f", result[0], result[1])
	}

	// Bar 2 should have r = 1.0 (perfect correlation in window)
	if math.Abs(result[2]-1.0) > 0.0001 {
		t.Errorf("result[2] = %f, want 1.0", result[2])
	}

	// Bars 3+ should be 0 (y ran out of data)
	for i := 3; i < len(result); i++ {
		if result[i] != 0 {
			t.Errorf("result[%d] = %f, want 0 (y data exhausted)", i, result[i])
		}
	}
}

// TestTSI_RisingPrices verifies TSI > 0 for rising prices.
func TestTSI_RisingPrices(t *testing.T) {
	closes := []float64{100, 110, 120, 130, 140, 150, 160, 170, 180, 190}
	period := 5

	result := TSI(closes, period)

	// Bars 0-3 should be 0 (insufficient data)
	for i := 0; i < period-1; i++ {
		if result[i] != 0 {
			t.Errorf("result[%d] = %f, want 0 (insufficient data)", i, result[i])
		}
	}

	// Bars 4+ should have positive correlation (rising prices vs bar index)
	for i := period - 1; i < len(result); i++ {
		if result[i] <= 0 {
			t.Errorf("result[%d] = %f, want > 0 (rising prices)", i, result[i])
		}
		if result[i] > 1.0 {
			t.Errorf("result[%d] = %f, want <= 1.0", i, result[i])
		}
	}
}

// TestTSI_FallingPrices verifies TSI < 0 for falling prices.
func TestTSI_FallingPrices(t *testing.T) {
	closes := []float64{200, 190, 180, 170, 160, 150, 140, 130, 120, 110}
	period := 5

	result := TSI(closes, period)

	// Bars 4+ should have negative correlation (falling prices vs bar index)
	for i := period - 1; i < len(result); i++ {
		if result[i] >= 0 {
			t.Errorf("result[%d] = %f, want < 0 (falling prices)", i, result[i])
		}
		if result[i] < -1.0 {
			t.Errorf("result[%d] = %f, want >= -1.0", i, result[i])
		}
	}
}

// TestTSI_FlatPrices verifies TSI ≈ 0 for flat prices.
func TestTSI_FlatPrices(t *testing.T) {
	closes := []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100}
	period := 5

	result := TSI(closes, period)

	// Bars 4+ should have correlation close to 0 (no trend)
	for i := period - 1; i < len(result); i++ {
		if math.Abs(result[i]) > 0.01 {
			t.Errorf("result[%d] = %f, want ~0 (flat prices)", i, result[i])
		}
	}
}

// TestTSI_EmptyInput verifies handling of empty input.
func TestTSI_EmptyInput(t *testing.T) {
	closes := []float64{}
	period := 5

	result := TSI(closes, period)

	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

// TestTSI_SingleBar verifies TSI with a single bar.
func TestTSI_SingleBar(t *testing.T) {
	closes := []float64{100}
	period := 5

	result := TSI(closes, period)

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}
	// Should be 0 (insufficient data)
	if result[0] != 0 {
		t.Errorf("result[0] = %f, want 0", result[0])
	}
}

// TestTSI_LinearTrend verifies TSI = 1.0 for perfect linear uptrend.
func TestTSI_LinearTrend(t *testing.T) {
	// Perfect linear relationship: closes[i] = 100 + 10*i
	closes := []float64{100, 110, 120, 130, 140, 150, 160, 170, 180, 190}
	period := 5

	result := TSI(closes, period)

	// Perfect linear uptrend should give r = 1.0
	for i := period - 1; i < len(result); i++ {
		if math.Abs(result[i]-1.0) > 0.0001 {
			t.Errorf("result[%d] = %f, want 1.0 (perfect linear trend)", i, result[i])
		}
	}
}

// TestTSI_PeriodBoundary verifies TSI at period boundary.
func TestTSI_PeriodBoundary(t *testing.T) {
	closes := []float64{100, 110, 120, 130, 140}
	period := 5

	result := TSI(closes, period)

	// Only bar 4 should have valid correlation
	for i := 0; i < period-1; i++ {
		if result[i] != 0 {
			t.Errorf("result[%d] = %f, want 0", i, result[i])
		}
	}

	// Bar 4 should have correlation
	if result[4] <= 0 {
		t.Errorf("result[4] = %f, want > 0", result[4])
	}
}

// TestTSI_Volatility verifies TSI handles volatile data.
func TestTSI_Volatility(t *testing.T) {
	// Noisy data with slight upward trend
	closes := []float64{100, 105, 95, 110, 90, 115, 85, 120, 80, 125}
	period := 5

	result := TSI(closes, period)

	// Should still have positive correlation (overall upward trend)
	lastIdx := len(result) - 1
	if result[lastIdx] <= 0 {
		t.Errorf("result[%d] = %f, want > 0 (overall upward trend)", lastIdx, result[lastIdx])
	}
}

// TestTSI_Precision verifies precision handling.
func TestTSI_Precision(t *testing.T) {
	closes := []float64{100, 110.5, 120.25, 130.125, 140.0625}
	period := 5

	result := TSI(closes, period)

	// All values should be finite numbers
	for i, tsi := range result {
		if math.IsNaN(tsi) || math.IsInf(tsi, 0) {
			t.Errorf("result[%d] = %f (NaN or Inf)", i, tsi)
		}
	}
}

// TestTSI_SmallPeriod verifies TSI with small period.
func TestTSI_SmallPeriod(t *testing.T) {
	closes := []float64{100, 110, 120, 130, 140}
	period := 2

	result := TSI(closes, period)

	// Bar 0 should be 0 (insufficient data for period 2)
	if result[0] != 0 {
		t.Errorf("result[0] = %f, want 0", result[0])
	}

	// Bars 1+ should have positive correlation
	for i := 1; i < len(result); i++ {
		if result[i] <= 0 {
			t.Errorf("result[%d] = %f, want > 0", i, result[i])
		}
		if result[i] > 1.0 {
			t.Errorf("result[%d] = %f, want <= 1.0", i, result[i])
		}
	}
}
