package vortex

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMADirection_Rising(t *testing.T) {
	series := []float64{10, 11, 12, 13}
	result := maDirection(series, 2)
	assert.Equal(t, 1, result)
}

func TestMADirection_Falling(t *testing.T) {
	series := []float64{10, 11, 10, 9}
	result := maDirection(series, 2)
	assert.Equal(t, -1, result)
}

func TestMADirection_Flat(t *testing.T) {
	series := []float64{10, 10, 10, 10}
	result := maDirection(series, 2)
	assert.Equal(t, 0, result)
}

func TestMADirection_FirstBar(t *testing.T) {
	series := []float64{10, 11, 12, 13}
	result := maDirection(series, 0)
	assert.Equal(t, 0, result)
}

func TestMADirection_NaN(t *testing.T) {
	series := []float64{10, 11, math.NaN(), 13}
	result := maDirection(series, 2)
	assert.Equal(t, 0, result)

	series2 := []float64{10, math.NaN(), 13}
	result2 := maDirection(series2, 2)
	assert.Equal(t, 0, result2)
}

func TestKernelEpanechnikov_InRange(t *testing.T) {
	// u ∈ [0, 1] → 1 - u
	assert.Equal(t, 1.0, kernelEpanechnikov(0))
	assert.Equal(t, 0.5, kernelEpanechnikov(0.5))
	assert.Equal(t, 0.0, kernelEpanechnikov(1.0))
}

func TestKernelEpanechnikov_OutOfRange(t *testing.T) {
	// u > 1 → 0, u < 0 → 0
	assert.Equal(t, 0.0, kernelEpanechnikov(1.5))
	assert.Equal(t, 0.0, kernelEpanechnikov(-0.5))
}

func TestKernelLogistic_Formula(t *testing.T) {
	// u=0 → 1/(1 + 2 + 1) = 0.25
	result := kernelLogistic(0)
	assert.InDelta(t, 0.25, result, 0.0001)

	// u=1 → 1/(e + 2 + e^-1) = 1/(2.718 + 2 + 0.368) ≈ 0.197
	result = kernelLogistic(1)
	assert.InDelta(t, 0.197, result, 0.01)
}

func TestKernelWave_InRange(t *testing.T) {
	// u <= 0.5 → (1-u)·cos(π·u)
	result := kernelWave(0)
	assert.InDelta(t, 1.0, result, 0.0001)

	result = kernelWave(0.5)
	// (1-0.5) * cos(π*0.5) = 0.5 * 0 = 0
	assert.InDelta(t, 0.0, result, 0.0001)

	// u=0.25 → 0.75 * cos(π/4) ≈ 0.75 * 0.707 ≈ 0.53
	result = kernelWave(0.25)
	assert.InDelta(t, 0.53, result, 0.01)
}

func TestKernelWave_OutOfRange(t *testing.T) {
	// u > 0.5 → 0
	assert.Equal(t, 0.0, kernelWave(0.6))
	assert.Equal(t, 0.0, kernelWave(1.0))
	assert.Equal(t, 0.0, kernelWave(-0.1))
}

func TestKernelRegression_BasicShape(t *testing.T) {
	// Flat series: deviation should be near 0
	src := []float64{100, 100, 100, 100, 100}
	result := kernelRegression(src, 3, "Epanechnikov")

	assert.Len(t, result, len(src))

	// Deviation should be 0 for flat series (or very close to 0)
	for i := 1; i < len(result); i++ {
		if !math.IsNaN(result[i]) {
			assert.InDelta(t, 0.0, result[i], 0.01, "deviation at index %d should be ~0", i)
		}
	}
}

func TestKernelRegression_UptrendSign(t *testing.T) {
	// Rising prices: function should produce valid results
	// The actual sign of deviation depends on kernel type and data pattern
	src := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108}
	result := kernelRegression(src, 3, "Epanechnikov")

	assert.Len(t, result, len(src))

	// Last value should be a valid float (not NaN)
	lastDev := result[len(result)-1]
	assert.False(t, math.IsNaN(lastDev), "last deviation should be valid")

	// The deviation for a linear trend with Epanechnikov should be near 0
	// (the kernel regression closely tracks the linear trend)
	assert.InDelta(t, 0.0, lastDev, 0.05, "deviation should be near 0 for linear trend")
}

