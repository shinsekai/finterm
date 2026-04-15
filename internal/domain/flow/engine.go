// Package flow provides the FLOW trend following engine that computes the
// Sebastine indicator from double-smoothed Heikin-Ashi candles and combines
// it with Dynamic RSI for trend following signals.
package flow

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/shinsekai/finterm/internal/domain/dynamo"
)

// OHLCV represents a single candlestick with open, high, low, close, and volume.
type OHLCV struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// Config holds configuration for the FLOW engine.
type Config struct {
	RSILength    int     // RSI period length (default: 14)
	RSIThreshold float64 // RSI threshold for trend confirmation (default: 55)
	FastLength   int     // First EMA smoothing length for OHLC (default: 45)
	SlowLength   int     // Second EMA smoothing length for HA candles (default: 50)
}

// DefaultConfig returns the default configuration for the FLOW engine.
func DefaultConfig() Config {
	return Config{
		RSILength:    14,
		RSIThreshold: 55,
		FastLength:   45,
		SlowLength:   50,
	}
}

// Result contains the computed signal and indicator values.
type Result struct {
	Score     int     // +1 (long), -1 (short), 0 (hold)
	Signal    string  // "LONG", "SHORT", "HOLD"
	Sebastine float64 // Latest Sebastine value (HA body ratio × 100)
	RSISmooth float64 // Latest smoothed RSI value
}

// Sebastine computes the Sebastine indicator from double-smoothed Heikin-Ashi candles.
// Requires all four slices (opens, highs, lows, closes) to be the same length.
//
// Steps:
// 1. First EMA smoothing of each OHLC component with fastLen.
// 2. Heikin-Ashi construction from smoothed data.
// 3. Second EMA smoothing of HA open/close with slowLen.
// 4. Sebastine = ((close2 / open2) - 1) * 100
func Sebastine(opens, highs, lows, closes []float64, fastLen, slowLen int) ([]float64, error) {
	// Validate input lengths
	if len(opens) != len(highs) || len(opens) != len(lows) || len(opens) != len(closes) {
		return nil, errors.New("all OHLC slices must have the same length")
	}

	if len(opens) == 0 {
		return []float64{}, nil
	}

	// Step 1: First EMA smoothing of each OHLC component
	o := dynamo.DynamicEMA(opens, fastLen)
	c := dynamo.DynamicEMA(closes, fastLen)
	h := dynamo.DynamicEMA(highs, fastLen)
	l := dynamo.DynamicEMA(lows, fastLen)

	n := len(closes)

	// Step 2: Heikin-Ashi construction
	haclose := make([]float64, n)
	haopen := make([]float64, n)
	hahigh := make([]float64, n)
	halow := make([]float64, n)

	for i := 0; i < n; i++ {
		haclose[i] = (o[i] + h[i] + l[i] + c[i]) / 4.0

		if i == 0 {
			// Seed with average of first open and close
			haopen[i] = (o[i] + c[i]) / 2.0
		} else {
			// Recursive formula: (previous HA open + previous HA close) / 2
			haopen[i] = (haopen[i-1] + haclose[i-1]) / 2.0
		}

		hahigh[i] = math.Max(h[i], math.Max(haopen[i], haclose[i]))
		halow[i] = math.Min(l[i], math.Min(haopen[i], haclose[i]))
	}

	// Step 3: Second EMA smoothing of HA open/close
	o2 := dynamo.DynamicEMA(haopen, slowLen)
	c2 := dynamo.DynamicEMA(haclose, slowLen)

	// Step 4: Sebastine computation
	seb := make([]float64, n)
	for i := 0; i < n; i++ {
		if o2[i] == 0 {
			// Avoid division by zero
			seb[i] = 0
		} else {
			seb[i] = ((c2[i] / o2[i]) - 1.0) * 100.0
		}
	}

	return seb, nil
}

// Compute calculates the FLOW signal from OHLC data using the Sebastine indicator
// and Dynamic RSI with smoothing.
//
// Returns a Result containing the latched score, signal label, and latest indicator values.
// Requires at least 2 data points for computation.
func Compute(data []OHLCV, cfg Config) (*Result, error) {
	if len(data) < 2 {
		return nil, errors.New("insufficient data: at least 2 data points required")
	}

	if len(data) == 0 {
		return nil, errors.New("no data provided")
	}

	// Extract OHLC components into separate slices
	opens := make([]float64, len(data))
	highs := make([]float64, len(data))
	lows := make([]float64, len(data))
	closes := make([]float64, len(data))

	for i, bar := range data {
		opens[i] = bar.Open
		highs[i] = bar.High
		lows[i] = bar.Low
		closes[i] = bar.Close
	}

	// Compute Sebastine indicator
	seb, err := Sebastine(opens, highs, lows, closes, cfg.FastLength, cfg.SlowLength)
	if err != nil {
		return nil, fmt.Errorf("computing Sebastine: %w", err)
	}

	// Compute Dynamic RSI
	rsi := dynamo.DynamicRSI(closes, cfg.RSILength)

	// Smooth RSI with EMA
	rsiSmooth := dynamo.DynamicEMA(rsi, cfg.RSILength)

	// Compute latching score
	score := 0
	n := len(closes)

	for i := 1; i < n; i++ {
		// Skip bars with NaN values
		if math.IsNaN(seb[i]) || math.IsNaN(rsiSmooth[i]) || math.IsNaN(rsiSmooth[i-1]) {
			continue
		}

		// Long signal: Sebastine > 0 AND RSI rising AND RSI above threshold
		longSignal := seb[i] > 0 &&
			rsiSmooth[i] > rsiSmooth[i-1] &&
			rsiSmooth[i] > cfg.RSIThreshold

		// Short signal: Sebastine < 0 OR (RSI falling AND RSI below threshold)
		shortSignal := seb[i] < 0 ||
			(rsiSmooth[i] < rsiSmooth[i-1] && rsiSmooth[i] < cfg.RSIThreshold)

		if longSignal && !shortSignal {
			score = 1
		}

		if shortSignal {
			score = -1
		}
	}

	// Build result with latest values
	result := &Result{
		Score:     score,
		Signal:    SignalLabel(score),
		Sebastine: seb[n-1],
		RSISmooth: rsiSmooth[n-1],
	}

	return result, nil
}

// SignalLabel converts a score to its string representation.
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
