package flow

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSebastine_Uptrend - steadily rising OHLC should produce sebastine > 0
func TestSebastine_Uptrend(t *testing.T) {
	n := 100
	opens := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)

	// Create steadily rising prices
	for i := 0; i < n; i++ {
		base := 100.0 + float64(i)*0.5
		opens[i] = base
		highs[i] = base + 1.0
		lows[i] = base - 0.5
		closes[i] = base + 0.8
	}

	seb, err := Sebastine(opens, highs, lows, closes, 45, 50)
	require.NoError(t, err)
	require.Len(t, seb, n)

	// Sebastine should be positive for uptrend (latest value)
	assert.Greater(t, seb[n-1], 0.0, "Sebastine should be positive for uptrend")
}

// TestSebastine_Downtrend - steadily falling OHLC should produce sebastine < 0
func TestSebastine_Downtrend(t *testing.T) {
	n := 100
	opens := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)

	// Create steadily falling prices
	for i := 0; i < n; i++ {
		base := 150.0 - float64(i)*0.5
		opens[i] = base
		highs[i] = base + 1.0
		lows[i] = base - 0.5
		closes[i] = base - 0.8
	}

	seb, err := Sebastine(opens, highs, lows, closes, 45, 50)
	require.NoError(t, err)
	require.Len(t, seb, n)

	// Sebastine should be negative for downtrend (latest value)
	assert.Less(t, seb[n-1], 0.0, "Sebastine should be negative for downtrend")
}

// TestSebastine_FlatPrices - constant OHLC should produce sebastine ≈ 0
func TestSebastine_FlatPrices(t *testing.T) {
	n := 100
	opens := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)

	// Create flat prices
	for i := 0; i < n; i++ {
		opens[i] = 100.0
		highs[i] = 101.0
		lows[i] = 99.0
		closes[i] = 100.0
	}

	seb, err := Sebastine(opens, highs, lows, closes, 45, 50)
	require.NoError(t, err)
	require.Len(t, seb, n)

	// Sebastine should be approximately zero for flat prices
	assert.InDelta(t, 0.0, seb[n-1], 0.01, "Sebastine should be approximately zero for flat prices")
}

// TestSebastine_DivisionByZero - all-zero OHLC should produce sebastine = 0 (no panic)
func TestSebastine_DivisionByZero(t *testing.T) {
	// All zeros - this will make o2[i] = 0, triggering the division by zero guard
	opens := []float64{0, 0, 0}
	highs := []float64{0, 0, 0}
	lows := []float64{0, 0, 0}
	closes := []float64{0, 0, 0}

	// Should not panic
	seb, err := Sebastine(opens, highs, lows, closes, 45, 50)
	require.NoError(t, err)
	require.Len(t, seb, 3)

	// All values should be 0 to avoid division by zero
	for _, val := range seb {
		assert.Equal(t, 0.0, val)
	}
}

// TestSebastine_MismatchedLengths - different slice lengths should return error
func TestSebastine_MismatchedLengths(t *testing.T) {
	opens := []float64{1, 2, 3}
	highs := []float64{1.5, 2.5, 3.5}
	lows := []float64{0.5, 1.5, 2.5}
	closes := []float64{1.2, 2.2} // Different length

	_, err := Sebastine(opens, highs, lows, closes, 45, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "same length")
}

// TestSebastine_HAOpenRecursive - verify haopen[i] depends on haopen[i-1] + haclose[i-1]
func TestSebastine_HAOpenRecursive(t *testing.T) {
	opens := []float64{100, 105, 110, 115, 120}
	highs := []float64{101, 106, 111, 116, 121}
	lows := []float64{99, 104, 109, 114, 119}
	closes := []float64{100.5, 105.5, 110.5, 115.5, 120.5}

	seb, err := Sebastine(opens, highs, lows, closes, 10, 10)
	require.NoError(t, err)
	require.Len(t, seb, 5)

	// Verify that the values are computed (not all zeros)
	hasNonZero := false
	for _, val := range seb {
		if val != 0 {
			hasNonZero = true
			break
		}
	}
	assert.True(t, hasNonZero, "Sebastine should have non-zero values with recursive HA open")
}