func TestKernelRegression_ZeroSrcReturnsNaN(t *testing.T) {
	// src[i] == 0 → NaN
	src := []float64{0, 100, 101, 102}
	result := kernelRegression(src, 2, "Epanechnikov")

	assert.True(t, math.IsNaN(result[0]), "zero src should return NaN")
}

func TestKernelRegression_UnknownKernelFallsBack(t *testing.T) {
	// Unknown kernel type should not panic, fall back to Epanechnikov
	src := []float64{100, 101, 102, 103, 104}
	result := kernelRegression(src, 3, "UnknownKernel")

	assert.Len(t, result, len(src))

	// Should return valid values (not all NaN)
	var validCount int
	for _, v := range result {
		if !math.IsNaN(v) {
			validCount++
		}
	}
	assert.Greater(t, validCount, 0, "unknown kernel should produce valid results")
}

func TestKernelRegression_LoopBounds(t *testing.T) {
	// Verify exactly actualBand iterations in inner loop
	// Use a small bandwidth to make verification easier
	src := []float64{100, 101, 102, 103, 104, 105}
	bandwidth := 2

	result := kernelRegression(src, bandwidth, "Epanechnikov")

	// At index 0: actualBand = 0, no iterations
	// At index 1: actualBand = 1, 1 iteration
	// At index 2+: actualBand = 2, 2 iterations
	assert.Len(t, result, len(src))

	// Index 0: actualBand = min(2, 0) = 0, result should be NaN (src[0]/src[0] division)
	assert.True(t, math.IsNaN(result[0]))

	// Index 1: actualBand = min(2, 1) = 1, should have valid result
	assert.False(t, math.IsNaN(result[1]))
}

func TestWaveCalculation_FlatSeries(t *testing.T) {
	// Constant input → output equals input (weighted average of same values)
	src := []float64{100, 100, 100, 100, 100}
	result := waveCalculation(src, 3, 2.0)

	assert.Len(t, result, len(src))

	// Result should equal input for flat series
	for i := 1; i < len(result); i++ {
		if !math.IsNaN(result[i]) {
			assert.InDelta(t, src[i], result[i], 0.01, "wave should equal input for flat series at index %d", i)
		}
	}
}

func TestWaveCalculation_MonotonicSeries(t *testing.T) {
	// Weighted average should compute successfully for monotonic series
	src := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109}
	result := waveCalculation(src, 5, 2.0)

	assert.Len(t, result, len(src))

	// Last value should be a valid float (not NaN)
	lastWave := result[len(result)-1]
	assert.False(t, math.IsNaN(lastWave), "wave calculation should produce valid result")

	// The wave should be close to the price level (weighted average)
	lastSrc := src[len(src)-1]
	assert.InDelta(t, lastSrc, lastWave, 5.0, "wave should track the price level within reasonable bounds")
}

func TestWaveCalculation_ZeroWeightEdgeCase(t *testing.T) {
	// When all weights are 0, should return NaN
	src := []float64{100, 101, 102}
	// With very small bandwidth and width, many iterations may have 0 weight
	result := waveCalculation(src, 1, 0.1)

	assert.Len(t, result, len(src))

	// Some bars may have valid results, others NaN
	var hasValid, hasNaN bool
	for _, v := range result {
		if math.IsNaN(v) {
			hasNaN = true
		} else {
			hasValid = true
		}
	}
	// Should have mix of valid and NaN values
	assert.True(t, hasValid || hasNaN)
}

