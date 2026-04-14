// Package destiny implements the DESTINY trend following engine.
// DESTINY computes the Trend Probability Indicator (TPI) from 7 moving averages,
// combines it with Dynamic RSI confirmation, and produces a trend score.
package destiny

import (
	"fmt"
	"math"

	"github.com/shinsekai/finterm/internal/domain/blitz"
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

// Config holds the DESTINY engine configuration parameters.
type Config struct {
	MALength     int     // Default: 45. Base period for all 7 MAs.
	RSILength    int     // Default: 18. Dynamic RSI lookback.
	RSIThreshold float64 // Default: 56. Smoothed RSI threshold.
	LSMAOffset   int     // Default: 6. LSMA projection offset.
}

// DefaultConfig returns the standard DESTINY configuration.
func DefaultConfig() Config {
	return Config{
		MALength:     45,
		RSILength:    18,
		RSIThreshold: 56,
		LSMAOffset:   6,
	}
}

// Result represents the output of the DESTINY engine computation.
type Result struct {
	Score     int     // +1 (long), -1 (short), 0 (hold)
	Signal    string  // "LONG", "SHORT", "HOLD"
	TPI       float64 // Trend Probability Indicator: avg of 7 MA direction scores (-1 to +1)
	RSISmooth float64 // Latest smoothed RSI value
	// Individual MA direction scores for detail view
	MAScores [7]int // [SMA, EMA, DEMA, TEMA, WMA, HMA, LSMA] each +1/0/-1
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
// It computes all 7 MAs using blitz package functions and averages their direction scores.
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

	// Compute all 7 MAs using blitz package functions
	sma := blitz.DynamicSMA(closes, maLength)
	ema := blitz.DynamicEMA(closes, maLength)
	dema := blitz.DynamicDEMA(closes, maLength*2)
	tema := blitz.DynamicTEMA(closes, maLength*3)
	wma := blitz.DynamicWMA(closes, maLength)
	hma := blitz.DynamicHMA(closes, maLength)
	lsma := blitz.DynamicLSMA(closes, maLength, lsmaOffset)

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

// Compute is the main entry point for DESTINY analysis.
// It takes full price history and configuration, then returns a trend signal result.
//
// Steps:
// 1. Extract close prices from OHLCV data
// 2. Compute TPI from 7 moving averages
// 3. Compute RSI and smooth it with EMA
// 4. Walk through bars to compute the latching score
// 5. Build and return Result
//
// The signal logic is asymmetric:
//   - LONG requires TPI > 0.5 AND RSI rising AND RSI > threshold (all 3 conditions)
//   - SHORT fires if TPI < -0.5 (even without RSI) OR if RSI is falling below threshold
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
	rsi := blitz.DynamicRSI(closes, cfg.RSILength)

	// Compute smoothed RSI using EMA
	rsiSmooth := blitz.DynamicEMA(rsi, cfg.RSILength)

	// Walk through bars to compute the latching score
	score := 0
	n := len(closes)
	for i := 1; i < n; i++ {
		if math.IsNaN(tpiSeries[i]) || math.IsNaN(rsiSmooth[i]) {
			continue
		}
		prevRSI := rsiSmooth[i-1]
		if math.IsNaN(prevRSI) {
			continue
		}

		// Long signal: all 3 conditions must be true (AND)
		longSignal := tpiSeries[i] > 0.5 &&
			rsiSmooth[i] > prevRSI &&
			rsiSmooth[i] > cfg.RSIThreshold

		// Short signal: TPI < -0.5 OR (RSI falling AND below threshold)
		shortSignal := tpiSeries[i] < -0.5 ||
			(rsiSmooth[i] < prevRSI && rsiSmooth[i] < cfg.RSIThreshold)

		if longSignal && !shortSignal {
			score = 1
		}
		if shortSignal {
			score = -1
		}
	}

	// Build result
	result := &Result{
		Score:     score,
		Signal:    SignalLabel(score),
		TPI:       tpiSeries[n-1],
		RSISmooth: rsiSmooth[n-1],
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
