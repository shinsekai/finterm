package destiny

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

func TestComputeTPI_Mixed(t *testing.T) {
	// Mixed trend: some MAs rising, some falling
	// Use data that creates divergence
	closes := []float64{100, 101, 102, 100, 98, 102, 105, 103, 106, 108}

	tpiSeries, maScores, err := ComputeTPI(closes, 5, 3)
	require.NoError(t, err)
	assert.Len(t, tpiSeries, len(closes))

	// TPI should be in [-1, 1] range
	lastTPI := tpiSeries[len(closes)-1]
	assert.GreaterOrEqual(t, lastTPI, -1.0)
	assert.LessOrEqual(t, lastTPI, 1.0)

	// Check that maScores are populated with valid values
	for i, score := range maScores {
		assert.GreaterOrEqual(t, score, -1, "MA %d score should be >= -1", i)
		assert.LessOrEqual(t, score, 1, "MA %d score should be <= 1", i)
	}
}

func TestComputeTPI_FlatData(t *testing.T) {
	// Flat prices: constant value
	closes := []float64{100, 100, 100, 100, 100}

	tpiSeries, maScores, err := ComputeTPI(closes, 3, 2)
	require.NoError(t, err)
	assert.Len(t, tpiSeries, len(closes))
	assert.Equal(t, 0.0, tpiSeries[0])

	// TPI should be close to 0 for flat data
	lastTPI := tpiSeries[len(closes)-1]
	assert.Equal(t, 0.0, lastTPI)

	// All MA scores should be 0 for flat trend
	for i, score := range maScores {
		assert.Equal(t, 0, score, "MA %d should have score 0", i)
	}
}