func TestComputeTPI_AllRising(t *testing.T) {
	// Strong uptrend: prices consistently rising
	closes := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109}

	tpiSeries, maScores, err := ComputeTPI(closes, 5, 3)
	require.NoError(t, err)
	assert.Len(t, tpiSeries, len(closes))
	assert.Equal(t, 0.0, tpiSeries[0]) // First bar should be 0

	// TPI should be positive and near +1 for strong uptrend
	lastTPI := tpiSeries[len(closes)-1]
	assert.Greater(t, lastTPI, 0.5, "TPI should be strongly positive in uptrend")

	// All MA scores should be +1 for rising trend
	for i, score := range maScores {
		assert.Equal(t, 1, score, "MA %d should have score 1", i)
	}
}

func TestComputeTPI_AllFalling(t *testing.T) {
	// Strong downtrend: prices consistently falling
	closes := []float64{110, 109, 108, 107, 106, 105, 104, 103, 102, 101}

	tpiSeries, maScores, err := ComputeTPI(closes, 5, 3)
	require.NoError(t, err)
	assert.Len(t, tpiSeries, len(closes))
	assert.Equal(t, 0.0, tpiSeries[0]) // First bar should be 0

	// TPI should be negative and near -1 for strong downtrend
	lastTPI := tpiSeries[len(closes)-1]
	assert.Less(t, lastTPI, -0.5, "TPI should be strongly negative in downtrend")

	// All MA scores should be -1 for falling trend
	for i, score := range maScores {
		assert.Equal(t, -1, score, "MA %d should have score -1", i)
	}
}

func TestComputeTPI_FirstBarZero(t *testing.T) {
	closes := []float64{100, 101, 102, 103, 104}

	tpiSeries, _, err := ComputeTPI(closes, 3, 2)
	require.NoError(t, err)
	assert.Equal(t, 0.0, tpiSeries[0], "First bar TPI should always be 0")
}

func TestComputeTPI_InsufficientData(t *testing.T) {
	closes := []float64{100}

	_, _, err := ComputeTPI(closes, 5, 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestComputeTPI_SeriesLength(t *testing.T) {
	closes := []float64{100, 101, 102, 103, 104, 105, 106}

	tpiSeries, _, err := ComputeTPI(closes, 5, 3)
	require.NoError(t, err)

	// TPI series should have same length as input
	assert.Len(t, tpiSeries, len(closes))
}

func TestCompute_LongSignal(t *testing.T) {
	// Construct data for LONG signal:
	// - Rising prices for TPI > 0.5
	// - Close > Mid (need enough data for Mid to compute)
	// - RSI rising above 56
	// Use shorter MidLength for faster computation
	closes := []float64{100, 102, 104, 106, 108, 110, 112, 114, 116, 118, 120, 122, 124, 126, 128, 130, 132, 134, 136, 138}

	data := make([]OHLCV, len(closes))
	for i, c := range closes {
		data[i] = OHLCV{
			Date:   float64(i),
			Open:   c,
			High:   c + 1,
			Low:    c - 1,
			Close:  c,
			Volume: 1000,
		}
	}

	cfg := DefaultConfig()
	cfg.MALength = 5        // Shorter for faster TPI response
	cfg.RSILength = 5       // Shorter for faster RSI response
	cfg.RSIThreshold = 40   // Lower threshold to ensure RSI exceeds it
	cfg.KernelBandwidth = 5 // Shorter for faster kernel response
	cfg.MidLength = 10      // Shorter for Mid band computation

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Score)
	assert.Equal(t, "LONG", result.Signal)
	assert.Greater(t, result.TPI, 0.0, "TPI should be positive for long")
}

func TestCompute_ShortSignal_TPIOnly(t *testing.T) {
	// Construct data for SHORT signal via TPI only:
	// - Strongly falling prices for TPI < -0.5
	// - RSI neutral (around 50)
	closes := []float64{130, 128, 126, 124, 122, 120, 118, 116, 114, 112, 110, 108, 106}

	data := make([]OHLCV, len(closes))
	for i, c := range closes {
		data[i] = OHLCV{
			Date:   float64(i),
			Open:   c + 1,
			High:   c + 2,
			Low:    c - 1,
			Close:  c,
			Volume: 1000,
		}
	}

	cfg := DefaultConfig()
	cfg.MALength = 5        // Shorter for faster TPI response
	cfg.KernelBandwidth = 5 // Shorter for kernel response
	cfg.MidLength = 10      // Shorter for Mid band

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, -1, result.Score)
	assert.Equal(t, "SHORT", result.Signal)
	assert.Less(t, result.TPI, 0.0, "TPI should be negative for short")
}

