package blitz

import (
	"math"
	"testing"
)

// TestDynamicLength_BarZero verifies that bar 0 returns length 1.
func TestDynamicLength_BarZero(t *testing.T) {
	tests := []struct {
		name      string
		maxLength int
		barIndex  int
		expected  int
	}{
		{"max 12, bar 0", 12, 0, 1},
		{"max 5, bar 0", 5, 0, 1},
		{"max 100, bar 0", 100, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DynamicLength(tt.maxLength, tt.barIndex)
			if got != tt.expected {
				t.Errorf("DynamicLength(%d, %d) = %d, want %d", tt.maxLength, tt.barIndex, got, tt.expected)
			}
		})
	}
}

// TestDynamicLength_GrowingPhase verifies the growing phase of the window.
func TestDynamicLength_GrowingPhase(t *testing.T) {
	tests := []struct {
		name      string
		maxLength int
		barIndex  int
		expected  int
	}{
		{"max 12, bar 5", 12, 5, 6},
		{"max 10, bar 3", 10, 3, 4},
		{"max 14, bar 7", 14, 7, 8},
		{"max 20, bar 9", 20, 9, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DynamicLength(tt.maxLength, tt.barIndex)
			if got != tt.expected {
				t.Errorf("DynamicLength(%d, %d) = %d, want %d", tt.maxLength, tt.barIndex, got, tt.expected)
			}
		})
	}
}

// TestDynamicLength_AtMax verifies the window stays at max once reached.
func TestDynamicLength_AtMax(t *testing.T) {
	tests := []struct {
		name      string
		maxLength int
		barIndex  int
		expected  int
	}{
		{"max 12, bar 12", 12, 12, 12},
		{"max 12, bar 20", 12, 20, 12},
		{"max 10, bar 10", 10, 10, 10},
		{"max 10, bar 100", 10, 100, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DynamicLength(tt.maxLength, tt.barIndex)
			if got != tt.expected {
				t.Errorf("DynamicLength(%d, %d) = %d, want %d", tt.maxLength, tt.barIndex, got, tt.expected)
			}
		})
	}
}

// TestDynamicLength_BeyondMax verifies behavior when index exceeds max.
func TestDynamicLength_BeyondMax(t *testing.T) {
	tests := []struct {
		name      string
		maxLength int
		barIndex  int
		expected  int
	}{
		{"max 12, bar 20", 12, 20, 12},
		{"max 5, bar 50", 5, 50, 5},
		{"max 14, bar 100", 14, 100, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DynamicLength(tt.maxLength, tt.barIndex)
			if got != tt.expected {
				t.Errorf("DynamicLength(%d, %d) = %d, want %d", tt.maxLength, tt.barIndex, got, tt.expected)
			}
		})
	}
}

// TestDynamicSMA_SingleBar verifies SMA with a single data point.
func TestDynamicSMA_SingleBar(t *testing.T) {
	data := []float64{10.0}
	maxLength := 5

	result := DynamicSMA(data, maxLength)

	if len(result) != 1 {
		t.Fatalf("Expected result length 1, got %d", len(result))
	}
	if result[0] != 10.0 {
		t.Errorf("DynamicSMA([10.0], 5)[0] = %f, want 10.0", result[0])
	}
}

// TestDynamicSMA_GrowingWindow verifies adaptive windows [1,2,3,3,3] for maxLength=3.
func TestDynamicSMA_GrowingWindow(t *testing.T) {
	data := []float64{10, 20, 30, 40, 50}
	maxLength := 3

	result := DynamicSMA(data, maxLength)

	// Bar 0: window [10], avg = 10
	if result[0] != 10 {
		t.Errorf("Bar 0: got %f, want 10", result[0])
	}

	// Bar 1: window [10, 20], avg = 15
	if result[1] != 15 {
		t.Errorf("Bar 1: got %f, want 15", result[1])
	}

	// Bar 2: window [10, 20, 30], avg = 20
	if result[2] != 20 {
		t.Errorf("Bar 2: got %f, want 20", result[2])
	}

	// Bar 3: window [20, 30, 40], avg = 30
	if result[3] != 30 {
		t.Errorf("Bar 3: got %f, want 30", result[3])
	}

	// Bar 4: window [30, 40, 50], avg = 40
	if result[4] != 40 {
		t.Errorf("Bar 4: got %f, want 40", result[4])
	}
}