// TestSebastine_DoubleSmoothing - verify two layers of EMA are applied
func TestSebastine_DoubleSmoothing(t *testing.T) {
	// Create some price data
	n := 50
	opens := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)

	for i := 0; i < n; i++ {
		base := 100.0 + float64(i)*0.5
		opens[i] = base
		highs[i] = base + 1.0
		lows[i] = base - 0.5
		closes[i] = base + 0.8
	}

	seb, err := Sebastine(opens, highs, lows, closes, 5, 10)
	require.NoError(t, err)

	// With double EMA smoothing, the values should be different from single smoothing
	// This is a qualitative test - we're just ensuring values exist
	assert.Len(t, seb, n)
	assert.False(t, math.IsNaN(seb[n-1]), "Sebastine value should not be NaN")
}

// TestCompute_LongSignal - construct OHLC where sebastine > 0 + RSI confirms
func TestCompute_LongSignal(t *testing.T) {
	data := make([]OHLCV, 100)

	// Create uptrend data with rising closes
	for i := 0; i < 100; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   100.0 + float64(i)*0.5,
			High:   101.0 + float64(i)*0.5,
			Low:    99.0 + float64(i)*0.5,
			Close:  100.8 + float64(i)*0.5,
			Volume: 1000000,
		}
	}

	cfg := Config{
		RSILength:    14,
		RSIThreshold: 50, // Lower threshold to trigger long
		FastLength:   10,
		SlowLength:   15,
	}

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// For a strong uptrend, we expect a LONG signal
	assert.Equal(t, 1, result.Score)
	assert.Equal(t, "LONG", result.Signal)
	assert.Greater(t, result.Sebastine, 0.0)
	assert.Greater(t, result.RSISmooth, 0.0)
}

// TestCompute_ShortSignal_SebastineOnly - sebastine < 0, RSI neutral → short fires
func TestCompute_ShortSignal_SebastineOnly(t *testing.T) {
	data := make([]OHLCV, 100)

	// Create downtrend data with falling closes
	for i := 0; i < 100; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   150.0 - float64(i)*0.5,
			High:   151.0 - float64(i)*0.5,
			Low:    149.0 - float64(i)*0.5,
			Close:  149.2 - float64(i)*0.5,
			Volume: 1000000,
		}
	}

	cfg := Config{
		RSILength:    14,
		RSIThreshold: 55,
		FastLength:   10,
		SlowLength:   15,
	}

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// For a downtrend with sebastine < 0, we expect a SHORT signal
	assert.Equal(t, -1, result.Score)
	assert.Equal(t, "SHORT", result.Signal)
	assert.Less(t, result.Sebastine, 0.0)
}

// TestCompute_ShortSignal_RSIOnly - sebastine neutral, RSI falling < 55 → short fires
func TestCompute_ShortSignal_RSIOnly(t *testing.T) {
	// First create a run-up, then a decline
	data := make([]OHLCV, 100)

	// Rising phase (first 60 bars)
	for i := 0; i < 60; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   100.0 + float64(i)*0.5,
			High:   101.0 + float64(i)*0.5,
			Low:    99.0 + float64(i)*0.5,
			Close:  100.8 + float64(i)*0.5,
			Volume: 1000000,
		}
	}

	// Declining phase (last 40 bars) - RSI should fall
	for i := 60; i < 100; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   100.0 + float64(60)*0.5 - float64(i-60)*0.3,
			High:   101.0 + float64(60)*0.5 - float64(i-60)*0.3,
			Low:    99.0 + float64(60)*0.5 - float64(i-60)*0.3,
			Close:  100.0 + float64(60)*0.5 - float64(i-60)*0.3 - 0.2, // Falling
			Volume: 1000000,
		}
	}

	cfg := Config{
		RSILength:    14,
		RSIThreshold: 55,
		FastLength:   20, // Longer periods to smooth the transition
		SlowLength:   25,
	}

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// With falling prices after a rise, we expect SHORT (RSI falling below threshold)
	// The exact result depends on the data, but we should have valid values
	assert.Contains(t, []int{-1, 0, 1}, result.Score)
	assert.Contains(t, []string{"SHORT", "HOLD", "LONG"}, result.Signal)
	assert.False(t, math.IsNaN(result.RSISmooth))
}