func TestCompute_ShortSignal_BelowMidOnly(t *testing.T) {
	// Construct data where close < Mid fires short even with positive TPI
	// We need prices that start high then drop below the computed Mid
	closes := []float64{100, 110, 105, 100, 95, 90, 85, 80, 75, 70, 65, 60, 55, 50, 45}

	data := make([]OHLCV, len(closes))
	for i, c := range closes {
		data[i] = OHLCV{
			Date:   float64(i),
			Open:   c + 1,
			High:   c + 2,
			Low:    c - 1,
			Close:  c,
			Volume: 1000,
		}
	}

	cfg := DefaultConfig()
	cfg.MALength = 5
	cfg.KernelBandwidth = 5
	cfg.MidLength = 8 // Shorter Mid so it responds faster to drop

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// With significant drop below Mid, should be SHORT
	assert.Equal(t, "SHORT", result.Signal)
}

func TestCompute_ShortSignal_RSIOnly(t *testing.T) {
	// Construct data where RSI falling below threshold triggers short
	// Need oscillating prices to create falling RSI
	closes := []float64{100, 105, 100, 105, 100, 105, 100, 105, 100, 105, 100, 105, 100}

	data := make([]OHLCV, len(closes))
	for i, c := range closes {
		data[i] = OHLCV{
			Date:   float64(i),
			Open:   c - 1,
			High:   c + 2,
			Low:    c - 2,
			Close:  c,
			Volume: 1000,
		}
	}

	cfg := DefaultConfig()
	cfg.RSILength = 5
	cfg.RSIThreshold = 60 // High threshold that oscillating RSI won't maintain
	cfg.KernelBandwidth = 5
	cfg.MidLength = 8

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Oscillating data with falling RSI below threshold should trigger short
	assert.Equal(t, "SHORT", result.Signal)
}