// TestDynamicSMA_FullLength verifies correct average once at max length.
func TestDynamicSMA_FullLength(t *testing.T) {
	data := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	maxLength := 5

	result := DynamicSMA(data, maxLength)

	// Once we reach bar 4+ with maxLength 5, we should have a rolling 5-period average
	// Bar 9: window [60, 70, 80, 90, 100], avg = 80
	expected := 80.0
	if result[9] != expected {
		t.Errorf("Bar 9 (full window): got %f, want %f", result[9], expected)
	}

	// Bar 8: window [50, 60, 70, 80, 90], avg = 70
	expected = 70.0
	if result[8] != expected {
		t.Errorf("Bar 8 (full window): got %f, want %f", result[8], expected)
	}
}

// TestDynamicSMA_EmptyInput verifies handling of empty input.
func TestDynamicSMA_EmptyInput(t *testing.T) {
	data := []float64{}
	maxLength := 5

	result := DynamicSMA(data, maxLength)

	if len(result) != 0 {
		t.Errorf("DynamicSMA([], %d) returned length %d, want 0", maxLength, len(result))
	}
}

// TestDynamicSMA_KnownValues verifies hand-calculated expected outputs.
func TestDynamicSMA_KnownValues(t *testing.T) {
	data := []float64{2, 4, 6, 8}
	maxLength := 2

	result := DynamicSMA(data, maxLength)

	// Bar 0: window [2], avg = 2
	// Bar 1: window [2, 4], avg = 3
	// Bar 2: window [4, 6], avg = 5
	// Bar 3: window [6, 8], avg = 7
	expected := []float64{2, 3, 5, 7}

	for i, got := range result {
		if got != expected[i] {
			t.Errorf("Bar %d: got %f, want %f", i, got, expected[i])
		}
	}
}

// TestDynamicSMA_NaNValues verifies NaN handling (should skip NaN values).
func TestDynamicSMA_NaNValues(t *testing.T) {
	data := []float64{10, math.NaN(), 30, 40, 50}
	maxLength := 3

	result := DynamicSMA(data, maxLength)

	// Bar 0: window [10], avg = 10
	if result[0] != 10 {
		t.Errorf("Bar 0: got %f, want 10", result[0])
	}

	// Bar 1: window [10, NaN], avg = 10 (NaN skipped, count=1)
	if result[1] != 10 {
		t.Errorf("Bar 1: got %f, want 10", result[1])
	}

	// Bar 2: window [10, NaN, 30], avg = 20 (NaN skipped, count=2)
	if result[2] != 20 {
		t.Errorf("Bar 2: got %f, want 20", result[2])
	}

	// Bar 3: window [NaN, 30, 40], avg = 35 (NaN skipped, count=2)
	if result[3] != 35 {
		t.Errorf("Bar 3: got %f, want 35", result[3])
	}

	// Bar 4: window [30, 40, 50], avg = 40
	if result[4] != 40 {
		t.Errorf("Bar 4: got %f, want 40", result[4])
	}
}

// TestDynamicRMA_SingleBar verifies RMA with a single data point.
func TestDynamicRMA_SingleBar(t *testing.T) {
	data := []float64{10.0}
	maxLength := 14

	result := DynamicRMA(data, maxLength)

	if len(result) != 1 {
		t.Fatalf("Expected result length 1, got %d", len(result))
	}
	if result[0] != 10.0 {
		t.Errorf("DynamicRMA([10.0], 14)[0] = %f, want 10.0 (seeded)", result[0])
	}
}