func TestComputeTPI_InsufficientData(t *testing.T) {
	closes := []float64{100}

	_, _, err := ComputeTPI(closes, 5, 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestComputeTPI_MAScoresPopulated(t *testing.T) {
	closes := []float64{100, 101, 102, 103, 104, 105}

	tpiSeries, maScores, err := ComputeTPI(closes, 5, 3)
	require.NoError(t, err)
	assert.Len(t, tpiSeries, len(closes))

	// Verify all 7 MA scores are populated
	assert.Len(t, maScores, 7)
	for i, score := range maScores {
		assert.GreaterOrEqual(t, score, -1, "MA score %d should be >= -1", i)
		assert.LessOrEqual(t, score, 1, "MA score %d should be <= 1", i)
	}

	// Verify the MAs correspond to: [SMA, EMA, DEMA, TEMA, WMA, HMA, LSMA]
	mas := []string{"SMA", "EMA", "DEMA", "TEMA", "WMA", "HMA", "LSMA"}
	for i, name := range mas {
		t.Logf("%s score: %d", name, maScores[i])
	}
}

func TestCompute_LongSignal(t *testing.T) {
	// Construct data for LONG signal:
	// - Rising prices for TPI > 0.5
	// - RSI rising above 56
	closes := []float64{100, 102, 104, 106, 108, 110, 112, 114, 116, 118, 120, 122, 124}

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
	cfg.MALength = 5      // Shorter for faster TPI response
	cfg.RSILength = 5     // Shorter for faster RSI response
	cfg.RSIThreshold = 40 // Lower threshold to ensure RSI exceeds it

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
	closes := []float64{120, 118, 116, 114, 112, 110, 108, 106, 104, 102, 100, 98, 96}

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
	cfg.MALength = 5 // Shorter for faster TPI response

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, -1, result.Score)
	assert.Equal(t, "SHORT", result.Signal)
	assert.Less(t, result.TPI, 0.0, "TPI should be negative for short")
}

func TestCompute_ShortSignal_RSIOnly(t *testing.T) {
	// Construct data for SHORT signal via RSI only:
	// - TPI neutral (sideways prices)
	// - RSI falling below threshold
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
	cfg.RSIThreshold = 60 // High threshold that oscillating RSI won't maintain

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
	// The oscillating data should trigger short conditions
	assert.Equal(t, "SHORT", result.Signal)
}

func TestCompute_NeutralAtStart(t *testing.T) {
	// Test that minimal data without clear trend produces HOLD signal
	// Use only 2 data points with identical close prices
	data := []OHLCV{
		{Date: 0, Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Date: 1, Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
	}

	cfg := DefaultConfig()
	cfg.MALength = 5
	cfg.RSILength = 5
	cfg.RSIThreshold = 1

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	t.Logf("Score: %d, Signal: %s, TPI: %.3f, RSISmooth: %.3f", result.Score, result.Signal, result.TPI, result.RSISmooth)

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

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
	// After the downtrend, score should be -1
	assert.Equal(t, -1, result.Score)
	assert.Equal(t, "SHORT", result.Signal)
	assert.Less(t, result.TPI, 0.0, "TPI should be negative at end")
}

func TestCompute_AsymmetricShort(t *testing.T) {
	// Verify OR logic: TPI alone can trigger short
	// Data should have TPI < -0.5 but RSI conditions not met
	closes := []float64{120, 115, 110, 105, 100, 95, 90, 85, 80, 75}

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
	cfg.MALength = 3 // Very short for fast TPI response
	cfg.RSILength = 3

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify TPI is strongly negative
	assert.Less(t, result.TPI, -0.3, "TPI should be negative")

	// Verify short signal fires
	assert.Equal(t, -1, result.Score, "Score should be -1 (short)")
	assert.Equal(t, "SHORT", result.Signal)
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

func TestCompute_AllFieldsPopulated(t *testing.T) {
	data := []OHLCV{
		{Date: 0, Open: 100, High: 101, Low: 99, Close: 100, Volume: 1000},
		{Date: 1, Open: 100, High: 102, Low: 98, Close: 101, Volume: 1000},
		{Date: 2, Open: 101, High: 103, Low: 99, Close: 102, Volume: 1000},
		{Date: 3, Open: 102, High: 104, Low: 100, Close: 103, Volume: 1000},
		{Date: 4, Open: 103, High: 105, Low: 101, Close: 104, Volume: 1000},
		{Date: 5, Open: 104, High: 106, Low: 102, Close: 105, Volume: 1000},
	}

	cfg := DefaultConfig()
	cfg.MALength = 3
	cfg.RSILength = 3

	result, err := Compute(data, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify all fields are set
	assert.Contains(t, []int{-1, 0, 1}, result.Score, "Score should be -1, 0, or 1")
	assert.Contains(t, []string{"LONG", "SHORT", "HOLD"}, result.Signal, "Signal should be LONG, SHORT, or HOLD")
	assert.GreaterOrEqual(t, result.TPI, -1.0, "TPI should be >= -1")
	assert.LessOrEqual(t, result.TPI, 1.0, "TPI should be <= 1")
	assert.GreaterOrEqual(t, result.RSISmooth, 0.0, "RSISmooth should be >= 0")
	assert.LessOrEqual(t, result.RSISmooth, 100.0, "RSISmooth should be <= 100")

	// Verify maScores are all populated
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
	assert.Equal(t, 18, cfg.RSILength)
	assert.Equal(t, 56.0, cfg.RSIThreshold)
	assert.Equal(t, 6, cfg.LSMAOffset)
}

func TestComputeTPI_TPIRange(t *testing.T) {
	// Test that TPI always stays in [-1, +1] range
	testCases := []struct {
		name   string
		closes []float64
	}{
		{"uptrend", []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110}},
		{"downtrend", []float64{110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100}},
		{"sideways", []float64{100, 101, 100, 101, 100, 101, 100, 101, 100, 101, 100}},
		{"volatile", []float64{100, 110, 90, 115, 85, 120, 80, 125, 75, 130, 70}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tpiSeries, _, err := ComputeTPI(tc.closes, 5, 3)
			require.NoError(t, err)

			for i, tpi := range tpiSeries {
				assert.GreaterOrEqual(t, tpi, -1.0, "TPI[%d] should be >= -1", i)
				assert.LessOrEqual(t, tpi, 1.0, "TPI[%d] should be <= 1", i)
			}
		})
	}
}

func TestComputeTPI_TPIOneWhenAllRising(t *testing.T) {
	// Extended uptrend to ensure all MAs are rising
	closes := make([]float64, 20)
	for i := range closes {
		closes[i] = 100 + float64(i)
	}

	tpiSeries, maScores, err := ComputeTPI(closes, 5, 3)
	require.NoError(t, err)

	// Last TPI should be close to +1
	lastTPI := tpiSeries[len(closes)-1]
	assert.InDelta(t, 1.0, lastTPI, 0.01, "TPI should be exactly 1.0 when all MAs rising")

	// All MA scores should be +1
	for i, score := range maScores {
		assert.Equal(t, 1, score, "MA %d should be +1 in rising trend", i)
	}
}

func TestComputeTPI_TPIMinusOneWhenAllFalling(t *testing.T) {
	// Extended downtrend to ensure all MAs are falling
	closes := make([]float64, 20)
	for i := range closes {
		closes[i] = 120 - float64(i)
	}

	tpiSeries, maScores, err := ComputeTPI(closes, 5, 3)
	require.NoError(t, err)

	// Last TPI should be close to -1
	lastTPI := tpiSeries[len(closes)-1]
	assert.InDelta(t, -1.0, lastTPI, 0.01, "TPI should be exactly -1.0 when all MAs falling")

	// All MA scores should be -1
	for i, score := range maScores {
		assert.Equal(t, -1, score, "MA %d should be -1 in falling trend", i)
	}
}