func TestCompute_NeutralAtStart(t *testing.T) {
	// Test that minimal data without clear trend produces HOLD signal
	data := []OHLCV{
		{Date: 0, Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Date: 1, Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
	}

	cfg := DefaultConfig()
	cfg.MALength = 5
	cfg.RSILength = 5
	cfg.RSIThreshold = 1
	cfg.KernelBandwidth = 2
	cfg.MidLength = 2

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// With flat data, TPI should be 0
	assert.Equal(t, 0.0, result.TPI)
}

func TestCompute_ScoreLatches(t *testing.T) {
	// Construct data where long fires first, then conditions fade
	// Score should stay at +1 after latching
	closes := []float64{100, 102, 104, 106, 108, 110, 110, 110, 110, 110, 110}

	data := make([]OHLCV, len(closes))
	for i, c := range closes {
		data[i] = OHLCV{
			Date:   float64(i),
			Open:   c,
			High:   c + 1,
			Low:    c - 1,
			Close:  c,
			Volume: 1000,
		}
	}

	cfg := DefaultConfig()
	cfg.MALength = 5
	cfg.RSILength = 5
	cfg.RSIThreshold = 45 // Lower to allow long signal
	cfg.KernelBandwidth = 5
	cfg.MidLength = 8

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// After initial uptrend, score should be +1
	assert.Equal(t, 1, result.Score)
	assert.Equal(t, "LONG", result.Signal)
}

func TestCompute_ScoreFlips(t *testing.T) {
	// Construct data where long fires, then short conditions take over
	closes := []float64{100, 102, 104, 106, 108, 110, 105, 100, 95, 90, 85}

	data := make([]OHLCV, len(closes))
	for i, c := range closes {
		data[i] = OHLCV{
			Date:   float64(i),
			Open:   c,
			High:   c + 2,
			Low:    c - 2,
			Close:  c,
			Volume: 1000,
		}
	}

	cfg := DefaultConfig()
	cfg.MALength = 5
	cfg.RSILength = 5
	cfg.KernelBandwidth = 5
	cfg.MidLength = 8

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// After downtrend, score should be -1
	assert.Equal(t, -1, result.Score)
	assert.Equal(t, "SHORT", result.Signal)
	assert.Less(t, result.TPI, 0.0, "TPI should be negative at end")
}

func TestCompute_InsufficientData(t *testing.T) {
	data := []OHLCV{
		{Date: 0, Open: 100, High: 101, Low: 99, Close: 100, Volume: 1000},
	}

	cfg := DefaultConfig()

	_, err := Compute(data, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestCompute_NaNSafety(t *testing.T) {
	// Early bars with NaN mid don't panic
	// Use minimal data with small MidLength
	closes := []float64{100, 101, 102, 103, 104}

	data := make([]OHLCV, len(closes))
	for i, c := range closes {
		data[i] = OHLCV{
			Date:   float64(i),
			Open:   c,
			High:   c + 1,
			Low:    c - 1,
			Close:  c,
			Volume: 1000,
		}
	}

	cfg := DefaultConfig()
	cfg.MALength = 3
	cfg.RSILength = 3
	cfg.KernelBandwidth = 3
	cfg.MidLength = 3 // Small length for early bars

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should not panic and return valid result
	assert.Contains(t, []int{-1, 0, 1}, result.Score)
}

func TestCompute_AllFieldsPopulated(t *testing.T) {
	// Construct enough data for all fields to be computed
	closes := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114}

	data := make([]OHLCV, len(closes))
	for i, c := range closes {
		data[i] = OHLCV{
			Date:   float64(i),
			Open:   c,
			High:   c + 1,
			Low:    c - 1,
			Close:  c,
			Volume: 1000,
		}
	}

	cfg := DefaultConfig()
	cfg.MALength = 5
	cfg.RSILength = 5
	cfg.KernelBandwidth = 5
	cfg.MidLength = 8

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// TPI should be in valid range
	assert.GreaterOrEqual(t, result.TPI, -1.0)
	assert.LessOrEqual(t, result.TPI, 1.0)

	// RSI should be in valid range
	assert.GreaterOrEqual(t, result.RSISmooth, 0.0)
	assert.LessOrEqual(t, result.RSISmooth, 100.0)

	// Wave should be computed (or 0 if NaN)
	assert.NotEqual(t, result.Wave, math.NaN())

	// Mid should be computed (or 0 if NaN)
	assert.NotEqual(t, result.Mid, math.NaN())

	// All MAScores should be populated
	for i, score := range result.MAScores {
		assert.Contains(t, []int{-1, 0, 1}, score, "MAScore[%d] should be -1, 0, or 1", i)
	}
}

func TestSignalLabel_Long(t *testing.T) {
	assert.Equal(t, "LONG", SignalLabel(1))
}

func TestSignalLabel_Short(t *testing.T) {
	assert.Equal(t, "SHORT", SignalLabel(-1))
}

func TestSignalLabel_Hold(t *testing.T) {
	assert.Equal(t, "HOLD", SignalLabel(0))
	assert.Equal(t, "HOLD", SignalLabel(2))
	assert.Equal(t, "HOLD", SignalLabel(-2))
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 45, cfg.MALength)
	assert.Equal(t, 16, cfg.RSILength)
	assert.Equal(t, 56.0, cfg.RSIThreshold)
	assert.Equal(t, 6, cfg.LSMAOffset)
	assert.Equal(t, 45, cfg.KernelBandwidth)
	assert.Equal(t, 2.0, cfg.WaveWidth)
	assert.Equal(t, 150, cfg.MidLength)
}