// TestDynamicRMA_FallbackToSMA verifies fallback to SMA when len==1.
func TestDynamicRMA_FallbackToSMA(t *testing.T) {
	data := []float64{10, 20, 30}
	maxLength := 1

	result := DynamicRMA(data, maxLength)

	// With maxLength=1, every bar should use SMA of length 1
	// which is just the data value itself
	for i, got := range result {
		if got != data[i] {
			t.Errorf("Bar %d: got %f, want %f (SMA fallback)", i, got, data[i])
		}
	}
}

// TestDynamicRMA_WilderFormula verifies the (prev*(len-1)+src)/len formula.
func TestDynamicRMA_WilderFormula(t *testing.T) {
	// Use constant data after first bar to verify formula with adaptive length
	data := []float64{100, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	maxLength := 14

	result := DynamicRMA(data, maxLength)

	// Bar 0: seeded with 100
	if result[0] != 100 {
		t.Errorf("Bar 0 (seed): got %f, want 100", result[0])
	}

	// Bar 1: len=2, (100 * 1 + 10) / 2 = 55
	expectedBar1 := (100.0*1 + 10) / 2
	if math.Abs(result[1]-expectedBar1) > 0.0001 {
		t.Errorf("Bar 1: got %f, want %f (Wilder formula with adaptive len=2)", result[1], expectedBar1)
	}

	// Bars after reaching max length should use Wilder formula with len=14
	// At bar 13: len=14, (prev * 13 + src) / 14
	// Let's verify bar 13 after the adaptive warm-up is complete
	// Bar 13: len=14, use Wilder formula with full length
	if result[13] <= 0 || result[13] > 100 {
		t.Errorf("Bar 13 (len=14): got %f, expected value between 0 and 100", result[13])
	}
}

// TestDynamicRMA_GrowingLength verifies adaptive behavior during warm-up.
func TestDynamicRMA_GrowingLength(t *testing.T) {
	data := []float64{10, 20, 30, 40, 50}
	maxLength := 3

	result := DynamicRMA(data, maxLength)

	// Bar 0: seeded with 10
	if result[0] != 10 {
		t.Errorf("Bar 0 (seed): got %f, want 10", result[0])
	}

	// Bar 1: len=2, (10 * 1 + 20) / 2 = 15
	expected := 15.0
	if result[1] != expected {
		t.Errorf("Bar 1 (len=2): got %f, want %f", result[1], expected)
	}

	// Bar 2: len=3, (15 * 2 + 30) / 3 = 60/3 = 20
	expected = 20.0
	if result[2] != expected {
		t.Errorf("Bar 2 (len=3): got %f, want %f", result[2], expected)
	}

	// Bar 3: len=3 (at max), (20 * 2 + 40) / 3 = 80/3 ≈ 26.6667
	expected = 80.0 / 3.0
	if math.Abs(result[3]-expected) > 0.0001 {
		t.Errorf("Bar 3 (len=3 max): got %f, want %f", result[3], expected)
	}

	// Bar 4: len=3, (26.6667 * 2 + 50) / 3 = 103.3334/3 ≈ 34.4445
	expected = (80.0/3.0*2 + 50) / 3.0
	if math.Abs(result[4]-expected) > 0.0001 {
		t.Errorf("Bar 4 (len=3 max): got %f, want %f", result[4], expected)
	}
}

// TestDynamicRMA_EmptyInput verifies handling of empty input.
func TestDynamicRMA_EmptyInput(t *testing.T) {
	data := []float64{}
	maxLength := 14

	result := DynamicRMA(data, maxLength)

	if len(result) != 0 {
		t.Errorf("DynamicRMA([], %d) returned length %d, want 0", maxLength, len(result))
	}
}

// TestDynamicRMA_KnownValues verifies hand-calculated expected outputs.
func TestDynamicRMA_KnownValues(t *testing.T) {
	data := []float64{10, 20, 30, 40}
	maxLength := 2

	result := DynamicRMA(data, maxLength)

	// Bar 0: seeded with 10
	// Bar 1: len=2, (10 * 1 + 20) / 2 = 15
	// Bar 2: len=2, (15 * 1 + 30) / 2 = 22.5
	// Bar 3: len=2, (22.5 * 1 + 40) / 2 = 31.25
	expected := []float64{10, 15, 22.5, 31.25}

	for i, got := range result {
		if math.Abs(got-expected[i]) > 0.0001 {
			t.Errorf("Bar %d: got %f, want %f", i, got, expected[i])
		}
	}
}

// TestDynamicRMA_NaNValues verifies NaN handling.
func TestDynamicRMA_NaNValues(t *testing.T) {
	data := []float64{100, math.NaN(), 30, 40}
	maxLength := 14

	result := DynamicRMA(data, maxLength)

	// Bar 0: seeded with 100
	if result[0] != 100 {
		t.Errorf("Bar 0 (seed): got %f, want 100", result[0])
	}

	// Bar 1: NaN - should carry forward previous RMA
	if result[1] != 100 {
		t.Errorf("Bar 1 (NaN): got %f, want 100 (carry forward)", result[1])
	}

	// Bar 2: len=3, (100 * 2 + 30) / 3 = 230/3 ≈ 76.6667
	expected := (100.0*2 + 30) / 3
	if math.Abs(result[2]-expected) > 0.0001 {
		t.Errorf("Bar 2: got %f, want %f (adaptive len=3)", result[2], expected)
	}

	// Bar 3: len=4, (76.6667 * 3 + 40) / 4 ≈ 67.5
	expected = (expected*3 + 40) / 4
	if math.Abs(result[3]-expected) > 0.0001 {
		t.Errorf("Bar 3: got %f, want %f (adaptive len=4)", result[3], expected)
	}
}

// TestDynamicEMA_SingleBar verifies EMA with a single data point.
func TestDynamicEMA_SingleBar(t *testing.T) {
	data := []float64{10.0}
	maxLength := 10

	result := DynamicEMA(data, maxLength)

	if len(result) != 1 {
		t.Fatalf("Expected result length 1, got %d", len(result))
	}
	if result[0] != 10.0 {
		t.Errorf("DynamicEMA([10.0], 10)[0] = %f, want 10.0 (seeded)", result[0])
	}
}

// TestDynamicEMA_AlphaCalculation verifies alpha = 2/(len+1) not 2+(len+1).
func TestDynamicEMA_AlphaCalculation(t *testing.T) {
	data := []float64{100, 100, 100}
	maxLength := 9 // So len at bar 0 = 1, alpha = 2/10 = 0.2

	result := DynamicEMA(data, maxLength)

	// Bar 0: len=1, i=0, i == len-1 (0==0), seed with 100
	if result[0] != 100 {
		t.Errorf("Bar 0 (seed): got %f, want 100", result[0])
	}

	// Bar 1: len=2, alpha = 2/3 ≈ 0.6667
	// Since i (1) == len-1 (1), seed with 100
	if result[1] != 100 {
		t.Errorf("Bar 1 (seed): got %f, want 100", result[1])
	}

	// Bar 2: len=3, alpha = 2/4 = 0.5
	// i (2) == len-1 (2), seed with 100
	if result[2] != 100 {
		t.Errorf("Bar 2 (seed): got %f, want 100", result[2])
	}
}

// TestDynamicEMA_AlphaCalculationDetailed verifies alpha calculation with different values.
func TestDynamicEMA_AlphaCalculationDetailed(t *testing.T) {
	// Use data where we can verify the EMA formula after seed point
	// With maxLength=9, seed point is at bar 8 (len=9, i=8)
	data := make([]float64, 10)
	for i := range data {
		data[i] = 100 // All same value
	}
	data[9] = 110 // Different value at bar 9 to verify alpha
	maxLength := 9

	result := DynamicEMA(data, maxLength)

	// Bar 8: seed point (len=9, i=8), should be 100
	if result[8] != 100 {
		t.Errorf("Bar 8 (seed point): got %f, want 100", result[8])
	}

	// Bar 9: len=9, alpha = 2/10 = 0.2
	// EMA = (110 - 100) * 0.2 + 100 = 2 + 100 = 102
	alpha := 2.0 / 10.0
	expected := (110.0-result[8])*alpha + result[8]
	if math.Abs(result[9]-expected) > 0.0001 {
		t.Errorf("Bar 9: alpha should be 2/(9+1)=0.2, got result %f, want %f", result[9], expected)
	}
}

// TestDynamicEMA_SeedPoint verifies first valid EMA value is seeded properly.
func TestDynamicEMA_SeedPoint(t *testing.T) {
	data := []float64{100, 110, 120, 130, 140, 150}
	maxLength := 5

	result := DynamicEMA(data, maxLength)

	// Bars 0-4 should be seeded with their respective data values
	// because i == len-1 for each:
	// Bar 0: len=1, i=0, 0==0 -> seed with data[0]=100
	// Bar 1: len=2, i=1, 1==1 -> seed with data[1]=110
	// Bar 2: len=3, i=2, 2==2 -> seed with data[2]=120
	// Bar 3: len=4, i=3, 3==3 -> seed with data[3]=130
	// Bar 4: len=5, i=4, 4==4 -> seed with data[4]=140
	for i := 0; i < 5; i++ {
		if result[i] != data[i] {
			t.Errorf("Bar %d (seed point): got %f, want %f", i, result[i], data[i])
		}
	}

	// Bar 5: len=5 (max), alpha=2/6=0.3333
	// EMA = (150 - 140) * 0.3333 + 140 = 3.3333 + 140 = 143.3333
	alpha := 2.0 / 6.0
	expected := (150-result[4])*alpha + result[4]
	if math.Abs(result[5]-expected) > 0.0001 {
		t.Errorf("Bar 5 (not seed point): got %f, want %f", result[5], expected)
	}
}

// TestDynamicEMA_GrowingLength verifies adaptive EMA during warm-up.
func TestDynamicEMA_GrowingLength(t *testing.T) {
	data := []float64{100, 110, 120, 130, 140}
	maxLength := 4

	result := DynamicEMA(data, maxLength)

	// All bars up to 3 are seed points (i == len-1)
	for i := 0; i < 4; i++ {
		if result[i] != data[i] {
			t.Errorf("Bar %d (growing phase): got %f, want %f (seed)", i, result[i], data[i])
		}
	}

	// Bar 4: len=4 (at max), alpha=2/5=0.4
	// EMA = (140 - 130) * 0.4 + 130 = 4 + 130 = 134
	alpha := 2.0 / 5.0
	expected := (140-result[3])*alpha + result[3]
	if math.Abs(result[4]-expected) > 0.0001 {
		t.Errorf("Bar 4 (at max length): got %f, want %f", result[4], expected)
	}
}

// TestDynamicEMA_EmptyInput verifies handling of empty input.
func TestDynamicEMA_EmptyInput(t *testing.T) {
	data := []float64{}
	maxLength := 10

	result := DynamicEMA(data, maxLength)

	if len(result) != 0 {
		t.Errorf("DynamicEMA([], %d) returned length %d, want 0", maxLength, len(result))
	}
}

// TestDynamicEMA_KnownValues verifies hand-calculated expected outputs.
func TestDynamicEMA_KnownValues(t *testing.T) {
	data := []float64{100, 110, 120, 130}
	maxLength := 3

	result := DynamicEMA(data, maxLength)

	// Bar 0: len=1, seed with 100
	// Bar 1: len=2, seed with 110
	// Bar 2: len=3, seed with 120
	// Bar 3: len=3 (at max), alpha=2/4=0.5
	// EMA = (130 - 120) * 0.5 + 120 = 5 + 120 = 125

	expected := []float64{100, 110, 120, 125}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Bar %d: got %f, want %f", i, result[i], exp)
		}
	}
}