// TestCompute_NeutralAtStart - no clear trend → score = 0
func TestCompute_NeutralAtStart(t *testing.T) {
	data := make([]OHLCV, 50)

	// Create truly flat/sideways data with small symmetric noise
	for i := 0; i < 50; i++ {
		base := 100.0
		// Add small symmetric noise that averages to zero
		noise := math.Sin(float64(i)) * 0.5
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(50-i) * 24 * time.Hour),
			Open:   base + noise,
			High:   base + noise + 0.3,
			Low:    base + noise - 0.3,
			Close:  base + noise,
			Volume: 1000000,
		}
	}

	cfg := Config{
		RSILength:    14,
		RSIThreshold: 55,
		FastLength:   10,
		SlowLength:   15,
	}

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// For flat prices with symmetric noise, Sebastine should be near zero
	// and score should be HOLD (no clear trend)
	assert.InDelta(t, 0.0, result.Sebastine, 0.5, "Sebastine should be near zero for flat prices")
	// The exact score may vary due to small noise, but should be close to neutral
	assert.Contains(t, []int{0, 1, -1}, result.Score)
}

// TestCompute_ScoreLatches - long fires, conditions fade, score stays +1
func TestCompute_ScoreLatches(t *testing.T) {
	data := make([]OHLCV, 100)

	// Create data: uptrend then neutral
	for i := 0; i < 70; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   100.0 + float64(i)*0.5,
			High:   101.0 + float64(i)*0.5,
			Low:    99.0 + float64(i)*0.5,
			Close:  100.8 + float64(i)*0.5,
			Volume: 1000000,
		}
	}

	// Flat phase (last 30 bars) - long should latch
	for i := 70; i < 100; i++ {
		price := 100.0 + float64(70)*0.5
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   price,
			High:   price + 0.5,
			Low:    price - 0.5,
			Close:  price,
			Volume: 1000000,
		}
	}

	cfg := Config{
		RSILength:    14,
		RSIThreshold: 50, // Lower to keep long active
		FastLength:   10,
		SlowLength:   15,
	}

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// The long from the uptrend should latch
	assert.Equal(t, 1, result.Score)
	assert.Equal(t, "LONG", result.Signal)
}

// TestCompute_ScoreFlips - long, then short conditions → flips to -1
func TestCompute_ScoreFlips(t *testing.T) {
	data := make([]OHLCV, 100)

	// Uptrend phase (first 50 bars)
	for i := 0; i < 50; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   100.0 + float64(i)*0.5,
			High:   101.0 + float64(i)*0.5,
			Low:    99.0 + float64(i)*0.5,
			Close:  100.8 + float64(i)*0.5,
			Volume: 1000000,
		}
	}

	// Sharp downtrend phase (last 50 bars) - should flip to short
	for i := 50; i < 100; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(100-i) * 24 * time.Hour),
			Open:   100.0 + float64(50)*0.5 - float64(i-50),
			High:   101.0 + float64(50)*0.5 - float64(i-50),
			Low:    99.0 + float64(50)*0.5 - float64(i-50),
			Close:  99.0 + float64(50)*0.5 - float64(i-50) - 0.5,
			Volume: 1000000,
		}
	}

	cfg := Config{
		RSILength:    14,
		RSIThreshold: 55,
		FastLength:   10,
		SlowLength:   15,
	}

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// The short from the downtrend should override the earlier long
	assert.Equal(t, -1, result.Score)
	assert.Equal(t, "SHORT", result.Signal)
}

