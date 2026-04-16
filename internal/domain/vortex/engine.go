// Package vortex implements the VORTEX trend following engine.
// VORTEX computes the Trend Probability Indicator (TPI) from 7 moving averages,
// combines it with Dynamic RSI confirmation, and adds a kernel-regression Mid band
// that gates the signal. The kernel filters detect price deviations from local trends,
// providing additional confirmation for trend entries and exits.
package vortex

import (
	"fmt"
	"math"

	"github.com/shinsekai/finterm/internal/domain/dynamo"
)

// OHLCV represents Open, High, Low, Close, Volume data for a single bar.
// This local type avoids import cycles with other packages.
type OHLCV struct {
	Date   float64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// Config holds the VORTEX engine configuration parameters.
type Config struct {
	MALength        int     // Default: 45. Base period for all 7 MAs.
	RSILength       int     // Default: 16. Dynamic RSI lookback.
	RSIThreshold    float64 // Default: 56. Smoothed RSI threshold.
	LSMAOffset      int     // Default: 6. LSMA projection offset.
	KernelBandwidth int     // Default: 45. Bandwidth for kernel regression filters.
	WaveWidth       float64 // Default: 2.0. Width factor for wave kernel calculation.
	MidLength       int     // Default: 150. SMA length for Mid band.
}

// DefaultConfig returns the standard VORTEX configuration.
func DefaultConfig() Config {
	return Config{
		MALength:        45,
		RSILength:       16,
		RSIThreshold:    56,
		LSMAOffset:      6,
		KernelBandwidth: 45,
		WaveWidth:       2.0,
		MidLength:       150,
	}
}

// Result represents the output of the VORTEX engine computation.
type Result struct {
	Score     int     // +1 (long), -1 (short), 0 (hold)
	Signal    string  // "LONG", "SHORT", "HOLD"
	TPI       float64 // Trend Probability Indicator: avg of 7 MA direction scores (-1 to +1)
	RSISmooth float64 // Latest smoothed RSI value
	Wave      float64 // Latest wave-weighted regression value (price level)
	Mid       float64 // Latest Mid band value (SMA of AV)
	// Individual MA direction scores for detail view
	MAScores [7]int // [SMA, EMA, DEMA, TEMA, WMA, HMA, LSMA] each +1/0/-1
}

// kernelEpanechnikov returns (1 - u) for u in [0, 1], else 0.
// Linear simplified variant used by common Pine kernel libraries.
func kernelEpanechnikov(u float64) float64 {
	if u >= 0 && u <= 1 {
		return 1 - u
	}
	return 0
}

// kernelLogistic returns 1 / (exp(u) + 2 + exp(-u)).
func kernelLogistic(u float64) float64 {
	return 1.0 / (math.Exp(u) + 2.0 + math.Exp(-u))
}

// kernelWave returns (1 - u) * cos(π·u) for u <= 0.5, else 0.
func kernelWave(u float64) float64 {
	if u <= 0.5 && u >= 0 {
		return (1 - u) * math.Cos(math.Pi*u)
	}
	return 0
}

// kernelRegression computes a kernel regression deviation series.
// For each bar i, it computes the weighted deviation of prices in the bandwidth
// using the specified kernel function.
//
// Parameters:
//   - src: Source price series
//   - bandwidth: Lookback window size
//   - kernelType: Kernel function name ("Epanechnikov", "Logistic", "Wave")
//
// Returns a series of deviation ratios, or NaN where computation fails.
func kernelRegression(src []float64, bandwidth int, kernelType string) []float64 {
	if len(src) == 0 {
		return []float64{}
	}

	result := make([]float64, len(src))

	// Select kernel function (fallback to Epanechnikov for unknown types)
	var kernel func(float64) float64
	switch kernelType {
	case "Epanechnikov":
		kernel = kernelEpanechnikov
	case "Logistic":
		kernel = kernelLogistic
	case "Wave":
		kernel = kernelWave
	default:
		// Fallback to Epanechnikov for unknown kernel types
		kernel = kernelEpanechnikov
	}

	bandwidthSq := float64(bandwidth * bandwidth)

	for i := range src {
		actualBand := bandwidth
		if i < bandwidth {
			actualBand = i
		}

		var sumKernels, sumWeightedY float64

		// Inner loop: exactly actualBand iterations
		for j := 0; j < actualBand; j++ {
			base := float64(j*j) / bandwidthSq
			kv := kernel(base)

			// Look back j bars: src[i-j]
			y := src[i-j]

			// Skip NaN source values
			if math.IsNaN(y) {
				continue
			}

			sumKernels += kv
			sumWeightedY += kv * y
		}

		// Compute deviation ratio
		if src[i] != 0 && sumKernels != 0 {
			result[i] = (src[i] - sumWeightedY/sumKernels) / src[i]
		} else {
			result[i] = math.NaN()
		}
	}

	return result
}

// waveCalculation computes a wave-weighted regression series.
// For each bar i, it computes a weighted average using the wave kernel
// with the specified width factor.
//
// Parameters:
//   - src: Source price series
//   - bandwidth: Lookback window size
//   - width: Width factor for kernel calculation
//
// Returns a series of weighted values, or NaN where computation fails.
func waveCalculation(src []float64, bandwidth int, width float64) []float64 {
	if len(src) == 0 {
		return []float64{}
	}

	result := make([]float64, len(src))

	bandwidthSq := float64(bandwidth * bandwidth)
	piWidth := math.Pi * width

	for i := range src {
		actualBand := bandwidth
		if i < bandwidth {
			actualBand = i
		}

		var sum, sumw float64

		// Inner loop: exactly actualBand + 1 iterations
		for j := 0; j <= actualBand; j++ {
			val := float64(j*j) / bandwidthSq / width

			var weight float64
			if val <= 0.5 && val >= 0 {
				weight = (1 - val) * math.Cos(piWidth*val)
			} else {
				weight = 0
			}

			// Look back j bars: src[i-j]
			y := src[i-j]

			// Skip NaN source values
			if math.IsNaN(y) {
				continue
			}

			sum += weight * y
			sumw += weight
		}

		if sumw != 0 {
			result[i] = sum / sumw
		} else {
			result[i] = math.NaN()
		}
	}

	return result
}

// maDirection compares current vs previous value of an MA series.
// Returns +1 if rising, -1 if falling, 0 if flat or invalid data.
func maDirection(series []float64, index int) int {
	if index < 1 || index >= len(series) {
		return 0
	}
	current := series[index]
	previous := series[index-1]

	// Check for NaN values
	if math.IsNaN(current) || math.IsNaN(previous) {
		return 0
	}

	switch {
	case current > previous:
		return 1
	case current < previous:
		return -1
	default:
		return 0
	}
}

// ComputeTPI calculates the Trend Probability Indicator from 7 moving averages.
// It computes all 7 MAs using dynamo package functions and averages their direction scores.
//
// Parameters:
//   - closes: Close price series
//   - maLength: Base period for MA calculations
//   - lsmaOffset: Projection offset for LSMA
//
// Returns:
//   - tpiSeries: TPI value for each bar (range [-1, +1])
//   - maScores: Direction scores for the last bar only [SMA, EMA, DEMA, TEMA, WMA, HMA, LSMA]
//   - err: Error if insufficient data
func ComputeTPI(closes []float64, maLength, lsmaOffset int) (tpiSeries []float64, maScores [7]int, err error) {
	if len(closes) < 2 {
		return nil, [7]int{}, fmt.Errorf("insufficient data: need at least 2 bars, got %d", len(closes))
	}

	// Compute all 7 MAs using dynamo package functions
	sma := dynamo.DynamicSMA(closes, maLength)
	ema := dynamo.DynamicEMA(closes, maLength)
	dema := dynamo.DynamicDEMA(closes, maLength*2)
	tema := dynamo.DynamicTEMA(closes, maLength*3)
	wma := dynamo.DynamicWMA(closes, maLength)
	hma := dynamo.DynamicHMA(closes, maLength)
	lsma := dynamo.DynamicLSMA(closes, maLength, lsmaOffset)

	maSeries := [][]float64{sma, ema, dema, tema, wma, hma, lsma}

	tpiSeries = make([]float64, len(closes))

	// For bar 0: no previous bar to compare, TPI = 0
	tpiSeries[0] = 0

	// For each bar i from 1 onward
	for i := 1; i < len(closes); i++ {
		var sumSig int

		// Score each MA's direction
		for _, series := range maSeries {
			sig := maDirection(series, i)
			sumSig += sig
		}

		// TPI is the arithmetic mean of 7 direction scores
		tpiSeries[i] = float64(sumSig) / 7.0
	}

	// Populate maScores for the last bar only
	lastIdx := len(closes) - 1
	for j, series := range maSeries {
		maScores[j] = maDirection(series, lastIdx)
	}

	return tpiSeries, maScores, nil
}

// Compute is the main entry point for VORTEX analysis.
// It takes full price history and configuration, then returns a trend signal result.
//
// Steps:
// 1. Extract close prices from OHLCV data
// 2. Compute TPI from 7 moving averages
// 3. Compute RSI and smooth it with EMA
// 4. Compute three kernel deviation series (Epanechnikov, Logistic, Wave)
// 5. Compute wave-weighted regression
// 6. Build AV series and compute Mid band (SMA of AV)
// 7. Walk through bars to compute the latching score with Mid gate
// 8. Build and return Result
//
// The signal logic:
//   - LONG requires TPI > 0.5 AND close > Mid AND RSI rising AND RSI > threshold (all 4 conditions)
//   - SHORT fires if TPI < -0.5 OR close < Mid OR (RSI falling AND RSI < threshold)
//
// Note: The Pine script applies rescale/descale via syminfo.mintick, but these cancel
// in the comparison since both close and Mid use the same scale.
func Compute(data []OHLCV, cfg Config) (*Result, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("insufficient data: need at least 2 bars, got %d", len(data))
	}

	// Extract closes from data
	closes := make([]float64, len(data))
	for i, d := range data {
		closes[i] = d.Close
	}

	// Compute TPI
	tpiSeries, maScores, err := ComputeTPI(closes, cfg.MALength, cfg.LSMAOffset)
	if err != nil {
		return nil, err
	}

	// Compute RSI
	rsi := dynamo.DynamicRSI(closes, cfg.RSILength)

	// Compute smoothed RSI using EMA
	rsiSmooth := dynamo.DynamicEMA(rsi, cfg.RSILength)

	// Compute three kernel deviation series
	ep := kernelRegression(closes, cfg.KernelBandwidth, "Epanechnikov")
	lo := kernelRegression(closes, cfg.KernelBandwidth, "Logistic")
	wa := kernelRegression(closes, cfg.KernelBandwidth, "Wave")

	// Compute wave-weighted regression
	wave := waveCalculation(closes, cfg.KernelBandwidth, cfg.WaveWidth)

	// Build AV series: average of three kernel deviations + closes
	av := make([]float64, len(closes))
	for i := range closes {
		if math.IsNaN(ep[i]) || math.IsNaN(lo[i]) || math.IsNaN(wa[i]) {
			// Fallback to closes for early bars where kernel output is NaN
			av[i] = closes[i]
		} else {
			av[i] = (ep[i]+lo[i]+wa[i])/3.0 + closes[i]
		}
	}

	// Compute Mid band: SMA of AV
	mid := dynamo.DynamicSMA(av, cfg.MidLength)

	// Walk through bars to compute the latching score
	score := 0
	n := len(closes)
	for i := 1; i < n; i++ {
		// Skip bars with NaN values
		if math.IsNaN(tpiSeries[i]) || math.IsNaN(rsiSmooth[i]) || math.IsNaN(mid[i]) {
			continue
		}
		prevRSI := rsiSmooth[i-1]
		if math.IsNaN(prevRSI) {
			continue
		}

		// Long signal: all 4 conditions must be true (AND)
		longSignal := tpiSeries[i] > 0.5 &&
			closes[i] > mid[i] &&
			rsiSmooth[i] > prevRSI &&
			rsiSmooth[i] > cfg.RSIThreshold

		// Short signal: TPI < -0.5 OR close < Mid OR (RSI falling AND below threshold)
		shortSignal := tpiSeries[i] < -0.5 ||
			closes[i] < mid[i] ||
			(rsiSmooth[i] < prevRSI && rsiSmooth[i] < cfg.RSIThreshold)

		if longSignal && !shortSignal {
			score = 1
		}
		if shortSignal {
			score = -1
		}
	}

	// Get latest Wave value (fallback to 0 if NaN)
	latestWave := wave[n-1]
	if math.IsNaN(latestWave) {
		latestWave = 0
	}

	// Get latest Mid value (carry NaN as 0 for UI safety)
	latestMid := mid[n-1]
	if math.IsNaN(latestMid) {
		latestMid = 0
	}

	// Build result
	result := &Result{
		Score:     score,
		Signal:    SignalLabel(score),
		TPI:       tpiSeries[n-1],
		RSISmooth: rsiSmooth[n-1],
		Wave:      latestWave,
		Mid:       latestMid,
		MAScores:  maScores,
	}

	return result, nil
}

// SignalLabel converts a numeric score to its string representation.
func SignalLabel(score int) string {
	switch score {
	case 1:
		return "LONG"
	case -1:
		return "SHORT"
	default:
		return "HOLD"
	}
}