// TestDynamicEMA_NaNValues verifies NaN handling.
func TestDynamicEMA_NaNValues(t *testing.T) {
	data := []float64{100, math.NaN(), 120, 130}
	maxLength := 3

	result := DynamicEMA(data, maxLength)

	// Bar 0: len=1, seed with 100
	if result[0] != 100 {
		t.Errorf("Bar 0: got %f, want 100", result[0])
	}

	// Bar 1: len=2, but NaN, so should carry forward
	// Since it's a seed point but data is NaN, we don't seed
	// The implementation should carry forward last known EMA
	if result[1] != 100 {
		t.Errorf("Bar 1 (NaN): got %f, want 100 (carry forward)", result[1])
	}

	// Bar 2: len=3, seed with 120 (valid data)
	if result[2] != 120 {
		t.Errorf("Bar 2: got %f, want 120 (seed)", result[2])
	}

	// Bar 3: len=3, alpha=0.5
	// EMA = (130 - 120) * 0.5 + 120 = 125
	if result[3] != 125 {
		t.Errorf("Bar 3: got %f, want 125", result[3])
	}
}

// TestDynamicEMA_MatchesStandardEMA verifies convergence to standard EMA.
func TestDynamicEMA_MatchesStandardEMA(t *testing.T) {
	data := make([]float64, 30)
	for i := range data {
		data[i] = float64(i * 10)
	}
	maxLength := 14

	result := DynamicEMA(data, maxLength)

	// After warm-up (bar 13+ with maxLength=14), we should have standard EMA behavior
	// Standard EMA formula: EMA = (src - prev) * alpha + prev
	// with alpha = 2/(period+1) = 2/15 ≈ 0.1333

	alpha := 2.0 / 15.0

	// Check bar 14: len=14, alpha=2/15
	// Bar 13 was seeded with data[13]=130
	// EMA[14] = (140 - 130) * alpha + 130
	expectedBar14 := (data[14]-data[13])*alpha + data[13]
	if math.Abs(result[14]-expectedBar14) > 0.0001 {
		t.Errorf("Bar 14 (post-warmup): got %f, want %f (standard EMA)", result[14], expectedBar14)
	}

	// Check bar 15
	expectedBar15 := (data[15]-result[14])*alpha + result[14]
	if math.Abs(result[15]-expectedBar15) > 0.0001 {
		t.Errorf("Bar 15 (post-warmup): got %f, want %f (standard EMA)", result[15], expectedBar15)
	}

	// Check bar 20
	expectedBar20 := (data[20]-result[19])*alpha + result[19]
	if math.Abs(result[20]-expectedBar20) > 0.0001 {
		t.Errorf("Bar 20 (post-warmup): got %f, want %f (standard EMA)", result[20], expectedBar20)
	}
}