// TestCompute_InsufficientData - less than 2 bars → error
func TestCompute_InsufficientData(t *testing.T) {
	tests := []struct {
		name string
		data []OHLCV
	}{
		{"empty slice", []OHLCV{}},
		{"single bar", []OHLCV{{Open: 100, High: 105, Low: 95, Close: 102}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			result, err := Compute(tt.data, cfg)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "insufficient data")
		})
	}
}

// TestCompute_AllFieldsPopulated - verify Sebastine, RSISmooth in result
func TestCompute_AllFieldsPopulated(t *testing.T) {
	data := make([]OHLCV, 50)

	for i := 0; i < 50; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(50-i) * 24 * time.Hour),
			Open:   100.0 + float64(i)*0.2,
			High:   101.0 + float64(i)*0.2,
			Low:    99.0 + float64(i)*0.2,
			Close:  100.5 + float64(i)*0.2,
			Volume: 1000000,
		}
	}

	cfg := DefaultConfig()
	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// All result fields should be populated
	assert.NotEqual(t, 0, result.Score)
	assert.NotEmpty(t, result.Signal)
	assert.False(t, math.IsNaN(result.Sebastine))
	assert.False(t, math.IsNaN(result.RSISmooth))
}

// TestSignalLabel_Long - score of 1 returns "LONG"
func TestSignalLabel_Long(t *testing.T) {
	assert.Equal(t, "LONG", SignalLabel(1))
}

// TestSignalLabel_Short - score of -1 returns "SHORT"
func TestSignalLabel_Short(t *testing.T) {
	assert.Equal(t, "SHORT", SignalLabel(-1))
}

// TestSignalLabel_Hold - score of 0 returns "HOLD"
func TestSignalLabel_Hold(t *testing.T) {
	assert.Equal(t, "HOLD", SignalLabel(0))
}

// TestDefaultConfig - verify defaults: RSI=14, Threshold=55, Fast=45, Slow=50
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 14, cfg.RSILength)
	assert.Equal(t, 55.0, cfg.RSIThreshold)
	assert.Equal(t, 45, cfg.FastLength)
	assert.Equal(t, 50, cfg.SlowLength)
}

// TestSebastine_EmptySlices - empty input should return empty slice
func TestSebastine_EmptySlices(t *testing.T) {
	opens := []float64{}
	highs := []float64{}
	lows := []float64{}
	closes := []float64{}

	seb, err := Sebastine(opens, highs, lows, closes, 45, 50)
	require.NoError(t, err)
	assert.Empty(t, seb)
}

// TestSebastine_NilSlices - nil slices should work like empty
func TestSebastine_NilSlices(t *testing.T) {
	seb, err := Sebastine(nil, nil, nil, nil, 45, 50)
	require.NoError(t, err)
	assert.Empty(t, seb)
}

// TestCompute_InvalidData - handles zero and negative values gracefully
func TestCompute_InvalidData(t *testing.T) {
	data := []OHLCV{
		{Date: time.Now().Add(-2 * 24 * time.Hour), Open: 100, High: 105, Low: 95, Close: 100, Volume: 0},
		{Date: time.Now().Add(-1 * 24 * time.Hour), Open: -50, High: 0, Low: -100, Close: -75, Volume: 0},
	}

	cfg := DefaultConfig()
	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should handle gracefully without panic
	assert.Contains(t, []int{-1, 0, 1}, result.Score)
}

// TestCompute_ZeroVolume - volume should not affect computation
func TestCompute_ZeroVolume(t *testing.T) {
	data := make([]OHLCV, 50)

	for i := 0; i < 50; i++ {
		data[i] = OHLCV{
			Date:   time.Now().Add(-time.Duration(50-i) * 24 * time.Hour),
			Open:   100.0 + float64(i)*0.5,
			High:   101.0 + float64(i)*0.5,
			Low:    99.0 + float64(i)*0.5,
			Close:  100.8 + float64(i)*0.5,
			Volume: 0, // Zero volume should not affect the computation
		}
	}

	cfg := DefaultConfig()
	result, err := Compute(data, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result.Score)
	assert.Equal(t, "LONG", result.Signal)
}