// TestDynamicEMA_AllNaN verifies handling when all data is NaN.
func TestDynamicEMA_AllNaN(t *testing.T) {
	data := []float64{math.NaN(), math.NaN(), math.NaN()}
	maxLength := 3

	result := DynamicEMA(data, maxLength)

	// All bars should be 0 (no seed yet)
	for i, got := range result {
		if got != 0 {
			t.Errorf("Bar %d (all NaN): got %f, want 0", i, got)
		}
	}
}

// TestDynamicSMA_AllNaN verifies SMA with all NaN values.
func TestDynamicSMA_AllNaN(t *testing.T) {
	data := []float64{math.NaN(), math.NaN(), math.NaN()}
	maxLength := 3

	result := DynamicSMA(data, maxLength)

	// All bars should be 0 (no valid data)
	for i, got := range result {
		if got != 0 {
			t.Errorf("Bar %d (all NaN): got %f, want 0", i, got)
		}
	}
}

// TestDynamicRMA_AllNaN verifies RMA with all NaN values.
func TestDynamicRMA_AllNaN(t *testing.T) {
	data := []float64{math.NaN(), math.NaN(), math.NaN()}
	maxLength := 14

	result := DynamicRMA(data, maxLength)

	// Bar 0: seed point but NaN, should be 0
	// Bars 1+: carry forward 0
	for i, got := range result {
		if got != 0 {
			t.Errorf("Bar %d (all NaN): got %f, want 0", i, got)
		}
	}
}
